package config

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/mattermost/mattermost-server/server/v8/model"
)

var ErrConfigInvalidPath = errors.New("selected path object is not valid")

func getValue(path []string, obj interface{}) (interface{}, bool) {
	r := reflect.ValueOf(obj)
	var val reflect.Value
	if r.Kind() == reflect.Map {
		val = r.MapIndex(reflect.ValueOf(path[0]))
		if val.IsValid() {
			val = val.Elem()
		}
	} else {
		val = r.FieldByName(path[0])
	}

	if !val.IsValid() {
		return nil, false
	}

	switch {
	case len(path) == 1:
		return val.Interface(), true
	case val.Kind() == reflect.Struct:
		return getValue(path[1:], val.Interface())
	case val.Kind() == reflect.Map:
		remainingPath := strings.Join(path[1:], ".")
		mapIter := val.MapRange()
		for mapIter.Next() {
			key := mapIter.Key().String()
			if strings.HasPrefix(remainingPath, key) {
				i := strings.Count(key, ".") + 2 // number of dots + a dot on each side
				mapVal := mapIter.Value()
				// if no sub field path specified, return the object
				if len(path[i:]) == 0 {
					return mapVal.Interface(), true
				}
				data := mapVal.Interface()
				if mapVal.Kind() == reflect.Ptr {
					data = mapVal.Elem().Interface() // if value is a pointer, dereference it
				}
				// pass subpath
				return getValue(path[i:], data)
			}
		}
	}
	return nil, false
}

func setValueWithConversion(val reflect.Value, newValue interface{}) error {
	switch val.Kind() {
	case reflect.Struct:
		val.Set(reflect.ValueOf(newValue))
		return nil
	case reflect.Slice:
		if val.Type().Elem().Kind() != reflect.String {
			return errors.New("unsupported type of slice")
		}
		v := reflect.ValueOf(newValue)
		if v.Kind() != reflect.Slice {
			return errors.New("target value is of type Array and provided value is not")
		}
		val.Set(v)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		bits := val.Type().Bits()
		v, err := strconv.ParseInt(newValue.(string), 10, bits)
		if err != nil {
			return fmt.Errorf("target value is of type %v and provided value is not", val.Kind())
		}
		val.SetInt(v)
		return nil
	case reflect.Float32, reflect.Float64:
		bits := val.Type().Bits()
		v, err := strconv.ParseFloat(newValue.(string), bits)
		if err != nil {
			return fmt.Errorf("target value is of type %v and provided value is not", val.Kind())
		}
		val.SetFloat(v)
		return nil
	case reflect.String:
		val.SetString(newValue.(string))
		return nil
	case reflect.Bool:
		v, err := strconv.ParseBool(newValue.(string))
		if err != nil {
			return errors.New("target value is of type Bool and provided value is not")
		}
		val.SetBool(v)
		return nil
	default:
		return errors.New("target value type is not supported")
	}
}

func setValue(path []string, obj reflect.Value, newValue interface{}) error {
	var val reflect.Value
	switch obj.Kind() {
	case reflect.Struct:
		val = obj.FieldByName(path[0])
	case reflect.Map:
		val = obj.MapIndex(reflect.ValueOf(path[0]))
		if val.IsValid() {
			val = val.Elem()
		}
	default:
		val = obj
	}

	if val.Kind() == reflect.Invalid {
		return ErrConfigInvalidPath
	}

	if len(path) == 1 {
		if val.Kind() == reflect.Ptr {
			return setValue(path, val.Elem(), newValue)
		} else if obj.Kind() == reflect.Map {
			// since we cannot set map elements directly, we clone the value, set it, and then put it back in the map
			mapKey := reflect.ValueOf(path[0])
			subVal := obj.MapIndex(mapKey)
			if subVal.IsValid() {
				tmpVal := reflect.New(subVal.Elem().Type())
				if err := setValueWithConversion(tmpVal.Elem(), newValue); err != nil {
					return err
				}
				obj.SetMapIndex(mapKey, tmpVal)
				return nil
			}
		}
		return setValueWithConversion(val, newValue)
	}

	if val.Kind() == reflect.Struct {
		return setValue(path[1:], val, newValue)
	} else if val.Kind() == reflect.Map {
		remainingPath := strings.Join(path[1:], ".")
		mapIter := val.MapRange()
		for mapIter.Next() {
			key := mapIter.Key().String()
			if strings.HasPrefix(remainingPath, key) {
				mapVal := mapIter.Value()

				if mapVal.Kind() == reflect.Ptr {
					mapVal = mapVal.Elem() // if value is a pointer, dereference it
				}
				i := len(strings.Split(key, ".")) + 1

				if i > len(path)-1 { // leaf element
					i = 1
					mapVal = val
				}
				// pass subpath
				return setValue(path[i:], mapVal, newValue)
			}
		}
	}
	return errors.New("path object type is not supported")
}

func SetConfigValue(path []string, config *model.Config, newValue []string) error {
	if len(newValue) > 1 {
		return setValue(path, reflect.ValueOf(config).Elem(), newValue)
	}
	return setValue(path, reflect.ValueOf(config).Elem(), newValue[0])
}

func resetConfigValue(path []string, config *model.Config, newValue interface{}) error {
	nv := reflect.ValueOf(newValue)
	if nv.Kind() == reflect.Ptr {
		switch nv.Elem().Kind() {
		case reflect.Int:
			return setValue(path, reflect.ValueOf(config).Elem(), strconv.Itoa(*newValue.(*int)))
		case reflect.Bool:
			return setValue(path, reflect.ValueOf(config).Elem(), strconv.FormatBool(*newValue.(*bool)))
		default:
			return setValue(path, reflect.ValueOf(config).Elem(), *newValue.(*string))
		}
	} else {
		return setValue(path, reflect.ValueOf(config).Elem(), newValue)
	}
}
