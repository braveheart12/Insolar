//
// Modified BSD 3-Clause Clear License
//
// Copyright (c) 2019 Insolar Technologies GmbH
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without modification,
// are permitted (subject to the limitations in the disclaimer below) provided that
// the following conditions are met:
//  * Redistributions of source code must retain the above copyright notice, this list
//    of conditions and the following disclaimer.
//  * Redistributions in binary form must reproduce the above copyright notice, this list
//    of conditions and the following disclaimer in the documentation and/or other materials
//    provided with the distribution.
//  * Neither the name of Insolar Technologies GmbH nor the names of its contributors
//    may be used to endorse or promote products derived from this software without
//    specific prior written permission.
//
// NO EXPRESS OR IMPLIED LICENSES TO ANY PARTY'S PATENT RIGHTS ARE GRANTED
// BY THIS LICENSE. THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS
// AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES,
// INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL
// THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT,
// INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING,
// BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS
// OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//
// Notwithstanding any other provisions of this license, it is prohibited to:
//    (a) use this software,
//
//    (b) prepare modifications and derivative works of this software,
//
//    (c) distribute this software (including without limitation in source code, binary or
//        object code form), and
//
//    (d) reproduce copies of this software
//
//    for any commercial purposes, and/or
//
//    for the purposes of making available this software to third parties as a service,
//    including, without limitation, any software-as-a-service, platform-as-a-service,
//    infrastructure-as-a-service or other similar online service, irrespective of
//    whether it competes with the products or services of Insolar Technologies GmbH.
//

// +build networktest

package tests

import (
	"context"
	"crypto"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/insolar/insolar/network/rules"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/insolar/insolar/insolar/bus"
	"github.com/insolar/insolar/network/servicenetwork"

	"github.com/insolar/insolar/certificate"
	"github.com/insolar/insolar/component"
	"github.com/insolar/insolar/configuration"
	"github.com/insolar/insolar/cryptography"
	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/log"
	"github.com/insolar/insolar/network"
	"github.com/insolar/insolar/network/consensusv1/packets"
	"github.com/insolar/insolar/network/consensusv1/phases"
	"github.com/insolar/insolar/network/nodenetwork"
	"github.com/insolar/insolar/network/transport"
	"github.com/insolar/insolar/platformpolicy"
	"github.com/insolar/insolar/testutils"
)

var (
	testNetworkPort uint32 = 10010
)

const (
	UseFakeTransport = true
	UseFakeBootstrap = true

	pulseTimeMs  int32 = 8000
	reqTimeoutMs int32 = 2000
	pulseDelta   int32 = 8

	Phase1Timeout  float64 = 0.30
	Phase2Timeout  float64 = 0.50
	Phase21Timeout float64 = 0.70
	Phase3Timeout  float64 = 0.90
)

type fixture struct {
	ctx            context.Context
	bootstrapNodes []*networkNode
	networkNodes   []*networkNode
	pulsar         TestPulsar

	discoveriesAreBootstrapped uint32
}

const cacheDir = "network_cache/"

func newFixture(t *testing.T) *fixture {
	return &fixture{
		ctx:            inslogger.TestContext(t),
		bootstrapNodes: make([]*networkNode, 0),
		networkNodes:   make([]*networkNode, 0),
	}
}

// testSuite is base test suite
type testSuite struct {
	suite.Suite
	fixtureMap     map[string]*fixture
	bootstrapCount int
	nodesCount     int
}

type consensusSuite struct {
	testSuite
}

func newTestSuite(bootstrapCount, nodesCount int) testSuite {
	return testSuite{
		Suite:          suite.Suite{},
		fixtureMap:     make(map[string]*fixture, 0),
		bootstrapCount: bootstrapCount,
		nodesCount:     nodesCount,
	}
}

func newConsensusSuite(bootstrapCount, nodesCount int) *consensusSuite {
	return &consensusSuite{
		testSuite: newTestSuite(bootstrapCount, nodesCount),
	}
}

func (s *testSuite) fixture() *fixture {
	return s.fixtureMap[s.T().Name()]
}

