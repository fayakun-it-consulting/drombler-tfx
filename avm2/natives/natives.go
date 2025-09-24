package natives

import (
	"fmt"
	"math"
	"time"

	"github.com/example/go-avm2-full/avm2/objects"
	"github.com/example/go-avm2-full/avm2/vm"
)

func BindNativeClasses(vmobj *vm.VM) {
	// Math object
	mathObj := &objects.ASObject{Traits: make(map[string]objects.Value)}
	mathObj.SetProperty("PI", math.Pi)
	// sin native
	sinFn := &objects.ASFunction{Name: "sin", Native: func(vm interface{}, this *objects.ASObject, args ...objects.Value) objects.Value {
		if len(args) == 0 { return 0.0 }
		return math.Sin(toFloat(args[0]))
	}}
	mathObj.SetProperty("sin", sinFn)
	vmobj.Globals["Math"] = mathObj

	// Date class with now
	dateClass := &objects.ASClass{Name: "Date", Traits: make(map[string]*objects.Trait)}
	nowFn := &objects.ASFunction{Name: "now", Native: func(vm interface{}, this *objects.ASObject, args ...objects.Value) objects.Value {
		return float64(time.Now().UnixNano() / 1e6)
	}}
	dateClass.Traits["now"] = &objects.Trait{Name: "now", Type: objects.Trait_Method, Method: nowFn}
	vmobj.Globals["Date"] = dateClass

	// Array class simplified
	arrayClass := &objects.ASClass{Name: "Array", Traits: make(map[string]*objects.Trait)}
	pushFn := &objects.ASFunction{Name: "push", Native: func(vm interface{}, this *objects.ASObject, args ...objects.Value) objects.Value {
		// use numeric keys
		lenv := 0
		if l, ok := this.Traits["length"].(int); ok { lenv = l }
		for _, a := range args {
			this.Traits[fmt.Sprintf("%d", lenv)] = a
			lenv++
		}
		this.Traits["length"] = lenv
		return lenv
	}}
	arrayClass.Traits["push"] = &objects.Trait{Name: "push", Type: objects.Trait_Method, Method: pushFn}
	vmobj.Globals["Array"] = arrayClass
}

func toFloat(v objects.Value) float64 {
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
