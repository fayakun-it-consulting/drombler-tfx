package main
import ("fmt"; "os"; "github.com/example/go-avm2-step3/avm2/abc"; "github.com/example/go-avm2-step3/avm2/vm"; "github.com/example/go-avm2-step3/avm2/natives")
func main(){ if len(os.Args)<2 { fmt.Println("usage"); return }; data,_:=os.ReadFile(os.Args[1]); af,_:=abc.ParseABC(data); r:=vm.NewVM(); r.LoadABC(af); natives.BindNativeClasses(r); fmt.Println("ok") }
