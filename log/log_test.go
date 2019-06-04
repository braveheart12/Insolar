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

package log

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insolar/insolar/configuration"
	"github.com/insolar/insolar/insolar"
)

func capture(f func()) string {
	var buf bytes.Buffer
	SetOutput(&buf)
	defer SetOutput(os.Stderr)

	f()
	return buf.String()
}

var logLevelEnvVarName = "INSOLAR_LOG_LEVEL"

func testWithEnvVar(t *testing.T) {
	val := strings.ToLower(os.Getenv(logLevelEnvVarName))

	assert.Containsf(t,
		capture(func() { Warn("HelloWorld") }),
		"HelloWorld", "Warn on level=%v by is set", val)
	assert.Containsf(t,
		capture(func() { Info("HelloWorld") }),
		"HelloWorld", "Info on level=%v is set", val)

	if val == "debug" {
		assert.Containsf(t,
			capture(func() { Debug("HelloWorld") }),
			"HelloWorld", "Debug should work on level %v", val)
	} else {
		assert.NotContainsf(t, capture(func() { Debug("HelloWorld") }),
			"HelloWorld", "Debug should not work on level %v", val)
	}
}

func TestLog_GlobalLogger_Env(t *testing.T) {
	if os.Getenv("__TestLoggerWithEnv__") == "1" {
		testWithEnvVar(t)
		return
	}

	levels := []string{"", "debug"}
	for _, val := range levels {
		name := val
		if name == "" {
			name = "empty"
		}
		t.Run(name, func(t *testing.T) {
			env := []string{"__TestLoggerWithEnv__=1"}
			for _, e := range os.Environ() {
				if strings.HasPrefix(e, logLevelEnvVarName+"=") {
					e = logLevelEnvVarName + "=" + val
				}
				env = append(env, e)
			}

			cmd := exec.Command(os.Args[0], "-test.run=TestLog_GlobalLogger_Env")
			cmd.Env = env
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if e, ok := err.(*exec.ExitError); ok && !e.Success() {
				exitCode := 0
				if status, ok := e.Sys().(syscall.WaitStatus); ok {
					exitCode = status.ExitStatus()
				}
				t.Fatalf("%v with env var %v=%v failed (status=%v, code=%v)",
					os.Args[0], logLevelEnvVarName, val, e.String(), exitCode)
			}
		})
	}
}

func assertHelloWorld(t *testing.T, out string) {
	assert.Contains(t, out, "HelloWorld")
}

func TestLog_GlobalLogger(t *testing.T) {

	assert.NoError(t, SetLevel("debug"))

	assertHelloWorld(t, capture(func() { Debug("HelloWorld") }))
	assertHelloWorld(t, capture(func() { Debugf("%s", "HelloWorld") }))

	assertHelloWorld(t, capture(func() { Info("HelloWorld") }))
	assertHelloWorld(t, capture(func() { Infof("%s", "HelloWorld") }))

	assertHelloWorld(t, capture(func() { Warn("HelloWorld") }))
	assertHelloWorld(t, capture(func() { Warnf("%s", "HelloWorld") }))

	assertHelloWorld(t, capture(func() { Error("HelloWorld") }))
	assertHelloWorld(t, capture(func() { Errorf("%s", "HelloWorld") }))

	assert.Panics(t, func() { capture(func() { Panic("HelloWorld") }) })
	assert.Panics(t, func() { capture(func() { Panicf("%s", "HelloWorld") }) })

	// can't catch os.exit() to test Fatal
	// Fatal("HelloWorld")
	// Fatalln("HelloWorld")
	// Fatalf("%s", "HelloWorld")
}

func TestLog_NewLog_Config(t *testing.T) {
	invalidtests := map[string]configuration.Log{
		"InvalidAdapter":   configuration.Log{Level: "Debug", Adapter: "invalid", Formatter: "text"},
		"InvalidLevel":     configuration.Log{Level: "Invalid", Adapter: "zerolog", Formatter: "text"},
		"InvalidFormatter": configuration.Log{Level: "Debug", Adapter: "zerolog", Formatter: "invalid"},
	}

	for name, test := range invalidtests {
		t.Run(name, func(t *testing.T) {
			logger, err := NewLog(test)
			assert.Nil(t, logger)
			assert.Error(t, err)
		})
	}

	validtests := map[string]configuration.Log{
		"WithAdapter": configuration.Log{Level: "Debug", Adapter: "zerolog", Formatter: "text"},
	}
	for name, test := range validtests {
		t.Run(name, func(t *testing.T) {
			logger, err := NewLog(test)
			assert.NotNil(t, logger)
			assert.NoError(t, err)
		})
	}
}

func TestLog_GlobalLogger_Level(t *testing.T) {
	defer SetLevel("info")
	assert.NoError(t, SetLevel("error"))
	assert.Error(t, SetLevel("errorrr"))
}

func TestLog_AddFields(t *testing.T) {
	errtxt1 := "~CHECK_ERROR_OUTPUT_WITH_FIELDS~"
	errtxt2 := "~CHECK_ERROR_OUTPUT_WITHOUT_FIELDS~"

	var (
		fieldname  = "TraceID"
		fieldvalue = "Trace100500"
	)
	tt := []struct {
		name    string
		fieldfn func(la insolar.Logger) insolar.Logger
	}{
		{
			name: "WithFields",
			fieldfn: func(la insolar.Logger) insolar.Logger {
				fields := map[string]interface{}{fieldname: fieldvalue}
				return la.WithFields(fields)
			},
		},
		{
			name: "WithField",
			fieldfn: func(la insolar.Logger) insolar.Logger {
				return la.WithField(fieldname, fieldvalue)
			},
		},
	}

	for _, tItem := range tt {
		t.Run(tItem.name, func(t *testing.T) {
			la, err := newZerologAdapter(configuration.NewLog())
			assert.NoError(t, err)

			var b bytes.Buffer
			logger := la.WithOutput(&b)

			tItem.fieldfn(logger).Error(errtxt1)
			logger.Error(errtxt2)

			var recitems []string
			for {
				line, err := b.ReadBytes('\n')
				if err != nil && err != io.EOF {
					require.NoError(t, err)
				}

				recitems = append(recitems, string(line))
				if err == io.EOF {
					break
				}
			}
			assert.Contains(t, recitems[0], errtxt1)
			assert.Contains(t, recitems[1], errtxt2)
			assert.Contains(t, recitems[0], fieldvalue)
			assert.NotContains(t, recitems[1], fieldvalue)
		})
	}
}

func TestLog_Timestamp(t *testing.T) {
	for _, adapter := range []string{"zerolog"} {
		adapter := adapter
		t.Run(adapter, func(t *testing.T) {
			log, err := NewLog(configuration.Log{Level: "info", Adapter: adapter, Formatter: "json"})
			require.NoError(t, err)
			require.NotNil(t, log)

			var buf bytes.Buffer
			log = log.WithOutput(&buf)

			log.Error("test")

			require.Regexp(t, regexp.MustCompile("[0-9][0-9]:[0-9][0-9]:[0-9][0-9]"), buf.String())
		})
	}
}
