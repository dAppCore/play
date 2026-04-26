//go:build engine_retroarch

package play

func init() {
	_ = RegisterEngine(RetroArchEngine{Binary: "retroarch"})
}
