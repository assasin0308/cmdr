/*
 * Copyright © 2019 Hedzr Yeh.
 */

package cmdr

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

// GetUsedConfigFile returns the main config filename (generally it's `<appname>.yml`)
func GetUsedConfigFile() string {
	return usedConfigFile
}

// GetUsedConfigSubDir returns the sub-directory `conf.d` of config files
func GetUsedConfigSubDir() string {
	return usedConfigSubDir
}

var rwlCfgReload = new(sync.RWMutex)

// AddOnConfigLoadedListener add an functor on config loaded and merged
func AddOnConfigLoadedListener(c ConfigReloaded) {
	defer rwlCfgReload.Unlock()
	rwlCfgReload.Lock()

	// rwlCfgReload.RLock()
	if _, ok := onConfigReloadedFunctions[c]; ok {
		// rwlCfgReload.RUnlock()
		return
	}

	// rwlCfgReload.RUnlock()
	// rwlCfgReload.Lock()

	// defer rwlCfgReload.Unlock()

	onConfigReloadedFunctions[c] = true
}

// RemoveOnConfigLoadedListener remove an functor on config loaded and merged
func RemoveOnConfigLoadedListener(c ConfigReloaded) {
	defer rwlCfgReload.Unlock()
	rwlCfgReload.Lock()
	delete(onConfigReloadedFunctions, c)
}

// SetOnConfigLoadedListener enable/disable an functor on config loaded and merged
func SetOnConfigLoadedListener(c ConfigReloaded, enabled bool) {
	defer rwlCfgReload.Unlock()
	rwlCfgReload.Lock()
	onConfigReloadedFunctions[c] = enabled
}

// LoadConfigFile Load a yaml config file and merge the settings into `rxxtOptions`
// and load files in the `conf.d` child directory too.
func LoadConfigFile(file string) (err error) {
	return rxxtOptions.LoadConfigFile(file)
}

// LoadConfigFile Load a yaml config file and merge the settings into `rxxtOptions`
// and load files in the `conf.d` child directory too.
func (s *Options) LoadConfigFile(file string) (err error) {
	if !FileExists(file) {
		// log.Warnf("%v NOT EXISTS. PWD=%v", file, GetCurrentDir())
		return // not error, just ignore loading
	}

	if err = s.loadConfigFile(file); err != nil {
		return
	}

	usedConfigFile = file
	dir := path.Dir(usedConfigFile)
	_ = os.Setenv("CFG_DIR", dir)

	usedConfigSubDir = path.Join(dir, "conf.d")
	if !FileExists(usedConfigSubDir) {
		usedConfigSubDir = ""
		return
	}

	err = filepath.Walk(usedConfigSubDir, s.visit)

	if err == nil {
		s.watchConfigDir(usedConfigSubDir)
	}
	// log.Fatalf("ERROR: filepath.Walk() returned %v\n", err)

	return
}

// Load a yaml config file and merge the settings into `Options`
func (s *Options) loadConfigFile(file string) (err error) {
	var (
		b  []byte
		m  map[string]interface{}
		mm map[string]map[string]interface{}
	)

	b, _ = ioutil.ReadFile(file)

	m = make(map[string]interface{})
	switch path.Ext(file) {
	case ".toml", ".ini", ".conf", "toml":
		mm = make(map[string]map[string]interface{})
		err = toml.Unmarshal(b, &mm)
		if err == nil {
			err = s.loopMapMap("", mm)
		}
		if err != nil {
			return
		}
		return

	case ".json", "json":
		err = json.Unmarshal(b, &m)
	default:
		err = yaml.Unmarshal(b, &m)
	}

	if err == nil {
		err = s.loopMap("", m)
	}
	if err != nil {
		return
	}

	return
}

func (s *Options) mergeConfigFile(fr io.Reader, ext string) (err error) {
	var (
		m   map[string]interface{}
		buf *bytes.Buffer
	)

	buf = new(bytes.Buffer)
	_, err = buf.ReadFrom(fr)

	m = make(map[string]interface{})
	switch ext {
	case ".toml", ".ini", ".conf", "toml":
		err = toml.Unmarshal(buf.Bytes(), &m)
	case ".json", "json":
		err = json.Unmarshal(buf.Bytes(), &m)
	default:
		err = yaml.Unmarshal(buf.Bytes(), &m)
	}

	if err == nil {
		err = s.loopMap("", m)
	}
	if err != nil {
		return
	}

	return
}

