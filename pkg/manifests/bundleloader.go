package manifests

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/operator-framework/api/pkg/metadata"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
)

// bundleLoader loads a bundle directory from disk
type bundleLoader struct {
	dir      string
	bundle   *Bundle
	foundCSV bool
}

func NewBundleLoader(dir string) bundleLoader {
	return bundleLoader{
		dir:    dir,
		bundle: &Bundle{},
	}
}

func (b *bundleLoader) LoadBundle() (err error) {
	b.bundle.Annotations, err = metadata.FindBundleMetadata(b.dir)
	if err != nil {
		return err
	}
	getter, err := metadata.NewAnnotationGetter(b.bundle.Annotations)
	if err != nil {
		return err
	}

	errs := make([]error, 0)
	err = filepath.Walk(getter.GetManifestsDir(), collectWalkErrs(b.LoadBundleWalkFunc, &errs))
	if err != nil {
		errs = append(errs, err)
	}
	if err := loadDependencies(b.bundle, getter.GetMetadataDir()); err != nil {
		errs = append(errs, err)
	}

	if !b.foundCSV {
		errs = append(errs, fmt.Errorf("unable to find a csv in bundle directory %s", b.dir))
	} else if b.bundle == nil {
		errs = append(errs, fmt.Errorf("unable to load bundle from directory %s", b.dir))
	}

	return utilerrors.NewAggregate(errs)
}

// collectWalkErrs calls the given walk func and appends any non-nil, non skip dir error returned to the given errors slice.
func collectWalkErrs(walk filepath.WalkFunc, errs *[]error) filepath.WalkFunc {
	return func(path string, f os.FileInfo, err error) (walkErr error) {
		if walkErr = walk(path, f, err); walkErr != nil && walkErr != filepath.SkipDir {
			*errs = append(*errs, walkErr)
			return nil
		}

		return walkErr
	}
}

func (b *bundleLoader) LoadBundleWalkFunc(path string, f os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if f == nil {
		return fmt.Errorf("invalid file: %v", f)
	}

	if f.IsDir() {
		if strings.HasPrefix(f.Name(), ".") {
			return filepath.SkipDir
		}
		return nil
	}

	if strings.HasPrefix(f.Name(), ".") {
		return nil
	}

	fileReader, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("unable to load file %s: %s", path, err)
	}
	defer fileReader.Close()

	decoder := yaml.NewYAMLOrJSONDecoder(fileReader, 30)
	csv := unstructured.Unstructured{}

	if err = decoder.Decode(&csv); err != nil {
		return nil
	}

	if csv.GetKind() != operatorsv1alpha1.ClusterServiceVersionKind {
		return nil
	}

	b.foundCSV = true
	b.bundle.Name = csv.GetName()

	var errs []error
	err = loadBundleManifests(b.bundle, filepath.Dir(path))
	if err != nil {
		errs = append(errs, fmt.Errorf("error loading objs in directory: %s", err))
	}

	if b.bundle.CSV == nil {
		errs = append(errs, fmt.Errorf("no bundle csv found"))
		return utilerrors.NewAggregate(errs)
	}

	return utilerrors.NewAggregate(errs)
}

