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

package hostnetwork

import (
	"context"
	"encoding/gob"
	"sync"
	"testing"
	"time"

	"github.com/insolar/insolar/component"
	"github.com/insolar/insolar/configuration"
	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/log"
	"github.com/insolar/insolar/network"
	"github.com/insolar/insolar/network/hostnetwork/host"
	"github.com/insolar/insolar/network/hostnetwork/packet/types"
	"github.com/insolar/insolar/network/transport"
	"github.com/insolar/insolar/network/utils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	InvalidPacket types.PacketType = 1024

	ID1       = "4K2V1kpVycZ6qSFsNdz2FtpNxnJs17eBNzf9rdCMcKoe"
	ID2       = "4NwnA4HWZurKyXWNowJwYmb9CwX4gBKzwQKov1ExMf8M"
	ID3       = "4Ss5JMkXAD9Z7cktFEdrqeMuT6jGMF1pVozTyPHZ6zT4"
	IDUNKNOWN = "4K3Mi2hyZ6QKgynGv33sR5n3zWmSzdo8zv5Em7X26r1w"
	DOMAIN    = ".4F7BsTMVPKFshM1MwLf6y23cid6fL3xMpazVoF9krzUw"
)

type MockResolver struct {
	mu       sync.RWMutex
	mapping  map[insolar.Reference]*host.Host
	smapping map[insolar.ShortNodeID]*host.Host
}

func (m *MockResolver) ResolveConsensus(id insolar.ShortNodeID) (*host.Host, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result, exist := m.smapping[id]
	if !exist {
		return nil, errors.New("failed to resolve")
	}
	return result, nil
}

func (m *MockResolver) ResolveConsensusRef(nodeID insolar.Reference) (*host.Host, error) {
	return m.Resolve(nodeID)
}

func (m *MockResolver) Resolve(nodeID insolar.Reference) (*host.Host, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result, exist := m.mapping[nodeID]
	if !exist {
		return nil, errors.New("failed to resolve")
	}
	return result, nil
}

func (m *MockResolver) AddToKnownHosts(h *host.Host)      {}
func (m *MockResolver) Rebalance(network.PartitionPolicy) {}

func (m *MockResolver) addMapping(key, value string) error {
	k, err := insolar.NewReferenceFromBase58(key)
	if err != nil {
		return err
	}
	h, err := host.NewHostN(value, *k)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.mapping[*k] = h
	return nil
}

func (m *MockResolver) addMappingHost(h *host.Host) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.mapping[h.NodeID] = h
	m.smapping[h.ShortID] = h
}

func newMockResolver() *MockResolver {
	return &MockResolver{
		mapping:  make(map[insolar.Reference]*host.Host),
		smapping: make(map[insolar.ShortNodeID]*host.Host),
	}
}

func TestNewHostNetwork_InvalidReference(t *testing.T) {
	n, err := NewHostNetwork("invalid reference")
	require.Error(t, err)
	require.Nil(t, n)
}

type hostSuite struct {
	t        *testing.T
	ctx      context.Context
	id1, id2 string
	n1, n2   network.HostNetwork
	resolver *MockResolver
}

func newHostSuite(t *testing.T) *hostSuite {
	resolver := newMockResolver()
	id1 := ID1 + DOMAIN
	id2 := ID2 + DOMAIN

	cm1 := component.NewManager(nil)
	f1 := transport.NewFactory(configuration.NewHostNetwork().Transport)
	n1, err := NewHostNetwork(id1)
	require.NoError(t, err)
	cm1.Inject(f1, n1, resolver)

	cm2 := component.NewManager(nil)
	f2 := transport.NewFactory(configuration.NewHostNetwork().Transport)
	n2, err := NewHostNetwork(id2)
	require.NoError(t, err)
	cm2.Inject(f2, n2, resolver)

	ctx := context.Background()

	err = n1.Init(ctx)
	require.NoError(t, err)
	err = n2.Init(ctx)
	require.NoError(t, err)

	return &hostSuite{
		t: t, ctx: ctx, id1: id1, id2: id2, n1: n1, n2: n2, resolver: resolver,
	}
}

func (s *hostSuite) Start() {
	// start the second hostNetwork before the first because most test cases perform sending packets first -> second,
	// so the second hostNetwork should be ready to receive packets when the first starts to send
	err := s.n2.Start(s.ctx)
	require.NoError(s.t, err)
	err = s.n1.Start(s.ctx)
	require.NoError(s.t, err)

	err = s.resolver.addMapping(s.id1, s.n1.PublicAddress())
	require.NoError(s.t, err, "failed to add mapping %s -> %s: %s", s.id1, s.n1.PublicAddress(), err)
	err = s.resolver.addMapping(s.id2, s.n2.PublicAddress())
	require.NoError(s.t, err, "failed to add mapping %s -> %s: %s", s.id2, s.n2.PublicAddress(), err)
}

