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

package adapters

import (
	"math/rand"

	"github.com/insolar/insolar/network/consensus/common/cryptkit"
	"github.com/insolar/insolar/network/consensus/common/longbits"
	"github.com/insolar/insolar/network/consensus/gcpv2/api/member"
	"github.com/insolar/insolar/network/consensus/gcpv2/api/transport"
)

type seqDigester struct {
	// TODO do test or a proper digest calc
	rnd      *rand.Rand
	lastSeed int64
}

func (s *seqDigester) AddNext(digest longbits.FoldableReader) {
	// it is a dirty emulation of digest
	if s.rnd == nil {
		s.rnd = rand.New(rand.NewSource(0))
	}
	s.lastSeed = int64(s.rnd.Uint64() ^ digest.FoldToUint64())
	s.rnd.Seed(s.lastSeed)
}

func (s *seqDigester) GetDigestMethod() cryptkit.DigestMethod {
	return SHA3512Digest
}

func (s *seqDigester) ForkSequence() cryptkit.SequenceDigester {
	cp := seqDigester{}
	if s.rnd != nil {
		cp.rnd = rand.New(rand.NewSource(s.lastSeed))
	}
	return &cp
}

func (s *seqDigester) FinishSequence() cryptkit.Digest {
	if s.rnd == nil {
		panic("nothing")
	}
	var res longbits.Bits512
	bits64 := longbits.NewBits64(s.rnd.Uint64())
	copy(res[:], bits64[:])
	s.rnd = nil
	return cryptkit.NewDigest(&res, s.GetDigestMethod())
}

type gshDigester struct {
	sd            cryptkit.SequenceDigester
	defaultDigest longbits.FoldableReader
}

func (p *gshDigester) AddNext(digest longbits.FoldableReader, fullRank member.FullRank) {
	if digest == nil {
		p.sd.AddNext(p.defaultDigest)
	} else {
		p.sd.AddNext(digest)
	}
}

func (p *gshDigester) GetDigestMethod() cryptkit.DigestMethod {
	return p.sd.GetDigestMethod()
}

func (p *gshDigester) ForkSequence() transport.StateDigester {
	return &gshDigester{p.sd.ForkSequence(), p.defaultDigest}
}

func (p *gshDigester) FinishSequence() cryptkit.Digest {
	return p.sd.FinishSequence()
}
