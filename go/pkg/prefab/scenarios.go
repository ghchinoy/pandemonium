package prefab

import "embed"

// BuiltinScenarios contains the default YAML templates.
//go:embed scenarios/*.yaml
var BuiltinScenarios embed.FS
