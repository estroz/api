package internal

import (
	"testing"

	"github.com/operator-framework/api/pkg/validation/errors"
	"github.com/operator-framework/operator-registry/pkg/registry"
)

func TestValidatePackageManifest(t *testing.T) {
	channels := []registry.PackageChannel{
		{Name: "foo", CurrentCSVName: "bar"},
	}
	pkgName := "test-package"
	pkg := registry.PackageManifest{
		Channels:           channels,
		DefaultChannelName: "foo",
		PackageName:        pkgName,
	}

	cases := []struct {
		validatorFuncTest
		operation func(*registry.PackageManifest)
	}{
		{
			validatorFuncTest{
				description: "successful validation",
			},
			nil,
		},
		{
			validatorFuncTest{
				description: "default channel does not exist in channels",
				wantErr:     true,
				errors: []errors.Error{
					errors.ErrInvalidPackageManifest(`default channel "baz" not found in the list of declared channels`, pkgName),
				},
				numErrs: 1,
			},
			func(pkg *registry.PackageManifest) {
				pkg.DefaultChannelName = "baz"
			},
		},
		{
			validatorFuncTest{
				description: "one channel's CSVName is empty",
				wantErr:     true,
				errors: []errors.Error{
					errors.ErrInvalidPackageManifest(`channel "foo" currentCSV is empty`, pkgName),
				},
				numErrs: 1,
			},
			func(pkg *registry.PackageManifest) {
				pkg.DefaultChannelName = pkg.Channels[0].Name
				pkg.Channels = make([]registry.PackageChannel, 1)
				copy(pkg.Channels, channels)
				pkg.Channels[0].CurrentCSVName = ""
			},
		},
		{
			validatorFuncTest{
				description: "duplicate channel name",
				wantErr:     true,
				errors: []errors.Error{
					errors.ErrInvalidPackageManifest(`duplicate package manifest channel name "foo"`, pkgName),
				},
				numErrs: 1,
			},
			func(pkg *registry.PackageManifest) {
				pkg.Channels = append(channels, channels...)
			},
		},
	}

	for _, c := range cases {
		if c.operation != nil {
			c.operation(&pkg)
		}
		result := validatePackageManifest(&pkg)
		c.check(t, result)
	}
}
