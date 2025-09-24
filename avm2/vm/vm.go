package vm

import (
	"github.com/example/go-avm2-step3/avm2/abc"
	"github.com/example/go-avm2-step3/avm2/objects"
)

type VM struct {
	ABC     *abc.ABCFile
	Stack   []objects.Value
	Globals map[string]interface{}
	Classes []*objects.ASClass
	Methods []*objects.ASFunction
}

func NewVM() *VM {
	return &VM{Globals: make(map[string]interface{}), Stack: make([]objects.Value, 0)}
}

func (vm *VM) resolveStringIndex(idx uint32) string {
	if vm.ABC == nil { return "" }
	if idx == 0 { return "" }
	if int(idx) < len(vm.ABC.CP.Strings) {
		return vm.ABC.CP.Strings[idx]
	}
	return ""
}

// CallFunction executes an ASFunction; supports Native funcs; bytecode execution not implemented yet.
func (vm *VM) CallFunction(fn *objects.ASFunction, args ...objects.Value) objects.Value {
	if fn == nil { return nil }
	if fn.Native != nil {
		return fn.Native(vm, nil, args...)
	}
	// Bytecode execution is not implemented in this step.
	return nil
}
