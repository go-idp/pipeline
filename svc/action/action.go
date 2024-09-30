package action

type Action struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

type Model[T any] struct {
	name   string
	encode func(pl T) ([]byte, error)
	decode func(payload []byte) (T, error)
}

func (s *Model[T]) Name() string {
	return s.name
}

func (s *Model[T]) Encode(pl T) ([]byte, error) {
	return s.encode(pl)
}

func (s *Model[T]) Decode(payload []byte) (T, error) {
	return s.decode(payload)
}

func Create[T any](name string, encode func(pl T) ([]byte, error), decode func(payload []byte) (T, error)) *Model[T] {
	return &Model[T]{
		name:   name,
		encode: encode,
		decode: decode,
	}
}
