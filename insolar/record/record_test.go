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
	"testing"

	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/gen"
)

func FuzzRandomID(t *insolar.ID, _ fuzz.Continue) {
	*t = gen.ID()
}

func FuzzRandomReference(t *insolar.Reference, _ fuzz.Continue) {
	*t = gen.Reference()
}

func fuzzer() *fuzz.Fuzzer {
	return fuzz.New().Funcs(FuzzRandomID, FuzzRandomReference).NumElements(50, 100).NilChance(0)
}

func TestMarshalUnmarshalRecord(t *testing.T) {

	t.Run("GenesisRecordTest", func(t *testing.T) {
		f := fuzzer()
		a := assert.New(t)
		t.Parallel()
		var record Genesis

		for i := 0; i < 1; i++ {
			f.Fuzz(&record)

			bin, err := record.Marshal()
			a.NoError(err)
			for i := 0; i < 2; i++ {
				binNew, err := record.Marshal()
				a.NoError(err)
				a.Equal(bin, binNew)

				var recordNew Genesis
				err = recordNew.Unmarshal(binNew)
				require.NoError(t, err)

				a.Equal(&record, &recordNew)
			}
		}
	})

	t.Run("ChildRecordTest", func(t *testing.T) {
		f := fuzzer()
		a := assert.New(t)
		t.Parallel()
		var record Child

		for i := 0; i < 10; i++ {
			f.Fuzz(&record)

			bin, err := record.Marshal()
			a.NoError(err)
			for i := 0; i < 2; i++ {
				binNew, err := record.Marshal()
				a.NoError(err)
				a.Equal(bin, binNew)

				var recordNew Child
				err = recordNew.Unmarshal(binNew)
				require.NoError(t, err)

				a.Equal(&record, &recordNew)
			}
		}
	})

	t.Run("JetRecordTest", func(t *testing.T) {
		f := fuzzer()
		a := assert.New(t)
		t.Parallel()
		var record Jet

		for i := 0; i < 10; i++ {
			f.Fuzz(&record)

			bin, err := record.Marshal()
			a.NoError(err)
			for i := 0; i < 2; i++ {
				binNew, err := record.Marshal()
				a.NoError(err)
				a.Equal(bin, binNew)

				var recordNew Jet
				err = recordNew.Unmarshal(binNew)
				require.NoError(t, err)

				a.Equal(&record, &recordNew)
			}
		}
	})

	t.Run("RequestRecordTest", func(t *testing.T) {
		f := fuzzer()
		a := assert.New(t)
		t.Parallel()
		var record IncomingRequest

		for i := 0; i < 10; i++ {
			f.Fuzz(&record)

			bin, err := record.Marshal()
			a.NoError(err)
			for i := 0; i < 2; i++ {
				binNew, err := record.Marshal()
				a.NoError(err)
				a.Equal(bin, binNew)

				var recordNew IncomingRequest
				err = recordNew.Unmarshal(binNew)
				require.NoError(t, err)

				a.Equal(&record, &recordNew)
			}
		}
	})

	t.Run("ResultRecordTest", func(t *testing.T) {
		f := fuzzer()
		a := assert.New(t)
		t.Parallel()
		var record Result

		for i := 0; i < 10; i++ {
			f.Fuzz(&record)

			bin, err := record.Marshal()
			a.NoError(err)
			for i := 0; i < 2; i++ {
				binNew, err := record.Marshal()
				a.NoError(err)
				a.Equal(bin, binNew)

				var recordNew Result
				err = recordNew.Unmarshal(binNew)
				require.NoError(t, err)

				a.Equal(&record, &recordNew)
			}
		}
	})

	t.Run("TypeRecordTest", func(t *testing.T) {
		f := fuzzer()
		a := assert.New(t)
		t.Parallel()
		var record Type

		for i := 0; i < 10; i++ {
			f.Fuzz(&record)

			bin, err := record.Marshal()
			a.NoError(err)
			for i := 0; i < 2; i++ {
				binNew, err := record.Marshal()
				a.NoError(err)
				a.Equal(bin, binNew)

				var recordNew Type
				err = recordNew.Unmarshal(binNew)
				require.NoError(t, err)

				a.Equal(&record, &recordNew)
			}
		}
	})

	t.Run("CodeRecordTest", func(t *testing.T) {
		f := fuzzer()
		a := assert.New(t)
		t.Parallel()
		var record Code

		for i := 0; i < 10; i++ {
			f.Fuzz(&record)

			bin, err := record.Marshal()
			a.NoError(err)
			for i := 0; i < 2; i++ {
				binNew, err := record.Marshal()
				a.NoError(err)
				a.Equal(bin, binNew)

				var recordNew Code
				err = recordNew.Unmarshal(binNew)
				require.NoError(t, err)

				a.Equal(&record, &recordNew)
			}
		}
	})

	t.Run("ActivateRecordTest", func(t *testing.T) {
		f := fuzzer()
		a := assert.New(t)
		t.Parallel()
		var record Activate

		for i := 0; i < 10; i++ {
			f.Fuzz(&record)

			bin, err := record.Marshal()
			a.NoError(err)
			for i := 0; i < 2; i++ {
				binNew, err := record.Marshal()
				a.NoError(err)
				a.Equal(bin, binNew)

				var recordNew Activate
				err = recordNew.Unmarshal(binNew)
				require.NoError(t, err)

				a.Equal(&record, &recordNew)
			}
		}
	})

	t.Run("AmendRecordTest", func(t *testing.T) {
		f := fuzzer()
		a := assert.New(t)
		t.Parallel()
		var record Amend

		for i := 0; i < 10; i++ {
			f.Fuzz(&record)

			bin, err := record.Marshal()
			a.NoError(err)
			for i := 0; i < 2; i++ {
				binNew, err := record.Marshal()
				a.NoError(err)
				a.Equal(bin, binNew)

				var recordNew Amend
				err = recordNew.Unmarshal(binNew)
				require.NoError(t, err)

				a.Equal(&record, &recordNew)
			}
		}
	})

	t.Run("DeactivateRecordTest", func(t *testing.T) {
		f := fuzzer()
		a := assert.New(t)
		t.Parallel()
		var record Deactivate

		for i := 0; i < 10; i++ {
			f.Fuzz(&record)

			bin, err := record.Marshal()
			a.NoError(err)
			for i := 0; i < 2; i++ {
				binNew, err := record.Marshal()
				a.NoError(err)
				a.Equal(bin, binNew)

				var recordNew Deactivate
				err = recordNew.Unmarshal(binNew)
				require.NoError(t, err)

				a.Equal(&record, &recordNew)
			}
		}
	})
}

func TestRequestInterface_IncomingRequest(t *testing.T) {
	t.Parallel()
	objref := gen.Reference()
	req := &IncomingRequest{
		Object: &objref,
		Reason: gen.Reference(),
	}
	iface := Request(req)
	require.Equal(t, req.Object, iface.AffinityRef())
	require.Equal(t, req.Reason, iface.ReasonRef())
}

func TestRequestInterface_OutgoingRequest(t *testing.T) {
	t.Parallel()
	objref := gen.Reference()
	callerref := gen.Reference()
	req := &OutgoingRequest{
		Caller: callerref,
		Object: &objref,
		Reason: gen.Reference(),
	}
	iface := Request(req)
	require.Equal(t, &req.Caller, iface.AffinityRef())
	require.Equal(t, req.Reason, iface.ReasonRef())
}
