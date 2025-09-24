package objects

type ASObject struct {
    Properties map[string]interface{}
}

type ASClass struct {
    Name    string
    Methods map[string]*ASFunction
}

type ASFunction struct {
    Name string
    Code func([]interface{}) interface{}
}

func NewASObject() *ASObject {
    return &ASObject{Properties: make(map[string]interface{})}
}

func NewASClass(name string) *ASClass {
    return &ASClass{Name: name, Methods: make(map[string]*ASFunction)}
}

func NewASFunction(name string, code func([]interface{}) interface{}) *ASFunction {
    return &ASFunction{Name: name, Code: code}
}
