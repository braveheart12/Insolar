///
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
///

package purgatory

//type purgatorySlot struct {
//	nodeID insolar.ShortNodeID
//	svf    cryptkit.SignatureVerifierFactory
//	//limiter
//
//	mutex sync.RWMutex
//	sv    cryptkit.SignatureVerifier
//	intro profiles.StaticProfile
//	index member.Index
//	mode  member.OpMode
//	pw    member.Power
//}
//
//func (p *purgatorySlot) GetStatic() profiles.StaticProfile {
//	p.mutex.RLock()
//	defer p.mutex.RUnlock()
//
//	return p.intro
//}
//
//func (p *purgatorySlot) SetNodeIntroProfile(nip profiles.StaticProfile) {
//
//	if p.nodeID != nip.GetStaticNodeID() {
//		panic("illegal value")
//	}
//
//	p.mutex.Lock()
//	defer p.mutex.Unlock()
//	if p.intro == nip {
//		return
//	}
//	p.intro = nip
//	p.sv = nil
//}
//
//func (p *purgatorySlot) GetDefaultEndpoint() endpoints.Outbound {
//	return p.GetStatic().GetDefaultEndpoint()
//}
//
//func (p *purgatorySlot) GetPublicKeyStore() cryptkit.PublicKeyStore {
//	return p.GetStatic().GetPublicKeyStore()
//}
//
//func (p *purgatorySlot) IsAcceptableHost(from endpoints.Inbound) bool {
//	return p.GetStatic().IsAcceptableHost(from)
//}
//
//func (p *purgatorySlot) GetNodeID() insolar.ShortNodeID {
//	return p.nodeID
//}
//
//func (p *purgatorySlot) GetStartPower() member.Power {
//	return p.GetStatic().GetStartPower()
//}
//
//func (p *purgatorySlot) GetPrimaryRole() member.PrimaryRole {
//	return p.GetStatic().GetPrimaryRole()
//}
//
//func (p *purgatorySlot) GetSpecialRoles() member.SpecialRole {
//	return p.GetStatic().GetSpecialRoles()
//}
//
//func (p *purgatorySlot) GetNodePublicKey() cryptkit.SignatureKeyHolder {
//	return p.GetStatic().GetNodePublicKey()
//}
//
//func (p *purgatorySlot) GetAnnouncementSignature() cryptkit.SignatureHolder {
//	return p.GetStatic().GetAnnouncementSignature()
//}
//
//func (p *purgatorySlot) GetIntroduction() profiles.NodeIntroduction {
//	return p.GetStatic().GetIntroduction()
//}
//
//func (p *purgatorySlot) GetSignatureVerifier() cryptkit.SignatureVerifier {
//	p.mutex.RLock()
//	if p.sv != nil || p.svf == nil {
//		return p.sv
//	}
//	p.mutex.RUnlock()
//	return p.createSignatureVerifier()
//}
//
//func (p *purgatorySlot) createSignatureVerifier() cryptkit.SignatureVerifier {
//	p.mutex.Lock()
//	defer p.mutex.Unlock()
//	if p.sv == nil {
//		p.sv = p.svf.GetSignatureVerifierWithPKS(p.intro.GetPublicKeyStore())
//	}
//	return p.sv
//}
//
//func (p *purgatorySlot) GetOpMode() member.OpMode {
//	if p.index.IsJoiner() {
//		return member.ModeNormal
//	} else {
//		return p.mode
//	}
//}
//
//func (p *purgatorySlot) GetIndex() member.Index {
//	return p.index.Ensure()
//}
//
//func (p *purgatorySlot) IsJoiner() bool {
//	return p.index.IsJoiner()
//}
//
//func (p *purgatorySlot) GetDeclaredPower() member.Power {
//	if p.index.IsJoiner() || p.mode.IsPowerless() {
//		return 0
//	} else {
//		return p.pw
//	}
//}
