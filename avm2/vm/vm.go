package vm

import (
	"errors"
	"fmt"
	"math"

	"github.com/example/go-avm2-full/avm2/abc"
	"github.com/example/go-avm2-full/avm2/objects"
)

// VM represents the runtime
type VM struct {
	ABC     *abc.ABCFile
	Stack   []objects.Value
	Frames  []*Frame
	Globals map[string]objects.Value
}

// Frame: simple execution frame
type Frame struct {
	Code   []byte
	PC     int
	Stack  []objects.Value
	Locals []objects.Value
	Scope  []*objects.ASObject
	// Exception table omitted (future)
}

func NewVM(af *abc.ABCFile) *VM {
	vm := &VM{ABC: af, Stack: make([]objects.Value, 0, 256), Globals: make(map[string]objects.Value)}
	return vm
}

func (vm *VM) push(v objects.Value) { vm.Stack = append(vm.Stack, v) }
func (vm *VM) pop() objects.Value {
	if len(vm.Stack) == 0 { return nil }
	v := vm.Stack[len(vm.Stack)-1]
	vm.Stack = vm.Stack[:len(vm.Stack)-1]
	return v
}

// Opcode constants (subset)
const (
	OP_pushbyte   = 0x24
	OP_pushstring = 0x2C
	OP_add        = 0x2A
	OP_return     = 0x48
	OP_getproperty = 0x66
	OP_setproperty = 0x67
	OP_callprop    = 0x4E
	OP_newclass    = 0x43
	OP_constructprop = 0x53
	OP_newfunction = 0x40
	OP_getlex = 0x62
	OP_initproperty = 0x46
	OP_pushscope = 0x9F
	OP_popscope = 0xA0
	OP_hasnext = 0x41
	OP_hasnext2 = 0x47
)

// helper to coerce to number
func toNumber(v objects.Value) float64 {
	switch t := v.(type) {
	case int:
		return float64(t)
	case float64:
		return t
	case string:
		var f float64
		fmt.Sscanf(t, "%f", &f)
		return f
	default:
		return 0
	}
}

