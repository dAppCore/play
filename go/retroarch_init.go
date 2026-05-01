//go:build engine_retroarch

package play

func init() {
	if err := RegisterEngine(RetroArchEngine{Binary: "retroarch"}); err != nil {
		panic(err)
	}
}
