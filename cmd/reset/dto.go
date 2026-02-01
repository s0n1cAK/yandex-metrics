package main

import "go/types"

type tmpEnum struct {
	PackageName string
	Structs     []tmpStruct
}

type tmpStruct struct {
	Name   string
	Fields []tmpStructFiled
}

type tmpStructFiled struct {
	Name string
	Type types.Type
}

func (f tmpStructFiled) TypeString() string {
	return types.TypeString(f.Type, nil)
}

func (f tmpStructFiled) IsPointer() bool {
	_, ok := f.Type.(*types.Pointer)
	return ok
}

func (f tmpStructFiled) IsPointerBasic() bool {
	pointer, ok := f.Type.(*types.Pointer)
	if !ok {
		return false
	}
	_, ok = pointer.Elem().(*types.Basic)
	return ok
}

func (f tmpStructFiled) IsSlice() bool {
	_, ok := f.Type.(*types.Slice)
	return ok
}

func (f tmpStructFiled) IsMap() bool {
	_, ok := f.Type.(*types.Map)
	return ok
}

func (f tmpStructFiled) IsBasic() bool {
	_, ok := f.Type.(*types.Basic)
	return ok
}

func (f tmpStructFiled) HasResetMethod() bool {
	var checkType = f.Type

	if pointer, ok := f.Type.(*types.Pointer); ok {
		checkType = pointer.Elem()
	}

	named, ok := checkType.(*types.Named)
	if !ok {
		return false
	}

	for i := 0; i < named.NumMethods(); i++ {
		method := named.Method(i)
		if method.Name() == "Reset" {
			sig := method.Type().(*types.Signature)
			if sig.Params().Len() == 0 && sig.Results().Len() == 0 {
				return true
			}
		}
	}

	return false
}

func (f tmpStructFiled) ZeroValue() string {
	switch t := f.Type.(type) {
	case *types.Basic:
		return f.zeroBasicValue(t)
	case *types.Pointer:
		switch elem := t.Elem().(type) {
		case *types.Basic:
			return f.zeroBasicValue(elem)
		default:
			return "nil"
		}
	default:
		return types.TypeString(f.Type, nil) + "{}"
	}
}

func (f tmpStructFiled) zeroBasicValue(t *types.Basic) string {
	switch t.Kind() {
	case types.Bool:
		return "false"
	case types.String:
		return `""`
	case types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
		types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64,
		types.Uintptr, types.Float32, types.Float64, types.Complex64, types.Complex128:
		return "0"
	default:
		return "nil"
	}
}
