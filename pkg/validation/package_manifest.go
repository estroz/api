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
		result := validator.ManifestResult{Name: key}
		result.Add(validatePackageManifest(pkg)...)
		results = append(results, result)
	}
	return results
}

func (v packageValidator) Name() string {
	return "Package Validator"
}

func validatePackageManifest(pkg registry.PackageManifest) (errs []validator.Error) {
	if pkg.PackageName == "" {
		errs = append(errs, validator.ErrInvalidPackageManifest("packageName empty", pkg.PackageName))
	}
	if len(pkg.Channels) == 0 {
		errs = append(errs, validator.ErrInvalidPackageManifest("channels empty", pkg.PackageName))
		return errs
	}
	if pkg.DefaultChannelName == "" {
		errs = append(errs, validator.WarnInvalidPackageManifest("default channel not found", pkg.PackageName))
	}

	seen := map[string]struct{}{}
	for i, c := range pkg.Channels {
		if c.Name == "" {
			errs = append(errs, validator.ErrInvalidPackageManifest(fmt.Sprintf("channel %d name cannot be empty", i), pkg.PackageName))
		}
		if c.CurrentCSVName == "" {
			errs = append(errs, validator.ErrInvalidPackageManifest(fmt.Sprintf("channel %q currentCSV cannot be empty", c.Name), pkg.PackageName))
		}
		if _, ok := seen[c.Name]; ok {
			errs = append(errs, validator.ErrInvalidPackageManifest(fmt.Sprintf("duplicate package manifest channel name %q; channel names must be unique", c.Name), pkg.PackageName))
		}
		seen[c.Name] = struct{}{}
	}
	if _, ok := seen[pkg.DefaultChannelName]; !ok && pkg.DefaultChannelName != "" {
		errs = append(errs, validator.ErrInvalidPackageManifest(fmt.Sprintf("default channel %s not found in the list of declared channels", pkg.DefaultChannelName), pkg.PackageName))
	}

	return errs
}
