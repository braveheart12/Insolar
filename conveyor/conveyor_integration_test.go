/*
 *    Copyright 2019 Insolar Technologies
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

package conveyor

import (
	"os"
	"testing"
	"time"

	"github.com/insolar/insolar/conveyor/fsm"
	"github.com/insolar/insolar/conveyor/generator/matrix"
	"github.com/insolar/insolar/conveyor/handler"
	"github.com/insolar/insolar/conveyor/slot"
	"github.com/insolar/insolar/insolar"
	"github.com/stretchr/testify/require"
)

const maxState = fsm.StateID(1000)

type mockStateMachineSet struct {
	stateMachine matrix.StateMachine
}

func (s *mockStateMachineSet) GetStateMachineByID(id int) matrix.StateMachine {
	return s.stateMachine
}

type mockStateMachineHolder struct{}

func (m *mockStateMachineHolder) makeSetAccessor() matrix.SetAccessor {
	return &mockStateMachineSet{
		stateMachine: m.GetStateMachinesByType(),
	}
}

func (m *mockStateMachineHolder) GetFutureConfig() matrix.SetAccessor {
	return m.makeSetAccessor()
}

func (m *mockStateMachineHolder) GetPresentConfig() matrix.SetAccessor {
	return m.makeSetAccessor()
}

func (m *mockStateMachineHolder) GetPastConfig() matrix.SetAccessor {
	return m.makeSetAccessor()
}

func (m *mockStateMachineHolder) GetInitialStateMachine() matrix.StateMachine {
	return m.GetStateMachinesByType()
}

func (m *mockStateMachineHolder) GetStateMachinesByType() matrix.StateMachine {

	sm := matrix.NewStateMachineMock(&testing.T{})
	sm.GetMigrationHandlerFunc = func(s fsm.StateID) (r handler.MigrationHandler) {
		return func(element fsm.SlotElementHelper) (interface{}, fsm.ElementState, error) {
			if s > maxState {
				s /= 2
			}
			return element.GetElementID(), fsm.NewElementState(fsm.ID(s%3), s+1), nil
		}
	}

	sm.GetTransitionHandlerFunc = func(s fsm.StateID) (r handler.TransitHandler) {
		return func(element fsm.SlotElementHelper) (interface{}, fsm.ElementState, error) {
			if s > maxState {
				s /= 2
			}
			return element.GetElementID(), fsm.NewElementState(fsm.ID(s%3), s+1), nil
		}
	}

	sm.GetResponseHandlerFunc = func(s fsm.StateID) (r handler.AdapterResponseHandler) {
		return func(element fsm.SlotElementHelper, response interface{}) (interface{}, fsm.ElementState, error) {
			if s > maxState {
				s /= 2
			}
			return element.GetPayload(), fsm.NewElementState(fsm.ID(s%3), s+1), nil
		}
	}

	return sm
}

func mockHandlerStorage() matrix.StateMachineHolder {
	return &mockStateMachineHolder{}
}

func setup() {
	slot.HandlerStorage = mockHandlerStorage()
}

func testMainWrapper(m *testing.M) int {
	setup()
	code := m.Run()
	return code
}

func TestMain(m *testing.M) {
	os.Exit(testMainWrapper(m))
}

func TestConveyor_ChangePulse(t *testing.T) {
	conveyor, err := NewPulseConveyor()
	require.NoError(t, err)
	callback := mockCallback()
	pulse := insolar.Pulse{PulseNumber: testRealPulse + testPulseDelta}
	err = conveyor.PreparePulse(pulse, callback)
	require.NoError(t, err)

	callback.(*mockSyncDone).GetResult()

	err = conveyor.ActivatePulse()
	require.NoError(t, err)
}

func TestConveyor_ChangePulseMultipleTimes(t *testing.T) {
	conveyor, err := NewPulseConveyor()
	require.NoError(t, err)

	pulseNumber := testRealPulse + testPulseDelta
	for i := 0; i < 20; i++ {
		callback := mockCallback()
		pulseNumber += testPulseDelta
		pulse := insolar.Pulse{PulseNumber: pulseNumber, NextPulseNumber: pulseNumber + testPulseDelta}
		err = conveyor.PreparePulse(pulse, callback)
		require.NoError(t, err)

		callback.(*mockSyncDone).GetResult()

		err = conveyor.ActivatePulse()
		require.NoError(t, err)
	}
}

func TestConveyor_ChangePulseMultipleTimes_WithEvents(t *testing.T) {
	conveyor, err := NewPulseConveyor()
	require.NoError(t, err)

	pulseNumber := testRealPulse + testPulseDelta
	for i := 0; i < 100; i++ {

		go func() {
			for j := 0; j < 1; j++ {
				// TODO: handle error checking.
				_ = conveyor.SinkPush(pulseNumber, "TEST")
				_ = conveyor.SinkPush(pulseNumber-testPulseDelta, "TEST")
				_ = conveyor.SinkPush(pulseNumber+testPulseDelta, "TEST")
				_ = conveyor.SinkPushAll(pulseNumber, []interface{}{"TEST", i * j})
			}
		}()

		go func() {
			for j := 0; j < 100; j++ {
				conveyor.GetState()
			}
		}()

		go func() {
			for j := 0; j < 100; j++ {
				conveyor.IsOperational()
			}
		}()

		callback := mockCallback()
		pulseNumber += testPulseDelta
		pulse := insolar.Pulse{PulseNumber: pulseNumber, NextPulseNumber: pulseNumber + testPulseDelta}
		err = conveyor.PreparePulse(pulse, callback)
		require.NoError(t, err)

		if i == 0 {
			require.Equal(t, 0, callback.(*mockSyncDone).GetResult())
		} else {
			require.Equal(t, 555, callback.(*mockSyncDone).GetResult())
		}

		err = conveyor.ActivatePulse()
		require.NoError(t, err)

		go func() {
			for j := 0; j < 10; j++ {
				require.NoError(t, conveyor.SinkPushAll(pulseNumber, []interface{}{"TEST", i}))
				require.NoError(t, conveyor.SinkPush(pulseNumber, "TEST"))
				require.NoError(t, conveyor.SinkPush(pulseNumber-testPulseDelta, "TEST"))
				// TODO: handle error check
				_ = conveyor.SinkPush(pulseNumber+testPulseDelta, "TEST")
			}
		}()
	}

	time.Sleep(time.Millisecond * 200)
}

// TODO: Add test on InitiateShutdown
