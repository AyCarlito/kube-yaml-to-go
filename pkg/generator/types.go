package generator

import (
	"fmt"
	"reflect"
	"regexp"
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

// outputTemplate is the template for output when verbose output is enabled.
var outputTemplate = `
package main 

import (
	%s
) 

%s`

// kubernetesTypePackage matches a package providing types for Kubernetes objects.
var kubernetesTypePackage = regexp.MustCompile(`v\d[a-zA-Z0-9]*$`)

// importWithAlias prepends a standard alias to a provided package import containing Kubernetes types.
func importWithAlias(i string) string {
	// Return the import unmutated.
	if !kubernetesTypePackage.Match([]byte(i)) {
		return fmt.Sprintf("\"%s\"", i)
	}
	packagePathParts := strings.Split(i, "/")
	alias := packagePathParts[len(packagePathParts)-2] + packagePathParts[len(packagePathParts)-1]
	return fmt.Sprintf("%s \"%s\"", alias, i)
}

// typeWithAlias replaces the package name in a type with the standard alias used when the package is imported.
// Kubernetes type files across multiple packages share the same package name e.g. v1, v1alpha1.
// Standardly, these are imported as:
//
//	"k8s.io/api/apps/v1" 					-> appsv1 "k8s.io/api/apps/v1"
//	"k8s.io/apimachinery/pkg/apis/meta/v1" 	-> metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
func typeWithAlias(v reflect.Value, path string) string {
	// Return the type unmutated.
	if !kubernetesTypePackage.Match([]byte(path)) {
		return v.Type().String()
	}
	packagePathParts := strings.Split(path, "/")
	alias := packagePathParts[len(packagePathParts)-2] + packagePathParts[len(packagePathParts)-1]
	return strings.Replace(v.Type().String(), "v1", alias, -1)
}
