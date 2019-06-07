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
	"encoding/gob"

	"github.com/pkg/errors"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/message"
	"github.com/insolar/insolar/insolar/reply"
	"github.com/insolar/insolar/instrumentation/inslogger"
)

type CaseRequest struct {
	Parcel     insolar.Parcel
	Request    insolar.Reference
	MessageBus insolar.MessageBus
	Reply      insolar.Reply
	Error      string
}

// CaseBinder is a whole result of executor efforts on every object it seen on this pulse
type CaseBind struct {
	Requests []CaseRequest
}

func NewCaseBind() *CaseBind {
	return &CaseBind{Requests: make([]CaseRequest, 0)}
}

func NewCaseBindFromValidateMessage(ctx context.Context, mb insolar.MessageBus, msg *message.ValidateCaseBind) *CaseBind {
	res := &CaseBind{
		Requests: make([]CaseRequest, len(msg.Requests)),
	}
	for i, req := range msg.Requests {
		// TODO: here we used message bus player
		res.Requests[i] = CaseRequest{
			Parcel:  req.Parcel,
			Request: req.Request,
			Reply:   req.Reply,
			Error:   req.Error,
		}
	}
	return res
}

func NewCaseBindFromExecutorResultsMessage(msg *message.ExecutorResults) *CaseBind {
	panic("not implemented")
}

func (cb *CaseBind) getCaseBindForMessage(_ context.Context) []message.CaseBindRequest {
	return make([]message.CaseBindRequest, 0)
	// TODO: we don't validate at the moment, just send empty case bind
	//
	//if cb == nil {
	//	return make([]message.CaseBindRequest, 0)
	//}
	//
	//requests := make([]message.CaseBindRequest, len(cb.Requests))
	//
	//for i, req := range cb.Requests {
	//	var buf bytes.Buffer
	//	err := req.MessageBus.(insolar.TapeWriter).WriteTape(ctx, &buf)
	//	if err != nil {
	//		panic("couldn't write tape: " + err.Error())
	//	}
	//	requests[i] = message.CaseBindRequest{
	//		Parcel:         req.Parcel,
	//		Request:        req.Request,
	//		MessageBusTape: buf.Bytes(),
	//		Reply:          req.Reply,
	//		Error:          req.Error,
	//	}
	//}
	//
	//return requests
}

func (cb *CaseBind) ToValidateMessage(ctx context.Context, ref Ref, pulse insolar.Pulse) *message.ValidateCaseBind {
	res := &message.ValidateCaseBind{
		RecordRef: ref,
		Requests:  cb.getCaseBindForMessage(ctx),
		Pulse:     pulse,
	}
	return res
}

func (cb *CaseBind) NewRequest(p insolar.Parcel, request Ref, mb insolar.MessageBus) *CaseRequest {
	res := CaseRequest{
		Parcel:     p,
		Request:    request,
		MessageBus: mb,
	}
	cb.Requests = append(cb.Requests, res)
	return &cb.Requests[len(cb.Requests)-1]
}

type CaseBindReplay struct {
	Pulse    insolar.Pulse
	CaseBind CaseBind
	Request  int
	Record   int
	Steps    int
	Fail     int
}

func NewCaseBindReplay(cb CaseBind) *CaseBindReplay {
	return &CaseBindReplay{
		CaseBind: cb,
		Request:  -1,
		Record:   -1,
	}
}

func (r *CaseBindReplay) NextRequest() *CaseRequest {
	if r.Request+1 >= len(r.CaseBind.Requests) {
		return nil
	}
	r.Request++
	return &r.CaseBind.Requests[r.Request]
}

func (lr *LogicRunner) Validate(ctx context.Context, ref Ref, p insolar.Pulse, cb CaseBind) (int, error) {
	//os := lr.UpsertObjectState(ref)
	//vs := os.StartValidation(ref)
	//
	//vs.Lock()
	//defer vs.Unlock()
	//
	//checker := &ValidationChecker{
	//	lr: lr,
	//	cb: NewCaseBindReplay(cb),
	//}
	//vs.Behaviour = checker
	//
	//for {
	//	request := checker.NextRequest()
	//	if request == nil {
	//		break
	//	}
	//
	//	traceID := "TODO" // FIXME
	//
	//	ctx = inslogger.ContextWithTrace(ctx, traceID)
	//
	//	// TODO: here we were injecting message bus into context
	//
	//	sender := request.Parcel.GetSender()
	//	vs.Current = &CurrentExecution{
	//		Context:       ctx,
	//		Request:       &request.Request,
	//		RequesterNode: &sender,
	//	}
	//
	//	rep, err := func() (insolar.Reply, error) {
	//		vs.Unlock()
	//		defer vs.Lock()
	//		return lr.executeOrValidate(ctx, vs, request.Parcel)
	//	}()
	//
	//	err = vs.Behaviour.Result(rep, err)
	//	if err != nil {
	//		return 0, errors.Wrap(err, "validation step failed")
	//	}
	//}
	return 1, nil
}

func (lr *LogicRunner) HandleValidateCaseBindMessage(ctx context.Context, inmsg insolar.Parcel) (insolar.Reply, error) {
	ctx = loggerWithTargetID(ctx, inmsg)
	inslogger.FromContext(ctx).Debug("LogicRunner.HandleValidateCaseBindMessage starts ...")
	msg, ok := inmsg.Message().(*message.ValidateCaseBind)
	if !ok {
		return nil, errors.New("Execute( ! message.ValidateCaseBindInterface )")
	}

	procCheckRole := CheckOurRole{
		msg:  msg,
		role: insolar.DynamicRoleVirtualValidator,
		lr:   lr,
	}
	if err := procCheckRole.Proceed(ctx); err != nil {
		return nil, errors.Wrap(err, "[ HandleValidateCaseBindMessage ] can't play role")
	}

	passedStepsCount, validationError := lr.Validate(
		ctx, msg.GetReference(), msg.GetPulse(), *NewCaseBindFromValidateMessage(ctx, lr.MessageBus, msg),
	)
	errstr := ""
	if validationError != nil {
		errstr = validationError.Error()
	}

	_, err := lr.MessageBus.Send(ctx, &message.ValidationResults{
		RecordRef:        msg.GetReference(),
		PassedStepsCount: passedStepsCount,
		Error:            errstr,
	}, nil)

	return &reply.OK{}, err
}

func (lr *LogicRunner) HandleValidationResultsMessage(ctx context.Context, inmsg insolar.Parcel) (insolar.Reply, error) {
	ctx = loggerWithTargetID(ctx, inmsg)
	inslogger.FromContext(ctx).Debug("LogicRunner.HandleValidationResultsMessage starts ...")
	msg, ok := inmsg.Message().(*message.ValidationResults)
	if !ok {
		return nil, errors.Errorf("HandleValidationResultsMessage got argument typed %t", inmsg)
	}

	c := lr.GetConsensus(ctx, msg.RecordRef)
	if err := c.AddValidated(ctx, inmsg, msg); err != nil {
		return nil, err
	}
	return &reply.OK{}, nil
}

func (lr *LogicRunner) HandleExecutorResultsMessage(ctx context.Context, inmsg insolar.Parcel) (insolar.Reply, error) {
	return lr.FlowDispatcher.WrapBusHandle(ctx, inmsg)
}

func init() {
	gob.Register(&CaseRequest{})
	gob.Register(&CaseBind{})
}
