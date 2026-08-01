package play

func ExampleNewService() {
	_ = NewService
}

func ExampleService_ListBundles() {
	_ = (*Service).ListBundles
}

func ExampleService_VerifyBundle() {
	_ = (*Service).VerifyBundle
}

func ExampleService_PreparePlay() {
	_ = (*Service).PreparePlay
}

func ExampleService_PlanBundle() {
	_ = (*Service).PlanBundle
}

func ExampleService_RenderBundle() {
	_ = (*Service).RenderBundle
}

func ExampleService_WriteBundle() {
	_ = (*Service).WriteBundle
}
