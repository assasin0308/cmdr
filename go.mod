module github.com/hedzr/cmdr

go 1.13

//replace github.com/hedzr/cmdr-base => ../cmdr-base

// replace github.com/hedzr/log => ../log

//replace github.com/hedzr/logex => ../logex

// replace gopkg.in/hedzr/errors.v2 => ../errors

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/fsnotify/fsnotify v1.4.9
	github.com/hedzr/cmdr-base v0.1.3
	github.com/hedzr/log v0.3.9
	github.com/hedzr/logex v1.3.9
	golang.org/x/crypto v0.0.0-20201002170205-7f63de1d35b0
	gopkg.in/hedzr/errors.v2 v2.1.3
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
)
