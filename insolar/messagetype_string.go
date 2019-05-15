// Code generated by "stringer -type=MessageType"; DO NOT EDIT.

package insolar

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TypeCallMethod-0]
	_ = x[TypeCallConstructor-1]
	_ = x[TypeReturnResults-2]
	_ = x[TypeExecutorResults-3]
	_ = x[TypeValidateCaseBind-4]
	_ = x[TypeValidationResults-5]
	_ = x[TypePendingFinished-6]
	_ = x[TypeStillExecuting-7]
	_ = x[TypeGetCode-8]
	_ = x[TypeGetObject-9]
	_ = x[TypeGetDelegate-10]
	_ = x[TypeGetChildren-11]
	_ = x[TypeUpdateObject-12]
	_ = x[TypeRegisterChild-13]
	_ = x[TypeSetRecord-14]
	_ = x[TypeValidateRecord-15]
	_ = x[TypeSetBlob-16]
	_ = x[TypeGetObjectIndex-17]
	_ = x[TypeGetPendingRequests-18]
	_ = x[TypeHotRecords-19]
	_ = x[TypeGetJet-20]
	_ = x[TypeAbandonedRequestsNotification-21]
	_ = x[TypeGetRequest-22]
	_ = x[TypeGetPendingRequestID-23]
	_ = x[TypeHeavyStartStop-24]
	_ = x[TypeHeavyPayload-25]
	_ = x[TypeGenesisRequest-26]
	_ = x[TypeNodeSignRequest-27]
}

const _MessageType_name = "TypeCallMethodTypeCallConstructorTypeReturnResultsTypeExecutorResultsTypeValidateCaseBindTypeValidationResultsTypePendingFinishedTypeStillExecutingTypeGetCodeTypeGetObjectTypeGetDelegateTypeGetChildrenTypeUpdateObjectTypeRegisterChildTypeSetRecordTypeValidateRecordTypeSetBlobTypeGetObjectIndexTypeGetPendingRequestsTypeHotRecordsTypeGetJetTypeAbandonedRequestsNotificationTypeGetRequestTypeGetPendingRequestIDTypeHeavyStartStopTypeHeavyPayloadTypeGenesisRequestTypeNodeSignRequest"

var _MessageType_index = [...]uint16{0, 14, 33, 50, 69, 89, 110, 129, 147, 158, 171, 186, 201, 217, 234, 247, 265, 276, 294, 316, 330, 340, 373, 387, 410, 428, 444, 462, 481}

func (i MessageType) String() string {
	if i >= MessageType(len(_MessageType_index)-1) {
		return "MessageType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _MessageType_name[_MessageType_index[i]:_MessageType_index[i+1]]
}
