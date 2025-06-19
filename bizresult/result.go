package bizresult

// Result represents either a success value or an error.
type Result[T any] struct {
	value T
	err   error
}

type ResultEmpty = Result[struct{}]

func New[T any]() *Result[T] {

	return &Result[T]{}
}

// Ok creates a new Result with a success value.
func Ok[T any](value T) *Result[T] {
	return &Result[T]{
		value: value,
		err:   nil,
	}
}

// Err creates a new Result with an error value.
func Err[T any, E error](err E) *Result[T] {
	var zero T

	return &Result[T]{
		value: zero,
		err:   err,
	}
}

func (r *Result[T]) Must() T {
	if r.err != nil {
		panic(r.err)
	}

	return r.value
}

func (r *Result[T]) MustOr(defaultValue T) T {
	if r.err == nil {
		return r.value
	}

	return defaultValue
}

// Raw returns the error value if Result is Err, panics otherwise.
func (r *Result[T]) Raw() (T, error) {

	return r.value, r.err
}

func (r *Result[T]) SetErr(err error) *Result[T] {
	r.err = err

	return r
}

func (r *Result[T]) SetValue(value T) *Result[T] {
	r.value = value

	return r
}

type WithExist[T any] struct {
	Data   T
	Exists bool
}