func (s *hostSuite) Stop() {
	// stop hostNetworks in the reverse order of their start
	_ = s.n1.Stop(s.ctx)
	_ = s.n2.Stop(s.ctx)
}

func TestNewHostNetwork(t *testing.T) {
	s := newHostSuite(t)
	defer s.Stop()

	count := 10
	wg := sync.WaitGroup{}
	wg.Add(count)

	handler := func(ctx context.Context, request network.Request) (network.Response, error) {
		log.Info("handler triggered")
		wg.Done()
		return s.n2.BuildResponse(ctx, request, nil), nil
	}
	s.n2.RegisterRequestHandler(types.Ping, handler)

	s.Start()

	for i := 0; i < count; i++ {
		request := s.n1.NewRequestBuilder().Type(types.Ping).Data(nil).Build()
		ref, err := insolar.NewReferenceFromBase58(ID2 + DOMAIN)
		require.NoError(t, err)
		_, err = s.n1.SendRequest(s.ctx, request, *ref)
		require.NoError(t, err)
	}

	wg.Wait()
}

func TestHostNetwork_SendRequestPacket(t *testing.T) {
	m := newMockResolver()
	ctx := context.Background()

	n1, err := NewHostNetwork(ID1 + DOMAIN)
	require.NoError(t, err)

	cm := component.NewManager(nil)
	cm.Register(m, n1, transport.NewFactory(configuration.NewHostNetwork().Transport))
	cm.Inject()
	err = cm.Init(ctx)
	require.NoError(t, err)
	err = cm.Start(ctx)
	require.NoError(t, err)

	defer func() {
		err = cm.Stop(ctx)
		assert.NoError(t, err)
	}()

	unknownID, err := insolar.NewReferenceFromBase58(IDUNKNOWN + DOMAIN)
	require.NoError(t, err)

	// should return error because cannot resolve NodeID -> Address
	request := n1.NewRequestBuilder().Type(types.Ping).Data(nil).Build()
	_, err = n1.SendRequest(ctx, request, *unknownID)
	require.Error(t, err)

	err = m.addMapping(ID2+DOMAIN, "abirvalg")
	require.Error(t, err)
	err = m.addMapping(ID3+DOMAIN, "127.0.0.1:7654")
	require.NoError(t, err)

	ref, err := insolar.NewReferenceFromBase58(ID2 + DOMAIN)
	require.NoError(t, err)
	// should return error because resolved address is invalid
	_, err = n1.SendRequest(ctx, request, *ref)
	require.Error(t, err)
}

func TestHostNetwork_SendRequestPacket2(t *testing.T) {
	s := newHostSuite(t)
	defer s.Stop()

	wg := sync.WaitGroup{}
	wg.Add(1)

	handler := func(ctx context.Context, r network.Request) (network.Response, error) {
		log.Info("handler triggered")
		ref, err := insolar.NewReferenceFromBase58(ID1 + DOMAIN)
		require.NoError(t, err)
		require.Equal(t, *ref, r.GetSender())
		require.Equal(t, s.n1.PublicAddress(), r.GetSenderHost().Address.String())
		wg.Done()
		return s.n2.BuildResponse(ctx, r, nil), nil
	}

	s.n2.RegisterRequestHandler(types.Ping, handler)

	s.Start()

	request := s.n1.NewRequestBuilder().Type(types.Ping).Data(nil).Build()
	ref, err := insolar.NewReferenceFromBase58(ID1 + DOMAIN)
	require.NoError(t, err)
	require.Equal(t, *ref, request.GetSender())
	require.Equal(t, s.n1.PublicAddress(), request.GetSenderHost().Address.String())

	ref, err = insolar.NewReferenceFromBase58(ID2 + DOMAIN)
	require.NoError(t, err)
	_, err = s.n1.SendRequest(s.ctx, request, *ref)
	require.NoError(t, err)

	wg.Wait()
}

