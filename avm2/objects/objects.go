package objects

import "fmt"

type Value interface{}

// NamespaceKind indicates visibility
type NamespaceKind int

const (
	NS_Public NamespaceKind = iota
	NS_Private
	NS_Protected
	NS_Explicit
	NS_PackageInternal
)

type Namespace struct {
	Name string
	Kind NamespaceKind
}

// TraitType describes trait kind (slot, method, getter, setter)
type TraitType int

const (
	Trait_Slot TraitType = iota
	Trait_Method
	Trait_Getter
	Trait_Setter
	Trait_Class
	Trait_Function
)

type Trait struct {
	Name      string
	Namespace *Namespace
	Type      TraitType
	Method    *ASFunction
	SlotID    int
}

// ASFunction can be a bytecode function or native.
type NativeFunc func(vm interface{}, this *ASObject, args ...Value) Value

type ASFunction struct {
	Name   string
	Code   []byte
	Max    int
	Locals int
	Scope  []*ASObject
	Native NativeFunc
}

type ASClass struct {
	Name        string
	SuperClass  *ASClass
	Traits      map[string]*Trait
	Constructor *ASFunction
}

type ASObject struct {
	Traits    map[string]Value
	Prototype *ASObject
	Class     *ASClass
}

func NewObject(class *ASClass) *ASObject {
	var proto *ASObject
	if class != nil && class.SuperClass != nil {
		proto = NewObject(class.SuperClass)
	}
	return &ASObject{
		Traits:    make(map[string]Value),
		Prototype: proto,
		Class:     class,
	}
}

func (o *ASObject) SetProperty(name string, v Value) {
	o.Traits[name] = v
}

func (o *ASObject) GetOwnProperty(name string) (Value, bool) {
	v, ok := o.Traits[name]
	return v, ok
}

func (o *ASObject) ResolveTrait(name string, ns *Namespace) (*Trait, bool) {
	// Look on self
	if t, ok := o.Traits[name]; ok {
		if trait, ok2 := t.(*Trait); ok2 {
			// simplistic namespace check
			if trait.Namespace == nil || trait.Namespace.Kind == NS_Public {
				return trait, true
			}
			if ns != nil && trait.Namespace == ns {
				return trait, true
			}
		} else {
			// not a Trait stored, but a direct value slot
			return &Trait{Name: name, Type: Trait_Slot}, true
		}
	}
	// prototype chain
	if o.Prototype != nil {
		return o.Prototype.ResolveTrait(name, ns)
	}
	return nil, false
}

func (o *ASObject) GetProperty(name string) (Value, bool) {
	// direct value or via prototype
	if v, ok := o.Traits[name]; ok {
		return v, true
	}
	if o.Prototype != nil {
		return o.Prototype.GetProperty(name)
	}
	return nil, false
}

func (o *ASObject) String() string {
	if o == nil {
		return "<nil>"
	}
	if o.Class != nil {
		return fmt.Sprintf("[object %s]", o.Class.Name)
	}
	return "[object Object]"
}
