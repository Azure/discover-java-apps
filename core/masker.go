package core

import (
	"bufio"
	"bytes"
	gookitconfig "github.com/gookit/config/v2"
	"strings"
)

const (
	yamlSeparator = "---\n"
	Mask          = "****"
)

type Masker interface {
	Mask(content string) (string, error)
}

func NewYamlMasker() Masker {
	return yamlMasker{}
}

func NewPropsMasker() Masker {
	return propsMasker{}
}

type yamlMasker struct {
}

type propsMasker struct {
}

var yamlSplitter bufio.SplitFunc = func(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	// Find the index of the input of a newline followed by a
	// pound sign.
	if i := strings.Index(string(data), yamlSeparator); i >= 0 {
		return i + 4, data[0:i], nil
	}

	// If at end of file with data return the data
	if atEOF {
		return len(data), data, nil
	}

	return
}

func mask(cfg *gookitconfig.Config) error {
	var err error
	var f func(parent string, data map[string]interface{})
	f = func(parent string, data map[string]interface{}) {
		var newKey string
		for key, value := range data {
			if len(parent) > 0 {
				newKey = parent + string(cfg.Options().Delimiter) + key
			} else {
				newKey = key
			}
			switch value.(type) {
			case map[string]interface{}:
				f(newKey, value.(map[string]interface{}))
			default:
				err = cfg.Set(newKey, "", true)
			}

			if err != nil {
				panic(err)
			}
		}
	}

	f("", cfg.Data())
	return nil
}

func (y yamlMasker) Mask(origin string) (string, error) {
	s := bufio.NewScanner(strings.NewReader(origin))
	s.Split(yamlSplitter)

	var parts []string
	for s.Scan() {
		cfg, err := ParseYaml(s.Text())
		if err != nil {
			return "", nil
		}

		err = mask(cfg)
		if err != nil {
			return "", err
		}

		buf := new(bytes.Buffer)
		_, err = cfg.DumpTo(buf, gookitconfig.Yaml)
		if err != nil {
			return "", err
		}
		parts = append(parts, buf.String())
	}

	return strings.Join(parts, yamlSeparator), nil
}

func (p propsMasker) Mask(content string) (string, error) {
	var result []string

	m := ParseProperties(content)
	for k, _ := range m {
		result = append(result, k)
	}

	return strings.Join(result, "\n"), nil
}
