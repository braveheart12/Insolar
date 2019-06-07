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

// THIS CODE IS AUTOGENERATED

package builtin

import (
	"github.com/pkg/errors"

	helloworld "github.com/insolar/insolar/logicrunner/builtin/contract/helloworld"

	XXX_insolar "github.com/insolar/insolar/insolar"
	XXX_rootdomain "github.com/insolar/insolar/insolar/rootdomain"
	XXX_artifacts "github.com/insolar/insolar/logicrunner/artifacts"
	XXX_preprocessor "github.com/insolar/insolar/logicrunner/preprocessor"
)

func InitializeContractMethods() map[string]XXX_preprocessor.ContractWrapper {
	return map[string]XXX_preprocessor.ContractWrapper{
		"helloworld": helloworld.Initialize(),
	}
}

func shouldLoadRef(strRef string) XXX_insolar.Reference {
	ref, err := XXX_insolar.NewReferenceFromBase58(strRef)
	if err != nil {
		panic(errors.Wrap(err, "Unexpected error, bailing out"))
	}
	return *ref
}

func InitializeCodeRefs() map[XXX_insolar.Reference]string {
	rv := make(map[XXX_insolar.Reference]string, 0)

	rv[shouldLoadRef("111A5w1GcnTsht82duVrnWdVHVNyrxCUVcSPLtgQCPR.11111111111111111111111111111111")] = "helloworld"

	return rv
}

func InitializeCodeDescriptors() []XXX_artifacts.CodeDescriptor {
	rv := make([]XXX_artifacts.CodeDescriptor, 0)

	// helloworld
	rv = append(rv, XXX_artifacts.NewCodeDescriptor(
		/* code:        */ nil,
		/* machineType: */ XXX_insolar.MachineTypeBuiltin,
		/* ref:         */ shouldLoadRef("111A5w1GcnTsht82duVrnWdVHVNyrxCUVcSPLtgQCPR.11111111111111111111111111111111"),
	))

	return rv
}

func InitializePrototypeDescriptors() []XXX_artifacts.ObjectDescriptor {
	rv := make([]XXX_artifacts.ObjectDescriptor, 0)

	{ // helloworld
		pRef := shouldLoadRef("111A85JAZugtAkQErbDe3eAaTw56DPLku8QGymJUCt2.11111111111111111111111111111111")
		cRef := shouldLoadRef("111A5w1GcnTsht82duVrnWdVHVNyrxCUVcSPLtgQCPR.11111111111111111111111111111111")
		rv = append(rv, XXX_artifacts.NewObjectDescriptor(
			/* head:         */ pRef,
			/* state:        */ *pRef.Record(),
			/* prototype:    */ &cRef,
			/* isPrototype:  */ true,
			/* childPointer: */ nil,
			/* memory:       */ nil,
			/* parent:       */ XXX_rootdomain.RootDomain.Ref(),
		))
	}

	return rv
}
