package validation

import (
	"fmt"

	"github.com/operator-framework/api/pkg/validation/validator"

	"github.com/operator-framework/operator-registry/pkg/registry"
)

type packageValidator struct {
	pkgs map[string]registry.PackageManifest
}

func NewPackageManifestValidator(pkgs ...registry.PackageManifest) validator.Validator {
	val := packageValidator{pkgs: map[string]registry.PackageManifest{}}
	for _, pkg := range pkgs {
		val.pkgs[pkg.PackageName] = pkg
	}
	return &val
}

func (v *packageValidator) Validate() (results []validator.ManifestResult) {
	for key, pkg := range v.pkgs {
		result := pkgInspect(pkg)
		result.Name = key
		results = append(results, result)
	}
	return results
}

func (v packageValidator) Name() string {
	return "Package Validator"
}

func pkgInspect(pkg registry.PackageManifest) (manifestResult validator.ManifestResult) {
	manifestResult = validator.ManifestResult{}
	present, manifestResult := isDefaultPresent(pkg, manifestResult)
	if !present {
		manifestResult.Errors = append(manifestResult.Errors, validator.InvalidDefaultChannel(fmt.Sprintf("Error: default channel %s not found in the list of declared channels", pkg.DefaultChannelName), pkg.DefaultChannelName))
	}
	return
}

func isDefaultPresent(pkg registry.PackageManifest, manifestResult validator.ManifestResult) (bool, validator.ManifestResult) {
	present := false
	for _, channel := range pkg.Channels {
		if pkg.DefaultChannelName == "" {
			manifestResult.Warnings = append(manifestResult.Warnings, validator.InvalidDefaultChannel(fmt.Sprintf("Warning: default channel not found in %s package manifest", pkg.PackageName), pkg.PackageName))
			return true, manifestResult
		} else if pkg.DefaultChannelName == channel.Name {
			present = true
		}
	}
	return present, manifestResult
}
