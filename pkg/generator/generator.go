package generator

import (
	"bytes"
	"context"
	"fmt"
	"go/format"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"

	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	"github.com/AyCarlito/kube-yaml-to-go/pkg/logger"
)

// Generator decodes Kubernetes objects from YAML and generates equivalent Go source code.
type Generator struct {
	ctx      context.Context
	scheme   *runtime.Scheme
	out      bytes.Buffer
	packages map[string]struct{}

	inputFilePath  string
	outputFilePath string
	verbose        bool
}

// NewGenerator returns a new *Generator.
func NewGenerator(ctx context.Context, inputFilePath, outputFilePath string, verbose bool) *Generator {
	return &Generator{
		ctx:      ctx,
		scheme:   runtime.NewScheme(),
		packages: make(map[string]struct{}),

		inputFilePath:  inputFilePath,
		outputFilePath: outputFilePath,
		verbose:        verbose,
	}
}

// addToScheme registers types with the runtime.Scheme, allowing serializing and deserializing of objects.
func (g *Generator) addToScheme() error {
	err := clientgoscheme.AddToScheme(g.scheme)
	if err != nil {
		return fmt.Errorf("failed to add client-go scheme to runtime scheme: %v", err)
	}

	err = monitoring.AddToScheme(g.scheme)
	if err != nil {
		return fmt.Errorf("failed to add monitoring scheme to runtime scheme: %v", err)
	}

	return nil
}

// deserialize deserializes the provided YAML document using the types registered with the scheme.
func (g *Generator) deserialize(document string) (runtime.Object, *schema.GroupVersionKind, error) {
	return serializer.NewCodecFactory(g.scheme).UniversalDeserializer().Decode([]byte(document), nil, nil)
}

// Generate decodes Kubernetes objects from YAML and generates equivalent Go source code.
func (g *Generator) Generate() error {
	log := logger.LoggerFromContext(g.ctx)
	log.Info("Starting generation")

	err := g.addToScheme()
	if err != nil {
		return fmt.Errorf("failed to add to scheme: %v", err)
	}

	f := os.Stdin
	if g.inputFilePath != "" {
		f, err = os.Open(g.inputFilePath)
		if err != nil {
			return fmt.Errorf("failed to open input file: %v", err)
		}
		defer f.Close()
	}

	fileBytes, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read input: %v", err)
	}

	documents := strings.Split(string(fileBytes), yamlDocumentDelimiter)
	for i, document := range documents {
		// Skip invalid documents.
		if document == "" {
			log.Info("Skipping invalid document")
			continue
		}
		obj, kind, err := g.deserialize(document)
		if err != nil {
			return fmt.Errorf("failed to deserialize document: %v", err)
		}
		log.Info("Generating source for: " + kind.String())

		// In verbose output, each converted document is stored as a variable.
		if g.verbose {
			g.out.Write([]byte(fmt.Sprintf("var %sDocumentIndex%d = ", kind.Kind, i)))
		}
		g.dump(reflect.ValueOf(obj))
		g.out.Write(newLineBytes)
	}

	f = os.Stdout
	if g.outputFilePath != "" {
		f, err = os.Create(g.outputFilePath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %v", err)
		}
		defer f.Close()
	}
	outputBytes := g.out.Bytes()

	// In verbose output, add the imports used across the types.
	if g.verbose {
		var imports []string
		for k := range g.packages {
			if k != "" {
				imports = append(imports, k)
			}
		}
		outputBytes = []byte(fmt.Sprintf(outputTemplate, strings.Join(imports, "\n"), string(outputBytes)))
	}

	// Run "gofmt" on the generated source. This accepts full and partial source files.
	formattedBytes, err := format.Source(outputBytes)
	if err != nil {
		return fmt.Errorf("failed to format generated source code: %v", err)
	}
	f.Write(formattedBytes)

	return nil
}

// dump recursively dumps the provided reflect.Value.
func (g *Generator) dump(v reflect.Value) {
	kind := v.Kind()
	if kind == reflect.Invalid {
		return
	}

	switch kind {
	case reflect.Ptr:
		g.out.Write(addressBytes)
		if v.Elem().Type().Kind() == reflect.Struct {
			g.dump(v.Elem())
			return
		}
		g.out.Write(openSquareBracketBytes)
		g.out.Write(closeSquareBracketBytes)
		g.out.Write([]byte(v.Elem().Type().String()))
		g.out.Write(openBraceBytes)
		g.dump(v.Elem())
		g.out.Write(closeBraceBytes)
		g.out.Write(zeroIndexBytes)

	case reflect.String:
		g.out.Write([]byte(strconv.Quote(v.String())))

	case reflect.Bool:
		g.out.Write([]byte(strconv.FormatBool(v.Bool())))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		g.out.Write([]byte(strconv.FormatInt(v.Int(), 10)))

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		g.out.Write([]byte(strconv.FormatUint(v.Uint(), 10)))

	case reflect.Float32:
		g.out.Write([]byte(strconv.FormatFloat(v.Float(), 'g', -1, 32)))

	case reflect.Float64:
		g.out.Write([]byte(strconv.FormatFloat(v.Float(), 'g', -1, 64)))

	case reflect.Slice:
		g.packages[importWithAlias(v.Type().Elem().PkgPath())] = struct{}{}
		g.out.Write([]byte(typeWithAlias(v, v.Type().Elem().PkgPath())))
		g.out.Write(openBraceBytes)
		g.out.Write(newLineBytes)
		numItems := v.Len()
		for i := range numItems {
			g.dump(v.Index(i))
			g.out.Write(commaBytes)
			g.out.Write(newLineBytes)
		}
		g.out.Write(newLineBytes)
		g.out.Write(closeBraceBytes)

	case reflect.Map:
		g.packages[importWithAlias(v.Type().PkgPath())] = struct{}{}
		g.out.Write([]byte(typeWithAlias(v, v.Type().PkgPath())))
		g.out.Write(openBraceBytes)
		g.out.Write(newLineBytes)
		keys := v.MapKeys()
		for _, key := range keys {
			g.dump(key)
			g.out.Write(colonBytes)
			g.dump(v.MapIndex(key))
			g.out.Write(commaBytes)
			g.out.Write(newLineBytes)
		}
		g.out.Write(newLineBytes)
		g.out.Write(closeBraceBytes)

	case reflect.Struct:
		g.packages[importWithAlias(v.Type().PkgPath())] = struct{}{}
		g.out.Write([]byte(typeWithAlias(v, v.Type().PkgPath())))
		g.out.Write(openBraceBytes)
		g.out.Write(newLineBytes)
		numFields := v.NumField()
		for i := range numFields {
			// Skip unexported fields.
			if !v.Type().Field(i).IsExported() {
				continue
			}

			// Skip zero fields.
			if v.Field(i).IsZero() {
				continue
			}

			g.out.Write([]byte(v.Type().Field(i).Name))
			g.out.Write(colonBytes)
			g.dump(v.Field(i))
			g.out.Write(commaBytes)
			g.out.Write(newLineBytes)
		}
		g.out.Write(newLineBytes)
		g.out.Write(closeBraceBytes)
	}

}
