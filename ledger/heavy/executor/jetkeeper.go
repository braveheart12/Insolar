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

package executor

import (
	"context"
	"sync"

	"github.com/insolar/insolar/insolar/pulse"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/network/storage"
	"github.com/pkg/errors"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/jet"
	"github.com/insolar/insolar/internal/ledger/store"
)

//go:generate minimock -i github.com/insolar/insolar/ledger/heavy/executor.JetKeeper -o ./ -s _gen_mock.go

// JetKeeper provides a method for adding jet to storage, checking pulse completion and getting access to highest synced pulse.
type JetKeeper interface {
	// AddJet performs adding jet to storage and checks pulse completion.
	AddJet(context.Context, insolar.PulseNumber, insolar.JetID) error
	// AddHotConfirmation performs adding hot confirmation to storage and checks pulse completion.
	AddHotConfirmation(context.Context, insolar.PulseNumber, insolar.JetID) error
	// TopSyncPulse provides access to highest synced (replicated) pulse.
	TopSyncPulse() insolar.PulseNumber
}

func NewJetKeeper(jets jet.Storage, db store.DB, pulses pulse.Calculator) JetKeeper {
	return &dbJetKeeper{
		jetTrees: jets,
		db:       db,
		pulses:   pulses,
	}
}

type dbJetKeeper struct {
	jetTrees jet.Storage

	pulses storage.PulseCalculator

	sync.RWMutex
	db store.DB
}

type jetInfo struct {
	JetID        insolar.JetID
	HotConfirmed bool
	JetConfirmed bool
}

func (j *jetInfo) isConfirmed() bool {
	return j.JetConfirmed && j.HotConfirmed
}

func (jk *dbJetKeeper) AddHotConfirmation(ctx context.Context, pn insolar.PulseNumber, id insolar.JetID) error {
	jk.Lock()
	defer jk.Unlock()

	inslogger.FromContext(ctx).Debug(">>>>>>>>>>>>>>>>> AddHotConfirmation HERE: pulse: ", pn, ". ID: ", id.DebugString())

	if err := jk.addHotConfirm(ctx, pn, id); err != nil {
		return errors.Wrapf(err, "failed to save updated jets")
	}

	inslogger.FromContext(ctx).Debug(">>>>>>>>>>>>>>>>> AddHotConfirmation AFTER addHotConfirm: pulse: ", pn, ". ID: ", id.DebugString())

	err := jk.propagateConsistency(ctx, pn, id)
	return errors.Wrapf(err, "AddHotConfirmation. propagateConsistency returns error")
}

func (jk *dbJetKeeper) AddJet(ctx context.Context, pn insolar.PulseNumber, id insolar.JetID) error {
	jk.Lock()
	defer jk.Unlock()

	inslogger.FromContext(ctx).Debug(">>>>>>>>>>>>>>>>> AddJet HERE: pulse: ", pn, ". ID: ", id.DebugString())

	if err := jk.addJet(ctx, pn, id); err != nil {
		return errors.Wrapf(err, "AddJet. failed to save updated jets")
	}

	inslogger.FromContext(ctx).Debug(">>>>>>>>>>>>>>>>> AddJet AFTER addJet: pulse: ", pn, ". ID: ", id.DebugString())

	err := jk.propagateConsistency(ctx, pn, id)

	return errors.Wrap(err, "propagateConsistency returns error")
}

func (jk *dbJetKeeper) propagateConsistency(ctx context.Context, pn insolar.PulseNumber, jetID insolar.JetID) error {
	logger := inslogger.FromContext(ctx)

	prev, err := jk.pulses.Backwards(ctx, pn, 1)
	if err != nil {
		return errors.Wrapf(err, "failed to get previous pulse for %d", pn)
	}

	top := jk.topSyncPulse()

	logger.Debug(">>>>>>>>>>>>>>>>>.. AFTER Backwards: pulse: ", pn, ". ID: ", jetID.DebugString(),
		". TOP: ", top, ". prev.PulseNumber: ", prev.PulseNumber)

	if prev.PulseNumber == top || prev.PulseNumber == insolar.GenesisPulse.PulseNumber {
		for jk.checkPulseConsistency(ctx, pn) {
			logger.Debug(">>>>>>>>>>>>>>>>>.. AFTER checkPulseConsistency: pulse: ", pn, ". ID: ", jetID.DebugString(),
				". TOP: ", top, ". prev.PulseNumber: ", prev.PulseNumber)
			err := jk.updateSyncPulse(pn)
			if err != nil {
				return errors.Wrapf(err, "failed to update consistent pulse")
			}
			logger.Debugf("pulse completed: %d", pn)

			next, err := jk.pulses.Forwards(ctx, pn, 1)
			if err == pulse.ErrNotFound {
				return nil
			}
			if err != nil {
				return errors.Wrapf(err, "failed to get next pulse for %d", pn)
			}
			pn = next.PulseNumber
		}
	}

	return nil
}

func (jk *dbJetKeeper) TopSyncPulse() insolar.PulseNumber {
	jk.RLock()
	defer jk.RUnlock()

	return jk.topSyncPulse()
}

func (jk *dbJetKeeper) topSyncPulse() insolar.PulseNumber {
	val, err := jk.db.Get(syncPulseKey(insolar.GenesisPulse.PulseNumber))
	if err != nil {
		return insolar.GenesisPulse.PulseNumber
	}
	return insolar.NewPulseNumber(val)
}

