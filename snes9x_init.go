//go:build engine_snes9x

package play

func init() {
	if err := RegisterEngine(Snes9xEngine{Binary: "snes9x"}); err != nil {
		panic(err)
	}
}
