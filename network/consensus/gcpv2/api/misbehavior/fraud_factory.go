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

package misbehavior

import (
	"fmt"

	"github.com/insolar/insolar/network/consensus/common/endpoints"
	"github.com/insolar/insolar/network/consensus/gcpv2/api/profiles"
)

func IsFraud(err error) bool {
	_, ok := err.(*FraudError)
	return ok
}

func FraudOf(err error) *FraudError {
	rep, ok := err.(*FraudError)
	if ok {
		return rep
	}
	return nil
}

var _ Report = &FraudError{}

type FraudError struct {
	fraudType    int
	msg          string
	violatorHost endpoints.InboundConnection
	violatorNode profiles.ActiveNode
	details      []interface{}
	captureMark  interface{}
}

func (p *FraudError) IsUnknown() bool {
	return p.fraudType == 0
}

func (p *FraudError) MisbehaviorType() Type {
	return Fraud.Of(p.fraudType)
}

func (p *FraudError) CaptureMark() interface{} {
	return p.captureMark
}

func (p *FraudError) Details() []interface{} {
	return p.details
}

func (p *FraudError) ViolatorNode() profiles.ActiveNode {
	return p.violatorNode
}

func (p *FraudError) ViolatorHost() endpoints.InboundConnection {
	return p.violatorHost
}

func (p FraudError) FraudType() int {
	return p.fraudType
}

func (p FraudError) Error() string {
	var v interface{} = p.violatorNode
	if v == nil {
		v = p.violatorHost
	}
	c := p.captureMark
	if c == nil {
		c = ""
	}
	return fmt.Sprintf("fraud: type=%v(%v)%v, violator=%v, details=%+v", p.fraudType, p.msg, c, v, p.details)
}

const (
	_ = iota
	FraudMultipleNsh
	MismatchedRank
	MismatchedNeighbour
)

func NewFraudFactory(capture ReportFunc) FraudFactory {
	return FraudFactory{capture: capture}
}

type FraudFactory struct {
	capture ReportFunc
}

func (p FraudFactory) NewFraud(fraudType int, msg string, violatorHost endpoints.Inbound, violatorNode profiles.ActiveNode, details ...interface{}) FraudError {
	err := FraudError{
		fraudType:    fraudType,
		msg:          msg,
		violatorNode: violatorNode,
		details:      details,
	}
	if violatorHost != nil {
		err.violatorHost = endpoints.NewHostIdentityFromHolder(violatorHost)
	}
	if p.capture != nil {
		err.captureMark = p.capture(&err)
	}
	return err
}

func (p FraudFactory) NewNodeFraud(fraudType int, msg string, violatorNode profiles.ActiveNode, details ...interface{}) FraudError {
	return p.NewFraud(fraudType, msg, nil, violatorNode, details...)
}

func (p FraudFactory) NewHostFraud(fraudType int, msg string, violatorHost endpoints.Inbound, details ...interface{}) FraudError {
	return p.NewFraud(fraudType, msg, violatorHost, nil, details...)
}

func (p FraudFactory) NewInconsistentMembershipAnnouncement(violator profiles.ActiveNode,
	evidence1 profiles.MembershipAnnouncement, evidence2 profiles.MembershipAnnouncement) FraudError {
	return p.NewNodeFraud(FraudMultipleNsh, "multiple membership profile", violator, evidence1, evidence2)
}

func (p FraudFactory) NewMismatchedMembershipRank(violator profiles.ActiveNode, mp profiles.MembershipProfile) FraudError {
	return p.NewNodeFraud(MismatchedRank, "mismatched membership profile rank", violator, mp)
}

func (p FraudFactory) NewMismatchedMembershipRankOrNodeCount(violator profiles.ActiveNode, mp profiles.MembershipProfile, nodeCount int) FraudError {
	return p.NewNodeFraud(MismatchedRank, "mismatched membership profile node count", violator, mp, nodeCount)
}

func (p FraudFactory) NewUnknownNeighbour(violator profiles.ActiveNode) error {
	return p.NewNodeFraud(MismatchedNeighbour, "unknown neighbour", violator)
}

func (p FraudFactory) NewMismatchedNeighbourRank(violator profiles.ActiveNode) error {
	return p.NewNodeFraud(MismatchedNeighbour, "mismatched neighbour rank", violator)
}

func (p FraudFactory) NewNeighbourMissingTarget(violator profiles.ActiveNode) error {
	return p.NewNodeFraud(MismatchedNeighbour, "neighbour must include target node", violator)
}

func (p FraudFactory) NewNeighbourContainsRource(violator profiles.ActiveNode) error {
	return p.NewNodeFraud(MismatchedNeighbour, "neighbour must NOT include source node", violator)
}
