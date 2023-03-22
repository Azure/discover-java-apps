package core

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"unicode/utf8"

	v1alpha1 "microsoft.com/azure-spring-discovery/api/v1alpha1"
)

type Output struct {
	fileName string
	format   string
}

func NewOutput(fileName string, format string) *Output {

	return &Output{
		fileName: fileName,
		format:   format,
	}
}

func (o *Output) Write(apps []v1alpha1.SpringBootAppSpec) error {
	if len(o.format) == 0 || o.format == "json" {
		return o.WriteJson(apps)
	} else if o.format == "csv" {
		return o.WritCSV(apps)
	}
	return nil
}
func (o *Output) WriteJson(records []v1alpha1.SpringBootAppSpec) error {
	b, err := json.Marshal(records)
	if err != nil {
		panic(err)
	}

	var out bytes.Buffer
	err = json.Indent(&out, b, "", "  ")
	if err != nil {
		panic(err)
	}
	if len(o.fileName) != 0 {
		file, err := os.OpenFile(o.fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		defer file.Close()
		file.WriteString(out.String())
	} else {
		fmt.Println(out.String())
	}
	return nil
}
func (o *Output) WritCSV(records []v1alpha1.SpringBootAppSpec) error {
	var writer *csv.Writer
	if len(o.fileName) != 0 {
		file, err := os.OpenFile(o.fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		defer file.Close()
		writer = csv.NewWriter(file)
	} else {
		writer = csv.NewWriter(os.Stdout)
	}
	delim, _ := utf8.DecodeRuneInString(",")
	writer.Comma = delim

	if len(records) == 0 {
		return nil
	}
	var header []string

	t := reflect.TypeOf(records[0])
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if canString(field.Type) {
			header = append(header, field.Name)
		}
	}
	writer.Write(header)
	for _, app := range records {
		var row []string
		for _, key := range header {
			value := reflect.ValueOf(app).FieldByName(key)
			row = append(row, toString(value))
		}
		writer.Write(row)
	}

	writer.Flush()
	return nil
}

func toString(v reflect.Value) string {
	switch k := v.Kind(); k {
	case reflect.Invalid:
		return "<invalid Value>"
	case reflect.String:
		return v.String()
	case reflect.Int:
	case reflect.Int8:
	case reflect.Int16:
	case reflect.Int32:
	case reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	}
	// If you call String on a reflect.Value of other type, it's better to
	// print something than to panic. Useful in debugging.
	return ""
}

func canString(t reflect.Type) bool {
	switch k := t.Kind(); k {
	case reflect.Invalid:
		return false
	case reflect.String:
		return true
	case reflect.Int:
	case reflect.Int8:
	case reflect.Int16:
	case reflect.Int32:
	case reflect.Int64:
		return true
	}
	// If you call String on a reflect.Value of other type, it's better to
	// print something than to panic. Useful in debugging.
	return false
}
