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
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/gofuzz"
	"github.com/insolar/insolar/insolar/pulse"
	"github.com/stretchr/testify/require"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/jet"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/internal/ledger/store"
)

func TestNewJetKeeper(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "bdb-test-")
	defer os.RemoveAll(tmpdir)
	require.NoError(t, err)

	db, err := store.NewBadgerDB(tmpdir)
	require.NoError(t, err)
	defer db.Stop(context.Background())
	jets := jet.NewDBStore(db)
	pulses := pulse.NewCalculatorMock(t)
	jetKeeper := NewJetKeeper(jets, db, pulses)
	require.NotNil(t, jetKeeper)
}

func TestDbJetKeeper_AddJet(t *testing.T) {
	ctx := inslogger.TestContext(t)

	tmpdir, err := ioutil.TempDir("", "bdb-test-")
	defer os.RemoveAll(tmpdir)
	require.NoError(t, err)

	db, err := store.NewBadgerDB(tmpdir)
	require.NoError(t, err)
	defer db.Stop(context.Background())
	jets := jet.NewDBStore(db)
	pulses := pulse.NewCalculatorMock(t)
	pulses.BackwardsFunc = func(p context.Context, p1 insolar.PulseNumber, p2 int) (r insolar.Pulse, r1 error) {
		return insolar.Pulse{PulseNumber: p1 - insolar.PulseNumber(p2)}, nil
	}
	jetKeeper := NewJetKeeper(jets, db, pulses)

	var (
		pulse insolar.PulseNumber
		jet   insolar.JetID
	)
	f := fuzz.New()
	f.Fuzz(&pulse)
	f.Fuzz(&jet)
	err = jetKeeper.AddJet(ctx, pulse, jet)
	require.NoError(t, err)
}

func TestDbJetKeeper_TopSyncPulse(t *testing.T) {
	ctx := inslogger.TestContext(t)

	tmpdir, err := ioutil.TempDir("", "bdb-test-")
	defer os.RemoveAll(tmpdir)
	require.NoError(t, err)

	db, err := store.NewBadgerDB(tmpdir)
	require.NoError(t, err)
	defer db.Stop(context.Background())
	jets := jet.NewDBStore(db)
	pulses := pulse.NewCalculatorMock(t)
	pulses.BackwardsFunc = func(p context.Context, p1 insolar.PulseNumber, p2 int) (r insolar.Pulse, r1 error) {
		return insolar.Pulse{PulseNumber: insolar.GenesisPulse.PulseNumber}, nil
	}
	pulses.ForwardsFunc = func(p context.Context, p1 insolar.PulseNumber, p2 int) (r insolar.Pulse, r1 error) {
		return insolar.Pulse{}, pulse.ErrNotFound
	}
	jetKeeper := NewJetKeeper(jets, db, pulses)

	require.Equal(t, insolar.GenesisPulse.PulseNumber, jetKeeper.TopSyncPulse())

	var (
		currentPulse insolar.PulseNumber
		futurePulse  insolar.PulseNumber
		jet          insolar.JetID
	)
	currentPulse = 10
	futurePulse = 20
	jet = insolar.ZeroJetID

	err = jets.Update(ctx, currentPulse, true, jet)
	require.NoError(t, err)
	err = jetKeeper.AddJet(ctx, currentPulse, jet)

	require.NoError(t, err)

	// it's still top confirmed
	require.Equal(t, insolar.GenesisPulse.PulseNumber, jetKeeper.TopSyncPulse())

	err = jetKeeper.AddHotConfirmation(ctx, currentPulse, jet)
	require.Equal(t, currentPulse, jetKeeper.TopSyncPulse())

	pulses.BackwardsFunc = func(p context.Context, p1 insolar.PulseNumber, p2 int) (r insolar.Pulse, r1 error) {
		return insolar.Pulse{PulseNumber: currentPulse}, nil
	}

	err = jets.Clone(ctx, currentPulse, futurePulse, true)
	require.NoError(t, err)
	left, right, err := jets.Split(ctx, futurePulse, jet)
	require.NoError(t, err)
	err = jetKeeper.AddJet(ctx, futurePulse, left)
	require.NoError(t, err)
	require.Equal(t, currentPulse, jetKeeper.TopSyncPulse())
	err = jetKeeper.AddJet(ctx, futurePulse, right)
	require.NoError(t, err)
	require.Equal(t, currentPulse, jetKeeper.TopSyncPulse())

	err = jetKeeper.AddHotConfirmation(ctx, futurePulse, right)
	require.NoError(t, err)
	require.Equal(t, currentPulse, jetKeeper.TopSyncPulse())
	err = jetKeeper.AddHotConfirmation(ctx, futurePulse, left)
	require.NoError(t, err)
	require.Equal(t, futurePulse, jetKeeper.TopSyncPulse())
}

