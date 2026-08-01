package play

func ExampleValidationIssue_Error() {
	_ = (*ValidationIssue).Error
}

func ExampleValidationErrors_Error() {
	_ = (*ValidationErrors).Error
}

func ExampleValidationErrors_HasIssues() {
	_ = (*ValidationErrors).HasIssues
}

func ExampleManifest_Validate() {
	_ = (*Manifest).Validate
}
