package interfaces

import "io"

type CacheManager interface {
	Load() map[string][]string
	Save(data map[string][]string) error
	Add(filename, categoryPath string)
	Clear(categoryPath string)
}

type Prompter interface {
	ReadLine() (string, error)
	ReadLineLower() (string, error)
	ReadLineLowerDefault(defaultValue string) (string, error)
}

type Writer interface {
	io.Writer
}
