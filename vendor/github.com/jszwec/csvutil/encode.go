package csvutil

import (
	"encoding"
	"encoding/base64"
	"reflect"
	"strconv"
)

var (
	textMarshaler = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	csvMarshaler  = reflect.TypeOf((*Marshaler)(nil)).Elem()
)

var (
	encodeFloat32 = encodeFloatN(32)
	encodeFloat64 = encodeFloatN(64)
)

type encodeFunc func(buf []byte, v reflect.Value, omitempty bool) ([]byte, error)

func encodeString(buf []byte, v reflect.Value, omitempty bool) ([]byte, error) {
	return append(buf, v.String()...), nil
}

func encodeInt(buf []byte, v reflect.Value, omitempty bool) ([]byte, error) {
	n := v.Int()
	if n == 0 && omitempty {
		return buf, nil
	}
	return strconv.AppendInt(buf, n, 10), nil
}

func encodeUint(buf []byte, v reflect.Value, omitempty bool) ([]byte, error) {
	n := v.Uint()
	if n == 0 && omitempty {
		return buf, nil
	}
	return strconv.AppendUint(buf, n, 10), nil
}

func encodeFloatN(bits int) encodeFunc {
	return func(buf []byte, v reflect.Value, omitempty bool) ([]byte, error) {
		f := v.Float()
		if f == 0 && omitempty {
			return buf, nil
		}
		return strconv.AppendFloat(buf, f, 'G', -1, bits), nil
	}
}

func encodeBool(buf []byte, v reflect.Value, omitempty bool) ([]byte, error) {
	t := v.Bool()
	if !t && omitempty {
		return buf, nil
	}
	return strconv.AppendBool(buf, t), nil
}

func encodeInterface(buf []byte, v reflect.Value, omitempty bool) ([]byte, error) {
	if !v.IsValid() || v.IsNil() || !v.Elem().IsValid() {
		return buf, nil
	}

	v = v.Elem()
	enc, err := encodeFn(v.Type())
	if err != nil {
		return nil, err
	}
	return enc(buf, walkPtr(v), omitempty)
}

func encodePtrMarshaler(buf []byte, v reflect.Value, omitempty bool) ([]byte, error) {
	if v.CanAddr() {
		return encodeMarshaler(buf, v.Addr(), omitempty)
	}
	return buf, nil
}

func encodeTextMarshaler(buf []byte, v reflect.Value, _ bool) ([]byte, error) {
	b, err := v.Interface().(encoding.TextMarshaler).MarshalText()
	if err != nil {
		return nil, &MarshalerError{Type: v.Type(), MarshalerType: "MarshalText", Err: err}
	}
	return append(buf, b...), nil
}

func encodePtrTextMarshaler(buf []byte, v reflect.Value, omitempty bool) ([]byte, error) {
	if v.CanAddr() {
		return encodeTextMarshaler(buf, v.Addr(), omitempty)
	}
	return buf, nil
}

func encodeMarshaler(buf []byte, v reflect.Value, _ bool) ([]byte, error) {
	b, err := v.Interface().(Marshaler).MarshalCSV()
	if err != nil {
		return nil, &MarshalerError{Type: v.Type(), MarshalerType: "MarshalCSV", Err: err}
	}
	return append(buf, b...), nil
}

func encodePtr(typ reflect.Type) (encodeFunc, error) {
	next, err := encodeFn(typ.Elem())
	if err != nil {
		return nil, err
	}
	return func(buf []byte, v reflect.Value, omitempty bool) ([]byte, error) {
		return next(buf, v, omitempty)
	}, nil
}

func encodeBytes(buf []byte, v reflect.Value, _ bool) ([]byte, error) {
	data := v.Bytes()

	l := len(buf)
	buf = append(buf, make([]byte, base64.StdEncoding.EncodedLen(len(data)))...)
	base64.StdEncoding.Encode(buf[l:], data)
	return buf, nil
}

func encodeFn(typ reflect.Type) (encodeFunc, error) {
	if typ.Implements(csvMarshaler) {
		return encodeMarshaler, nil
	}
	if reflect.PtrTo(typ).Implements(csvMarshaler) {
		return encodePtrMarshaler, nil
	}

	if typ.Implements(textMarshaler) {
		return encodeTextMarshaler, nil
	}
	if reflect.PtrTo(typ).Implements(textMarshaler) {
		return encodePtrTextMarshaler, nil
	}

	switch typ.Kind() {
	case reflect.String:
		return encodeString, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return encodeInt, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return encodeUint, nil
	case reflect.Float32:
		return encodeFloat32, nil
	case reflect.Float64:
		return encodeFloat64, nil
	case reflect.Bool:
		return encodeBool, nil
	case reflect.Interface:
		return encodeInterface, nil
	case reflect.Ptr:
		return encodePtr(typ)
	case reflect.Slice:
		if typ.Elem().Kind() == reflect.Uint8 {
			return encodeBytes, nil
		}
	}

	return nil, &UnsupportedTypeError{Type: typ}
}
