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

package replication

import (
	"testing"

	fuzz "github.com/google/gofuzz"
	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/gen"
	"github.com/insolar/insolar/insolar/message"
	"github.com/insolar/insolar/insolar/record"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/ledger/blob"
	"github.com/insolar/insolar/ledger/drop"
	"github.com/insolar/insolar/ledger/object"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestDataGatherer_ForPulseAndJet(t *testing.T) {
	ctx := inslogger.TestContext(t)
	pn := gen.PulseNumber()
	jetID := gen.JetID()

	da := drop.NewAccessorMock(t)
	d := drop.Drop{
		JetID: gen.JetID(),
		Pulse: gen.PulseNumber(),
		Hash:  []byte{4, 2, 3},
	}
	da.ForPulseMock.Expect(ctx, jetID, pn).Return(d, nil)

	ba := blob.NewCollectionAccessorMock(t)
	b := blob.Blob{
		JetID: gen.JetID(),
		Value: []byte{1, 2, 3, 4, 5, 6, 7},
	}
	ba.ForPulseMock.Expect(ctx, jetID, pn).Return([]blob.Blob{b})

	ra := object.NewRecordCollectionAccessorMock(t)
	rec := makeItemRecord()
	ra.ForPulseMock.Expect(ctx, jetID, pn).Return([]record.Item{
		rec,
	})

	ia := object.NewIndexBucketAccessorMock(t)
	idx := object.Lifeline{
		JetID:        gen.JetID(),
		ChildPointer: insolar.NewID(gen.PulseNumber(), nil),
		LatestState:  insolar.NewID(gen.PulseNumber(), nil),
	}
	idxID := gen.ID()
	bucks := []object.FilamentIndex{
		{
			ObjID:    idxID,
			Lifeline: idx,
		},
	}
	ia.ForPNAndJetMock.Return(bucks)

	s := record.Store{
		JetID:   rec.JetID,
		Virtual: record.ToVirtual(rec.Virtual),
	}

	recData, err := s.Marshal()
	require.NoError(t, err)

	expectedMsg := &message.HeavyPayload{
		JetID:        jetID,
		PulseNum:     pn,
		IndexBuckets: convertIndexBuckets(ctx, bucks),
		Drop:         drop.MustEncode(&d),
		Blobs:        [][]byte{blob.MustEncode(&b)},
		Records:      [][]byte{recData},
	}

	dataGatherer := NewDataGatherer(da, ba, ra, ia)

	msg, err := dataGatherer.ForPulseAndJet(ctx, pn, jetID)

	require.NoError(t, err)
	require.Equal(t, expectedMsg, msg)
}

func TestDataGatherer_ForPulseAndJet_DropFetchingFailed(t *testing.T) {
	da := drop.NewAccessorMock(t)
	da.ForPulseMock.Return(drop.Drop{}, errors.New("everything is broken"))

	dataGatherer := NewDataGatherer(da, nil, nil, nil)
	_, err := dataGatherer.ForPulseAndJet(inslogger.TestContext(t), gen.PulseNumber(), gen.JetID())

	require.Error(t, err, errors.New("everything is broken"))
}

func TestLightDataGatherer_convertIndexBuckets(t *testing.T) {
	var idxs []object.FilamentIndex
	fuzz.New().NilChance(0).NumElements(500, 1000).Funcs(func(elem *object.FilamentIndex, c fuzz.Continue) {
		elem.Lifeline = object.Lifeline{
			JetID:        gen.JetID(),
			LatestUpdate: gen.PulseNumber(),
		}
		elem.LifelineLastUsed = gen.PulseNumber()
	}).Fuzz(&idxs)

	var expected [][]byte

	for _, idx := range idxs {
		buff, err := idx.Marshal()
		require.NoError(t, err)
		expected = append(expected, buff)
	}

	resp := convertIndexBuckets(inslogger.TestContext(t), idxs)

	require.Equal(t, resp, expected)

}

func TestDataGatherer_convertBlobs(t *testing.T) {
	var blobs []blob.Blob
	fuzz.New().NilChance(0).NumElements(500, 1000).Fuzz(&blobs)
	var expected [][]byte
	for _, b := range blobs {
		temp := b
		expected = append(expected, blob.MustEncode(&temp))
	}

	resp := convertBlobs(blobs)

	require.Equal(t, resp, expected)
}

func TestDataGatherer_convertRecords(t *testing.T) {
	ctx := inslogger.TestContext(t)

	var itemRecs []record.Item
	fuzz.New().NilChance(0).NumElements(500, 1000).Funcs(
		func(elem *record.Item, c fuzz.Continue) {
			elem.JetID = gen.JetID()
			elem.Virtual = makeVirtualRecord()
		},
	).Fuzz(&itemRecs)

	var expected [][]byte
	for _, r := range itemRecs {
		s := r.ToStore()
		data, err := s.Marshal()
		require.NoError(t, err)
		expected = append(expected, data)
	}

	gotConverted := convertRecords(ctx, itemRecs)
	require.Equal(t, expected, gotConverted)
}

// makeVirtualRecord generates random Virtual record
func makeVirtualRecord() record.Record {
	var requestRecord record.Request

	obj := gen.Reference()
	requestRecord.Object = &obj

	return &requestRecord
}

// makeItemRecord generates random Store record
func makeItemRecord() record.Item {
	virtRec := makeVirtualRecord()
	return record.Item{
		JetID:   gen.JetID(),
		Virtual: virtRec,
	}
}
