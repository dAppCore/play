//go:build engine_dosboxx

package play

func init() {
	if err := RegisterEngine(DOSBoxXEngine{Binary: "dosbox-x"}); err != nil {
		panic(err)
	}
}
