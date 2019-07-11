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

package census

import (
	"fmt"
	"github.com/insolar/insolar/network/consensus/common/pulse_data"
	"github.com/insolar/insolar/network/consensus/gcpv2/api"
	"sync"
)

func NewLocalChronicles() LocalConsensusChronicles {
	return &localChronicles{}
}

type LocalConsensusChronicles interface {
	api.ConsensusChronicles
	makeActive(ce api.ExpectedCensus, ca localActiveCensus)
}

var _ api.ConsensusChronicles = &localChronicles{}

type localActiveCensus interface {
	api.ActiveCensus
	getVersionedRegistries() api.VersionedRegistries
	setVersionedRegistries(vr api.VersionedRegistries)
}

type localChronicles struct {
	rw                  sync.RWMutex
	active              localActiveCensus
	expected            api.ExpectedCensus
	expectedPulseNumber pulse_data.PulseNumber
}

func (c *localChronicles) GetLatestCensus() api.OperationalCensus {
	c.rw.RLock()
	defer c.rw.RUnlock()

	if c.expected != nil {
		return c.expected
	}
	return c.active
}

func (c *localChronicles) GetRecentCensus(pn pulse_data.PulseNumber) api.OperationalCensus {
	c.rw.RLock()
	defer c.rw.RUnlock()

	if c.expected != nil && pn == c.expected.GetPulseNumber() {
		return c.expected
	}

	if pn == c.active.GetPulseNumber() {
		return c.active
	}
	panic(fmt.Sprintf("recent census is missing for pulse (%v)", pn))
}

func (c *localChronicles) GetActiveCensus() api.ActiveCensus {
	c.rw.RLock()
	defer c.rw.RUnlock()

	return c.active
}

func (c *localChronicles) GetExpectedCensus() api.ExpectedCensus {
	c.rw.RLock()
	defer c.rw.RUnlock()

	return c.expected
}

func (c *localChronicles) makeActive(ce api.ExpectedCensus, ca localActiveCensus) {
	c.rw.Lock()
	defer c.rw.Unlock()

	if c.expected != ce {
		panic("illegal state")
	}
	if !c.expectedPulseNumber.IsUnknown() && c.expectedPulseNumber != ca.GetPulseNumber() {
		panic("unexpected pulse number")
	}
	pd := ca.GetPulseData()

	if c.active != nil {
		pd.EnsurePulseData()
		registries := c.active.getVersionedRegistries().CommitNextPulse(pd, ca.GetOnlinePopulation())
		c.expectedPulseNumber = pd.GetNextPulseNumber() // should go before any updates as it may panic
		ca.setVersionedRegistries(registries)
	} else {
		switch {
		case ca.getVersionedRegistries() == nil:
			panic("versioned registries are nil")
		case pd.IsExpectedPulse():
			c.expectedPulseNumber = pd.GetPulseNumber()
		case pd.IsValidPulseData():
			c.expectedPulseNumber = pd.GetNextPulseNumber()
		default:
			c.expectedPulseNumber = pulse_data.UnknownPulseNumber
		}
	}

	c.active = ca
	c.expected = nil
}

func (c *localChronicles) makeExpected(ce api.ExpectedCensus) {
	c.rw.Lock()
	defer c.rw.Unlock()

	if c.expected != nil {
		panic("illegal state")
	}

	if !c.expectedPulseNumber.IsUnknown() && c.expectedPulseNumber != ce.GetPulseNumber() {
		panic("unexpected pulse number")
	}

	c.expected = ce
}
