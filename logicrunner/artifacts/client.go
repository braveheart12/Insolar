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

package artifacts

import (
	"context"
	"fmt"
	"time"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/bus"
	"github.com/insolar/insolar/insolar/jet"
	"github.com/insolar/insolar/insolar/message"
	"github.com/insolar/insolar/insolar/payload"
	"github.com/insolar/insolar/insolar/pulse"
	"github.com/insolar/insolar/insolar/record"
	"github.com/insolar/insolar/insolar/reply"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/instrumentation/instracer"
	"github.com/insolar/insolar/messagebus"

	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

const (
	getChildrenChunkSize = 10 * 1000
)

type localStorage struct {
	initialized bool
	storage     map[insolar.Reference]interface{}
}

func (s *localStorage) Initialized() {
	s.initialized = true
}

func (s *localStorage) StoreObject(reference insolar.Reference, descriptor interface{}) {
	if s.initialized {
		panic("Trying to initialize singleFlightCache after initialization was finished")
	}
	s.storage[reference] = descriptor
}

func (s *localStorage) Code(reference insolar.Reference) CodeDescriptor {
	codeDescI, ok := s.storage[reference]
	if !ok {
		return nil
	}
	codeDesc, ok := codeDescI.(CodeDescriptor)
	if !ok {
		return nil
	}
	return codeDesc
}

func (s *localStorage) Object(reference insolar.Reference) ObjectDescriptor {
	objectDescI, ok := s.storage[reference]
	if !ok {
		return nil
	}
	objectDesc, ok := objectDescI.(ObjectDescriptor)
	if !ok {
		return nil
	}
	return objectDesc
}

func newLocalStorage() *localStorage {
	return &localStorage{
		initialized: false,
		storage:     make(map[insolar.Reference]interface{}),
	}
}

// Client provides concrete API to storage for processing module.
type client struct {
	JetStorage     jet.Storage                        `inject:""`
	DefaultBus     insolar.MessageBus                 `inject:""`
	PCS            insolar.PlatformCryptographyScheme `inject:""`
	PulseAccessor  pulse.Accessor                     `inject:""`
	JetCoordinator jet.Coordinator                    `inject:""`

	sender               bus.Sender
	getChildrenChunkSize int
	senders              *messagebus.Senders
	localStorage         *localStorage
}

// State returns hash state for artifact manager.
func (m *client) State() []byte {
	// This is a temporary stab to simulate real hash.
	return m.PCS.IntegrityHasher().Hash([]byte{1, 2, 3})
}

// NewClient creates new client instance.
func NewClient(sender bus.Sender) *client { // nolint
	return &client{
		getChildrenChunkSize: getChildrenChunkSize,
		senders:              messagebus.NewSenders(),
		sender:               sender,
		localStorage:         newLocalStorage(),
	}
}

// registerRequest registers incoming or outgoing request.
func (m *client) registerRequest(
	ctx context.Context, req record.Record, msgPayload payload.Payload, callType record.CallType,
	objectRef *insolar.Reference, retriesNumber int,
) (*insolar.ID, error) {
	var err error
	ctx, span := instracer.StartSpan(ctx, "artifactmanager.registerRequest")
	instrumenter := instrument(ctx, "registerRequest").err(&err)
	defer func() {
		if err != nil {
			span.AddAttributes(trace.StringAttribute("error", err.Error()))
		}
		span.End()
		instrumenter.end()
	}()

	virtRec := record.Wrap(req)
	buf, err := virtRec.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "registerRequest: failed to marshal record")
	}

	h := m.PCS.ReferenceHasher()
	_, err = h.Write(buf)
	if err != nil {
		return nil, errors.Wrap(err, "registerRequest: failed to calculate hash")
	}

	var recRef *insolar.Reference
	switch callType {
	case record.CTMethod:
		recRef = objectRef
	case record.CTSaveAsChild, record.CTSaveAsDelegate, record.CTGenesis:
		recRef, err = m.genReferenceForCallTypeOtherThanCTMethod(ctx)
	default:
		err = errors.New("registerRequest: not supported call type " + callType.String())
	}
	if err != nil {
		return nil, err
	}

	pl, err := m.sendWithRetry(ctx, msgPayload, insolar.DynamicRoleLightExecutor, *recRef, retriesNumber)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't register request")
	}

	switch p := pl.(type) {
	case *payload.RequestInfo:
		return &p.RequestID, nil
	case *payload.Error:
		return nil, errors.New(p.Text)
	default:
		return nil, fmt.Errorf("registerRequest: unexpected reply: %#v", p)
	}
}