func TestDbJetKeeper_TopSyncPulse_FinalizeMultiple(t *testing.T) {
	ctx := inslogger.TestContext(t)

	tmpdir, err := ioutil.TempDir("", "bdb-test-")
	defer os.RemoveAll(tmpdir)
	require.NoError(t, err)

	db, err := store.NewBadgerDB(tmpdir)
	require.NoError(t, err)
	defer db.Stop(context.Background())
	jets := jet.NewDBStore(db)
	pulses := pulse.NewDB(db)
	err = pulses.Append(ctx, insolar.Pulse{PulseNumber: insolar.GenesisPulse.PulseNumber})
	require.NoError(t, err)

	jetKeeper := NewJetKeeper(jets, db, pulses)

	require.Equal(t, insolar.GenesisPulse.PulseNumber, jetKeeper.TopSyncPulse())

	var (
		currentPulse insolar.PulseNumber
		futurePulse  insolar.PulseNumber
		nextPulse    insolar.PulseNumber
		jet          insolar.JetID
	)
	currentPulse = insolar.GenesisPulse.PulseNumber + 10
	nextPulse = insolar.GenesisPulse.PulseNumber + 20
	futurePulse = insolar.GenesisPulse.PulseNumber + 30
	jet = insolar.ZeroJetID

	inslogger.FromContext(ctx).Debug("INIT: JET: ", jet.DebugString())

	err = jets.Update(ctx, currentPulse, true, jet)
	require.NoError(t, err)

	err = pulses.Append(ctx, insolar.Pulse{PulseNumber: currentPulse})
	require.NoError(t, err)

	// Complete currentPulse pulse
	{
		err = jetKeeper.AddHotConfirmation(ctx, currentPulse, jet)
		require.NoError(t, err)
		err = jetKeeper.AddJet(ctx, currentPulse, jet)
		require.NoError(t, err)
		require.Equal(t, currentPulse, jetKeeper.TopSyncPulse())
	}

	err = jets.Clone(ctx, currentPulse, nextPulse, true)
	require.NoError(t, err)
	left, right, err := jets.Split(ctx, nextPulse, jet)
	require.NoError(t, err)

	err = pulses.Append(ctx, insolar.Pulse{PulseNumber: nextPulse})
	require.NoError(t, err)
	err = pulses.Append(ctx, insolar.Pulse{PulseNumber: futurePulse})
	require.NoError(t, err)

	// Complete future pulse
	{
		err = jets.Clone(ctx, nextPulse, futurePulse, true)
		require.NoError(t, err)
		leftFuture, rightFuture, err := jets.Split(ctx, futurePulse, left)
		require.NoError(t, err)
		err = jetKeeper.AddJet(ctx, futurePulse, leftFuture)
		require.NoError(t, err)
		require.Equal(t, currentPulse, jetKeeper.TopSyncPulse())
		err = jetKeeper.AddJet(ctx, futurePulse, rightFuture)
		require.NoError(t, err)
		require.Equal(t, currentPulse, jetKeeper.TopSyncPulse())
		err = jetKeeper.AddJet(ctx, futurePulse, right)
		require.Equal(t, currentPulse, jetKeeper.TopSyncPulse())

		err = jetKeeper.AddHotConfirmation(ctx, futurePulse, rightFuture)
		require.NoError(t, err)
		require.Equal(t, currentPulse, jetKeeper.TopSyncPulse())
		err = jetKeeper.AddHotConfirmation(ctx, futurePulse, leftFuture)
		require.NoError(t, err)
		require.Equal(t, currentPulse, jetKeeper.TopSyncPulse())
		err = jetKeeper.AddHotConfirmation(ctx, futurePulse, right)
		require.NoError(t, err)
		require.Equal(t, currentPulse, jetKeeper.TopSyncPulse())
	}

	// complete next pulse
	{
		err = jetKeeper.AddJet(ctx, nextPulse, left)
		require.NoError(t, err)
		require.Equal(t, currentPulse, jetKeeper.TopSyncPulse())
		err = jetKeeper.AddJet(ctx, nextPulse, right)
		require.NoError(t, err)

		err = jetKeeper.AddHotConfirmation(ctx, nextPulse, left)
		require.NoError(t, err)
		require.Equal(t, currentPulse, jetKeeper.TopSyncPulse())
		err = jetKeeper.AddHotConfirmation(ctx, nextPulse, right)
		require.NoError(t, err)
	}

	require.Equal(t, futurePulse, jetKeeper.TopSyncPulse())

}

func TestDbJetKeeper_Add_CantGetPulse(t *testing.T) {
	ctx := inslogger.TestContext(t)
	dbMock := store.NewDBMock(t)

	pn := insolar.GenesisPulse.PulseNumber

	dbMock.GetMock.Expect(jetKeeperKey(pn)).Return([]byte{}, nil)

	jets := jet.NewStorageMock(t)
	pulses := pulse.NewCalculatorMock(t)

	jetKeeper := NewJetKeeper(jets, dbMock, pulses)
	err := jetKeeper.AddHotConfirmation(ctx, pn, insolar.ZeroJetID)
	require.Error(t, err)

	err = jetKeeper.AddJet(ctx, pn, insolar.ZeroJetID)
	require.Error(t, err)

}