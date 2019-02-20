package saga

import (
	"fmt"
	"github.com/itimofeev/go-saga/storage"
	"reflect"
)

// MarshalParam convert args into ParamData.
// This method will lookup typeName in given SEC.
func MarshalParam(sec *ExecutionCoordinator, args []interface{}) []storage.ParamData {
	p := make([]storage.ParamData, 0, len(args))
	for _, arg := range args {
		typ := sec.MustFindParamName(reflect.ValueOf(arg).Type())
		p = append(p, storage.ParamData{
			ParamType: typ,
			Data:      arg,
		})
	}
	return p
}

// UnmarshalParam convert ParamData back to parameter values to function call usage.
// This method will lookup reflect.Type in given SEC.
func UnmarshalParam(sec *ExecutionCoordinator, paramData []storage.ParamData) []reflect.Value {
	var values []reflect.Value
	for _, param := range paramData {
		ptyp := sec.MustFindParamType(param.ParamType)
		obj := reflect.New(ptyp).Interface()
		mustUnmarshal(param.Data, obj)
		objV := reflect.ValueOf(obj)
		if objV.Type().Kind() == reflect.Ptr && objV.Type() != ptyp {
			objV = objV.Elem()
		}
		values = append(values, objV)
	}
	return values
}

func mustUnmarshal(data interface{}, v interface{}) {
	fmt.Println("hello(")
}