func (m *client) genReferenceForCallTypeOtherThanCTMethod(ctx context.Context) (*insolar.Reference, error) {
	h := m.PCS.ReferenceHasher()
	currentPN, err := m.pulse(ctx)
	if err != nil {
		return nil, err
	}
	recID := *insolar.NewID(currentPN, h.Sum(nil))
	return insolar.NewReference(recID), nil
}

// RegisterIncomingRequest sends message for incoming request registration,
// returns request record Ref if request successfully created or already exists.
func (m *client) RegisterIncomingRequest(
	ctx context.Context, request *record.IncomingRequest,
) (*insolar.ID, error) {
	incomingRequest := &payload.SetIncomingRequest{Request: record.Wrap(request)}

	// retriesNumber is zero, because we don't retry registering of incoming requests - the caller should
	// re-send the request instead.
	id, err := m.registerRequest(ctx, request, incomingRequest, request.CallType, request.AffinityRef(), 0)
	if err != nil {
		return id, errors.Wrap(err, "RegisterIncomingRequest")
	}
	return id, err
}

// RegisterOutgoingRequest sends message for outgoing request registration,
// returns request record Ref if request successfully created or already exists.
func (m *client) RegisterOutgoingRequest(ctx context.Context, request *record.OutgoingRequest) (*insolar.ID, error) {
	outgoingRequest := &payload.SetOutgoingRequest{Request: record.Wrap(request)}
	id, err := m.registerRequest(ctx, request, outgoingRequest, request.CallType, request.AffinityRef(), 3)
	if err != nil {
		return id, errors.Wrap(err, "RegisterOutgoingRequest")
	}
	return id, err
}

// GetCode returns code from code record by provided reference according to provided machine preference.
//
// This method is used by VM to fetch code for execution.
func (m *client) GetCode(
	ctx context.Context, code insolar.Reference,
) (CodeDescriptor, error) {
	var (
		desc CodeDescriptor
		err  error
	)

	desc = m.localStorage.Code(code)
	if desc != nil {
		return desc, nil
	}

	instrumenter := instrument(ctx, "GetCode").err(&err)
	ctx, span := instracer.StartSpan(ctx, "artifactmanager.GetCode")
	defer func() {
		if err != nil {
			span.AddAttributes(trace.StringAttribute("error", err.Error()))
		}
		span.End()
		instrumenter.end()
	}()

	msg, err := payload.NewMessage(&payload.GetCode{
		CodeID: *code.Record(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal message")
	}

	r := bus.NewRetrySender(m.sender, 3)
	reps, done := r.SendRole(ctx, msg, insolar.DynamicRoleLightExecutor, code)
	defer done()

	rep, ok := <-reps
	if !ok {
		return nil, ErrNoReply
	}

	pl, err := payload.UnmarshalFromMeta(rep.Payload)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal reply")
	}

	switch p := pl.(type) {
	case *payload.Code:
		rec := record.Material{}
		err := rec.Unmarshal(p.Record)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal record")
		}
		virtual := record.Unwrap(rec.Virtual)
		codeRecord, ok := virtual.(*record.Code)
		if !ok {
			return nil, errors.Wrapf(err, "unexpected record %T", virtual)
		}
		desc = &codeDescriptor{
			ref:         code,
			machineType: codeRecord.MachineType,
			code:        codeRecord.Code,
		}
		return desc, nil
	case *payload.Error:
		return nil, errors.New(p.Text)
	default:
		return nil, fmt.Errorf("GetObject: unexpected reply: %#v", p)
	}
}