// loadBundleManifests takes the directory that a CSV is in and assumes the rest of the objects in that directory
// are part of the bundle.
func loadBundleManifests(bundle *Bundle, dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	var errs []error
	for _, f := range files {
		path := filepath.Join(dir, f.Name())

		if f.IsDir() {
			errs = append(errs, fmt.Errorf("bundle manifests dir contains directory: %s", path))
			continue
		}

		if strings.HasPrefix(f.Name(), ".") {
			errs = append(errs, fmt.Errorf("bundle manifests dir has hidden file: %s", path))
			continue
		}

		fileReader, err := os.Open(path)
		if err != nil {
			errs = append(errs, fmt.Errorf("unable to load file %s: %s", path, err))
			continue
		}
		defer fileReader.Close()

		decoder := yaml.NewYAMLOrJSONDecoder(fileReader, 30)
		obj := &unstructured.Unstructured{}
		if err = decoder.Decode(obj); err != nil {
			errs = append(errs, fmt.Errorf("unable to decode object: %s", err))
			continue
		}

		bundle.Objects = append(bundle.Objects, obj)

		// Reset the reader so we can decode it into a typed object.
		if err = resetFile(fileReader); err != nil {
			errs = append(errs, err)
			continue
		}

		switch kind := obj.GetKind(); kind {
		case "ClusterServiceVersion":
			if bundle.CSV != nil {
				return fmt.Errorf("invalid bundle: contains multiple CSVs")
			}
			csv := operatorsv1alpha1.ClusterServiceVersion{}
			err := decoder.Decode(&csv)
			if err != nil {
				return fmt.Errorf("unable to parse CSV %s: %s", f.Name(), err.Error())
			}
			bundle.CSV = &csv
		case "CustomResourceDefinition":
			version := obj.GetAPIVersion()
			if version == apiextensionsv1beta1.SchemeGroupVersion.String() {
				crd := apiextensionsv1beta1.CustomResourceDefinition{}
				err := decoder.Decode(&crd)
				if err != nil {
					return fmt.Errorf("unable to parse CRD %s: %s", f.Name(), err.Error())
				}
				bundle.V1beta1CRDs = append(bundle.V1beta1CRDs, &crd)
			} else if version == apiextensionsv1.SchemeGroupVersion.String() {
				crd := apiextensionsv1.CustomResourceDefinition{}
				err := decoder.Decode(&crd)
				if err != nil {
					return fmt.Errorf("unable to parse CRD %s: %s", f.Name(), err.Error())
				}
				bundle.V1CRDs = append(bundle.V1CRDs, &crd)
			} else {
				return fmt.Errorf("unsupported CRD version %s for %s", version, f.Name())
			}
		}
	}

	return utilerrors.NewAggregate(errs)
}

// resetFile seeks f to read from 0, assuming it is read-only.
func resetFile(f *os.File) error {
	r, err := f.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("unable to reset file %s: %v", f.Name(), err)
	}
	if r != 0 {
		return fmt.Errorf("unable to reset file %s: seek is %d not 0", f.Name(), r)
	}
	return nil
}

func loadDependencies(bundle *Bundle, dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	var errs []error
	for _, f := range files {
		path := filepath.Join(dir, f.Name())

		if f.IsDir() {
			errs = append(errs, fmt.Errorf("bundle manifests dir contains directory: %s", path))
			continue
		}

		if strings.HasPrefix(f.Name(), ".") {
			errs = append(errs, fmt.Errorf("bundle manifests dir has hidden file: %s", path))
			continue
		}

		fileReader, err := os.Open(path)
		if err != nil {
			errs = append(errs, fmt.Errorf("unable to load file %s: %s", path, err))
			continue
		}
		defer fileReader.Close()

		decoder := yaml.NewYAMLOrJSONDecoder(fileReader, 30)
		obj := &unstructured.Unstructured{}
		if err = decoder.Decode(obj); err != nil {
			errs = append(errs, fmt.Errorf("unable to decode object: %s", err))
			continue
		}

		bundle.Objects = append(bundle.Objects, obj)

		// Reset the reader so we can decode it into a typed object.
		if err = resetFile(fileReader); err != nil {
			errs = append(errs, err)
			continue
		}

		switch kind := obj.GetKind(); kind {
		case "ClusterServiceVersion":
			if bundle.CSV != nil {
				return fmt.Errorf("invalid bundle: contains multiple CSVs")
			}
			csv := operatorsv1alpha1.ClusterServiceVersion{}
			err := decoder.Decode(&csv)
			if err != nil {
				return fmt.Errorf("unable to parse CSV %s: %s", f.Name(), err.Error())
			}
			bundle.CSV = &csv
		case "CustomResourceDefinition":
			version := obj.GetAPIVersion()
			if version == apiextensionsv1beta1.SchemeGroupVersion.String() {
				crd := apiextensionsv1beta1.CustomResourceDefinition{}
				err := decoder.Decode(&crd)
				if err != nil {
					return fmt.Errorf("unable to parse CRD %s: %s", f.Name(), err.Error())
				}
				bundle.V1beta1CRDs = append(bundle.V1beta1CRDs, &crd)
			} else if version == apiextensionsv1.SchemeGroupVersion.String() {
				crd := apiextensionsv1.CustomResourceDefinition{}
				err := decoder.Decode(&crd)
				if err != nil {
					return fmt.Errorf("unable to parse CRD %s: %s", f.Name(), err.Error())
				}
				bundle.V1CRDs = append(bundle.V1CRDs, &crd)
			} else {
				return fmt.Errorf("unsupported CRD version %s for %s", version, f.Name())
			}
		}
	}

	return utilerrors.NewAggregate(errs)
}
