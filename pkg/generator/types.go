package generator

import (
	"reflect"
	"strings"
)

const (
	// yamlDocumentDelimiter defines the delimiter on which YAML documents should be split.
	// The use of the standard "---" delimiter is avoided to prevent invalid documents if an RSA private key is present.
	yamlDocumentDelimiter string = "\n---\n"
)

var (
	addressBytes            = []byte(`&`)
	openSquareBracketBytes  = []byte(`[`)
	closeSquareBracketBytes = []byte(`]`)
	openBraceBytes          = []byte(`{`)
	closeBraceBytes         = []byte(`}`)
	zeroIndexBytes          = []byte(`[0]`)
	commaBytes              = []byte(`,`)
	colonBytes              = []byte(`:`)
	newLineBytes            = []byte("\n")
)

// aliasedType is an alias for a type.
type aliasedType = string

// newAliasedType returns a new aliasedType.
// Kubernetes type files across multiple packages share the same package name e.g. v1, v1alpha1.
// Types in the dumped output must reflect how these packages are typically aliased when imported:
//
//	"k8s.io/api/apps/v1" 					-> appsv1 "k8s.io/api/apps/v1"
//	"k8s.io/apimachinery/pkg/apis/meta/v1" 	-> metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
func newAliasedType(v reflect.Value, path string) aliasedType {
	if !strings.Contains(path, "k8s.io") {
		return v.Type().String()
	}
	packagePathParts := strings.Split(path, "/")
	alias := packagePathParts[len(packagePathParts)-2] + packagePathParts[len(packagePathParts)-1]
	return strings.Replace(v.Type().String(), "v1", alias, -1)
}
