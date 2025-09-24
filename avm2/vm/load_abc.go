package vm

import (
	"fmt"

	"github.com/example/go-avm2-step3/avm2/abc"
	"github.com/example/go-avm2-step3/avm2/objects"
)

// LoadABC maps parsed ABC structures into VM runtime objects (ASClass, ASFunction)
func (vm *VM) LoadABC(af *abc.ABCFile) error {
	if af == nil {
		return fmt.Errorf("nil ABCFile")
	}
	vm.ABC = af

	// Create classes for each instance (using string names if available)
	for i, inst := range af.Instances {
		name := vm.resolveStringIndex(inst.NameIndex)
		cls := &objects.ASClass{Name: name, Traits: make(map[string]*objects.Trait)}
		// attach placeholder constructor; real constructor may be in Classes/MethodBodies
		vm.Classes = append(vm.Classes, cls)
		vm.Globals[name] = cls
		_ = i
	}

	// Link superclasses (best-effort)
	for i := range af.Instances {
		inst := af.Instances[i]
		if inst.SuperNameIndex != 0 && int(inst.SuperNameIndex) < len(af.CP.Strings) {
			superName := vm.resolveStringIndex(inst.SuperNameIndex)
			// find class
			for _, c := range vm.Classes {
				if c.Name == superName {
					vm.Classes[i].SuperClass = c
					break
				}
			}
		}
	}

	// Map method bodies to ASFunction objects
	for _, mb := range af.MethodBodies {
		fn := &objects.ASFunction{Name: fmt.Sprintf("method_%d", mb.Method), Code: mb.Code, Max: int(mb.MaxStack), Locals: int(mb.LocalCount)}
		vm.Methods = append(vm.Methods, fn)
	}

	// Attach methods to traits based on TraitInfo in instances & classes
	for idx, inst := range af.Instances {
		for _, tr := range inst.Traits {
			name := vm.resolveStringIndex(tr.NameIndex)
			if tr.Tag == abc.Trait_Method || tr.Tag == abc.Trait_Getter || tr.Tag == abc.Trait_Setter {
				methodIdx := tr.Method
				// find function by Method index
				var fn *objects.ASFunction
				for _, f := range vm.Methods {
					// compare by name pattern "method_X" where X==methodIdx
					if f.Name == fmt.Sprintf("method_%d", methodIdx) {
						fn = f; break
					}
				}
				if fn != nil && idx < len(vm.Classes) {
					vm.Classes[idx].Traits[name] = &objects.Trait{Name: name, Type: objects.Trait_Method, Method: fn}
				}
			} else if tr.Tag == abc.Trait_Slot {
				// default slot handling
				if idx < len(vm.Classes) {
					vm.Classes[idx].Traits[name] = &objects.Trait{Name: name, Type: objects.Trait_Slot, SlotID: int(tr.SlotID)}
				}
			}
		}
	}

	return nil
}