func (s *Options) visit(path string, f os.FileInfo, e error) (err error) {
	// fmt.Printf("Visited: %s, e: %v\n", path, e)
	if f != nil && !f.IsDir() && e == nil {
		// log.Infof("    path: %v, ext: %v", path, filepath.Ext(path))
		ext := filepath.Ext(path)
		switch ext {
		case ".yml", ".yaml", ".json", ".toml", ".ini", ".conf": // , "yml", "yaml":
			var file *os.File
			file, err = os.Open(path)
			// if err != nil {
			// log.Warnf("ERROR: os.Open() returned %v", err)
			// } else {
			if err == nil {
				defer file.Close()
				if err = s.mergeConfigFile(bufio.NewReader(file), ext); err != nil {
					err = fmt.Errorf("Error in merging config file '%s': %v", path, err)
					return
				}
				configFiles = append(configFiles, path)
				// env := viper.Get("app.registrar.env")
				// key := fmt.Sprintf("app.registrar.consul.%s.addr", env)
				// log.Infof("%s = %s", key, viper.Get(key))
			} else {
				err = fmt.Errorf("Error in merging config file '%s': %v", path, err)
			}
		}
	} else {
		err = e
	}
	return
}

func (s *Options) reloadConfig(e fsnotify.Event) {
	// log.Debugf("\n\nConfig file changed: %s\n", e.String())

	defer rwlCfgReload.RUnlock()
	rwlCfgReload.RLock()

	for x, ok := range onConfigReloadedFunctions {
		if ok {
			x.OnConfigReloaded()
		}
	}
}

func (s *Options) watchConfigDir(configDir string) {
	initWG := sync.WaitGroup{}
	initWG.Add(1)
	go func() {
		watcher, err := fsnotify.NewWatcher()
		if err == nil {
			defer watcher.Close()

			eventsWG := &sync.WaitGroup{}
			eventsWG.Add(1)
			go s.watchRunner(configDir, watcher, eventsWG)
			_ = watcher.Add(configDir)
			initWG.Done()   // done initializing the watch in this go routine, so the parent routine can move on...
			eventsWG.Wait() // now, wait for event loop to end in this go-routine...
		}
	}()
	initWG.Wait() // make sure that the go routine above fully ended before returning
}

func (s *Options) watchRunner(configDir string, watcher *fsnotify.Watcher, eventsWG *sync.WaitGroup) {
	defer func() {
		eventsWG.Done()
	}()
	for {
		select {
		case event, ok := <-watcher.Events:
			// ok == false: 'Events' channel is closed
			if ok {
				// log.Debugf("ooo | fsnotify.watcher |%v", event.String())
				// currentConfigFile, _ := filepath.EvalSymlinks(filename)
				// we only care about the config file with the following cases:
				// 1 - if the config file was modified or created
				// 2 - if the real path to the config file changed (eg: k8s ConfigMap replacement)
				const writeOrCreateMask = fsnotify.Write | fsnotify.Create
				if strings.HasPrefix(filepath.Clean(event.Name), configDir) &&
					event.Op&writeOrCreateMask != 0 &&
					(testCfgSuffix(event.Name)) {
					file, err := os.Open(event.Name)
					if err != nil {
						log.Printf("ERROR: os.Open() returned %v\n", err)
					} else {
						err = s.mergeConfigFile(bufio.NewReader(file), path.Ext(event.Name))
						if err != nil {
							log.Printf("ERROR: os.Open() returned %v\n", err)
						}
						s.reloadConfig(event)
						file.Close()
					}
				}
			}

		case err, ok := <-watcher.Errors:
			if ok { // 'Errors' channel is not closed
				// log.Printf("watcher error: %v\n", err)
				log.Printf("Watcher error: %v\n", err)
			}
			return
		}
	}
}

func testCfgSuffix(name string) bool {
	for _, suf := range []string{".yaml", ".yml", ".json", ".toml", ".ini", ".conf"} {
		if strings.HasSuffix(name, suf) {
			return true
		}
	}
	return false
}
