package natives

import (
	"math"
	"time"

	"github.com/example/go-avm2-step3/avm2/objects"
	"github.com/example/go-avm2-step3/avm2/vm"
)

func BindNativeClasses(vmobj *vm.VM) {
	// Math object
	mathObj := &objects.ASObject{Traits: make(map[string]objects.Value)}
	mathObj.SetProperty("PI", math.Pi)
	sinFn := &objects.ASFunction{Name: "sin", Native: func(vm interface{}, this *objects.ASObject, args ...objects.Value) objects.Value {
		if len(args) == 0 { return 0.0 }
		switch t := args[0].(type) {
		case float64:
			return math.Sin(t)
		case int:
			return math.Sin(float64(t))
		default:
			return 0.0
		}
	}}
	mathObj.SetProperty("sin", sinFn)
	vmobj.Globals["Math"] = mathObj

	// Date class
	dateClass := &objects.ASClass{Name: "Date", Traits: make(map[string]*objects.Trait)}
	nowFn := &objects.ASFunction{Name: "now", Native: func(vm interface{}, this *objects.ASObject, args ...objects.Value) objects.Value {
		return float64(time.Now().UnixNano() / 1e6)
	}}
	dateClass.Traits["now"] = &objects.Trait{Name: "now", Type: objects.Trait_Method, Method: nowFn}
	vmobj.Globals["Date"] = dateClass
}
