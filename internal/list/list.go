package list

func Filter[T any](v []T, f func(T) bool) []T {
	var res []T
	for _, i := range v {
		if !f(i) {
			continue
		}

		res = append(res, i)
	}

	return res
}

func Map[T, K any](v []T, f func(T) K) []K {
	var res []K
	for _, i := range v {
		res = append(res, f(i))
	}
	return res
}

func FlatMap[T, K any](v []T, f func(T) []K) []K {
	var res []K
	for _, i := range v {
		res = append(res, f(i)...)
	}
	return res
}

func Contains[T any](v []T, f func(T) bool) bool {
	return len(Filter(v, f)) > 0
}

func All[T any](v []T, f func(T) bool) bool {
	return len(Filter(v, f)) == len(v)
}

func First[T any](v []T, f func(T) bool) T {
	for _, i := range v {
		if f(i) {
			return i
		}
	}

	panic("no element satisifies predicate")
}

func FirstOrDefault[T any](v []T, def T, f func(T) bool) T {
	for _, i := range v {
		if f(i) {
			return i
		}
	}

	return def
}

func Fold[T, K any](v []T, acc K, f func(T, K) K) K {
	for _, i := range v {
		acc = f(i, acc)
	}
	return acc
}
