package metadata

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/operator-framework/api/pkg/metadata/annotations"
	"github.com/operator-framework/api/pkg/metadata/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	ManifestsDirName    = "manifests"
	MetadataDirName     = "metadata"
	AnnotationsFileName = "annotations.yaml"
	DockerfileName      = "bundle.Dockerfile"
)

type AnnotationGetter interface {
	fmt.Stringer
	SetAnnotations(annotations.File)
	GetRequiredKeys() []annotations.Key
	GetManifestsDir() string
	GetMetadataDir() string
	GetChannels() []string
	GetDefaultChannel() string
}

var mediaTypers = map[string]AnnotationGetter{
	v1.RegistryV1Type{}.String(): v1.RegistryV1Type{},
	v1.PlainType{}.String():      v1.PlainType{},
	v1.HelmType{}.String():       v1.HelmType{},
}

var mediaTypeKeys = []annotations.Key{
	v1.MediaTypeKey,
}

func NewAnnotationGetter(a annotations.File) (AnnotationGetter, error) {
	if len(a.Annotations) == 0 {
		return nil, errors.New("no annotations in annotations file")
	}

	for _, mtKey := range mediaTypeKeys {
		if mt, hasMTKey := a.GetValue(mtKey); hasMTKey {
			if mediaTyper, hasMediaTyper := mediaTypers[mt]; hasMediaTyper {
				mediaTyper.SetAnnotations(a)
				for _, reqKey := range mediaTyper.GetRequiredKeys() {
					if _, hasReqKey := a.GetValue(reqKey); !hasReqKey {
						return nil, fmt.Errorf("required key %s for mediaType %s not present in annotations",
							reqKey, mediaTyper)
					}
				}
				return mediaTyper, nil
			}
		}
	}
	return nil, errors.New("no supported mediaType keys found in annotations file")
}

// FindBundleMetadata walks bundleRoot searching for metadata, and returns that directory if found.
// If one is not found, an error is returned.
func FindBundleMetadata(bundleRoot string) (annotations.File, error) {
	// Check the default path first.
	defaultPath := filepath.Join(bundleRoot, MetadataDirName, AnnotationsFileName)
	annotationsFile := annotations.File{
		Annotations: make(annotations.Annotations),
	}
	err := decodeFile(defaultPath, &annotationsFile)
	if err == nil && len(annotationsFile.Annotations) > 0 {
		return annotationsFile, nil
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return annotations.File{}, err
	}

	// Annotations are not at the default path, so search recursively.
	var annotationsFound bool
	err = filepath.Walk(bundleRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if annotationsFound {
			return nil
		}

		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("unable to load file %s: %v", path, err)
		}
		defer f.Close()

		err = decodeFile(path, annotationsFile)
		if err == nil && len(annotationsFile.Annotations) > 0 {
			annotationsFound = true
		}
		return nil
	})
	if err != nil {
		return annotations.File{}, err
	}

	if !annotationsFound {
		return annotations.File{}, fmt.Errorf("no annotations file found in %s", bundleRoot)
	}

	return annotationsFile, nil
}

// decodeFile decodes the file at a path into the given interface.
func decodeFile(path string, into interface{}) error {
	if into == nil {
		panic("internal error: decode destination must be instantiated before decode")
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("unable to read file %s: %s", path, err)
	}
	defer f.Close()

	decoder := yaml.NewYAMLOrJSONDecoder(f, 30)
	return decoder.Decode(into)
}
