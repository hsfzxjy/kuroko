package dffi

const (
	CALL_SPREAD int32 = 0 << 0
	CALL_ARRAY  int32 = 1 << 0

	CALL_WITHOUT_CODE int32 = 0 << 1
	CALL_WITH_CODE    int32 = 1 << 1

	CALL_ONCE  int32 = 0 << 2
	CALL_MULTI int32 = 1 << 2

	FUT_RESOLVED int32 = 0 << 3
	FUT_REJECTED int32 = 1 << 3
)

const (
	RETCODE_OK    int32 = 0
	RETCODE_ERROR int32 = 1
)