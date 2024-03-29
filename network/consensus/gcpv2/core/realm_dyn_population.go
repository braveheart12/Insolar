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

package core

import (
	"context"
	"fmt"
	"sync"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/network/consensus/common/consensuskit"
	"github.com/insolar/insolar/network/consensus/gcpv2/api/profiles"
)

func NewDynamicRealmPopulation(baselineWeight uint32, local profiles.ActiveNode, nodeCountHint int, phase2ExtLimit uint8,
	fn NodeInitFunc) *DynamicRealmPopulation {

	r := &DynamicRealmPopulation{
		nodeInit:       fn,
		baselineWeight: baselineWeight,
		phase2ExtLimit: phase2ExtLimit,
	}
	r.initPopulation(local, nodeCountHint)

	return r
}

var _ RealmPopulation = &DynamicRealmPopulation{}

type DynamicRealmPopulation struct {
	nodeInit       NodeInitFunc
	baselineWeight uint32
	phase2ExtLimit uint8
	self           *NodeAppearance

	rw sync.RWMutex

	joinerCount  int
	indexedCount int

	nodeIndex    []*NodeAppearance
	nodeShuffle  []*NodeAppearance // excluding self
	dynamicNodes map[insolar.ShortNodeID]*NodeAppearance
}

func (r *DynamicRealmPopulation) initPopulation(local profiles.ActiveNode, nodeCountHint int) {
	r.self = r.CreateNodeAppearance(context.Background(), local)
	r.dynamicNodes = make(map[insolar.ShortNodeID]*NodeAppearance, nodeCountHint)
}

func (r *DynamicRealmPopulation) GetSelf() *NodeAppearance {
	return r.self
}

func (r *DynamicRealmPopulation) GetNodeCount() int {
	r.rw.RLock()
	defer r.rw.RUnlock()
	return len(r.nodeIndex)
}

func (r *DynamicRealmPopulation) GetJoinersCount() int {
	r.rw.RLock()
	defer r.rw.RUnlock()
	return r.joinerCount
}

func (r *DynamicRealmPopulation) GetOthersCount() int {
	return r.GetNodeCount() - 1
}

func (r *DynamicRealmPopulation) GetBftMajorityCount() int {
	return consensuskit.BftMajority(r.GetNodeCount())
}

func (r *DynamicRealmPopulation) GetNodeAppearance(id insolar.ShortNodeID) *NodeAppearance {
	r.rw.RLock()
	defer r.rw.RUnlock()

	return r.dynamicNodes[id]
}

func (r *DynamicRealmPopulation) GetActiveNodeAppearance(id insolar.ShortNodeID) *NodeAppearance {
	na := r.GetNodeAppearance(id)
	if !na.GetProfile().IsJoiner() {
		return na
	}
	return nil
}

func (r *DynamicRealmPopulation) GetJoinerNodeAppearance(id insolar.ShortNodeID) *NodeAppearance {
	na := r.GetNodeAppearance(id)
	if !na.GetProfile().IsJoiner() {
		return nil
	}
	return na
}

func (r *DynamicRealmPopulation) GetNodeAppearanceByIndex(idx int) *NodeAppearance {
	if idx < 0 {
		panic("illegal value")
	}

	r.rw.RLock()
	defer r.rw.RUnlock()

	if idx >= len(r.nodeIndex) {
		return nil
	}
	return r.nodeIndex[idx]
}

func (r *DynamicRealmPopulation) GetShuffledOtherNodes() []*NodeAppearance {
	r.rw.RLock()
	defer r.rw.RUnlock()

	return r.nodeShuffle
}

func (r *DynamicRealmPopulation) IsComplete() bool {
	r.rw.RLock()
	defer r.rw.RUnlock()

	return len(r.nodeIndex) == r.indexedCount
}

func (r *DynamicRealmPopulation) GetIndexedNodes() []*NodeAppearance {
	cp, _ := r.GetIndexedNodesWithCheck()
	// if !ok {
	//	panic("node set is incomplete")
	// }
	return cp
}

func (r *DynamicRealmPopulation) GetIndexedNodesWithCheck() ([]*NodeAppearance, bool) {
	r.rw.RLock()
	defer r.rw.RUnlock()

	cp := make([]*NodeAppearance, len(r.nodeIndex))
	copy(cp, r.nodeIndex)

	return cp, len(r.nodeIndex) == r.indexedCount
}

func (r *DynamicRealmPopulation) CreateNodeAppearance(ctx context.Context, np profiles.ActiveNode) *NodeAppearance {

	n := &NodeAppearance{}
	n.init(np, nil, r.baselineWeight, r.phase2ExtLimit)
	r.nodeInit(ctx, n)

	return n
}

func (r *DynamicRealmPopulation) AddToDynamics(n *NodeAppearance) *NodeAppearance {
	nip := n.profile.GetStatic()

	if nip.GetIntroduction() == nil {
		panic("illegal value")
	}

	r.rw.Lock()
	defer r.rw.Unlock()

	id := nip.GetStaticNodeID()

	na := r.dynamicNodes[id]
	if na != nil {
		return na
	}

	if n.profile.IsJoiner() {
		r.joinerCount++
	} else {
		ni := n.profile.GetIndex()
		switch {
		case ni.AsInt() == len(r.nodeIndex):
			r.nodeIndex = append(r.nodeIndex, n)
		case ni.AsInt() > len(r.nodeIndex):
			r.nodeIndex = append(r.nodeIndex, make([]*NodeAppearance, 1+ni.AsInt()-len(r.nodeIndex))...)
			r.nodeIndex[ni] = n
		default:
			if r.nodeIndex[ni] != nil {
				panic(fmt.Sprintf("duplicate node id(%v)", ni))
			}
			r.nodeIndex[ni] = n
		}
		r.indexedCount++
		r.nodeShuffle = append(r.nodeShuffle, n)
	}
	return n
}

//
func (r *DynamicRealmPopulation) CreateVectorHelper() *RealmVectorHelper {
	r.rw.RLock()
	defer r.rw.RUnlock()

	v := &RealmVectorHelper{realmPopulation: r}
	v.setArrayNodes(r.nodeIndex, r.dynamicNodes, r.self.callback.GetPopulationVersion())
	v.realmPopulation = r
	return v
}
