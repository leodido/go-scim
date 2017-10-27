package shared

import (
	"context"
	"reflect"
	"sync"
)

func Mapping(subj *Resource, sch *Schema, ctx context.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case error:
				err = r.(error)
			default:
				err = Error.Text("%v", r)
			}
		}
	}()

	mappingInstance.executeMapping(subj, sch, ctx)

	err = nil
	return
}

type mapping struct {
	attrs map[string]string
}

func (mp *mapping) executeMapping(subj *Resource, sch *Schema, ctx context.Context) {

	mp.traverse(reflect.ValueOf(subj.Complex), sch.ToAttribute(), ctx)
}

func (mp *mapping) throw(err error, ctx context.Context) {
	panic(err)
}

var (
	singleMapping   sync.Once
	mappingInstance *mapping
)

func init() {
	singleMapping.Do(func() {
		mappingInstance = &mapping{make(map[string]string)}
	})
}

func (mp *mapping) traverse(v reflect.Value, attr *Attribute, ctx context.Context) {
	if !v.IsValid() {
		return
	}

	switch v.Kind() {
	case reflect.Interface, reflect.Ptr:
		v = v.Elem()
	}

	switch v.Kind() {

	case reflect.Slice, reflect.Array:
		mp.addAttr(v, attr, ctx)

		if attr.ExpectsComplexArray() {
			subAttr := attr.Clone()
			subAttr.MultiValued = false
			for i := 0; i < v.Len(); i++ {
				mp.traverse(v.Index(i), subAttr, ctx)
			}
		}

	case reflect.Map:
		mp.addAttr(v, attr, ctx)
		for _, k := range v.MapKeys() {
			p, err := NewPath(k.String())
			if err != nil {
				mp.throw(err, ctx)
			}
			subAttr := attr.GetAttribute(p, true)
			if subAttr == nil {
				mp.throw(Error.NoAttribute(p.Value()), ctx)
			}
			mp.traverse(v.MapIndex(k), subAttr, ctx)
		}

	default:
		mp.addAttr(v, attr, ctx)
	}

}

func (mp *mapping) addAttr(v reflect.Value, attr *Attribute, ctx context.Context) {
	mp.attrs[attr.Name] = attr.Assist.JSONName
}
