package annotations

// File holds annotation information about a bundle
type File struct {
	// annotations is a list of annotations for a given bundle
	Annotations Annotations `json:"annotations" yaml:"annotations"`
}

// Annotations are a mapping of bundle annotation keys to values. These mappings are potentially
// from outside those defined in the bundle spec, since any key:value mappings can be defined in
// an annotations file.
type Annotations map[string]string

// GetValue returns a value for key from File and true if an entry exists.
// If not, an empty string and false are returned.
func (a File) GetValue(key Key) (string, bool) {
	if a.Annotations == nil {
		return "", false
	}
	value, hasKey := a.Annotations[string(key)]
	return value, hasKey
}

type Key string

type MediaType struct {
	Annotations File
}
