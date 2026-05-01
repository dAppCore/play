//go:build engine_fuse

package play

func init() {
	if err := RegisterEngine(FUSEEngine{Binary: "fuse"}); err != nil {
		panic(err)
	}
}
