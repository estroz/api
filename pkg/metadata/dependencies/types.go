package dependencies

// File holds dependency information about a bundle
type File struct {
	// Dependencies is a list of dependencies for a given bundle
	Dependencies []Dependency `json:"dependencies" yaml:"dependencies"`
}

// Dependencies is a list of dependencies for a given bundle
type Dependency struct {
	// The type of dependency. It can be `olm.package` for operator-version based
	// dependency or `olm.gvk` for gvk based dependency. This field is required.
	Type string `json:"type" yaml:"type"`

	// The value of the dependency (either GVKDependency or PackageDependency)
	Value string `json:"value" yaml:"value"`
}
