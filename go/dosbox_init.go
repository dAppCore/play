//go:build engine_dosbox

package play

func init() {
	if err := RegisterEngine(DOSBoxEngine{Binary: "dosbox"}); err != nil {
		panic(err)
	}
}
