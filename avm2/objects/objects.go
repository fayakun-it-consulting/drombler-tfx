package objects

type Value interface{}

type NamespaceKind int

const (
	NS_Public NamespaceKind = iota
	NS_Private
	NS_Protected
	NS_Explicit
)

type Namespace struct {
	Name string
	Kind NamespaceKind
}

type TraitType int

const (
	Trait_Slot TraitType = iota
	Trait_Method
	Trait_Getter
	Trait_Setter
	Trait_Class
)

type Trait struct {
	Name      string
	Namespace *Namespace
	Type      TraitType
	Method    *ASFunction
	SlotID    int
}

type ASFunction struct {
	Name   string
	Code   []byte
	Max    int
	Locals int
	Scope  []*ASObject
	Native func(vm interface{}, this *ASObject, args ...Value) Value
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
	return &ASObject{Traits: make(map[string]Value), Prototype: proto, Class: class}
}

func (o *ASObject) SetProperty(name string, v Value) { o.Traits[name] = v }
func (o *ASObject) GetProperty(name string) (Value, bool) {
	if v, ok := o.Traits[name]; ok { return v, true }
	if o.Prototype != nil { return o.Prototype.GetProperty(name) }
	return nil, false
}
