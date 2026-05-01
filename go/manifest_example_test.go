package play

func ExampleResourceLimits_IsZero() {
	_ = (*ResourceLimits).IsZero
}

func ExampleLoadManifest() {
	_ = LoadManifest
}

func ExampleManifestMigration_Migrated() {
	_ = (*ManifestMigration).Migrated
}

func ExampleMigrateManifest() {
	_ = MigrateManifest
}

func ExampleParseError_Error() {
	_ = (*ParseError).Error
}
