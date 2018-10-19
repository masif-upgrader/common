package common

type LazyLogString struct {
	Generator func() string
}

func (s LazyLogString) String() string {
	return s.Generator()
}

func (s LazyLogString) MarshalText() (text []byte, err error) {
	return []byte(s.Generator()), nil
}