// GetObject returns descriptor for provided state.
//
// If provided state is nil, the latest state will be returned (with deactivation check). Returned descriptor will
// provide methods for fetching all related data.
func (m *client) GetObject(
	ctx context.Context,
	head insolar.Reference,
) (ObjectDescriptor, error) {
	var (
		err error
	)

	if desc := m.localStorage.Object(head); desc != nil {
		return desc, nil
	}

	ctx, span := instracer.StartSpan(ctx, "artifactmanager.Getobject")
	instrumenter := instrument(ctx, "GetObject").err(&err)
	defer func() {
		if err != nil {
			span.AddAttributes(trace.StringAttribute("error", err.Error()))
		}
		span.End()
		if err != nil && err == ErrObjectDeactivated {
			err = nil // megahack: threat it 2xx
		}
		instrumenter.end()
	}()

	logger := inslogger.FromContext(ctx).WithField("object", head.Record().DebugString())

	msg, err := payload.NewMessage(&payload.GetObject{
		ObjectID: *head.Record(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal message")
	}

	r := bus.NewRetrySender(m.sender, 3)
	reps, done := r.SendRole(ctx, msg, insolar.DynamicRoleLightExecutor, head)
	defer done()

	var (
		index        *record.Lifeline
		statePayload *payload.State
	)
	success := func() bool {
		return index != nil && statePayload != nil
	}

	for rep := range reps {
		replyPayload, err := payload.UnmarshalFromMeta(rep.Payload)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal reply")
		}

		switch p := replyPayload.(type) {
		case *payload.Index:
			logger.Debug("reply index")
			index = &record.Lifeline{}
			err := index.Unmarshal(p.Index)
			if err != nil {
				return nil, errors.Wrap(err, "failed to unmarshal index")
			}
		case *payload.State:
			logger.Debug("reply state")
			statePayload = p
		case *payload.Error:
			logger.Debug("reply error: ", p.Text)
			switch p.Code {
			case payload.CodeDeactivated:
				return nil, insolar.ErrDeactivated
			default:
				return nil, errors.New(p.Text)
			}
		default:
			return nil, fmt.Errorf("GetObject: unexpected reply: %#v", p)
		}

		if success() {
			break
		}
	}
	if !success() {
		logger.Error(ErrNoReply)
		return nil, ErrNoReply
	}

	rec := record.Material{}
	err = rec.Unmarshal(statePayload.Record)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal state")
	}
	virtual := record.Unwrap(rec.Virtual)
	s, ok := virtual.(record.State)
	if !ok {
		return nil, errors.New("wrong state record")
	}
	state := s

	desc := &objectDescriptor{
		head:         head,
		state:        *index.LatestState,
		prototype:    state.GetImage(),
		isPrototype:  state.GetIsPrototype(),
		childPointer: index.ChildPointer,
		memory:       statePayload.Memory,
		parent:       index.Parent,
	}
	return desc, err
}

func (m *client) GetIncomingRequest(
	ctx context.Context, object, reqRef insolar.Reference,
) (*record.IncomingRequest, error) {
	var err error
	instrumenter := instrument(ctx, "GetRequest").err(&err)
	ctx, span := instracer.StartSpan(ctx, "artifactmanager.GetRequest")
	defer func() {
		if err != nil {
			span.AddAttributes(trace.StringAttribute("error", err.Error()))
		}
		span.End()
		instrumenter.end()
	}()

	requestID := reqRef.Record()
	node, err := m.JetCoordinator.NodeForObject(ctx, *object.Record(), requestID.Pulse())
	if err != nil {
		return nil, err
	}

	msg, err := payload.NewMessage(&payload.GetRequest{
		RequestID: *requestID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a message")
	}
	reps, done := m.sender.SendTarget(ctx, msg, *node)
	defer done()
	res, ok := <-reps
	if !ok {
		return nil, errors.New("no reply while fetching request")
	}

	pl, err := payload.UnmarshalFromMeta(res.Payload)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal reply")
	}
	req, ok := pl.(*payload.Request)
	if !ok {
		return nil, fmt.Errorf("unexpected reply %T", pl)
	}

	concrete := record.Unwrap(&req.Request)
	castedRecord, ok := concrete.(*record.IncomingRequest)
	if !ok {
		return nil, fmt.Errorf("GetPendingRequest: unexpected message: %#v", concrete)
	}

	return castedRecord, nil
}

