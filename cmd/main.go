package main

import (
	"fmt"
	"os"

	"github.com/example/go-avm2-full/avm2/abc"
	"github.com/example/go-avm2-full/avm2/natives"
	"github.com/example/go-avm2-full/avm2/objects"
	"github.com/example/go-avm2-full/avm2/vm"
)

func main() {
	// Demo 1: simple.abc
	fmt.Println("--- Demo: simple.abc ---")
	data, err := os.ReadFile("examples/simple.abc")
	if err == nil {
		af, err := abc.ParseABC(data)
		if err == nil {
			r := vm.NewVM(af)
			// bind natives for completeness
			natives.BindNativeClasses(r)
			res, err := r.RunFirstMethod()
			if err == nil { fmt.Println("Result:", res) } else { fmt.Println("VM error:", err) }
		} else { fmt.Println("Parse error:", err) }
	} else { fmt.Println("examples/simple.abc not found") }

	// Demo 2: simple_math.abc
	fmt.Println("--- Demo: simple_math.abc ---")
	data2, err := os.ReadFile("examples/simple_math.abc")
	if err == nil {
		af, err := abc.ParseABC(data2)
		if err == nil {
			r := vm.NewVM(af)
			res, err := r.RunFirstMethod()
			if err == nil { fmt.Println("Result:", res) } else { fmt.Println("VM error:", err) }
		}
	}

	// Demo 3: programmatic class + method
	fmt.Println("--- Demo: programmatic class demo ---")
	r := vm.NewVM(nil)
	natives.BindNativeClasses(r)
	// create class Person
	personClass := &objects.ASClass{Name: "Person", Traits: make(map[string]*objects.Trait)}
	// add a method 'greet' as native function
	greet := &objects.ASFunction{Name: "greet", Native: func(_ interface{}, this *objects.ASObject, args ...objects.Value) objects.Value {
		name, _ := this.GetProperty("name")
		return fmt.Sprintf("Hello, %v!", name)
	}}
	personClass.Traits["greet"] = &objects.Trait{Name: "greet", Type: objects.Trait_Method, Method: greet}
	// instantiate
	inst := objects.NewObject(personClass)
	inst.SetProperty("name", "Alice")
	// call greet
	if t, ok := inst.ResolveTrait("greet", nil); ok {
		if t.Method != nil {
			res := r.CallFunction(t.Method, nil)
			fmt.Println("Person.greet ->", res)
		}
	}
}
