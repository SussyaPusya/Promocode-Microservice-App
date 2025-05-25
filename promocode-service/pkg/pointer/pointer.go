package pointer

func To[T any](t T) *T {
	return &t
}

func ToUint32(i uint32) *uint32 {
	return &i
}

func ToInt64(i int64) *int64 {
	return &i
}
