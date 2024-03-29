// Code generated by "stringer -type=MessageType"; DO NOT EDIT.

package insolar

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TypeCallMethod-0]
	_ = x[TypeReturnResults-1]
	_ = x[TypeExecutorResults-2]
	_ = x[TypeValidateCaseBind-3]
	_ = x[TypeValidationResults-4]
	_ = x[TypePendingFinished-5]
	_ = x[TypeAdditionalCallFromPreviousExecutor-6]
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
	_ = x[TypeGetJet-19]
	_ = x[TypeAbandonedRequestsNotification-20]
	_ = x[TypeGetPendingRequestID-21]
	_ = x[TypeGetOpenRequests-22]
	_ = x[TypeHeavyStartStop-23]
	_ = x[TypeHeavyPayload-24]
	_ = x[TypeGenesisRequest-25]
}

const _MessageType_name = "TypeCallMethodTypeReturnResultsTypeExecutorResultsTypeValidateCaseBindTypeValidationResultsTypePendingFinishedTypeAdditionalCallFromPreviousExecutorTypeStillExecutingTypeGetCodeTypeGetObjectTypeGetDelegateTypeGetChildrenTypeUpdateObjectTypeRegisterChildTypeSetRecordTypeValidateRecordTypeSetBlobTypeGetObjectIndexTypeGetPendingRequestsTypeGetJetTypeAbandonedRequestsNotificationTypeGetPendingRequestIDTypeGetOpenRequestsTypeHeavyStartStopTypeHeavyPayloadTypeGenesisRequest"

var _MessageType_index = [...]uint16{0, 14, 31, 50, 70, 91, 110, 148, 166, 177, 190, 205, 220, 236, 253, 266, 284, 295, 313, 335, 345, 378, 401, 420, 438, 454, 472}

func (i MessageType) String() string {
	if i >= MessageType(len(_MessageType_index)-1) {
		return "MessageType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _MessageType_name[_MessageType_index[i]:_MessageType_index[i+1]]
}
