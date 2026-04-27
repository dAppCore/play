//go:build engine_snes9x

package play

func init() {
	_ = RegisterEngine(Snes9xEngine{Binary: "snes9x"})
}
