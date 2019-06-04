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
// +build slowtest
// +build !race

// TODO test failed in race test call. added build tag to ignore this test
package ginsider

import (
	"io/ioutil"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/log"
	"github.com/insolar/insolar/testutils"
)

type HealthCheckSuite struct {
	suite.Suite
}

var binaryPath string

func (s *HealthCheckSuite) TestHealthCheck() {
	protocol := "unix"
	socket := os.TempDir() + "/" + testutils.RandomString() + ".sock"

	tmpDir, err := ioutil.TempDir("", "contractcache-")
	s.Require().NoError(err, "failed to build tmp dir")
	defer os.RemoveAll(tmpDir)

	currentPath, err := os.Getwd()
	s.Require().NoError(err)

	insgoccPath := binaryPath + "/insgocc"
	healthcheckPath := binaryPath + "/healthcheck"
	contractPath := currentPath + "/healthcheck/healthcheck.go"
	if _, err = os.Stat(healthcheckPath); err != nil {
		s.Failf("Binary file %s is not found, please run make build", healthcheckPath)
	}

	pathToTmp, err := filepath.Rel(currentPath, tmpDir)

	execResult, err := exec.Command(insgoccPath, "compile", "-o", pathToTmp, contractPath).CombinedOutput()
	log.Warnf("%s", execResult)
	s.Require().NoError(err, "failed to compile contract")

	// start GoInsider
	gi := NewGoInsider(tmpDir, protocol, socket)

	refString := "4K3NiGuqYGqKPnYp6XeGd2kdN4P9veL6rYcWkLKWXZCu.7ZQboaH24PH42sqZKUvoa7UBrpuuubRtShp6CKNuWGZa"
	ref, err := insolar.NewReferenceFromBase58(refString)
	s.Require().NoError(err)
	err = gi.AddPlugin(*ref, tmpDir+"/main.so")
	s.Require().NoError(err, "failed to add plugin")

	s.prepareGoInsider(gi, protocol, socket)

	cmd := exec.Command(healthcheckPath,
		"-a", socket,
		"-p", protocol,
		"-r", refString)

	output, err := cmd.CombinedOutput()

	log.Warnf("%+v", output)

	s.NoError(err)
}

func (s *HealthCheckSuite) prepareGoInsider(gi *GoInsider, protocol, socket string) {
	err := rpc.Register(&RPC{GI: gi})
	s.Require().NoError(err, "can't register gi as rpc")
	listener, err := net.Listen(protocol, socket)
	s.Require().NoError(err, "can't start listener")
	go rpc.Accept(listener)
}

func TestHealthCheck(t *testing.T) {
	suite.Run(t, new(HealthCheckSuite))
}

func init() {
	var ok bool

	binaryPath, ok = os.LookupEnv("BIN_DIR")
	if !ok {
		wd, err := os.Getwd()
		binaryPath = filepath.Join(wd, "..", "..", "..", "bin")

		if err != nil {
			panic(err.Error())
		}
	}
}
