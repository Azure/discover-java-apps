package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Output struct {
	writer io.Writer
	format string
}

type FieldWithTag struct {
	name string
	tag  string
}

type FieldWithTags []FieldWithTag

func (f FieldWithTags) headers() []string {
	var headers []string
	for _, fwt := range f {
		if len(fwt.tag) == 0 {
			headers = append(headers, fwt.name)
		} else {
			headers = append(headers, fwt.tag)
		}
	}
	return headers
}

func (f FieldWithTags) fields() []string {
	var fields []string
	for _, fwt := range f {
		fields = append(fields, fwt.name)
	}
	return fields
}

func NewOutput(filename string, format string) (*Output, error) {
	var writer io.Writer
	var err error
	if len(filename) == 0 {
		writer = os.Stdout
	} else {
		writer, err = fileWriter(filename)
		if err != nil {
			return nil, err
		}
	}
	return &Output{writer: writer, format: format}, nil
}

func fileWriter(filename string) (io.Writer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (o *Output) Write(records any) error {
	var err error
	switch strings.ToLower(strings.TrimSpace(o.format)) {
	case "":
	case "json":
		err = o.writeJson(records, o.writer)
	case "csv":
		err = o.writCSV(records, o.writer)
	}
	return err
}

func (o *Output) writeJson(records any, writer io.Writer) error {
	b, err := json.Marshal(records)
	if err != nil {
		return err
	}

	var out bytes.Buffer
	err = json.Indent(&out, b, "", "  ")
	if err != nil {
		return err
	}

	_, err = writer.Write(out.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func (o *Output) writCSV(records any, writer io.Writer) error {
	var csvWriter = csv.NewWriter(writer)
	defer csvWriter.Flush()
	csvWriter.Comma = ','

	var content [][]string
	var fieldWithTags FieldWithTags

	var refTyp = reflect.TypeOf(records)
	refVal := reflect.ValueOf(records)
	var values []reflect.Value
	switch refTyp.Kind() {
	case reflect.Slice:
		for i := 0; i < refVal.Len(); i++ {
			if refVal.Index(i).Kind() == reflect.Ptr {
				values = append(values, refVal.Index(i).Elem())
			} else {
				values = append(values, refVal.Index(i))
			}
		}
	case reflect.Ptr:
		values = append(values, refVal.Elem())
	default:
		values = append(values, refVal)
	}

	for i := 0; i < values[0].Type().NumField(); i++ {
		field := values[0].Type().Field(i)
		fieldWithTags = append(fieldWithTags, FieldWithTag{name: field.Name, tag: field.Tag.Get("csv")})
	}
	content = append(content, fieldWithTags.headers())
	fields := fieldWithTags.fields()
	for _, v := range values {
		var row []string
		for _, field := range fields {
			value := v.FieldByName(field)
			row = append(row, toString(value))
		}
		content = append(content, row)
	}
	for _, record := range content {
		err := csvWriter.Write(record)
		if err != nil {
			return err
		}
	}

	return nil
}

func toString(v reflect.Value) string {
	switch k := v.Kind(); k {
	case reflect.Invalid:
		return "<invalid Value>"
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%.2f", v.Float())
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	}
	if v.Type().String() == "time.Time" {
		return v.Interface().(time.Time).String()
	}
	// If you call String on a reflect.Value of other type, it's better to
	// print something than to panic. Useful in debugging.
	return ""
}
