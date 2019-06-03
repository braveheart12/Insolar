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

package insolar

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	genesisIDHex  = "00010001ac000000000000000000000000000000000000000000000000000000"
	genesisRefHex = genesisIDHex + "0000000000000000000000000000000000000000000000000000000000000000"
)

func TestGenesisRecordID(t *testing.T) {
	require.Equal(t, genesisIDHex, hex.EncodeToString(GenesisRecord.ID().Bytes()), "genesis ID should always be the same")
}

func TestReference(t *testing.T) {
	require.Equal(t, genesisRefHex, hex.EncodeToString(GenesisRecord.Ref().Bytes()), "genesisRef should always be the same")
}
