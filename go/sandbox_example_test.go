package play

func ExampleSandboxPolicy_ValidateLaunch() {
	_ = (*SandboxPolicy).ValidateLaunch
}

func ExamplePrepareSandbox() {
	_ = PrepareSandbox
}

func ExampleSandboxError_Error() {
	_ = (*SandboxError).Error
}
