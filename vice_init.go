//go:build engine_vice

package play

func init() {
	_ = RegisterEngine(VICEEngine{Binary: "x64sc"})
}
