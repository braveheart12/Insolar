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

package api

import (
	"context"
	"time"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/network/consensus/common/capacity"
	"github.com/insolar/insolar/network/consensus/common/cryptkit"
	"github.com/insolar/insolar/network/consensus/common/endpoints"
	"github.com/insolar/insolar/network/consensus/common/pulse"
	"github.com/insolar/insolar/network/consensus/gcpv2/api/census"
	"github.com/insolar/insolar/network/consensus/gcpv2/api/member"
	"github.com/insolar/insolar/network/consensus/gcpv2/api/power"
	"github.com/insolar/insolar/network/consensus/gcpv2/api/profiles"
	"github.com/insolar/insolar/network/consensus/gcpv2/api/transport"
)

type ConsensusController interface {
	ProcessPacket(ctx context.Context, payload transport.PacketParser, from endpoints.Inbound) error

	/* Ungraceful stop */
	Abort()
	/* Graceful exit, actual moment of leave will be indicated via Upstream */
	// RequestLeave()

	/* This node power in the active population, and pulse number of such. Without active population returns (0,0) */
	GetActivePowerLimit() (member.Power, pulse.Number)
}

//go:generate minimock -i github.com/insolar/insolar/network/consensus/gcpv2/api.CandidateControlFeeder -o . -s _mock.go
type CandidateControlFeeder interface {
	PickNextJoinCandidate() profiles.CandidateProfile
	RemoveJoinCandidate(candidateAdded bool, nodeID insolar.ShortNodeID) bool
}

type TrafficControlFeeder interface {
	/* Application traffic should be stopped or throttled down for the given duration
	LevelMax and LevelNormal should be considered equal, and duration doesnt apply to them
	*/
	SetTrafficLimit(level capacity.Level, duration time.Duration)

	/* Application traffic can be resumed at full */
	ResumeTraffic()
}

type ConsensusControlFeeder interface {
	TrafficControlFeeder

	GetRequiredPowerLevel() power.Request
	OnAppliedPowerLevel(pw member.Power, effectiveSince pulse.Number)

	GetRequiredGracefulLeave() (bool, uint32)
	OnAppliedGracefulLeave(exitCode uint32, effectiveSince pulse.Number)

	/* Called on receiving seem-to-be-valid Pulsar or Phase0 packets. Can be called multiple time in sequence.
	Application MUST NOT consider it as a new pulse. */
	PulseDetected()

	/* Consensus is finished. If expectedCensus == 0 then this node was evicted from consensus.	*/
	ConsensusFinished(report UpstreamReport, expectedCensus census.Operational)

	// /* Consensus has stopped abnormally	*/
	// ConsensusFailed(report UpstreamReport)
}

type RoundController interface {
	HandlePacket(ctx context.Context, packet transport.PacketParser, from endpoints.Inbound) error
	StopConsensusRound()
	StartConsensusRound(upstream UpstreamController)
}

type RoundControllerFactory interface {
	CreateConsensusRound(chronicle ConsensusChronicles, controlFeeder ConsensusControlFeeder,
		candidateFeeder CandidateControlFeeder, prevPulseRound RoundController) RoundController
	GetLocalConfiguration() LocalNodeConfiguration
}

type LocalNodeConfiguration interface {
	GetConsensusTimings(nextPulseDelta uint16, isJoiner bool) RoundTimings
	GetSecretKeyStore() cryptkit.SecretKeyStore
	GetParentContext() context.Context
}