func (jk *dbJetKeeper) addJet(ctx context.Context, pulse insolar.PulseNumber, id insolar.JetID) error {
	return jk.updateJet(ctx, pulse, id, true, false)
}

func (jk *dbJetKeeper) updateJet(ctx context.Context, pulse insolar.PulseNumber, id insolar.JetID, jetConfirmed bool, hotConfirmed bool) error {
	logger := inslogger.FromContext(ctx)
	jets, err := jk.get(pulse)
	var exists bool
	if err == nil {
		logger.Debug("pulse complete: addHotConfirm: update existing: ", pulse, ". Jet:", id.DebugString())

		for i, _ := range jets {
			if jets[i].JetID.Equal(id) {
				exists = true
				if hotConfirmed {
					jets[i].HotConfirmed = hotConfirmed
				}
				if jetConfirmed {
					jets[i].JetConfirmed = jetConfirmed
				}
				break
			}
		}
	} else if err != store.ErrNotFound {
		return errors.Wrapf(err, "can't get pulse: %d", pulse)
	}
	if !exists {
		jets = append(jets, jetInfo{JetID: id, HotConfirmed: hotConfirmed, JetConfirmed: jetConfirmed})
		logger.Debug("pulse complete: addHotConfirm: not exists: ", pulse, ". Jet:", id.DebugString())
	}
	return jk.set(pulse, jets)
}

func (jk *dbJetKeeper) addHotConfirm(ctx context.Context, pulse insolar.PulseNumber, id insolar.JetID) error {
	return jk.updateJet(ctx, pulse, id, false, true)
}

func (jk *dbJetKeeper) checkPulseConsistency(ctx context.Context, pulse insolar.PulseNumber) bool {
	infoToSet := func(s []jetInfo) (map[insolar.JetID]struct{}, bool) {
		r := make(map[insolar.JetID]struct{}, len(s))
		for _, el := range s {
			if !el.isConfirmed() {
				return nil, false
			}
			r[el.JetID] = struct{}{}
		}
		return r, true
	}

	infoToList := func(s []jetInfo) []insolar.JetID {
		r := make([]insolar.JetID, len(s))
		for i, el := range s {
			r[i] = el.JetID
		}
		return r
	}

	expectedJets := jk.jetTrees.All(ctx, pulse)
	actualJets := jk.all(pulse)

	if len(expectedJets) != len(actualJets) {
		if len(actualJets) > len(expectedJets) {
			inslogger.FromContext(ctx).Warn("num actual jets is more then expected. it's too bad. Pulse: ", pulse,
				". Expected: ", insolar.JetIDCollection(expectedJets).DebugString(),
				". Actual: ", insolar.JetIDCollection(infoToList(actualJets)).DebugString())
		}
		return false
	}

	actualJetsSet, allConfirmed := infoToSet(actualJets)
	if !allConfirmed {
		return false
	}

	for _, expID := range expectedJets {
		if _, ok := actualJetsSet[expID]; !ok {
			inslogger.FromContext(ctx).Error("jet sets are different. it's too bad. Pulse: ", pulse,
				". Expected: ", insolar.JetIDCollection(expectedJets).DebugString(),
				". Actual: ", insolar.JetIDCollection(infoToList(actualJets)).DebugString())
			return false
		}
	}

	return true
}

func (jk *dbJetKeeper) all(pulse insolar.PulseNumber) []jetInfo {
	jets, err := jk.get(pulse)
	if err != nil {
		jets = []jetInfo{}
	}
	return jets
}

const (
	jetKeeperKeyPrefix = 0x01
	syncPulseKeyPrefix = 0x02
)

type jetKeeperKey insolar.PulseNumber

func (k jetKeeperKey) Scope() store.Scope {
	return store.ScopeJetKeeper
}

func (k jetKeeperKey) ID() []byte {
	return append([]byte{jetKeeperKeyPrefix}, insolar.PulseNumber(k).Bytes()...)
}

type syncPulseKey insolar.PulseNumber

func (k syncPulseKey) Scope() store.Scope {
	return store.ScopeJetKeeper
}

func (k syncPulseKey) ID() []byte {
	return append([]byte{syncPulseKeyPrefix}, insolar.PulseNumber(k).Bytes()...)
}

func (jk *dbJetKeeper) get(pn insolar.PulseNumber) ([]jetInfo, error) {
	serializedJets, err := jk.db.Get(jetKeeperKey(pn))
	if err != nil {
		if err == store.ErrNotFound {
			return nil, err
		}
		return nil, errors.Wrapf(err, "failed to get jets by pulse=%v", pn)
	}

	var jets []jetInfo
	err = insolar.Deserialize(serializedJets, &jets)
	if err != nil {
		return nil, errors.Wrap(err, "failed to deserialize jets")
	}
	return jets, nil
}

func (jk *dbJetKeeper) set(pn insolar.PulseNumber, jets []jetInfo) error {
	key := jetKeeperKey(pn)

	serialized, err := insolar.Serialize(jets)
	if err != nil {
		return errors.Wrap(err, "failed to serialize jets")
	}

	return jk.db.Set(key, serialized)
}

func (jk *dbJetKeeper) updateSyncPulse(pn insolar.PulseNumber) error {
	err := jk.db.Set(syncPulseKey(insolar.GenesisPulse.PulseNumber), pn.Bytes())
	return errors.Wrapf(err, "failed to set up new sync pulse")
}