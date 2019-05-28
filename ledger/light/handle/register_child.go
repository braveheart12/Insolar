//
// Copyright 2019 Insolar Technologies GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package handle

import (
	"context"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/flow"
	"github.com/insolar/insolar/insolar/flow/bus"
	"github.com/insolar/insolar/insolar/message"
	"github.com/insolar/insolar/ledger/light/proc"
)

type RegisterChild struct {
	dep     *proc.Dependencies
	replyTo chan<- bus.Reply
	message *message.RegisterChild
	pulse   insolar.PulseNumber
}

func NewRegisterChild(dep *proc.Dependencies, rep chan<- bus.Reply, msg *message.RegisterChild, pulse insolar.PulseNumber) *RegisterChild {
	return &RegisterChild{
		dep:     dep,
		replyTo: rep,
		message: msg,
		pulse:   pulse,
	}
}

func (s *RegisterChild) Present(ctx context.Context, f flow.Flow) error {
	jet := proc.NewFetchJet(*s.message.DefaultTarget().Record(), flow.Pulse(ctx), s.replyTo)
	s.dep.FetchJet(jet)
	if err := f.Procedure(ctx, jet, true); err != nil {
		return err
	}

	hot := proc.NewWaitHot(jet.Result.Jet, flow.Pulse(ctx), s.replyTo)
	s.dep.WaitHot(hot)
	if err := f.Procedure(ctx, hot, true); err != nil {
		return err
	}

	getIndex := proc.NewGetIndex(s.message.Parent, jet.Result.Jet, s.replyTo, flow.Pulse(ctx))
	s.dep.GetIndex(getIndex)
	err := f.Procedure(ctx, getIndex, true)
	if err != nil {
		return err
	}

	registerChild := proc.NewRegisterChild(jet.Result.Jet, s.message, s.pulse, getIndex.Result.Index, s.replyTo)
	s.dep.RegisterChild(registerChild)
	return f.Procedure(ctx, registerChild, false)
}
