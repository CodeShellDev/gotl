package configutils

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"text/template"

	t "github.com/codeshelldev/gotl/pkg/configutils/types"
	"github.com/codeshelldev/gotl/pkg/templating"
	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var configLock sync.Mutex

var DELIM string = "."

var DEFAULT_HOOKS = []mapstructure.DecodeHookFunc{t.NilSentinelHook}

type Config struct {
	Layer *koanf.Koanf
	ReloadFunc func(string)
}

// Create a New Config with Args
func NewWith(delim string, reloadFunc func(string)) *Config {
	return &Config{
		Layer: koanf.New(delim),
		ReloadFunc: reloadFunc,
	}
}

// Create a New Config
func New() *Config {
	return &Config{
		Layer: koanf.New(DELIM),
		ReloadFunc: nil,
	}
}

// Set ReloadFunc
func (config *Config) OnReload(reloadFunc func(string)) {
	config.ReloadFunc = reloadFunc
}

// Watch file with file provider
func (config *Config) WatchFile(fileProvider *file.File, path string) {
	watchFile(fileProvider, path, config.ReloadFunc)
}

// Load file with parser into Config
func (config *Config) LoadFile(path string, parser koanf.Parser) (*file.File, error) {
	f := file.Provider(path)

	err := config.Layer.Load(f, parser)
	
	if err != nil {
		return nil, err
	}

	if config.ReloadFunc != nil {
		config.WatchFile(f, path)
	}

	return f, err
}

// Load files inside of dir with parser into Config path (default: ext="")
func (config *Config) LoadDir(path string, dir string, ext string, parser koanf.Parser, transform func(*Config, string)) error {
	files, err := filepath.Glob(filepath.Join(dir, "*" + ext))

	if err != nil {
		return nil
	}

	var array []any

	for _, f := range files {
		tmp := New()

		tmp.OnReload(config.ReloadFunc)

		_, err := tmp.LoadFile(f, parser)

		if err != nil {
			return err
		}

		transform(tmp, f)

		array = append(array, tmp.Layer.Raw())
	}

	return config.Load(array, path)
}

// Load data into Config path
func (config *Config) Load(data any, path string) error {
	parts := strings.Split(path, DELIM)

	if len(parts) <= 0 {
		return errors.New("invalid path")
	}

	res := map[string]any{}

	if path == "" {
		mapData, ok := data.(map[string]any)

		if ok {
			res = mapData
		}

		parts = []string{}
	}

	for i, key := range parts {
		if key == "" {
			continue
		}

		if i == 0 {
			res[key] = data
		} else {
			sub := map[string]any{}

			sub[key] = res

			res = sub
		}
	}

	return config.Layer.Load(confmap.Provider(res, DELIM), nil)
}

// Load environment into Config with transformFunc
func (config *Config) LoadEnv(transformFunc func(key string, value string) (string, any)) (*env.Env, error) {
	e := env.Provider(DELIM, env.Opt{
		TransformFunc: transformFunc,
	})

	err := config.Layer.Load(e, nil)

	return e, err
}

// Template Config with environment + variables
func (config *Config) TemplateConfig(variables map[string]any) error {
	return config.Load(config.GetTemplated(variables), "")
}

// Alternative to TemplateConfig(), doesn't modify the config
func (config *Config) GetTemplated(variables map[string]any) any {
	data := config.Layer.All()

	envMap := environMap()

	vars := map[string]any{
		"env": envMap,
		"vars": variables,
	}

	templated, err := templating.TemplateDataRecursively("", data, vars, template.New("").Delims("${{", "}}"))

	if err != nil {
		return nil
	}

	return templated
}

// Get tag from scheme field by using a pointer of said field
func GetSchemeTagByFieldPointer(config any, tag string, fieldPointer any) string {
	v := reflect.ValueOf(config)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	fieldValue := reflect.ValueOf(fieldPointer).Elem()

	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Addr().Interface() == fieldValue.Addr().Interface() {
			field := v.Type().Field(i)

			return field.Tag.Get(tag)
		}
	}

	return ""
}

func environMap() map[string]any {
	env := os.Environ()

	m := make(map[string]any, len(env))

	for _, kv := range env {
		parts := strings.SplitN(kv, "=", 2)

		key := parts[0]

		value := ""

		if len(parts) > 1 {
			value = parts[1]
		}
		
		m[key] = value
	}

	return m
}

// Merge layers into Config
func (config *Config) MergeLayers(layers ...*koanf.Koanf) error {
	for _, layer := range layers {
		err := config.Layer.Merge(layer)

		if err != nil {
			return err
		}
	}

	return nil
}

func (config *Config) Unmarshal(path string, schema any) error {
	return config.UnmarshalWithHooks(path, schema, t.NilSentinelHook)
}

func (config *Config) UnmarshalWithHooks(path string, schema any, hooks ...mapstructure.DecodeHookFunc) error {
	return config.UnmarshalWith(path, schema, koanf.UnmarshalConf{
		DecoderConfig: &mapstructure.DecoderConfig{
			DecodeNil: true,
			DecodeHook: mapstructure.ComposeDecodeHookFunc(hooks...),
			Metadata:         nil,
			WeaklyTypedInput: true,
		},
	})
}

func (config *Config) UnmarshalWith(path string, schema any, c koanf.UnmarshalConf) error {
	return config.Layer.UnmarshalWithConf(path, schema, c)
}

func watchFile(f *file.File, path string, loadFunc func(string)) {
	f.Watch(func(event any, err error) {
		if err != nil {
			return
		}

		configLock.Lock()
		defer configLock.Unlock()

		f.Unwatch()

		loadFunc(path)
	})
}

// Walks schema and calls fn for every field with its path, schema field and raw + typed value
func WalkSchema(schema reflect.Type, value reflect.Value, raw any, path []string, fn func(path string, field reflect.StructField, raw any, value reflect.Value)) {
	if schema == nil {
		return
	}

	// unwrap interfaces and pointers
	for value.IsValid() && (value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer) {

		if value.IsNil() {
			value = reflect.Value{}
			break
		}

		value = value.Elem()
	}

	switch schema.Kind() {
	case reflect.Struct:
		rawMap, _ := raw.(map[string]any)

		for field := range schema.Fields() {
			var child reflect.Value
			if value.IsValid() && value.Kind() == reflect.Struct {
				child = value.FieldByName(field.Name)
			}

			// source of truth, no koanf tag, no actual path
			key := field.Tag.Get("koanf")

			rawChild := raw
			nextPath := path

			if key != "" {
				nextPath = append(path, key)

				if rawMap != nil {
					rawChild = rawMap[key]
				}
			} else {
				// no koanf tag
				t := field.Type
				for t.Kind() == reflect.Pointer {
					t = t.Elem()
				}

				// if field is struct => leaf field, try other
				if t.Kind() != reflect.Struct {
					continue
				}
			}

			WalkSchema(
				field.Type,
				child,
				rawChild,
				nextPath,
				fn,
			)
		}

	default:
		fn(
			joinPaths(path...),
			reflect.StructField{
				Type: schema,
			},
			raw,
			value,
		)
	}
}