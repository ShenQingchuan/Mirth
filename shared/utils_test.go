package shared

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTernaryHelper(t *testing.T) {
	Convey("ternary helper", t, func() {
		So(Ternary(true, 1, 2), ShouldEqual, 1)
		So(Ternary(false, 1, 2), ShouldEqual, 2)
	})
}

func TestUnicodePointsToString(t *testing.T) {
	Convey("unicode points to string", t, func() {
		So(UnicodePointToString("77e5").Unwrap(), ShouldResemble, "知")
		So(UnicodePointToString("94F8").Unwrap(), ShouldResemble, "铸")
		So(UnicodePointToString("72fc").Unwrap(), ShouldResemble, "狼")
		So(UnicodePointToString("2F").Unwrap(), ShouldResemble, "/")
	})
}
