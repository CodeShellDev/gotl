package configutils

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var configLock sync.Mutex

var DELIM string = "."

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

// Template Config with environment variables
func (config *Config) TemplateConfig() error {
	data := config.Layer.All()

	for key, value := range data {
		str, isStr := value.(string)

		if isStr {
			templated := os.ExpandEnv(str)

			if templated != "" {
				data[key] = templated
			}
		}
	}

	return config.Load(data, "")
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