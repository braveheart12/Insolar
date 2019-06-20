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

package record

import (
	"fmt"
)

// ToVirtual converts Record struct to protobuf friendly Virtual record.
func ToVirtual(record Record) Virtual {
	switch generic := record.(type) {
	case *Genesis:
		return Virtual{
			Union: &Virtual_Genesis{
				Genesis: generic,
			},
		}
	case *Child:
		return Virtual{
			Union: &Virtual_Child{
				Child: generic,
			},
		}
	case *Jet:
		return Virtual{
			Union: &Virtual_Jet{
				Jet: generic,
			},
		}
	case *Request:
		return Virtual{
			Union: &Virtual_Request{
				Request: generic,
			},
		}
	case *Result:
		return Virtual{
			Union: &Virtual_Result{
				Result: generic,
			},
		}
	case *Type:
		return Virtual{
			Union: &Virtual_Type{
				Type: generic,
			},
		}
	case *Code:
		return Virtual{
			Union: &Virtual_Code{
				Code: generic,
			},
		}
	case *Activate:
		return Virtual{
			Union: &Virtual_Activate{
				Activate: generic,
			},
		}
	case *Amend:
		return Virtual{
			Union: &Virtual_Amend{
				Amend: generic,
			},
		}
	case *Deactivate:
		return Virtual{
			Union: &Virtual_Deactivate{
				Deactivate: generic,
			},
		}
	case *PendingFilament:
		return Virtual{
			Union: &Virtual_PendingFilament{
				PendingFilament: generic,
			},
		}
	default:
		panic(fmt.Sprintf("%T record is not registered", generic))
	}
}

// FromVirtual converts protobuf friendly Virtual record to Record struct.
func FromVirtual(v Virtual) Record {
	switch r := v.Union.(type) {
	case *Virtual_Genesis:
		return r.Genesis
	case *Virtual_Child:
		return r.Child
	case *Virtual_Jet:
		return r.Jet
	case *Virtual_Request:
		return r.Request
	case *Virtual_Result:
		return r.Result
	case *Virtual_Type:
		return r.Type
	case *Virtual_Code:
		return r.Code
	case *Virtual_Activate:
		return r.Activate
	case *Virtual_Amend:
		return r.Amend
	case *Virtual_Deactivate:
		return r.Deactivate
	case *Virtual_PendingFilament:
		return r.PendingFilament
	default:
		panic(fmt.Sprintf("%T virtual record unknown type", r))
	}
}
