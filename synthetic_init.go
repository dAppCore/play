//go:build engine_synthetic

package play

func init() {
	_ = RegisterEngine(SyntheticEngine{})
}
