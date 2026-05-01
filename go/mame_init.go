//go:build engine_mame

package play

func init() {
	if err := RegisterEngine(MAMEEngine{Binary: "mame"}); err != nil {
		panic(err)
	}
}
