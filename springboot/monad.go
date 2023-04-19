package springboot

import (
	"context"
	"reflect"
)

type MonadConstraints interface {
	*SpringBootApp | *Runtime | *Artifact
}

type Monadic[T MonadConstraints] struct {
	t       T
	jarFile JarFile
	process JavaProcess
	slice   []*Monad
}

type Monad struct {
	val   any
	err   error
	field string
}

type StepFunc func(process JavaProcess, jarFile JarFile) *Monad

func NewMonadic[T MonadConstraints](process JavaProcess, jarFile JarFile) *Monadic[T] {
	return &Monadic[T]{t: New[T](), process: process, jarFile: jarFile}
}

func Of[T any](actual T, extras ...any) *Monad {
	var err error
	if len(extras) > 0 {
		if e, ok := extras[0].(error); ok {
			err = e
		}
	}
	return &Monad{val: actual, err: err}
}

func (m *Monadic[T]) Apply(f StepFunc) *Monadic[T] {
	monad := f(m.process, m.jarFile)
	m.slice = append(m.slice, monad)
	return m
}

func (m *Monadic[T]) Final(monad *Monad) error {
	if monad.err != nil {
		return monad.err
	}

	var applyField = reflect.ValueOf(m.t)
	if applyField.Kind() == reflect.Ptr {
		applyField = applyField.Elem()
	}
	if applyField.FieldByName(monad.field).CanSet() {
		applyField.FieldByName(monad.field).Set(reflect.ValueOf(monad.val))
	}
	return nil

}

func (m *Monadic[T]) Get() (T, error) {

	err := FromSlice(context.Background(), m.slice).ForEach(m.Final)
	if err != nil {
		return nil, err
	}

	return m.t, nil
}

func (m *Monad) Map(mapFunc func(val any) any) *Monad {
	if m.err != nil {
		return m
	}
	m.val = mapFunc(m.val)
	return m
}

func (m *Monad) Field(field string) *Monad {
	m.field = field
	return m
}

func wrap[T any, U any](strFunc func(val T) U) func(any) any {
	return func(val any) any {
		if str, ok := val.(T); ok {
			return strFunc(str)
		}

		return val
	}
}
