//go:build engine_synthetic

package play

func init() {
	if err := RegisterEngine(SyntheticEngine{}); err != nil {
		panic(err)
	}
}
