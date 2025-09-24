package vm

import (
	"fmt"

	"github.com/example/go-avm2-step3/avm2/abc"
	"github.com/example/go-avm2-step3/avm2/objects"
)

func (vm *VM) LoadABC(af *abc.ABCFile) error {
	if af == nil { return fmt.Errorf("nil ABCFile") }
	vm.ABC = af

	// Create classes for instances
	vm.Classes = make([]*objects.ASClass, len(af.Instances))
	for i, inst := range af.Instances {
		name := vm.resolveStringIndex(inst.NameIndex)
		if name == "" { name = fmt.Sprintf("class_%d", i) }
		cls := &objects.ASClass{Name: name, Traits: make(map[string]*objects.Trait)}
		vm.Classes[i] = cls
		vm.Globals[name] = cls
	}

	// Link superclasses by name
	for i, inst := range af.Instances {
		if inst.SuperNameIndex != 0 {
			superName := vm.resolveStringIndex(inst.SuperNameIndex)
			for _, c := range vm.Classes {
				if c.Name == superName {
					vm.Classes[i].SuperClass = c
					break
				}
			}
		}
	}

	// Map method bodies to ASFunction
	vm.Methods = make([]*objects.ASFunction, len(af.MethodBodies))
	for i, mb := range af.MethodBodies {
		fn := &objects.ASFunction{
			Name: fmt.Sprintf("method_%d", mb.Method),
			Code: mb.Code,
			Max:  int(mb.MaxStack),
			Locals: int(mb.LocalCount),
		}
		vm.Methods[i] = fn
	}

	// Attach traits from instances (instance-side)
	for ci, inst := range af.Instances {
		cls := vm.Classes[ci]
		for _, tr := range inst.Traits {
			name := vm.resolveStringIndex(tr.NameIndex)
			switch tr.Tag {
			case abc.Trait_Method, abc.Trait_Getter, abc.Trait_Setter:
				// find method body by index
				var fn *objects.ASFunction
				for _, m := range vm.Methods {
					// method index stored in Trait.Method refers to method_info index; match by name pattern
					if m.Name == fmt.Sprintf("method_%d", tr.Method) {
						fn = m; break
					}
				}
				if fn != nil {
					cls.Traits[name] = &objects.Trait{Name: name, Type: objects.Trait_Method, Method: fn}
				}
			case abc.Trait_Slot:
				cls.Traits[name] = &objects.Trait{Name: name, Type: objects.Trait_Slot, SlotID: int(tr.SlotID)}
			default:
				// ignore other trait types for now
			}
		}
	}

	// Attach traits from class side (static)
	for ci, cl := range af.Classes {
		cls := vm.Classes[ci]
		for _, tr := range cl.Traits {
			name := vm.resolveStringIndex(tr.NameIndex)
			switch tr.Tag {
			case abc.Trait_Method:
				var fn *objects.ASFunction
				for _, m := range vm.Methods {
					if m.Name == fmt.Sprintf("method_%d", tr.Method) {
						fn = m; break
					}
				}
				if fn != nil {
					cls.Traits[name] = &objects.Trait{Name: name, Type: objects.Trait_Method, Method: fn}
				}
			case abc.Trait_Class:
				// class slot: attach referenced class
				if int(tr.ClassIndex) < len(vm.Classes) {
					ref := vm.Classes[tr.ClassIndex]
					cls.Traits[name] = &objects.Trait{Name: name, Type: objects.Trait_Class}
					// store class as property on Globals too
					vm.Globals[ref.Name] = ref
				}
			}
		}
	}

	return nil
}
