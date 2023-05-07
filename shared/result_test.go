package shared

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type TestError struct {
	message string
}

func (e TestError) Error() string {
	return e.message
}

func TestResultMonad(t *testing.T) {
	Convey("Result Monad", t, func() {
		r := ResultOk(1)
		e := ResultErr[int](
			&TestError{"This is an test error"},
		)

		So(r.Ok, ShouldBeTrue)
		So(r.Value, ShouldEqual, 1)
		So(r.Err, ShouldBeNil)

		So(e.Ok, ShouldBeFalse)
		So(e.Value, ShouldEqual, 0) // 0 is the default value for int
		So(e.Err, ShouldNotBeNil)
		So(e.Err.Error(), ShouldEqual, "This is an test error")
	})

	Convey("Result Unwrap", t, func() {
		r := ResultOk(1)
		e := ResultErr[int](
			&TestError{"This is an test error"},
		)

		So(r.Unwrap(), ShouldEqual, 1)
		So(func() {
			e.Unwrap()
		}, ShouldPanic)
	})

	Convey("Result UnwrapOr", t, func() {
		r := ResultOk(1)
		e := ResultErr[int](
			&TestError{"This is an test error"},
		)

		So(r.UnwrapOr(2), ShouldEqual, 1)
		So(e.UnwrapOr(2), ShouldEqual, 2)
	})

	Convey("Result UnwrapOrElse", t, func() {
		r := ResultOk(1)
		e := ResultErr[int](
			&TestError{"This is an test error"},
		)

		So(r.UnwrapOrElse(func() int { return 2 }), ShouldEqual, 1)
		So(e.UnwrapOrElse(func() int { return 2 }), ShouldEqual, 2)
	})
}
