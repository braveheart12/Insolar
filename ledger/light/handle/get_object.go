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
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/insolar/insolar/insolar/payload"
	"github.com/pkg/errors"

	"github.com/insolar/insolar/insolar/flow"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/ledger/light/proc"
)

type GetObject struct {
	dep *proc.Dependencies

	message *message.Message
	passed  bool
}

func NewGetObject(dep *proc.Dependencies, msg *message.Message, passed bool) *GetObject {
	return &GetObject{
		dep:     dep,
		message: msg,
		passed:  passed,
	}
}

func (s *GetObject) Present(ctx context.Context, f flow.Flow) error {
	pl, err := payload.UnmarshalFromMeta(s.message.Payload)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal payload")
	}
	msg, ok := pl.(*payload.GetObject)
	if !ok {
		return fmt.Errorf("unexpected payload type: %T", pl)
	}

	ctx, _ = inslogger.WithField(ctx, "object", msg.ObjectID.DebugString())

	passIfNotExecutor := !s.passed
	jet := proc.NewCheckJet(msg.ObjectID, flow.Pulse(ctx), s.message, passIfNotExecutor)
	s.dep.CheckJet(jet)
	if err := f.Procedure(ctx, jet, false); err != nil {
		return err
	}
	objJetID := jet.Result.Jet

	hot := proc.NewWaitHotWM(objJetID, flow.Pulse(ctx), s.message)
	s.dep.WaitHotWM(hot)
	if err := f.Procedure(ctx, hot, false); err != nil {
		return err
	}

	idx := proc.NewGetIndexWM(msg.ObjectID, objJetID, s.message)
	s.dep.GetIndexWM(idx)
	if err := f.Procedure(ctx, idx, false); err != nil {
		return err
	}

	send := proc.NewSendObject(s.message, msg.ObjectID, idx.Result.Index)
	s.dep.SendObject(send)
	return f.Procedure(ctx, send, false)
}
