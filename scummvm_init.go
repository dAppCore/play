//go:build engine_scummvm

package play

func init() {
	_ = RegisterEngine(ScummVMEngine{Binary: "scummvm"})
}