// GetPendings returns a list of pending requests
func (m *client) GetPendings(ctx context.Context, object insolar.Reference) ([]insolar.Reference, error) {
	var err error
	instrumenter := instrument(ctx, "GetPendings").err(&err)
	ctx, span := instracer.StartSpan(ctx, "artifactmanager.GetPendings")
	defer func() {
		if err != nil {
			span.AddAttributes(trace.StringAttribute("error", err.Error()))
		}
		span.End()
		instrumenter.end()
	}()

	msg, err := payload.NewMessage(&payload.GetPendings{
		ObjectID: *object.Record(),
	})

	if err != nil {
		return nil, errors.Wrap(err, "GetPendings: failed to create a message")
	}

	reps, done := m.sender.SendRole(ctx, msg, insolar.DynamicRoleLightExecutor, object)

	defer done()
	res, ok := <-reps
	if !ok {
		return []insolar.Reference{}, errors.New("GetPendings: no reply")
	}

	pl, err := payload.UnmarshalFromMeta(res.Payload)
	if err != nil {
		return []insolar.Reference{}, errors.Wrap(err, "GetPendings: failed to unmarshal reply")
	}

	switch concrete := pl.(type) {
	case *payload.IDs:
		res := make([]insolar.Reference, len(concrete.IDs))
		for i := range concrete.IDs {
			res[i] = *insolar.NewReference(concrete.IDs[i])
		}
		return res, nil
	case *payload.Error:
		if concrete.Code == payload.CodeNoPendings {
			return []insolar.Reference{}, insolar.ErrNoPendingRequest
		}
		return []insolar.Reference{}, errors.New(concrete.Text)
	default:
		return []insolar.Reference{}, fmt.Errorf("unexpected reply %T", pl)
	}
}

// HasPendingRequests returns true if object has unclosed requests.
func (m *client) HasPendingRequests(
	ctx context.Context,
	object insolar.Reference,
) (bool, error) {
	ctx, span := instracer.StartSpan(ctx, "artifactmanager.HasPendingRequests")
	defer span.End()

	sender := messagebus.BuildSender(
		m.DefaultBus.Send,
		messagebus.RetryIncorrectPulse(m.PulseAccessor),
		messagebus.RetryJetSender(m.JetStorage),
	)

	genericReact, err := sender(ctx, &message.GetPendingRequests{Object: object}, nil)

	if err != nil {
		return false, err
	}

	switch rep := genericReact.(type) {
	case *reply.HasPendingRequests:
		return rep.Has, nil
	case *reply.Error:
		return false, rep.Error()
	default:
		return false, fmt.Errorf("HasPendingRequests: unexpected reply: %#v", rep)
	}
}

// GetDelegate returns provided object's delegate reference for provided prototype.
//
// Object delegate should be previously created for this object. If object delegate does not exist, an error will
// be returned.
func (m *client) GetDelegate(
	ctx context.Context, head, asType insolar.Reference,
) (*insolar.Reference, error) {
	var err error
	ctx, span := instracer.StartSpan(ctx, "artifactmanager.GetDelegate")
	instrumenter := instrument(ctx, "GetDelegate").err(&err)
	defer func() {
		if err != nil {
			span.AddAttributes(trace.StringAttribute("error", err.Error()))
		}
		span.End()
		instrumenter.end()
	}()

	sender := messagebus.BuildSender(
		m.DefaultBus.Send,
		messagebus.RetryIncorrectPulse(m.PulseAccessor),
		messagebus.FollowRedirectSender(m.DefaultBus),
		messagebus.RetryJetSender(m.JetStorage),
	)
	genericReact, err := sender(ctx, &message.GetDelegate{
		Head:   head,
		AsType: asType,
	}, nil)
	if err != nil {
		return nil, err
	}

	switch rep := genericReact.(type) {
	case *reply.Delegate:
		return &rep.Head, nil
	case *reply.Error:
		return nil, rep.Error()
	default:
		return nil, fmt.Errorf("GetDelegate: unexpected reply: %#v", rep)
	}
}

