//go:build engine_dosbox

package play

func init() {
	_ = RegisterEngine(DOSBoxEngine{Binary: "dosbox"})
}
