package modbus

type VariableType string

const (
	VariableTypeString  VariableType = "string"
	VariableTypeInt8    VariableType = "int8"
	VariableTypeInt16   VariableType = "int16"
	VariableTypeInt32   VariableType = "int32"
	VariableTypeInt64   VariableType = "int64"
	VariableTypeUint8   VariableType = "uint8"
	VariableTypeUint16  VariableType = "uint16"
	VariableTypeUint32  VariableType = "uint32"
	VariableTypeUint64  VariableType = "uint64"
	VariableTypeFloat32 VariableType = "float32"
	VariableTypeFloat64 VariableType = "float64"
	VariableTypeBool    VariableType = "bool"
)

type Var interface {
	GetValue() interface{}
}
type Variable[T string | int8 | int16 | int32 | uint8 | uint16 | uint32 | int64 | uint64 | float32 | float64 | bool] struct {
	Type          VariableType
	LE            bool
	ByteSwap      bool
	StringReverse bool
	Offset        int
	Length        int
	Buffer        *[]byte
	Value         T
}

func (v *Variable[T]) GetValue() T {
	switch v.Type {
	case VariableTypeString:

	}
	return v.Value
}
