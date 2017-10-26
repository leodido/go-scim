package shared

import (
	"context"
	"fmt"
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

	err = nil
	return
}

type mapping struct{}

func (mp *mapping) executeMapping(subj *Resource, sch *Schema, ctx context.Context) {

	r := reflect.ValueOf(subj)
	f := reflect.Indirect(r).FieldByName("urn:ietf:params:scim:schemas:extension:enterprise:2.0:User")
	i := f.Interface()
	m := i.(map[string]interface{})

	fmt.Println(m)

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
		mappingInstance = &mapping{}
	})
}
