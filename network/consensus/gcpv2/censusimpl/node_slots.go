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

package censusimpl

import (
	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/network/consensus/common/cryptkit"
	"github.com/insolar/insolar/network/consensus/gcpv2/api/member"
	"github.com/insolar/insolar/network/consensus/gcpv2/api/profiles"
)

var _ profiles.LocalNode = &NodeProfileSlot{}

type NodeProfileSlot struct {
	profiles.StaticProfile
	verifier cryptkit.SignatureVerifier
	index    member.Index
	mode     member.OpMode
	power    member.Power
}

func (c *NodeProfileSlot) GetNodeID() insolar.ShortNodeID {
	return c.GetStaticNodeID()
}

func (c *NodeProfileSlot) GetStatic() profiles.StaticProfile {
	return c.StaticProfile
}

func NewNodeProfile(index member.Index, p profiles.StaticProfile, verifier cryptkit.SignatureVerifier, pw member.Power) NodeProfileSlot {

	return NodeProfileSlot{index: index.Ensure(), StaticProfile: p, verifier: verifier, power: pw}
}

func NewJoinerProfile(p profiles.StaticProfile, verifier cryptkit.SignatureVerifier, pw member.Power) NodeProfileSlot {

	return NodeProfileSlot{index: member.JoinerIndex, StaticProfile: p, verifier: verifier, power: pw}
}

func (c *NodeProfileSlot) GetDeclaredPower() member.Power {
	return c.power
}

func (c *NodeProfileSlot) GetOpMode() member.OpMode {
	return c.mode
}

func (c *NodeProfileSlot) LocalNodeProfile() {
}

func (c *NodeProfileSlot) GetIndex() member.Index {
	return c.index.Ensure()
}

func (c *NodeProfileSlot) IsJoiner() bool {
	return c.index.IsJoiner()
}

func (c *NodeProfileSlot) GetSignatureVerifier() cryptkit.SignatureVerifier {
	return c.verifier
}

var _ profiles.Updatable = &updatableSlot{}

type updatableSlot struct {
	NodeProfileSlot
	leaveReason uint32
}

func (c *updatableSlot) AsActiveNode() profiles.ActiveNode {
	return &c.NodeProfileSlot
}

func (c *updatableSlot) SetRank(index member.Index, m member.OpMode, power member.Power) {
	c.index = index.Ensure()
	c.power = power
	c.mode = m
}

func (c *updatableSlot) SetPower(power member.Power) {
	c.power = power
}

func (c *updatableSlot) SetOpMode(m member.OpMode) {
	c.mode = m
}

func (c *updatableSlot) SetOpModeAndLeaveReason(leaveReason uint32) {
	c.mode = member.ModeEvictedGracefully
	c.leaveReason = leaveReason
}

func (c *updatableSlot) GetLeaveReason() uint32 {
	if c.mode != member.ModeEvictedGracefully {
		return 0
	}
	return c.leaveReason
}

func (c *updatableSlot) SetIndex(index member.Index) {
	c.index = index.Ensure()
}

func (c *updatableSlot) SetSignatureVerifier(verifier cryptkit.SignatureVerifier) {
	c.verifier = verifier
}
