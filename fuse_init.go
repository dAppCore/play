//go:build engine_fuse

package play

func init() {
	_ = RegisterEngine(FUSEEngine{Binary: "fuse"})
}