// GetChildren returns children iterator.
//
// During iteration children refs will be fetched from remote source (parent object).
func (m *client) GetChildren(
	ctx context.Context, parent insolar.Reference, pulse *insolar.PulseNumber,
) (RefIterator, error) {
	var err error

	ctx, span := instracer.StartSpan(ctx, "artifactmanager.GetChildren")
	instrumenter := instrument(ctx, "GetChildren").err(&err)
	defer func() {
		if err != nil {
			span.AddAttributes(trace.StringAttribute("error", err.Error()))
		}
		span.End()
		instrumenter.end()
	}()

	sender := messagebus.BuildSender(
		m.DefaultBus.Send,
		messagebus.RetryIncorrectPulse(m.PulseAccessor),
		messagebus.FollowRedirectSender(m.DefaultBus),
		messagebus.RetryJetSender(m.JetStorage),
	)
	iter, err := NewChildIterator(ctx, sender, parent, pulse, m.getChildrenChunkSize)
	return iter, err
}

// DeployCode creates new code record in storage.
//
// CodeRef records are used to activate prototype or as migration code for an object.
func (m *client) DeployCode(
	ctx context.Context,
	domain insolar.Reference,
	request insolar.Reference,
	code []byte,
	machineType insolar.MachineType,
) (*insolar.ID, error) {
	var err error
	ctx, span := instracer.StartSpan(ctx, "artifactmanager.DeployCode")
	instrumenter := instrument(ctx, "DeployCode").err(&err)
	defer func() {
		if err != nil {
			span.AddAttributes(trace.StringAttribute("error", err.Error()))
		}
		span.End()
		instrumenter.end()
	}()

	currentPN, err := m.pulse(ctx)
	if err != nil {
		return nil, err
	}

	codeRec := record.Code{
		Domain:      domain,
		Request:     request,
		Code:        code,
		MachineType: machineType,
	}
	virtual := record.Wrap(codeRec)
	buf, err := virtual.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal record")
	}

	h := m.PCS.ReferenceHasher()
	_, err = h.Write(buf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to calculate hash")
	}
	recID := *insolar.NewID(currentPN, h.Sum(nil))

	psc := &payload.SetCode{
		Record: buf,
	}

	pl, err := m.sendWithRetry(ctx, psc, insolar.DynamicRoleLightExecutor, *insolar.NewReference(recID), 3)
	if err != nil {
		return nil, err
	}
	switch p := pl.(type) {
	case *payload.ID:
		return &p.ID, nil
	case *payload.Error:
		return nil, errors.New(p.Text)
	default:
		return nil, fmt.Errorf("DeployCode: unexpected reply: %#v", p)
	}
}

// sendWithRetry sends given Payload to the specified DynamicRole with provided `retriesNumber`.
// If retriesNumber is zero the methods sends the message only once.
func (m *client) sendWithRetry(
	ctx context.Context, ppl payload.Payload, role insolar.DynamicRole, // nolint: unparam
	ref insolar.Reference, retriesNumber int) (payload.Payload, error) {
	var lastPulse insolar.PulseNumber

	for retriesNumber >= 0 {
		currentPulse, err := m.PulseAccessor.Latest(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "[ sendWithRetry ] Can't get latest pulse")
		}

		if currentPulse.PulseNumber == lastPulse {
			inslogger.FromContext(ctx).Debugf("[ sendWithRetry ]  wait for pulse change. Current: %d", currentPulse)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		lastPulse = currentPulse.PulseNumber

		msg, err := payload.NewMessage(ppl)

		if err != nil {
			return nil, err
		}

		reps, done := m.sender.SendRole(ctx, msg, role, ref)
		rep, ok := <-reps
		done()

		if !ok {
			return nil, ErrNoReply
		}
		pl, err := payload.UnmarshalFromMeta(rep.Payload)
		if err != nil {
			return nil, errors.Wrap(err, "[ sendWithRetry ] failed to unmarshal reply")
		}

		if p, ok := pl.(*payload.Error); !ok || p.Code != payload.CodeFlowCanceled {
			return pl, nil
		}
		inslogger.FromContext(ctx).Debug("[ sendWithRetry ] flow cancelled, retrying")
		retriesNumber--
	}
	return nil, fmt.Errorf("[ sendWithRetry ] flow cancelled, retries exceeded")
}

