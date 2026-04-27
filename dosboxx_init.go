//go:build engine_dosboxx

package play

func init() {
	_ = RegisterEngine(DOSBoxXEngine{Binary: "dosbox-x"})
}
