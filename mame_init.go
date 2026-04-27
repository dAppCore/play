//go:build engine_mame

package play

func init() {
	_ = RegisterEngine(MAMEEngine{Binary: "mame"})
}
