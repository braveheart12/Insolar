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

package object

import (
	"crypto/sha256"
	"math/rand"
	"testing"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/gen"
	"github.com/insolar/insolar/insolar/record"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecordStorage_NewStorageMemory(t *testing.T) {
	t.Parallel()

	recordStorage := NewRecordMemory()
	assert.Equal(t, 0, len(recordStorage.recsStor))
}

func TestRecordStorage_ForID(t *testing.T) {
	t.Parallel()

	ctx := inslogger.TestContext(t)

	id := gen.ID()
	rec := getStoreRecord()

	t.Run("returns correct record-value", func(t *testing.T) {
		t.Parallel()

		recordStorage := NewRecordMemory()
		recordStorage.recsStor[id] = rec

		resultRec, err := recordStorage.ForID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, rec, resultRec)
	})

	t.Run("returns error when no record-value for id", func(t *testing.T) {
		t.Parallel()

		recordStorage := NewRecordMemory()
		recordStorage.recsStor[id] = rec

		_, err := recordStorage.ForID(ctx, gen.ID())
		require.Error(t, err)
		assert.Equal(t, ErrNotFound, err)
	})
}

func TestRecordStorage_Set(t *testing.T) {
	t.Parallel()

	ctx := inslogger.TestContext(t)

	id := gen.ID()
	rec := getStoreRecord()

	t.Run("saves correct record-value", func(t *testing.T) {
		t.Parallel()

		recordStorage := NewRecordMemory()

		err := recordStorage.Set(ctx, id, rec)
		require.NoError(t, err)
		assert.Equal(t, 1, len(recordStorage.recsStor))
		assert.Equal(t, rec, recordStorage.recsStor[id])
	})

	t.Run("returns override error when saving with the same id", func(t *testing.T) {
		t.Parallel()

		recordStorage := NewRecordMemory()

		err := recordStorage.Set(ctx, id, rec)
		require.NoError(t, err)

		err = recordStorage.Set(ctx, id, rec)
		require.Error(t, err)
		assert.Equal(t, ErrOverride, err)
	})
}

func TestRecordStorage_Delete(t *testing.T) {
	t.Parallel()

	ctx := inslogger.TestContext(t)

	firstPulse := gen.PulseNumber()
	secondPulse := firstPulse + 1

	t.Run("delete all records for selected pulse", func(t *testing.T) {
		t.Parallel()

		recordStorage := NewRecordMemory()

		countFirstPulse := rand.Int31n(256)
		countSecondPulse := rand.Int31n(256)

		for i := int32(0); i < countFirstPulse; i++ {
			randID := gen.ID()
			id := insolar.NewID(firstPulse, randID.Hash())
			err := recordStorage.Set(ctx, *id, record.Store{})
			require.NoError(t, err)
		}

		for i := int32(0); i < countSecondPulse; i++ {
			randID := gen.ID()
			id := insolar.NewID(secondPulse, randID.Hash())
			err := recordStorage.Set(ctx, *id, record.Store{})
			require.NoError(t, err)
		}
		assert.Equal(t, countFirstPulse+countSecondPulse, int32(len(recordStorage.recsStor)))

		recordStorage.DeleteForPN(ctx, firstPulse)
		assert.Equal(t, countSecondPulse, int32(len(recordStorage.recsStor)))
	})
}

func TestRecordStorage_ForPulse(t *testing.T) {
	t.Parallel()

	ctx := inslogger.TestContext(t)
	recordMemory := NewRecordMemory()

	searchJetID := gen.JetID()
	searchPN := gen.PulseNumber()

	searchRecs := map[insolar.ID]struct{}{}
	for i := int32(0); i < rand.Int31n(256); i++ {
		rec := getStoreRecord()
		rec.JetID = searchJetID

		h := sha256.New()
		hash := record.HashVirtual(h, *rec.Virtual)

		id := insolar.NewID(searchPN, hash)

		searchRecs[*id] = struct{}{}
		err := recordMemory.Set(ctx, *id, rec)
		require.NoError(t, err)
	}

	for i := int32(0); i < rand.Int31n(512); i++ {
		rec := getStoreRecord()

		randID := gen.ID()
		rID := insolar.NewID(gen.PulseNumber(), randID.Hash())
		err := recordMemory.Set(ctx, *rID, rec)
		require.NoError(t, err)
	}

	res := recordMemory.ForPulse(ctx, searchJetID, searchPN)
	require.Equal(t, len(searchRecs), len(res))

	for _, r := range res {
		h := sha256.New()
		hash := record.HashVirtual(h, *r.Virtual)

		rID := insolar.NewID(searchPN, hash)
		_, ok := searchRecs[*rID]
		require.Equal(t, true, ok)
	}
}

// getVirtualRecord generates random Virtual record
func getVirtualRecord() record.Virtual {
	var requestRecord record.Request

	obj := gen.Reference()
	requestRecord.Object = &obj

	virtualRecord := record.Virtual{
		Union: &record.Virtual_Request{
			Request: &requestRecord,
		},
	}

	return virtualRecord
}

// getStoreRecord generates random Store record
func getStoreRecord() record.Store {
	virtRec := getVirtualRecord()
	return record.Store{
		Virtual: &virtRec,
		JetID:   gen.JetID(),
	}
}