// CheckDiscoveryCount skips test if bootstrap nodes count less then consensusMin
func (s *consensusSuite) CheckBootstrapCount() {
	if len(s.fixture().bootstrapNodes) < consensusMin {
		s.T().Skip(consensusMinMsg)
	}
}

// SetupSuite creates and run network with bootstrap and common nodes once before run all tests in the suite
func (s *consensusSuite) SetupTest() {
	s.fixtureMap[s.T().Name()] = newFixture(s.T())
	var err error
	s.fixture().pulsar, err = NewTestPulsar(pulseTimeMs, reqTimeoutMs, pulseDelta)
	s.Require().NoError(err)

	log.Info("SetupTest")

	for i := 0; i < s.bootstrapCount; i++ {
		s.fixture().bootstrapNodes = append(s.fixture().bootstrapNodes, s.newNetworkNode(fmt.Sprintf("bootstrap_%d", i)))
	}

	for i := 0; i < s.nodesCount; i++ {
		s.fixture().networkNodes = append(s.fixture().networkNodes, s.newNetworkNode(fmt.Sprintf("node_%d", i)))
	}

	pulseReceivers := make([]string, 0)
	for _, node := range s.fixture().bootstrapNodes {
		pulseReceivers = append(pulseReceivers, node.host)
	}

	log.Info("Setup bootstrap nodes")
	s.SetupNodesNetwork(s.fixture().bootstrapNodes)
	s.StartNodesNetwork(s.fixture().bootstrapNodes)

	expectedBootstrapsCount := len(s.fixture().bootstrapNodes)
	retries := 100
	for {
		activeNodes := s.fixture().bootstrapNodes[0].serviceNetwork.NodeKeeper.GetAccessor().GetActiveNodes()
		if expectedBootstrapsCount == len(activeNodes) {
			break
		}

		retries--
		if retries == 0 {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	activeNodes := s.fixture().bootstrapNodes[0].serviceNetwork.NodeKeeper.GetAccessor().GetActiveNodes()
	s.Require().Equal(len(s.fixture().bootstrapNodes), len(activeNodes))

	if len(s.fixture().networkNodes) > 0 {
		log.Info("Setup network nodes")
		s.SetupNodesNetwork(s.fixture().networkNodes)
		s.StartNodesNetwork(s.fixture().networkNodes)

		s.waitForConsensus(2)

		// active nodes count verification
		activeNodes1 := s.fixture().networkNodes[0].serviceNetwork.NodeKeeper.GetAccessor().GetActiveNodes()
		activeNodes2 := s.fixture().networkNodes[0].serviceNetwork.NodeKeeper.GetAccessor().GetActiveNodes()

		s.Require().Equal(s.getNodesCount(), len(activeNodes1))
		s.Require().Equal(s.getNodesCount(), len(activeNodes2))
	}
	fmt.Println("=================== SetupTest() Done")
	log.Info("Start test pulsar")
	err = s.fixture().pulsar.Start(s.fixture().ctx, pulseReceivers)
	s.Require().NoError(err)

}

func (s *testSuite) waitResults(results chan error, expected int) {
	count := 0
	for count < expected {
		err := <-results
		s.Require().NoError(err)
		count++
	}
}

func (s *testSuite) SetupNodesNetwork(nodes []*networkNode) {
	for _, node := range nodes {
		s.preInitNode(node)
	}

	results := make(chan error, len(nodes))
	initNode := func(node *networkNode) {
		err := node.init()
		results <- err
	}

	log.Info("Init nodes")
	for _, node := range nodes {
		go initNode(node)
	}
	s.waitResults(results, len(nodes))
}

func (s *testSuite) StartNodesNetwork(nodes []*networkNode) {
	log.Info("Start nodes")

	results := make(chan error, len(nodes))
	startNode := func(node *networkNode) {
		err := node.componentManager.Start(node.ctx)
		results <- err
	}

	for _, node := range nodes {
		go startNode(node)
	}
	s.waitResults(results, len(nodes))
	atomic.StoreUint32(&s.fixture().discoveriesAreBootstrapped, 1)
}

// TearDownSuite shutdowns all nodes in network, calls once after all tests in suite finished
func (s *consensusSuite) TearDownTest() {
	log.Info("=================== TearDownTest()")
	log.Info("Stop network nodes")
	for _, n := range s.fixture().networkNodes {
		err := n.componentManager.Stop(n.ctx)
		s.NoError(err)
	}
	log.Info("Stop bootstrap nodes")
	for _, n := range s.fixture().bootstrapNodes {
		err := n.componentManager.Stop(n.ctx)
		s.NoError(err)
	}
	log.Info("Stop test pulsar")
	s.fixture().pulsar.Stop(s.fixture().ctx)
}

func (s *consensusSuite) waitForConsensus(consensusCount int) {
	for i := 0; i < consensusCount; i++ {
		for i, n := range s.fixture().bootstrapNodes {
			err := <-n.consensusResult
			require.NoError(s.T(), err, "Failed to pass consensus on bootstrapNodes[%d] %s ", i, n.host)
		}

		for _, n := range s.fixture().networkNodes {
			err := <-n.consensusResult
			require.NoError(s.T(), err, "Failed to pass consensus on networkNodes[%d] %s ", i, n.host)
		}
	}
}

func (s *consensusSuite) waitForConsensusExcept(consensusCount int, exception insolar.Reference) {
	for i := 0; i < consensusCount; i++ {
		for _, n := range s.fixture().bootstrapNodes {
			if n.id.Equal(exception) {
				continue
			}
			err := <-n.consensusResult
			require.NoError(s.T(), err, "Failed to pass consensus on bootstrapNodes[%d] %s ", i, n.host)
		}

		for _, n := range s.fixture().networkNodes {
			if n.id.Equal(exception) {
				continue
			}
			err := <-n.consensusResult
			require.NoError(s.T(), err, "Failed to pass consensus on networkNodes[%d] %s ", i, n.host)
		}
	}
}

// nodesCount returns count of nodes in network without testNode
func (s *testSuite) getNodesCount() int {
	return len(s.fixture().bootstrapNodes) + len(s.fixture().networkNodes)
}

func (s *testSuite) InitNode(node *networkNode) {
	if node.componentManager != nil {
		err := node.init()
		s.Require().NoError(err)
	}
}

func (s *testSuite) StartNode(node *networkNode) {
	if node.componentManager != nil {
		err := node.componentManager.Start(node.ctx)
		s.Require().NoError(err)
	}
}

func (s *testSuite) StopNode(node *networkNode) {
	if node.componentManager != nil {
		err := node.componentManager.Stop(s.fixture().ctx)
		s.NoError(err)
	}
}

type networkNode struct {
	id                  insolar.Reference
	role                insolar.StaticRole
	privateKey          crypto.PrivateKey
	cryptographyService insolar.CryptographyService
	host                string
	ctx                 context.Context

	componentManager   *component.Manager
	serviceNetwork     *servicenetwork.ServiceNetwork
	terminationHandler *testutils.TerminationHandlerMock
	consensusResult    chan error
}

// newNetworkNode returns networkNode initialized only with id, host address and key pair
func (s *testSuite) newNetworkNode(name string) *networkNode {
	key, err := platformpolicy.NewKeyProcessor().GeneratePrivateKey()
	s.Require().NoError(err)
	address := "127.0.0.1:" + strconv.Itoa(incrementTestPort())

	nodeContext, _ := inslogger.WithField(s.fixture().ctx, "nodeName", name)
	return &networkNode{
		id:                  testutils.RandomRef(),
		role:                RandomRole(),
		privateKey:          key,
		cryptographyService: cryptography.NewKeyBoundCryptographyService(key),
		host:                address,
		ctx:                 nodeContext,
		consensusResult:     make(chan error, 30),
	}
}

func incrementTestPort() int {
	result := atomic.AddUint32(&testNetworkPort, 1)
	return int(result)
}

// init calls Init for node component manager and wraps PhaseManager
func (n *networkNode) init() error {
	err := n.componentManager.Init(n.ctx)
	n.serviceNetwork.PhaseManager = &phaseManagerWrapper{original: n.serviceNetwork.PhaseManager, result: n.consensusResult}
	return err
}

func (s *testSuite) initCrypto(node *networkNode) (*certificate.CertificateManager, insolar.CryptographyService) {
	pubKey, err := node.cryptographyService.GetPublicKey()
	s.Require().NoError(err)

	// init certificate

	proc := platformpolicy.NewKeyProcessor()
	publicKey, err := proc.ExportPublicKeyPEM(pubKey)
	s.Require().NoError(err)

	cert := &certificate.Certificate{}
	cert.PublicKey = string(publicKey[:])
	cert.Reference = node.id.String()
	cert.Role = node.role.String()
	cert.BootstrapNodes = make([]certificate.BootstrapNode, 0)

	for _, b := range s.fixture().bootstrapNodes {
		pubKey, _ := b.cryptographyService.GetPublicKey()
		pubKeyBuf, err := proc.ExportPublicKeyPEM(pubKey)
		s.Require().NoError(err)

		bootstrapNode := certificate.NewBootstrapNode(
			pubKey,
			string(pubKeyBuf[:]),
			b.host,
			b.id.String())

		cert.BootstrapNodes = append(cert.BootstrapNodes, *bootstrapNode)
	}

	// dump cert and read it again from json for correct private files initialization
	jsonCert, err := cert.Dump()
	s.Require().NoError(err)
	log.Infof("cert: %s", jsonCert)

	cert, err = certificate.ReadCertificateFromReader(pubKey, proc, strings.NewReader(jsonCert))
	s.Require().NoError(err)
	return certificate.NewCertificateManager(cert), node.cryptographyService
}

func RandomRole() insolar.StaticRole {
	i := rand.Int()%3 + 1
	return insolar.StaticRole(i)
}

type pulseManagerMock struct {
	pulse insolar.Pulse
	lock  sync.Mutex

	keeper network.NodeKeeper
}

func newPulseManagerMock(keeper network.NodeKeeper) *pulseManagerMock {
	return &pulseManagerMock{pulse: *insolar.GenesisPulse, keeper: keeper}
}

func (p *pulseManagerMock) ForPulseNumber(context.Context, insolar.PulseNumber) (insolar.Pulse, error) {
	panic("not implemented")
}

func (p *pulseManagerMock) Latest(ctx context.Context) (insolar.Pulse, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.pulse, nil
}

func (p *pulseManagerMock) Set(ctx context.Context, pulse insolar.Pulse) error {
	p.lock.Lock()
	p.pulse = pulse
	p.lock.Unlock()

	return p.keeper.MoveSyncToActive(ctx, pulse.PulseNumber)
}

type staterMock struct {
	stateFunc func() []byte
}

func (m staterMock) State() []byte {
	return m.stateFunc()
}

type PublisherMock struct{}

func (p *PublisherMock) Publish(topic string, messages ...*message.Message) error {
	return nil
}

func (p *PublisherMock) Close() error {
	return nil
}

// preInitNode inits previously created node with mocks and external dependencies
func (s *testSuite) preInitNode(node *networkNode) {
	t := s.T()
	cfg := configuration.NewConfiguration()
	cfg.Pulsar.PulseTime = pulseTimeMs
	cfg.Host.Transport.Address = node.host
	cfg.Service.Skip = 5
	cfg.Service.CacheDirectory = cacheDir + node.host
	cfg.Service.Consensus.Phase1Timeout = Phase1Timeout
	cfg.Service.Consensus.Phase2Timeout = Phase2Timeout
	cfg.Service.Consensus.Phase21Timeout = Phase21Timeout
	cfg.Service.Consensus.Phase3Timeout = Phase3Timeout

	node.componentManager = &component.Manager{}
	node.componentManager.Register(platformpolicy.NewPlatformCryptographyScheme(), rules.NewRules())
	serviceNetwork, err := servicenetwork.NewServiceNetwork(cfg, node.componentManager)
	s.Require().NoError(err)

	amMock := staterMock{
		stateFunc: func() []byte {
			return make([]byte, packets.HashLength)
		},
	}

	certManager, cryptographyService := s.initCrypto(node)

	realKeeper, err := nodenetwork.NewNodeNetwork(cfg.Host.Transport, certManager.GetCertificate())
	s.Require().NoError(err)
	terminationHandler := testutils.NewTerminationHandlerMock(t)
	terminationHandler.LeaveFunc = func(p context.Context, p1 insolar.PulseNumber) {}
	terminationHandler.OnLeaveApprovedFunc = func(p context.Context) {}
	terminationHandler.AbortFunc = func(reason string) { log.Error(reason) }

	keyProc := platformpolicy.NewKeyProcessor()
	pubMock := &PublisherMock{}
	senderMock := bus.NewSenderMock(t)
	if UseFakeTransport {
		// little hack: this Register will override transport.Factory
		// in servicenetwork internal component manager with fake factory
		node.componentManager.Register(transport.NewFakeFactory(cfg.Host.Transport))
	}
	if UseFakeBootstrap {
		// little hack: this Register will override DiscoveryBootstrapper
		// in servicenetwork internal component manager with fakeBootstrap
		node.componentManager.Register(newFakeBootstrap(s.fixture()))
	}

	mb := testutils.NewMessageBusMock(t)
	mb.MustRegisterMock.Return()

	node.componentManager.Inject(realKeeper, newPulseManagerMock(realKeeper.(network.NodeKeeper)), pubMock,
		&amMock, certManager, cryptographyService, serviceNetwork, keyProc, terminationHandler,
		mb, testutils.NewContractRequesterMock(t), senderMock)

	serviceNetwork.SetOperableFunc(func(ctx context.Context, operable bool) {
	})
	serviceNetwork.SetGateway(NewFakeGateway())

	node.serviceNetwork = serviceNetwork
	node.terminationHandler = terminationHandler
}

func (s *testSuite) SetCommunicationPolicy(policy CommunicationPolicy) {
	if policy == FullTimeout {
		s.fixture().pulsar.Pause()
		defer s.fixture().pulsar.Continue()

		wrapper := s.fixture().bootstrapNodes[1].serviceNetwork.PhaseManager.(*phaseManagerWrapper)
		wrapper.original = &FullTimeoutPhaseManager{}
		s.fixture().bootstrapNodes[1].serviceNetwork.PhaseManager = wrapper
		return
	}

	ref := s.fixture().bootstrapNodes[0].id // TODO: should we declare argument to select this node?
	s.SetCommunicationPolicyForNode(ref, policy)
}

func (s *testSuite) SetCommunicationPolicyForNode(nodeID insolar.Reference, policy CommunicationPolicy) {
	nodes := s.fixture().bootstrapNodes
	timedOutNodesCount := 0
	switch policy {
	case PartialNegative1Phase, PartialNegative2Phase, PartialNegative3Phase, PartialNegative23Phase:
		timedOutNodesCount = int(float64(len(nodes)) * 0.6)
	case PartialPositive1Phase, PartialPositive2Phase, PartialPositive3Phase, PartialPositive23Phase:
		timedOutNodesCount = int(float64(len(nodes)) * 0.2)
	case SplitCase:
		timedOutNodesCount = int(float64(len(nodes)) * 0.5)
	}

	s.fixture().pulsar.Pause()
	defer s.fixture().pulsar.Continue()

	for i := 1; i <= timedOutNodesCount; i++ {
		comm := nodes[i].serviceNetwork.PhaseManager.(*phaseManagerWrapper).original.(*phases.Phases).FirstPhase.(*phases.FirstPhaseImpl).Communicator
		wrapper := &CommunicatorMock{communicator: comm, ignoreFrom: nodeID, policy: policy}
		phasemanager := nodes[i].serviceNetwork.PhaseManager.(*phaseManagerWrapper).original.(*phases.Phases)
		phasemanager.FirstPhase.(*phases.FirstPhaseImpl).Communicator = wrapper
		phasemanager.SecondPhase.(*phases.SecondPhaseImpl).Communicator = wrapper
		phasemanager.ThirdPhase.(*phases.ThirdPhaseImpl).Communicator = wrapper
	}
}

func (s *testSuite) AssertActiveNodesCountDelta(delta int) {
	activeNodes := s.fixture().bootstrapNodes[0].serviceNetwork.NodeKeeper.GetAccessor().GetActiveNodes()
	s.Equal(s.getNodesCount()+delta, len(activeNodes))
}

func (s *testSuite) AssertWorkingNodesCountDelta(delta int) {
	workingNodes := s.fixture().bootstrapNodes[0].serviceNetwork.NodeKeeper.GetAccessor().GetWorkingNodes()
	s.Equal(s.getNodesCount()+delta, len(workingNodes))
}
