package model

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	rdef "go.polydawn.net/repeatr/api/def"
)

func Test(t *testing.T) {
	Convey("Given a Project", t, func() {
		proj := &Project{}
		proj.Init()

		Convey("Putting names manually dtrt", func() {
			proj.PutManualTag("x", rdef.Ware{Hash: "a"})
			proj.PutManualTag("y", rdef.Ware{Hash: "b"})
			So(len(proj.Tags), ShouldEqual, 2)
			So(len(proj.RunRecords), ShouldEqual, 0)
			So(len(proj.Memos), ShouldEqual, 0)

			Convey("Putting colliding names dtrt", func() {
				proj.PutManualTag("x", rdef.Ware{Hash: "q"})
				So(len(proj.Tags), ShouldEqual, 2)
				So(len(proj.RunRecords), ShouldEqual, 0)
				So(len(proj.Memos), ShouldEqual, 0)
			})
		})

		Convey("Putting a RunRecord fills in names", func() {
			rr := &rdef.RunRecord{
				HID:        "rr1",
				FormulaHID: "f1",
				Results: rdef.ResultGroup{
					"name1": &rdef.Result{Ware: rdef.Ware{"tar", "h1"}},
				},
			}
			proj.PutResult("tag1", "name1", rr)
			So(len(proj.Tags), ShouldEqual, 1)
			Convey("...and memoizes the runRecord", func() {
				So(len(proj.RunRecords), ShouldEqual, 1)
				So(len(proj.Memos), ShouldEqual, 1)
			})
		})
	})
}
