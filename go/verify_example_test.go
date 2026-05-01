package play

func ExampleBundle_Verify() {
	_ = (*Bundle).Verify
}

func ExampleBundle_VerifyWithRegistry() {
	_ = (*Bundle).VerifyWithRegistry
}

func ExampleParseChecksumFile() {
	_ = ParseChecksumFile
}

func ExampleChecksumParseError_Error() {
	_ = (*ChecksumParseError).Error
}