// RunFirstMethod executes ABC.Methods[0]
func (vm *VM) RunFirstMethod() (objects.Value, error) {
	if vm.ABC == nil || len(vm.ABC.Methods) == 0 { return nil, errors.New("no methods") }
	code := vm.ABC.Methods[0].Code
	pc := 0
	for pc < len(code) {
		op := code[pc]
		pc++
		switch op {
		case OP_pushbyte:
			if pc >= len(code) { return nil, errors.New("pushbyte missing") }
			vm.push(int(code[pc])); pc++
		case OP_pushstring:
			if pc >= len(code) { return nil, errors.New("pushstring missing") }
			idx := int(code[pc]); pc++
			var s string
			if vm.ABC != nil && idx >=0 && idx < len(vm.ABC.Strings) { s = vm.ABC.Strings[idx] }
			vm.push(s)
		case OP_add:
			b := vm.pop(); a := vm.pop()
			vm.push(toNumber(a) + toNumber(b))
		case OP_return:
			return vm.pop(), nil
		case OP_getproperty:
			if pc >= len(code) { return nil, errors.New("getproperty missing") }
			idx := int(code[pc]); pc++
			name := ""
			if vm.ABC != nil && idx>=0 && idx < len(vm.ABC.Strings) { name = vm.ABC.Strings[idx] }
			objv := vm.pop()
			if obj, ok := objv.(*objects.ASObject); ok {
				if val, ok := obj.GetProperty(name); ok { vm.push(val); } else { vm.push(nil) }
			} else {
				vm.push(nil)
			}
		case OP_setproperty:
			if pc >= len(code) { return nil, errors.New("setproperty missing") }
			idx := int(code[pc]); pc++
			name := ""
			if vm.ABC != nil && idx>=0 && idx < len(vm.ABC.Strings) { name = vm.ABC.Strings[idx] }
			val := vm.pop()
			objv := vm.pop()
			if obj, ok := objv.(*objects.ASObject); ok {
				obj.SetProperty(name, val)
				vm.push(val)
			} else {
				vm.push(nil)
			}
		case OP_newclass:
			if pc >= len(code) { return nil, errors.New("newclass missing") }
			idx := int(code[pc]); pc++
			// simplified: create a class with name from constant pool
			name := ""
			if vm.ABC != nil && idx>=0 && idx < len(vm.ABC.Strings) { name = vm.ABC.Strings[idx] }
			cls := &objects.ASClass{Name: name, Traits: make(map[string]*objects.Trait)}
			vm.push(cls)
		case OP_constructprop:
			// pop class ref
			clsRef := vm.pop()
			cls, ok := clsRef.(*objects.ASClass)
			if !ok { vm.push(nil); break }
			// optional args: not from bytecode here; create instance
			inst := objects.NewObject(cls)
			// call constructor if exists
			if cls.Constructor != nil {
				// set 'this' in scope and call
				cls.Constructor.Scope = []*objects.ASObject{inst}
				vm.callFunction(cls.Constructor, nil)
			}
			vm.push(inst)
		case OP_callprop:
			if pc >= len(code) { return nil, errors.New("callprop missing") }
			nameIdx := int(code[pc]); pc++
			argCount := 0
			// In our simple encoding, next byte may be arg count if present
			if pc < len(code) { argCount = int(code[pc]); pc++ }
			// pop args
			args := make([]objects.Value, argCount)
			for i := argCount-1; i>=0; i-- { args[i] = vm.pop() }
			// pop object
			objv := vm.pop()
			name := ""
			if vm.ABC != nil && nameIdx>=0 && nameIdx < len(vm.ABC.Strings) { name = vm.ABC.Strings[nameIdx] }
			if obj, ok := objv.(*objects.ASObject); ok {
				if t, ok := obj.ResolveTrait(name, nil); ok {
					if t.Method != nil {
						res := vm.callFunction(t.Method, args)
						vm.push(res)
					} else {
						// slot
						if v, ok := obj.GetProperty(name); ok { vm.push(v) } else { vm.push(nil) }
					}
				} else {
					vm.push(nil)
				}
			} else {
				vm.push(nil)
			}
		case OP_newfunction:
			if pc >= len(code) { return nil, errors.New("newfunction missing") }
			idx := int(code[pc]); pc++
			// create a function using method idx if possible
			if vm.ABC != nil && idx>=0 && idx < len(vm.ABC.Methods) {
				m := vm.ABC.Methods[idx]
				fn := &objects.ASFunction{Name: fmt.Sprintf("fn%d", idx), Code: m.Code, Max: 0, Locals: 0}
				vm.push(fn)
			} else {
				vm.push(nil)
			}
		case OP_getlex:
			if pc >= len(code) { return nil, errors.New("getlex missing") }
			idx := int(code[pc]); pc++
			name := ""
			if vm.ABC != nil && idx>=0 && idx < len(vm.ABC.Strings) { name = vm.ABC.Strings[idx] }
			// check globals
			if g, ok := vm.Globals[name]; ok { vm.push(g) } else { vm.push(nil) }
		case OP_initproperty:
			if pc >= len(code) { return nil, errors.New("initproperty missing") }
			idx := int(code[pc]); pc++
			name := ""
			if vm.ABC != nil && idx>=0 && idx < len(vm.ABC.Strings) { name = vm.ABC.Strings[idx] }
			val := vm.pop()
			objv := vm.pop()
			if obj, ok := objv.(*objects.ASObject); ok { obj.SetProperty(name, val); vm.push(val) } else { vm.push(nil) }
		case OP_pushscope:
			ov := vm.pop()
			if o, ok := ov.(*objects.ASObject); ok {
				// push to frame scope (we're using global stack frame for demos)
				// not implemented per-frame here
				vm.push(o)
			} else { vm.push(nil) }
		case OP_popscope:
			// no-op in this minimal model
		case OP_hasnext, OP_hasnext2:
			// very simplified: expects object then index
			idxv := vm.pop()
			objv := vm.pop()
			index := 0
			if ii, ok := idxv.(int); ok { index = ii }
			if obj, ok := objv.(*objects.ASObject); ok {
				keys := make([]string, 0, len(obj.Traits))
				for k := range obj.Traits { keys = append(keys, k) }
				if index < len(keys) {
					vm.push(true); vm.push(keys[index]); vm.push(index+1)
				} else { vm.push(false) }
			} else { vm.push(false) }
		default:
			return nil, fmt.Errorf("unhandled opcode 0x%02X", op)
		}
	}
	return nil, nil
}

func (vm *VM) callFunction(fn *objects.ASFunction, args []objects.Value) objects.Value {
	if fn == nil { return nil }
	if fn.Native != nil {
		// For native, pass vm as interface{}, and set this to nil for simplicity
		return fn.Native(vm, nil, args...)
	}
	// For bytecode functions, create a temporary frame and execute its code
	f := &Frame{Code: fn.Code, PC: 0, Stack: make([]objects.Value,0), Locals: make([]objects.Value, fn.Locals)}
	// arguments -> locals starting at 0
	for i := 0; i < len(args) && i < len(f.Locals); i++ { f.Locals[i] = args[i] }
	// naive execution: reuse RunFrame-like loop
	for f.PC < len(f.Code) {
		op := f.Code[f.PC]; f.PC++
		switch op {
		case OP_return:
			if len(f.Stack) == 0 { return nil }
			return f.Stack[len(f.Stack)-1]
		case OP_pushbyte:
			if f.PC < len(f.Code) { f.Stack = append(f.Stack, int(f.Code[f.PC])); f.PC++ }
		case OP_pushstring:
			if f.PC < len(f.Code) { idx := int(f.Code[f.PC]); f.PC++; var s string; if vm.ABC!=nil && idx>=0 && idx < len(vm.ABC.Strings) { s = vm.ABC.Strings[idx] }; f.Stack = append(f.Stack, s) }
		case OP_add:
			b := f.Stack[len(f.Stack)-1]; f.Stack = f.Stack[:len(f.Stack)-1]
			a := f.Stack[len(f.Stack)-1]; f.Stack = f.Stack[:len(f.Stack)-1]
			f.Stack = append(f.Stack, toNumber(a)+toNumber(b))
		default:
			// unsupported in nested function
			return nil
		}
	}
	return nil
}