// ActivatePrototype creates activate object record in storage. Provided prototype reference will be used as objects prototype
// memory as memory of created object. If memory is not provided, the prototype default memory will be used.
//
// Request reference will be this object's identifier and referred as "object head".
func (m *client) ActivatePrototype(
	ctx context.Context,
	object, parent, code insolar.Reference,
	memory []byte,
) error {
	var err error
	ctx, span := instracer.StartSpan(ctx, "artifactmanager.ActivatePrototype")
	instrumenter := instrument(ctx, "ActivatePrototype").err(&err)
	defer func() {
		if err != nil {
			span.AddAttributes(trace.StringAttribute("error", err.Error()))
		}
		span.End()
		instrumenter.end()
	}()
	err = m.activateObject(ctx, object, code, true, parent, false, memory)
	return err
}

// RegisterValidation marks provided object state as approved or disapproved.
//
// When fetching object, validity can be specified.
func (m *client) RegisterValidation(
	ctx context.Context,
	object insolar.Reference,
	state insolar.ID,
	isValid bool,
	validationMessages []insolar.Message,
) error {
	var err error
	ctx, span := instracer.StartSpan(ctx, "artifactmanager.RegisterValidation")
	instrumenter := instrument(ctx, "RegisterValidation").err(&err)
	defer func() {
		if err != nil {
			span.AddAttributes(trace.StringAttribute("error", err.Error()))
		}
		span.End()
		instrumenter.end()
	}()

	msg := message.ValidateRecord{
		Object:             object,
		State:              state,
		IsValid:            isValid,
		ValidationMessages: validationMessages,
	}

	sender := messagebus.BuildSender(
		m.DefaultBus.Send,
		messagebus.RetryJetSender(m.JetStorage),
	)
	_, err = sender(ctx, &msg, nil)

	return err
}

// pulse returns current PulseNumber for artifact manager
func (m *client) pulse(ctx context.Context) (pn insolar.PulseNumber, err error) {
	pulse, err := m.PulseAccessor.Latest(ctx)
	if err != nil {
		return
	}

	pn = pulse.PulseNumber
	return
}

func (m *client) activateObject(
	ctx context.Context,
	obj insolar.Reference,
	prototype insolar.Reference,
	isPrototype bool,
	parent insolar.Reference,
	asDelegate bool,
	memory []byte,
) error {
	parentDesc, err := m.GetObject(ctx, parent)
	if err != nil {
		return err
	}

	activate := record.Activate{
		Request:     obj,
		Memory:      memory,
		Image:       prototype,
		IsPrototype: isPrototype,
		Parent:      parent,
		IsDelegate:  asDelegate,
	}

	result := record.Result{
		Object:  *obj.Record(),
		Request: obj,
	}

	virtActivate := record.Wrap(activate)
	virtResult := record.Wrap(result)

	activateBuf, err := virtActivate.Marshal()
	if err != nil {
		return errors.Wrap(err, "ActivateObject: can't serialize record")
	}
	resultBuf, err := virtResult.Marshal()
	if err != nil {
		return errors.Wrap(err, "ActivateObject: can't serialize record")
	}

	pa := &payload.Activate{
		Record: activateBuf,
		Result: resultBuf,
	}

	pl, err := m.sendWithRetry(ctx, pa, insolar.DynamicRoleLightExecutor, obj, 3)
	if err != nil {
		return errors.Wrap(err, "can't send activation and result records")
	}

	switch p := pl.(type) {
	case *payload.ResultInfo:
		var (
			asType *insolar.Reference
		)
		child := record.Child{Ref: obj}
		if parentDesc.ChildPointer() != nil {
			child.PrevChild = *parentDesc.ChildPointer()
		}
		if asDelegate {
			asType = &prototype
		}
		virtChild := record.Wrap(child)

		err = m.registerChild(
			ctx,
			virtChild,
			parent,
			obj,
			asType,
		)
		if err != nil {
			return errors.Wrap(err, "failed to register as child while activating")
		}
		return nil
	case *payload.Error:
		return errors.New(p.Text)
	default:
		return fmt.Errorf("ActivateObject: unexpected reply: %#v", p)
	}
}

