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

package proc

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/insolar/insolar/insolar/bus"
	"github.com/insolar/insolar/insolar/payload"
	"github.com/insolar/insolar/insolar/record"
	"github.com/insolar/insolar/ledger/blob"
	"github.com/insolar/insolar/ledger/object"
)

type GetCode struct {
	message payload.Meta

	Dep struct {
		RecordAccessor object.RecordAccessor
		BlobAccessor   blob.Accessor
		Sender         bus.Sender
	}
}

func NewGetCode(msg payload.Meta) *GetCode {
	return &GetCode{
		message: msg,
	}
}

func (p *GetCode) Proceed(ctx context.Context) error {
	getCode := payload.GetCode{}
	err := getCode.Unmarshal(p.message.Payload)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal GetCode message")
	}

	item, err := p.Dep.RecordAccessor.ForID(ctx, getCode.CodeID)
	if err != nil {
		return errors.Wrap(err, "failed to fetch record")
	}

	rec := item.Virtual
	code, ok := rec.(*record.Code)
	if !ok {
		return fmt.Errorf("expect code record, but got type %T", rec)
	}

	virtual := record.ToVirtual(rec)
	buf, err := virtual.Marshal()
	if err != nil {
		return errors.Wrap(err, "failed to marshal record")
	}

	msg, err := payload.NewMessage(&payload.Code{
		Record: buf,
		Code:   code.Code,
	})
	if err != nil {
		return errors.Wrap(err, "failed to create message")
	}

	go p.Dep.Sender.Reply(ctx, p.message, msg)

	return nil
}
