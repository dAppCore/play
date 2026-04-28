//go:build engine_scummvm

package play

func init() {
	if err := RegisterEngine(ScummVMEngine{Binary: "scummvm"}); err != nil {
		panic(err)
	}
}