func (m *client) registerChild(
	ctx context.Context,
	rec record.Virtual,
	parent insolar.Reference,
	child insolar.Reference,
	asType *insolar.Reference,
) error {
	data, err := rec.Marshal()
	if err != nil {
		return errors.Wrap(err, "setRecord: can't serialize record")
	}
	sender := messagebus.BuildSender(
		m.DefaultBus.Send,
		messagebus.RetryIncorrectPulse(m.PulseAccessor),
		messagebus.RetryJetSender(m.JetStorage),
		messagebus.RetryFlowCancelled(m.PulseAccessor),
	)
	genericReact, err := sender(ctx, &message.RegisterChild{
		Record: data,
		Parent: parent,
		Child:  child,
		AsType: asType,
	}, nil)

	if err != nil {
		return err
	}

	switch rep := genericReact.(type) {
	case *reply.ID:
		return nil
	case *reply.Error:
		return rep.Error()
	default:
		return fmt.Errorf("registerChild: unexpected reply: %#v", rep)
	}
}

func (m *client) InjectCodeDescriptor(reference insolar.Reference, descriptor CodeDescriptor) {
	m.localStorage.StoreObject(reference, descriptor)
}

func (m *client) InjectObjectDescriptor(reference insolar.Reference, descriptor ObjectDescriptor) {
	m.localStorage.StoreObject(reference, descriptor)
}

func (m *client) InjectFinish() {
	m.localStorage.Initialized()
}

