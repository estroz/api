package validator

import "fmt"

// ManifestResult represents verification result for each of the yaml files
// from the operator manifest.
type ManifestResult struct {
	// Name is some piece of information identifying the manifest.
	Name string
	// Errors pertain to issues with the manifest that must be corrected.
	Errors []Error
	// Warnings pertain to issues with the manifest that are optional to correct.
	Warnings []Error
}

func (r *ManifestResult) Add(errs ...Error) {
	for _, err := range errs {
		if err.Level == LevelError {
			r.Errors = append(r.Errors, err)
		} else {
			r.Warnings = append(r.Warnings, err)
		}
	}
}

func (r ManifestResult) HasError() bool {
	return len(r.Errors) == 0
}

func (r ManifestResult) HasWarn() bool {
	return len(r.Warnings) == 0
}

// Error is an implementation of the 'error' interface, which represents a
// warning or an error in a yaml file. Error type is taken as is from
// https://github.com/operator-framework/operator-registry/blob/master/vendor/k8s.io/apimachinery/pkg/util/validation/field/errors.go#L31
// to maintain compatibility with upstream.
type Error struct {
	// Type is the ErrorType string constant that represents the kind of
	// error, ex. "MandatoryStructMissing", "I/O".
	Type ErrorType
	// Level is the severity of the Error.
	Level Level
	// Field is the dot-hierarchical YAML path of the missing data.
	Field string
	// BadValue is the field or file that caused an error or warning.
	BadValue interface{}
	// Message represents the error message as a string.
	Message string
}

// Error strut implements the 'error' interface to define custom error formatting.
func (e Error) Error() string {
	msg := e.Message
	if msg != "" {
		msg = fmt.Sprintf(": %s", msg)
	}
	if e.Field != "" && e.BadValue != nil {
		msg = fmt.Sprintf("Field %s, Value %v%s", e.Field, e.BadValue, msg)
	} else if e.Field != "" {
		msg = fmt.Sprintf("Field %s%s", e.Field, msg)
	} else if e.BadValue != nil {
		msg = fmt.Sprintf("Value %v%s", e.BadValue, msg)
	}
	if msg != "" {
		return fmt.Sprintf("%s: %s", e.Level, msg)
	}
	return "ErrMessageMissing"
}

type Level string

const (
	LevelWarn  = "Warning"
	LevelError = "Error"
)

func NewError(t ErrorType, msg, field string, v interface{}) Error {
	return Error{Level: LevelError, Type: t, Message: msg, Field: field, BadValue: v}
}

func NewWarn(t ErrorType, msg, field string, v interface{}) Error {
	return Error{Level: LevelWarn, Type: t, Message: msg, Field: field, BadValue: v}
}

type ErrorType string

func ErrInvalidBundle(msg string, value interface{}) Error {
	return invalidBundle(LevelError, msg, value)
}

func WarnInvalidBundle(msg string, value interface{}) Error {
	return invalidBundle(LevelError, msg, value)
}

func invalidBundle(lvl Level, msg string, value interface{}) Error {
	return Error{ErrorInvalidBundle, lvl, "", value, msg}
}

func ErrInvalidManifestStructure(msg string) Error {
	return invalidManifestStructure(LevelError, msg)
}

func WarnInvalidManifestStructure(msg string) Error {
	return invalidManifestStructure(LevelWarn, msg)
}

func invalidManifestStructure(lvl Level, msg string) Error {
	return Error{ErrorInvalidManifestStructure, lvl, "", "", msg}
}

func ErrInvalidCSV(msg, csvName string) Error {
	return invalidCSV(LevelError, msg, csvName)
}

func WarnInvalidCSV(msg, csvName string) Error {
	return invalidCSV(LevelWarn, msg, csvName)
}

func invalidCSV(lvl Level, msg, csvName string) Error {
	return Error{ErrorInvalidCSV, lvl, "", "", fmt.Sprintf("(%s) %s", csvName, msg)}
}

func ErrFieldMissing(msg string, field string, value interface{}) Error {
	return fieldMissing(LevelError, msg, field, value)
}

func WarnFieldMissing(msg string, field string, value interface{}) Error {
	return fieldMissing(LevelWarn, msg, field, value)
}

func fieldMissing(lvl Level, msg string, field string, value interface{}) Error {
	return Error{ErrorFieldMissing, lvl, field, value, msg}
}

func ErrUnsupportedType(msg string) Error {
	return unsupportedType(LevelError, msg)
}

func WarnUnsupportedType(msg string) Error {
	return unsupportedType(LevelWarn, msg)
}

func unsupportedType(lvl Level, msg string) Error {
	return Error{ErrorUnsupportedType, lvl, "", "", msg}
}

// TODO: see if more information can be extracted out of 'unmarshall/parsing' errors.
func ErrInvalidParse(msg string, value interface{}) Error {
	return invalidParse(LevelError, msg, value)
}

func WarnInvalidParse(msg string, value interface{}) Error {
	return invalidParse(LevelWarn, msg, value)
}

func invalidParse(lvl Level, msg string, value interface{}) Error {
	return Error{ErrorInvalidParse, lvl, "", value, msg}
}

func ErrInvalidPackageManifest(msg string, pkgName string) Error {
	return invalidPackageManifest(LevelError, msg, pkgName)
}

func WarnInvalidPackageManifest(msg string, pkgName string) Error {
	return invalidPackageManifest(LevelWarn, msg, pkgName)
}

func invalidPackageManifest(lvl Level, msg string, pkgName string) Error {
	return Error{ErrorInvalidPackageManifest, lvl, "", "", fmt.Sprintf("(%s) %s", pkgName, msg)}
}

func ErrIOError(msg string, value interface{}) Error {
	return iOError(LevelError, msg, value)
}

func WarnIOError(msg string, value interface{}) Error {
	return iOError(LevelWarn, msg, value)
}

func iOError(lvl Level, msg string, value interface{}) Error {
	return Error{ErrorIO, lvl, "", value, msg}
}

func ErrFailedValidation(msg string, value interface{}) Error {
	return failedValidation(LevelError, msg, value)
}

func WarnFailedValidation(msg string, value interface{}) Error {
	return failedValidation(LevelWarn, msg, value)
}

func failedValidation(lvl Level, msg string, value interface{}) Error {
	return Error{ErrorFailedValidation, lvl, "", value, msg}
}

func ErrInvalidOperation(msg string, value interface{}) Error {
	return invalidOperation(LevelError, msg, value)
}

func WarnInvalidOperation(msg string, value interface{}) Error {
	return invalidOperation(LevelWarn, msg, value)
}

func invalidOperation(lvl Level, msg string, value interface{}) Error {
	return Error{ErrorInvalidOperation, lvl, "", value, msg}
}

const (
	ErrorInvalidCSV               ErrorType = "CSVFileNotValid"
	ErrorFieldMissing             ErrorType = "FieldNotFound"
	ErrorUnsupportedType          ErrorType = "FieldTypeNotSupported"
	ErrorInvalidParse             ErrorType = "ParseError"
	ErrorIO                       ErrorType = "FileReadError"
	ErrorFailedValidation         ErrorType = "ValidationFailed"
	ErrorInvalidOperation         ErrorType = "OperationFailed"
	ErrorInvalidManifestStructure ErrorType = "ManifestStructureNotValid"
	ErrorInvalidBundle            ErrorType = "BundleNotValid"
	ErrorInvalidPackageManifest   ErrorType = "PackageManifestNotValid"
)
