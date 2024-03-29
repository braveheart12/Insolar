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

// Package reply represents responses to messages of the messagebus
package reply

import (
	"bytes"
	"encoding/gob"
	"io"

	"github.com/insolar/insolar/insolar"
	"github.com/pkg/errors"
)

const (
	// Generic

	// TypeError is reply with error.
	TypeError = insolar.ReplyType(iota + 1)
	// TypeOK is a generic reply for signaling a positive result.
	TypeOK
	// TypeNotOK is a generic reply for signaling a negative result.
	TypeNotOK

	// TypeGetCodeRedirect is a redirect reply for code-call
	TypeGetCodeRedirect
	// TypeGetObjectRedirect is a redirect reply for object-call
	TypeGetObjectRedirect
	// TypeGetChildrenRedirect is a redirect reply for children-call
	TypeGetChildrenRedirect

	// Logicrunner

	// TypeCallMethod - two binary fields: data and results.
	TypeCallMethod
	// TypeCallConstructor - reference on created object
	TypeCallConstructor
	// TypeRegisterRequest - request for execution was registered
	TypeRegisterRequest

	// Ledger

	// TypeCode is code from storage.
	TypeCode
	// TypeObject is object from storage.
	TypeObject
	// TypeDelegate is delegate reference from storage.
	TypeDelegate
	// TypeID is common reply for methods returning id to lifeline states.
	TypeID
	// TypeChildren is a reply for fetching objects children in chunks.
	TypeChildren
	// TypeObjectIndex contains serialized object index. It can be stored in DB without processing.
	TypeObjectIndex
	// TypeJetMiss is returned for miscalculated jets due to incomplete jet tree.
	TypeJetMiss
	// TypePendingRequests contains unclosed requests for an object.
	TypePendingRequests
	// TypeJet contains jet.
	TypeJet
	// TypeOpenRequestsOnHeavy returns open requests from a heavy
	TypeOpenRequestsOnHeavy
	// TypeHeavyError carries heavy record sync
	TypeHeavyError
	// TypeIDs is common reply for methods returning list of IDs.
	TypeIDs
)

// ErrType is used to determine and compare reply errors.
type ErrType int

const (
	// ErrDeactivated returned when requested object is deactivated.
	ErrDeactivated = iota + 1
	// ErrStateNotAvailable is returned when requested object is deactivated.
	ErrStateNotAvailable
	// ErrHotDataTimeout is returned when no hot data received for a specific jet
	ErrHotDataTimeout
	// ErrNoPendingRequests is returned when there are no pending requests on current LME
	ErrNoPendingRequests
	// ErrTooManyPendingRequests is returned when a limit of pending requests has been reached
	ErrTooManyPendingRequests
	FlowCancelled
)

func getEmptyReply(t insolar.ReplyType) (insolar.Reply, error) {
	switch t {
	case TypeCallMethod:
		return &CallMethod{}, nil
	case TypeCallConstructor:
		return &CallConstructor{}, nil
	case TypeRegisterRequest:
		return &RegisterRequest{}, nil
	case TypeCode:
		return &Code{}, nil
	case TypeObject:
		return &Object{}, nil
	case TypeDelegate:
		return &Delegate{}, nil
	case TypeID:
		return &ID{}, nil
	case TypeIDs:
		return &IDs{}, nil
	case TypeChildren:
		return &Children{}, nil
	case TypeError:
		return &Error{}, nil
	case TypeHeavyError:
		return &HeavyError{}, nil
	case TypeOK:
		return &OK{}, nil
	case TypeObjectIndex:
		return &ObjectIndex{}, nil
	case TypeGetCodeRedirect:
		return &GetCodeRedirectReply{}, nil
	case TypeGetChildrenRedirect:
		return &GetChildrenRedirectReply{}, nil
	case TypeJetMiss:
		return &JetMiss{}, nil
	case TypePendingRequests:
		return &HasPendingRequests{}, nil
	case TypeJet:
		return &Jet{}, nil

	default:
		return nil, errors.Errorf("unimplemented reply type: '%d'", t)
	}
}

// Serialize returns encoded reply.
func Serialize(reply insolar.Reply) (io.Reader, error) {
	buff := &bytes.Buffer{}
	_, err := buff.Write([]byte{byte(reply.Type())})
	if err != nil {
		return nil, err
	}

	enc := gob.NewEncoder(buff)
	err = enc.Encode(reply)
	return buff, err
}

// Deserialize returns decoded reply.
func Deserialize(buff io.Reader) (insolar.Reply, error) {
	b := make([]byte, 1)
	_, err := buff.Read(b)
	if err != nil {
		return nil, errors.New("too short input to deserialize a message reply")
	}

	reply, err := getEmptyReply(insolar.ReplyType(b[0]))
	if err != nil {
		return nil, err
	}
	enc := gob.NewDecoder(buff)
	err = enc.Decode(reply)
	return reply, err
}

// ToBytes deserializes reply to bytes.
func ToBytes(rep insolar.Reply) []byte {
	repBuff, err := Serialize(rep)
	if err != nil {
		panic("failed to serialize reply")
	}
	return repBuff.(*bytes.Buffer).Bytes()
}

func init() {
	gob.Register(&CallMethod{})
	gob.Register(&CallConstructor{})
	gob.Register(&RegisterRequest{})
	gob.Register(&Code{})
	gob.Register(&Object{})
	gob.Register(&Delegate{})
	gob.Register(&ID{})
	gob.Register(&IDs{})
	gob.Register(&Children{})
	gob.Register(&Error{})
	gob.Register(&OK{})
	gob.Register(&ObjectIndex{})
	gob.Register(&GetCodeRedirectReply{})
	gob.Register(&GetChildrenRedirectReply{})
	gob.Register(&HeavyError{})
	gob.Register(&JetMiss{})
	gob.Register(&HasPendingRequests{})
}
