package vm

import (
    "go-avm2/avm2/objects"
)

type VM struct {
    Globals map[string]*objects.ASObject
}

func NewVM() *VM {
    return &VM{Globals: make(map[string]*objects.ASObject)}
}

func (vm *VM) LoadABC(name string) {
    cls := objects.NewASClass(name)
    vm.Globals[name] = &objects.ASObject{Properties: map[string]interface{}{"class": cls}}
}
