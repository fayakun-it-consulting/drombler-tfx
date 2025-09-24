package main

import (
	"fmt"
	"log"
	"os"

	"github.com/example/go-avm2-step3/avm2/abc"
	"github.com/example/go-avm2-step3/avm2/natives"
	"github.com/example/go-avm2-step3/avm2/vm"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <abcfile>", os.Args[0])
	}
	path := os.Args[1]
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("read error: %v", err)
	}
	af, err := abc.ParseABC(data)
	if err != nil {
		log.Fatalf("parse error: %v", err)
	}
	runtime := vm.NewVM()
	if err := runtime.LoadABC(af); err != nil {
		log.Fatalf("LoadABC error: %v", err)
	}
	natives.BindNativeClasses(runtime)
	fmt.Printf("Loaded ABC: %d instances, %d classes, %d method bodies\n", len(af.Instances), len(af.Classes), len(af.MethodBodies))
	fmt.Println("Globals:")
	for k := range runtime.Globals {
		fmt.Println(" -", k)
	}
}
