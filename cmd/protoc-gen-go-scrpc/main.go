package main

import (
	"flag"
	"fmt"
	"google.golang.org/protobuf/compiler/protogen"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

// package declarations
var (
	scrpcPkg     = protogen.GoImportPath("github.com/victor-leee/scrpc")
	netPkg       = protogen.GoImportPath("net")
	protoPkg     = protogen.GoImportPath("google.golang.org/protobuf/proto")
	pluginGenPkg = protogen.GoImportPath("github.com/victor-leee/plugin/github.com/victor-leee/plugin")
	scrpcGenPkg  = protogen.GoImportPath("github.com/victor-leee/scrpc/github.com/victor-leee/scrpc")
	ioPkg        = protogen.GoImportPath("io")
	ctxPkg       = protogen.GoImportPath("context")
)

var file *protogen.File

type Config struct {
	Service string `json:"service" yaml:"service"`
}

var defaultCfg *Config

func main() {
	defaultPath := "./.scrpc.yml"
	cfg, parseErr := parseConfig(defaultPath)
	if parseErr != nil {
		panic(parseErr)
	}
	defaultCfg = cfg

	var flags flag.FlagSet
	isServer := flags.Bool("server", false, "")

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(plugin *protogen.Plugin) error {
		for _, f := range plugin.Files {
			if !f.Generate {
				continue
			}
			file = f
			if *isServer {
				generateServer(plugin, f)
				continue
			}
			generateClient(plugin, f)
		}

		return nil
	})
}

func parseConfig(file string) (*Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("parseConfig.Open:%w", err)
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("parseConfig.ReadAll:%w", err)
	}
	cfg := &Config{}
	if err = yaml.Unmarshal(b, cfg); err != nil {
		return nil, fmt.Errorf("parseConfig.Unmarshal:%w", err)
	}

	return cfg, nil
}
