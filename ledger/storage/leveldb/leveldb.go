/*
 *    Copyright 2018 INS Ecosystem
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package leveldb

import (
	"path/filepath"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/insolar/insolar/ledger/index"
	"github.com/insolar/insolar/ledger/record"
)

const (
	dbDirPath = "_db"
)

// LevelLedger represents ledger's LevelDB storage.
type LevelLedger struct {
	ldb     *leveldb.DB
	pulseFn func() record.PulseNum
}

const (
	scopeIDLifeline byte = 1
	scopeIDRecord   byte = 2
)

// InitDB returns LevelLedger with LevelDB initialized with default settings.
func InitDB() (*LevelLedger, error) {
	// Options struct doc: https://godoc.org/github.com/syndtr/goleveldb/leveldb/opt#Options.
	opts := &opt.Options{
		AltFilters:  nil,
		BlockCacher: opt.LRUCacher,
		// BlockCacheCapacity increased to 32MiB from default 8 MiB.
		// BlockCacheCapacity defines the capacity of the 'sorted table' block caching.
		BlockCacheCapacity:                    32 * 1024 * 1024,
		BlockRestartInterval:                  16,
		BlockSize:                             4 * 1024,
		CompactionExpandLimitFactor:           25,
		CompactionGPOverlapsFactor:            10,
		CompactionL0Trigger:                   4,
		CompactionSourceLimitFactor:           1,
		CompactionTableSize:                   2 * 1024 * 1024,
		CompactionTableSizeMultiplier:         1.0,
		CompactionTableSizeMultiplierPerLevel: nil,
		// CompactionTotalSize increased to 32MiB from default 10 MiB.
		// CompactionTotalSize limits total size of 'sorted table' for each level.
		// The limits for each level will be calculated as:
		//   CompactionTotalSize * (CompactionTotalSizeMultiplier ^ Level)
		CompactionTotalSize:                   32 * 1024 * 1024,
		CompactionTotalSizeMultiplier:         10.0,
		CompactionTotalSizeMultiplierPerLevel: nil,
		Comparer:                     comparer.DefaultComparer,
		Compression:                  opt.DefaultCompression,
		DisableBufferPool:            false,
		DisableBlockCache:            false,
		DisableCompactionBackoff:     false,
		DisableLargeBatchTransaction: false,
		ErrorIfExist:                 false,
		ErrorIfMissing:               false,
		Filter:                       nil,
		IteratorSamplingRate:         1 * 1024 * 1024,
		NoSync:                       false,
		NoWriteMerge:                 false,
		OpenFilesCacher:              opt.LRUCacher,
		OpenFilesCacheCapacity:       500,
		ReadOnly:                     false,
		Strict:                       opt.DefaultStrict,
		WriteBuffer:                  16 * 1024 * 1024, // Default is 4 MiB
		WriteL0PauseTrigger:          12,
		WriteL0SlowdownTrigger:       8,
	}

	absPath, err := filepath.Abs(dbDirPath)
	if err != nil {
		return nil, err
	}
	db, err := leveldb.OpenFile(absPath, opts)
	if err != nil {
		return nil, err
	}

	return &LevelLedger{
		ldb: db,
		// FIXME: temporary pulse implementation
		pulseFn: func() record.PulseNum {
			return record.PulseNum(time.Now().Unix() / 10)
		},
	}, nil
}

// GetRecordByKey returns record from leveldb by record.Key
// It just a simple wrapper for GetRecord method.
func (ll *LevelLedger) GetRecordByKey(k record.Key) (record.Record, error) {
	return ll.GetRecord(record.Key2ID(k))
}

// GetRecord returns record from leveldb by record.ID
// It returns ErrNotFound if the DB does not contains the key.
func (ll *LevelLedger) GetRecord(id record.ID) (record.Record, error) {
	getkey := append([]byte{scopeIDRecord}, id[:]...)
	buf, err := ll.ldb.Get(getkey, nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	raw, err := record.DecodeToRaw(buf)
	if err != nil {
		return nil, err
	}
	return raw.ToRecord(), nil
}

func (ll *LevelLedger) setRawRecordByKey(k record.Key, raw *record.Raw) (record.ID, error) {
	id := record.Key2ID(k)
	putkey := append([]byte{scopeIDRecord}, id[:]...)
	return id, ll.ldb.Put(putkey, record.MustEncodeRaw(raw), nil)
}

// SetRecord stores record in leveldb
func (ll *LevelLedger) SetRecord(rec record.Record) (record.ID, error) {
	raw, err := record.EncodeToRaw(rec)
	if err != nil {
		return record.ID{}, err
	}
	key := record.Key{Pulse: ll.pulseFn(), Hash: raw.Hash()}
	return ll.setRawRecordByKey(key, raw)
}

// GetIndex fetches lifeline index from leveldb (records and lifeline indexes have the same id, but different scopes)
func (ll *LevelLedger) GetIndex(id record.ID) (*index.Lifeline, bool) {
	buf, err := ll.ldb.Get(append([]byte{scopeIDLifeline}, id[:]...), nil)
	if err != nil {
		return nil, false
	}
	idx := index.DecodeLifeline(buf)
	return &idx, true
}

// SetIndex stores lifeline index into leveldb (records and lifeline indexes have the same id, but different scopes)
func (ll *LevelLedger) SetIndex(id record.ID, idx *index.Lifeline) error {
	err := ll.ldb.Put(append([]byte{scopeIDLifeline}, id[:]...), index.EncodeLifeline(idx), nil)
	return err
}

// Close terminates db connection
func (ll *LevelLedger) Close() error {
	return ll.ldb.Close()
}
