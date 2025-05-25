package functional

func Map[T any, R any](input []T, mapper func(T) R) []R {
	output := make([]R, len(input))
	for i, v := range input {
		output[i] = mapper(v)
	}
	return output
}