// RegisterResult saves VM method call result with it's side effects
func (m *client) RegisterResult(
	ctx context.Context,
	request insolar.Reference,
	result RequestResult,
) error {

	sendResult := func(
		payloadInput payload.Payload,
		obj insolar.Reference,
	) (*insolar.ID, error) {
		payloadOutput, err := m.sendWithRetry(ctx, payloadInput, insolar.DynamicRoleLightExecutor, obj, 3)
		if err != nil {
			return nil, err
		}

		switch p := payloadOutput.(type) {
		case *payload.ResultInfo:
			return &payloadOutput.(*payload.ResultInfo).ResultID, nil
		case *payload.Error:
			return nil, errors.New(p.Text)
		default:
			return nil, fmt.Errorf("RegisterResult: unexpected reply: %#v", p)
		}
	}

	var (
		pl         payload.Payload
		parentDesc ObjectDescriptor
		err        error
	)

	ctx, span := instracer.StartSpan(ctx, "artifactmanager.RegisterResult")
	instrumenter := instrument(ctx, "RegisterResult").err(&err)
	defer func() {
		if err != nil {
			span.AddAttributes(trace.StringAttribute("error", err.Error()))
		}
		span.End()
		instrumenter.end()
	}()
	span.AddAttributes(trace.StringAttribute("SideEffect", result.Type().String()))

	objReference := result.ObjectReference()
	resultRecord := record.Result{
		Object:  *objReference.Record(),
		Request: request,
		Payload: result.Result(),
	}

	switch result.Type() {
	// ActivateObject creates activate object record in storage. Provided prototype reference will be used as objects prototype
	// memory as memory of created object. If memory is not provided, the prototype default memory will be used.
	//
	// Request reference will be this object's identifier and referred as "object head".
	case RequestSideEffectActivate:
		parentRef, imageRef, asDelegate, memory := result.Activate()

		parentDesc, err = m.GetObject(ctx, parentRef)
		if err != nil {
			return err
		}

		vResultRecord := record.Wrap(resultRecord)
		vActivateRecord := record.Wrap(record.Activate{
			Request:     request,
			Memory:      memory,
			Image:       imageRef,
			IsPrototype: false,
			Parent:      parentRef,
			IsDelegate:  asDelegate,
		})

		plTyped := payload.Activate{}
		plTyped.Record, err = vActivateRecord.Marshal()
		if err != nil {
			return errors.Wrap(err, "RegisterResult: can't serialize Activate record")
		}
		plTyped.Result, err = vResultRecord.Marshal()
		if err != nil {
			return errors.Wrap(err, "RegisterResult: can't serialize Result record")
		}
		pl = &plTyped

	// UpdateObject creates amend object record in storage. Provided reference should be a reference to the head of the
	// object. Provided memory well be the new object memory.
	//
	// Returned reference will be the latest object state (exact) reference.
	case RequestSideEffectAmend:
		objectStateID, objectImage, memory := result.Amend()

		vResultRecord := record.Wrap(resultRecord)
		vAmendRecord := record.Wrap(record.Amend{
			Request:     request,
			Memory:      memory,
			Image:       objectImage,
			IsPrototype: false,
			PrevState:   objectStateID,
		})

		plTyped := payload.Update{}
		plTyped.Record, err = vAmendRecord.Marshal()
		if err != nil {
			return errors.Wrap(err, "RegisterResult: can't serialize Amend record")
		}
		plTyped.Result, err = vResultRecord.Marshal()
		if err != nil {
			return errors.Wrap(err, "RegisterResult: can't serialize Result record")
		}
		pl = &plTyped

	// DeactivateObject creates deactivate object record in storage. Provided reference should be a reference to the head
	// of the object. If object is already deactivated, an error should be returned.
	//
	// Deactivated object cannot be changed.
	case RequestSideEffectDeactivate:
		objectStateID := result.Deactivate()

		vResultRecord := record.Wrap(resultRecord)
		vDeactivateRecord := record.Wrap(record.Deactivate{
			Request:   request,
			PrevState: objectStateID,
		})

		plTyped := payload.Deactivate{}
		plTyped.Record, err = vDeactivateRecord.Marshal()
		if err != nil {
			return errors.Wrap(err, "RegisterResult: can't serialize Deactivate record")
		}
		plTyped.Result, err = vResultRecord.Marshal()
		if err != nil {
			return errors.Wrap(err, "RegisterResult: can't serialize Result record")
		}
		pl = &plTyped

	case RequestSideEffectNone:
		vResultRecord := record.Wrap(resultRecord)

		plTyped := payload.SetResult{}
		plTyped.Result, err = vResultRecord.Marshal()
		if err != nil {
			return errors.Wrap(err, "RegisterResult: can't serialize Result record")
		}
		pl = &plTyped

	default:
		return errors.Errorf("RegisterResult: Unknown side effect %d", result.Type())
	}

	_, err = sendResult(pl, result.ObjectReference())
	if err != nil {
		return errors.Wrapf(err, "RegisterResult: Failed to send results: %s", result.Type().String())
	}

	if result.Type() == RequestSideEffectActivate {
		parentRef, imageRef, asDelegate, _ := result.Activate()

		var asType *insolar.Reference

		child := record.Child{
			Ref: result.ObjectReference(),
		}

		if parentDesc.ChildPointer() != nil {
			child.PrevChild = *parentDesc.ChildPointer()
		}
		if asDelegate {
			asType = &imageRef
		}

		err = m.registerChild(
			ctx,
			record.Wrap(child),
			parentRef,
			result.ObjectReference(),
			asType,
		)
		if err != nil {
			return errors.Wrap(err, "failed to register as child while activating")
		}
	}

	return nil
}
