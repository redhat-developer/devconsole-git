package test

func S(values ...string) SliceOfStrings {
	return func() []string {
		return values
	}
}

type SliceOfStrings func() []string
