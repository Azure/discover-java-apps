package core

import (
	"microsoft.com/azure-spring-discovery/api/v1alpha1"
	"reflect"
)

type ApplyMode int

const (
	NotSet ApplyMode = iota
	Spec
	ObjectMeta
	Status
)

type Monadic struct {
	jarFile JarFile
	process JavaProcess
	app     *v1alpha1.SpringBootApp
	err     error
}

type Monad struct {
	val       any
	err       error
	applyMode ApplyMode
	field     string
}

type StepFunc func(app *v1alpha1.SpringBootApp, process JavaProcess, jarFile JarFile) *Monad

func NewMonadic(process JavaProcess, jarFile JarFile, app *v1alpha1.SpringBootApp) *Monadic {
	return &Monadic{
		process: process,
		jarFile: jarFile,
		app:     app,
	}
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

func (m *Monadic) Then(f StepFunc) *Monadic {
	if m.err != nil {
		return m
	}

	monad := f(m.app, m.process, m.jarFile)
	if monad.err != nil {
		m.err = monad.err
	} else {
		if monad.applyMode != NotSet {
			var applyField reflect.Value
			switch monad.applyMode {
			case Spec:
				applyField = reflect.ValueOf(&m.app.Spec)
			case ObjectMeta:
				applyField = reflect.ValueOf(&m.app.ObjectMeta)
			case Status:
				applyField = reflect.ValueOf(&m.app.Status)
			}
			if applyField.Elem().FieldByName(monad.field).CanSet() {
				applyField.Elem().FieldByName(monad.field).Set(reflect.ValueOf(monad.val))
			}
		}
	}

	return m
}

func (m *Monad) OrElse(val any) *Monad {
	if m.err != nil || reflect.ValueOf(m.val).IsZero() {
		m.val = val
		m.err = nil
	}
	return m
}

func (m *Monad) Map(mapFunc func(val any) any) *Monad {
	if m.err != nil {
		return m
	}
	m.val = mapFunc(m.val)
	return m
}

func (m *Monad) Spec(field string) *Monad {
	m.applyMode = Spec
	m.field = field
	return m
}

func (m *Monad) ObjectMeta(field string) *Monad {
	m.applyMode = ObjectMeta
	m.field = field
	return m
}

func (m *Monad) Status(field string) *Monad {
	m.applyMode = Status
	m.field = field
	return m
}

func (m *Monad) NotSet(field string) *Monad {
	m.applyMode = NotSet
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
