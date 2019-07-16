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

package logicrunner

import (
	"context"

	"github.com/pkg/errors"
	"go.opencensus.io/trace"

	"github.com/insolar/insolar/insolar"

	"github.com/insolar/insolar/insolar/bus"
	"github.com/insolar/insolar/insolar/flow"
	"github.com/insolar/insolar/insolar/payload"

	"github.com/insolar/insolar/insolar/message"
	"github.com/insolar/insolar/insolar/reply"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/instrumentation/instracer"
)

type initializeAbandonedRequestsNotificationExecutionState struct {
	LR  *LogicRunner
	msg *message.AbandonedRequestsNotification
}

// Proceed initializes or sets LedgerHasMoreRequests to right value
func (p *initializeAbandonedRequestsNotificationExecutionState) Proceed(ctx context.Context) error {
	ref := *p.msg.DefaultTarget()

	broker := p.LR.StateStorage.UpsertExecutionState(ref)

	broker.executionState.Lock()
	if broker.executionState.pending == insolar.PendingUnknown {
		broker.executionState.pending = insolar.InPending
		broker.executionState.PendingConfirmed = false
	}
	broker.executionState.LedgerHasMoreRequests = true
	broker.executionState.Unlock()

	return nil
}

type HandleAbandonedRequestsNotification struct {
	dep *Dependencies

	Message payload.Meta
	Parcel  insolar.Parcel
}

func (h *HandleAbandonedRequestsNotification) Present(ctx context.Context, f flow.Flow) error {
	ctx = loggerWithTargetID(ctx, h.Parcel)
	replyOk := bus.ReplyAsMessage(ctx, &reply.OK{})
	h.dep.Sender.Reply(ctx, h.Message, replyOk)
	return nil

	logger := inslogger.FromContext(ctx)

	logger.Debug("HandleAbandonedRequestsNotification.Present starts ...")

	msg, ok := h.Parcel.Message().(*message.AbandonedRequestsNotification)
	if !ok {
		return errors.New("HandleAbandonedRequestsNotification( ! message.AbandonedRequestsNotification )")
	}

	ctx, span := instracer.StartSpan(ctx, "HandleAbandonedRequestsNotification.Present")
	span.AddAttributes(trace.StringAttribute("msg.Type", msg.Type().String()))
	defer span.End()

	procInitializeExecutionState := initializeAbandonedRequestsNotificationExecutionState{
		LR:  h.dep.lr,
		msg: msg,
	}
	if err := f.Procedure(ctx, &procInitializeExecutionState, false); err != nil {
		err := errors.Wrap(err, "[ HandleExecutorResults ] Failed to initialize execution state")
		rep, newErr := payload.NewMessage(&payload.Error{Text: err.Error()})
		if newErr != nil {
			return newErr
		}
		go h.dep.Sender.Reply(ctx, h.Message, rep)
		return err
	}
	replyOk = bus.ReplyAsMessage(ctx, &reply.OK{})
	go h.dep.Sender.Reply(ctx, h.Message, replyOk)
	return nil
}
