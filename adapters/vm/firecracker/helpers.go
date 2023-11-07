package firecracker

func intTo64Ptr(val int) *int64 {
	converted := int64(val)

	return &converted
}

func boolPtr(val bool) *bool {
	return &val
}

func strPtr(val string) *string {
	return &val
}
