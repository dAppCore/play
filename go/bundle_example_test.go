package play

func ExampleLoadBundle() {
	_ = LoadBundle
}

func ExampleBundle_Validate() {
	_ = (*Bundle).Validate
}

func ExamplePathError_Error() {
	_ = (*PathError).Error
}
