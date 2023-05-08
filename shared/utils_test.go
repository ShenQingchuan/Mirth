package shared

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnicodePointsToString(t *testing.T) {
	Convey("unicode points to string", t, func() {
		So(UnicodePointToString("77e5").Unwrap(), ShouldResemble, "知")
		So(UnicodePointToString("94F8").Unwrap(), ShouldResemble, "铸")
		So(UnicodePointToString("72fc").Unwrap(), ShouldResemble, "狼")
		So(UnicodePointToString("2F").Unwrap(), ShouldResemble, "/")
	})
}
