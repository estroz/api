package validation

import (
	"fmt"

	"github.com/operator-framework/api/pkg/validation/errors"
	interfaces "github.com/operator-framework/api/pkg/validation/interfaces"

	"github.com/operator-framework/operator-registry/pkg/registry"
)

type PackageManifestValidator struct{}

func (f PackageManifestValidator) GetFuncs(objs ...interface{}) (funcs interfaces.ValidatorFuncs) {
	for _, obj := range objs {
		switch v := obj.(type) {
		case *registry.PackageManifest:
			funcs = append(funcs, func() errors.ManifestResult {
				return validatePackageManifest(v)
			})
		}
	}
	return funcs
}

func validatePackageManifest(pkg *registry.PackageManifest) errors.ManifestResult {
	result := errors.ManifestResult{Name: pkg.PackageName}
	result.Add(validateChannels(pkg)...)
	return result
}

func validateChannels(pkg *registry.PackageManifest) (errs []errors.Error) {
	if pkg.PackageName == "" {
		errs = append(errs, errors.ErrInvalidPackageManifest("packageName empty", pkg.PackageName))
	}
	if len(pkg.Channels) == 0 {
		errs = append(errs, errors.ErrInvalidPackageManifest("channels empty", pkg.PackageName))
		return errs
	}
	if pkg.DefaultChannelName == "" {
		errs = append(errs, errors.WarnInvalidPackageManifest("default channel not found", pkg.PackageName))
	}

	seen := map[string]struct{}{}
	for i, c := range pkg.Channels {
		if c.Name == "" {
			errs = append(errs, errors.ErrInvalidPackageManifest(fmt.Sprintf("channel %d name cannot be empty", i), pkg.PackageName))
		}
		if c.CurrentCSVName == "" {
			errs = append(errs, errors.ErrInvalidPackageManifest(fmt.Sprintf("channel %q currentCSV cannot be empty", c.Name), pkg.PackageName))
		}
		if _, ok := seen[c.Name]; ok {
			errs = append(errs, errors.ErrInvalidPackageManifest(fmt.Sprintf("duplicate package manifest channel name %q; channel names must be unique", c.Name), pkg.PackageName))
		}
		seen[c.Name] = struct{}{}
	}
	if _, ok := seen[pkg.DefaultChannelName]; !ok && pkg.DefaultChannelName != "" {
		errs = append(errs, errors.ErrInvalidPackageManifest(fmt.Sprintf("default channel %s not found in the list of declared channels", pkg.DefaultChannelName), pkg.PackageName))
	}

	return errs
}
