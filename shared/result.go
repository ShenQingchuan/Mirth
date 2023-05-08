package shared

type IError interface {
	error
	comparable
}

type Result[T any, E IError] struct {
	Value T
	Err   E
	Ok    bool
	Pass  bool
}

type IResult[T any, E IError] interface {
	ResultOk(value T) *Result[T, error]
	ResultErr(err E) *Result[T, E]
}

func ResultPass[E IError]() *Result[any, E] {
	return &Result[any, E]{Ok: true, Pass: true}
}

func ResultOk[T any, E IError](value T) *Result[T, E] {
	return &Result[T, E]{Value: value, Ok: true}
}

func ResultErr[T any, E IError](err E) *Result[T, E] {
	return &Result[T, E]{Err: err}
}

func (r *Result[T, E]) Unwrap() T {
	if !r.Ok {
		panic(r.Err)
	}
	return r.Value
}
func (r *Result[T, E]) UnwrapOr(or T) T {
	return Ternary(
		r.Ok,
		r.Value,
		or,
	)
}
func (r *Result[T, E]) UnwrapOrElse(orElse func() T) T {
	return Ternary(
		r.Ok,
		r.Value,
		orElse(),
	)
}
