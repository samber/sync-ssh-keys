package sources

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	logger "sync-ssh-keys/logger"

	"github.com/thoas/go-funk"
)

type LocalSource struct {
	path string
}

func NewLocalSource(path string) *LocalSource {
	return &LocalSource{
		path: path,
	}
}

func (l LocalSource) GetName() string {
	return "Local"
}

func (l LocalSource) CheckInputErrors() string {
	_, err := os.Stat(l.path)
	if os.IsNotExist(err) {
		return fmt.Sprintf("File does not exist: %s\n", l.path)
	}
	return ""
}

func (l LocalSource) GetKeys() []string {
	b, err := ioutil.ReadFile(l.path)
	if err != nil {
		logger.Warning(err, fmt.Sprintf("Failed to read file: %s\n", l.path))
		return []string{}
	}

	keys := strings.Split(string(b), "\n")
	return funk.Filter(keys, func(key string) bool {
		return len(key) > 0
	}).([]string)
}