func TestHostNetwork_SendRequestPacket3(t *testing.T) {
	s := newHostSuite(t)
	defer s.Stop()

	type Data struct {
		Number int
	}
	gob.Register(&Data{})

	handler := func(ctx context.Context, r network.Request) (network.Response, error) {
		log.Info("handler triggered")
		d := r.GetData().(*Data)
		return s.n2.BuildResponse(ctx, r, &Data{Number: d.Number + 1}), nil
	}
	s.n2.RegisterRequestHandler(types.Ping, handler)

	s.Start()

	magicNumber := 42
	request := s.n1.NewRequestBuilder().Type(types.Ping).Data(&Data{Number: magicNumber}).Build()
	ref, err := insolar.NewReferenceFromBase58(ID2 + DOMAIN)
	require.NoError(t, err)
	f, err := s.n1.SendRequest(s.ctx, request, *ref)
	require.NoError(t, err)
	require.Equal(t, f.Request().GetSender(), request.GetSender())

	r, err := f.WaitResponse(time.Minute)
	require.NoError(t, err)

	d := r.GetData().(*Data)
	require.Equal(t, magicNumber+1, d.Number)

	magicNumber = 666
	request = s.n1.NewRequestBuilder().Type(types.Ping).Data(&Data{Number: magicNumber}).Build()
	f, err = s.n1.SendRequest(s.ctx, request, *ref)
	require.NoError(t, err)

	r = <-f.Response()
	d = r.GetData().(*Data)
	require.Equal(t, magicNumber+1, d.Number)
}

func TestHostNetwork_SendRequestPacket_errors(t *testing.T) {
	s := newHostSuite(t)
	defer s.Stop()

	handler := func(ctx context.Context, r network.Request) (network.Response, error) {
		log.Info("handler triggered")
		time.Sleep(time.Second)
		return s.n2.BuildResponse(ctx, r, nil), nil
	}
	s.n2.RegisterRequestHandler(types.Ping, handler)

	s.Start()

	request := s.n1.NewRequestBuilder().Type(types.Ping).Data(nil).Build()
	ref, err := insolar.NewReferenceFromBase58(ID2 + DOMAIN)
	require.NoError(t, err)
	f, err := s.n1.SendRequest(s.ctx, request, *ref)
	require.NoError(t, err)

	_, err = f.WaitResponse(time.Millisecond)
	require.Error(t, err)

	f, err = s.n1.SendRequest(s.ctx, request, *ref)
	require.NoError(t, err)

	_, err = f.WaitResponse(time.Minute)
	require.NoError(t, err)
}

func TestHostNetwork_WrongHandler(t *testing.T) {
	s := newHostSuite(t)
	defer s.Stop()

	wg := sync.WaitGroup{}
	wg.Add(1)

	handler := func(ctx context.Context, r network.Request) (network.Response, error) {
		log.Info("handler triggered")
		wg.Done()
		return s.n2.BuildResponse(ctx, r, nil), nil
	}
	s.n2.RegisterRequestHandler(InvalidPacket, handler)

	s.Start()

	request := s.n1.NewRequestBuilder().Type(types.Ping).Build()
	ref, err := insolar.NewReferenceFromBase58(ID2 + DOMAIN)
	require.NoError(t, err)
	_, err = s.n1.SendRequest(s.ctx, request, *ref)
	require.NoError(t, err)

	// should timeout because there is no handler set for Ping packet
	result := utils.WaitTimeout(&wg, time.Millisecond*100)
	require.False(t, result)
}

func TestStartStopSend(t *testing.T) {
	s := newHostSuite(t)
	defer s.Stop()

	wg := sync.WaitGroup{}
	wg.Add(2)

	handler := func(ctx context.Context, r network.Request) (network.Response, error) {
		log.Info("handler triggered")
		wg.Done()
		return s.n2.BuildResponse(ctx, r, nil), nil
	}
	s.n2.RegisterRequestHandler(types.Ping, handler)

	s.Start()

	send := func() {
		request := s.n1.NewRequestBuilder().Type(types.Ping).Build()
		ref, err := insolar.NewReferenceFromBase58(ID2 + DOMAIN)
		require.NoError(t, err)
		f, err := s.n1.SendRequest(s.ctx, request, *ref)
		require.NoError(t, err)
		<-f.Response()
	}

	send()

	err := s.n1.Stop(s.ctx)
	require.NoError(t, err)
	err = s.n1.Start(s.ctx)
	require.NoError(t, err)

	send()
	wg.Wait()
}

func TestHostNetwork_SendRequestToHost_NotStarted(t *testing.T) {
	hn, err := NewHostNetwork(ID1 + DOMAIN)
	require.NoError(t, err)

	_, err = hn.SendRequestToHost(context.Background(), nil, nil)
	require.EqualError(t, err, "host network is not started")
}
