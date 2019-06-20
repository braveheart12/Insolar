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
	"context"
	"sync"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/record"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/internal/ledger/store"
	"github.com/pkg/errors"
	"go.opencensus.io/stats"
)

// IndexDB is a db-based storage, that stores a collection of IndexBuckets
type IndexDB struct {
	lock sync.RWMutex
	db   store.DB

	recordStore *RecordDB
}

type indexKey struct {
	pn    insolar.PulseNumber
	objID insolar.ID
}

func (k indexKey) Scope() store.Scope {
	return store.ScopeIndex
}

func (k indexKey) ID() []byte {
	return append(k.pn.Bytes(), k.objID.Bytes()...)
}

type lastKnownIndexPNKey struct {
	objID insolar.ID
}

func (k lastKnownIndexPNKey) Scope() store.Scope {
	return store.ScopeLastKnownIndexPN
}

func (k lastKnownIndexPNKey) ID() []byte {
	return k.objID.Bytes()
}

// NewIndexDB creates a new instance of IndexDB
func NewIndexDB(db store.DB) *IndexDB {
	return &IndexDB{db: db, recordStore: NewRecordDB(db)}
}

// Set sets a lifeline to a bucket with provided pulseNumber and ID
func (i *IndexDB) Set(ctx context.Context, pn insolar.PulseNumber, objID insolar.ID, lifeline Lifeline) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	if lifeline.Delegates == nil {
		lifeline.Delegates = []LifelineDelegate{}
	}

	buc, err := i.getBucket(pn, objID)
	if err == ErrIndexBucketNotFound {
		buc = &FilamentIndex{}
	} else if err != nil {
		return err
	}

	buc.Lifeline = lifeline
	err = i.setBucket(pn, objID, buc)
	if err != nil {
		return err
	}

	err = i.setLastKnownPN(pn, objID)
	if err != nil {
		return err
	}

	stats.Record(ctx,
		statBucketAddedCount.M(1),
	)

	inslogger.FromContext(ctx).Debugf("[Set] lifeline for obj - %v was set successfully", objID.DebugString())

	return nil
}

// SetIndex adds a bucket with provided pulseNumber and ID
func (i *IndexDB) SetIndex(ctx context.Context, pn insolar.PulseNumber, bucket FilamentIndex) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	err := i.setBucket(pn, bucket.ObjID, &bucket)
	if err != nil {
		return err
	}

	stats.Record(ctx,
		statBucketAddedCount.M(1),
	)

	inslogger.FromContext(ctx).Debugf("[SetIndex] bucket for obj - %v was set successfully", bucket.ObjID.DebugString())
	return i.setLastKnownPN(pn, bucket.ObjID)
}

// ForID returns a lifeline from a bucket with provided PN and ObjID
func (i *IndexDB) ForID(ctx context.Context, pn insolar.PulseNumber, objID insolar.ID) (Lifeline, error) {
	var buck *FilamentIndex
	buck, err := i.getBucket(pn, objID)
	if err == ErrIndexBucketNotFound {
		lastPN, err := i.getLastKnownPN(objID)
		if err != nil {
			return Lifeline{}, ErrLifelineNotFound
		}

		buck, err = i.getBucket(lastPN, objID)
		if err != nil {
			return Lifeline{}, err
		}
	} else if err != nil {
		return Lifeline{}, err
	}

	return buck.Lifeline, nil
}

func (i *IndexDB) setBucket(pn insolar.PulseNumber, objID insolar.ID, bucket *FilamentIndex) error {
	key := indexKey{pn: pn, objID: objID}

	buff, err := bucket.Marshal()
	if err != nil {
		return err
	}

	return i.db.Set(key, buff)
}

func (i *IndexDB) getBucket(pn insolar.PulseNumber, objID insolar.ID) (*FilamentIndex, error) {
	buff, err := i.db.Get(indexKey{pn: pn, objID: objID})
	if err == store.ErrNotFound {
		return nil, ErrIndexBucketNotFound
	}
	if err != nil {
		return nil, err
	}
	bucket := FilamentIndex{}
	err = bucket.Unmarshal(buff)
	return &bucket, err
}

func (i *IndexDB) setLastKnownPN(pn insolar.PulseNumber, objID insolar.ID) error {
	key := lastKnownIndexPNKey{objID: objID}
	return i.db.Set(key, pn.Bytes())
}

func (i *IndexDB) getLastKnownPN(objID insolar.ID) (insolar.PulseNumber, error) {
	buff, err := i.db.Get(lastKnownIndexPNKey{objID: objID})
	if err != nil {
		return insolar.FirstPulseNumber, err
	}
	return insolar.NewPulseNumber(buff), err
}

func (i *IndexDB) filament(b *FilamentIndex) ([]record.CompositeFilamentRecord, error) {
	tempRes := make([]record.CompositeFilamentRecord, len(b.PendingRecords))
	for idx, metaID := range b.PendingRecords {
		metaRec, err := i.recordStore.get(metaID)
		if err != nil {
			return nil, err
		}
		pend := record.Unwrap(metaRec.Virtual).(*record.PendingFilament)
		rec, err := i.recordStore.get(pend.RecordID)
		if err != nil {
			return nil, err
		}

		tempRes[idx] = record.CompositeFilamentRecord{
			Meta:     metaRec,
			MetaID:   metaID,
			Record:   rec,
			RecordID: pend.RecordID,
		}
	}

	return tempRes, nil
}

func (i *IndexDB) nextFilament(b *FilamentIndex) (canContinue bool, nextPN insolar.PulseNumber, err error) {
	firstRecord := b.PendingRecords[0]
	metaRec, err := i.recordStore.get(firstRecord)
	if err != nil {
		return false, insolar.PulseNumber(0), err
	}
	pf := record.Unwrap(metaRec.Virtual).(*record.PendingFilament)
	if pf.PreviousRecord != nil {
		return true, pf.PreviousRecord.Pulse(), nil
	} else {
		return false, insolar.PulseNumber(0), nil
	}
}

func (i *IndexDB) Records(ctx context.Context, readFrom insolar.PulseNumber, readUntil insolar.PulseNumber, objID insolar.ID) ([]record.CompositeFilamentRecord, error) {
	currentPN := readFrom
	var res []record.CompositeFilamentRecord

	if readUntil > readFrom {
		return nil, errors.New("readUntil can't be more then readFrom")
	}

	hasFilamentBehind := true
	for hasFilamentBehind && currentPN >= readUntil {
		b, err := i.getBucket(currentPN, objID)
		if err != nil {
			return nil, err
		}
		if len(b.PendingRecords) == 0 {
			return nil, errors.New("can't fetch pednings from index")
		}

		tempRes, err := i.filament(b)
		if err != nil {
			return nil, err
		}
		if len(tempRes) == 0 {
			return nil, errors.New("can't fetch pednings from index")
		}
		res = append(tempRes, res...)

		hasFilamentBehind, currentPN, err = i.nextFilament(b)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}
