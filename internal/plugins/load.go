package plugins

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Plugin struct {
	Name         string `yaml:"name"`
	FileLocation string `yaml:"location"`
}

func LoadPlugins(pluginDirectory string) ([]Plugin, error) {
	plugins := make([]Plugin, 1)
	files, err := ioutil.ReadDir(pluginDirectory)
	if err != nil {
		log.Fatalln(err)
	}
	for _, f := range files {
		if !f.IsDir() {
			fullFilename := filepath.Join(pluginDirectory, f.Name())
			fbuf, err := ioutil.ReadFile(fullFilename)
			print(string(fbuf))
			if err != nil {
				log.Fatalln(err)
			}
			p := Plugin{}
			err = yaml.Unmarshal(fbuf, &p)
			if err != nil {
				log.Fatalln(err)
			}
			plugins = append(plugins, p)
		}
	}
	return plugins, nil
}
