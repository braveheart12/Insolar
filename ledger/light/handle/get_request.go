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

	"github.com/insolar/insolar/insolar/flow"
	"github.com/insolar/insolar/insolar/payload"
	"github.com/insolar/insolar/ledger/light/proc"
	"github.com/pkg/errors"
)

type GetRequest struct {
	dep *proc.Dependencies

	message payload.Meta
}

func NewGetRequest(dep *proc.Dependencies, msg payload.Meta) *GetRequest {
	return &GetRequest{
		dep:     dep,
		message: msg,
	}
}

func (s *GetRequest) Present(ctx context.Context, f flow.Flow) error {
	msg := payload.GetRequest{}
	err := msg.Unmarshal(s.message.Payload)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal GetRequest message")
	}

	req := proc.NewGetRequest(s.message, msg.RequestID)
	s.dep.GetRequest(req)
	return f.Procedure(ctx, req, false)
}
