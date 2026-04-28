//go:build engine_vice

package play

func init() {
	if err := RegisterEngine(VICEEngine{Binary: "x64sc"}); err != nil {
		panic(err)
	}
}
