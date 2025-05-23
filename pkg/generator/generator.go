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
	ctx           context.Context
	inputFilePath string
	scheme        *runtime.Scheme
}

// NewGenerator returns a new *Generator.
func NewGenerator(ctx context.Context, inputFilePath string) *Generator {
	return &Generator{
		ctx:           ctx,
		inputFilePath: inputFilePath,
		scheme:        runtime.NewScheme(),
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
		return fmt.Errorf("failed to add scheduling scheme to runtime scheme: %v", err)
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
	for _, document := range documents {
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

		var buf bytes.Buffer
		g.dump(&buf, reflect.ValueOf(obj))
		formattedBytes, err := format.Source(buf.Bytes())
		if err != nil {
			return fmt.Errorf("failed to format generated source code: %v", err)
		}

		// TODO: Need the option to write to file (or several files)
		fmt.Println(string(formattedBytes))
	}
	return nil
}

// dump recursively dumps the provided reflect.Value.
func (g *Generator) dump(w io.Writer, v reflect.Value) {
	kind := v.Kind()
	if kind == reflect.Invalid {
		return
	}

	switch kind {
	case reflect.Ptr:
		w.Write(addressBytes)
		if v.Elem().Type().Kind() == reflect.Struct {
			g.dump(w, v.Elem())
			return
		}
		w.Write(openSquareBracketBytes)
		w.Write(closeSquareBracketBytes)
		w.Write([]byte(v.Elem().Type().String()))
		w.Write(openBraceBytes)
		g.dump(w, v.Elem())
		w.Write(closeBraceBytes)
		w.Write(zeroIndexBytes)

	case reflect.String:
		w.Write([]byte(strconv.Quote(v.String())))

	case reflect.Bool:
		w.Write([]byte(strconv.FormatBool(v.Bool())))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		w.Write([]byte(strconv.FormatInt(v.Int(), 10)))

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		w.Write([]byte(strconv.FormatUint(v.Uint(), 10)))

	case reflect.Float32:
		w.Write([]byte(strconv.FormatFloat(v.Float(), 'g', -1, 32)))

	case reflect.Float64:
		w.Write([]byte(strconv.FormatFloat(v.Float(), 'g', -1, 64)))

	case reflect.Slice:
		w.Write([]byte(newAliasedType(v, v.Type().Elem().PkgPath())))
		w.Write(openBraceBytes)
		w.Write(newLineBytes)
		numItems := v.Len()
		for i := range numItems {
			g.dump(w, v.Index(i))
			w.Write(commaBytes)
			w.Write(newLineBytes)
		}
		w.Write(newLineBytes)
		w.Write(closeBraceBytes)

	case reflect.Map:
		w.Write([]byte(newAliasedType(v, v.Type().PkgPath())))
		w.Write(openBraceBytes)
		w.Write(newLineBytes)
		keys := v.MapKeys()
		for _, key := range keys {
			g.dump(w, key)
			w.Write(colonBytes)
			g.dump(w, v.MapIndex(key))
			w.Write(commaBytes)
			w.Write(newLineBytes)
		}
		w.Write(newLineBytes)
		w.Write(closeBraceBytes)

	case reflect.Struct:
		w.Write([]byte(newAliasedType(v, v.Type().PkgPath())))
		w.Write(openBraceBytes)
		w.Write(newLineBytes)
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

			w.Write([]byte(v.Type().Field(i).Name))
			w.Write(colonBytes)
			g.dump(w, v.Field(i))
			w.Write(commaBytes)
			w.Write(newLineBytes)
		}
		w.Write(newLineBytes)
		w.Write(closeBraceBytes)
	}

}
