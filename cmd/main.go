package main

import (
    "fmt"
    "log"
    "os"

    "go-avm2/avm2/abc"
)

func main() {
    if len(os.Args) < 2 {
        log.Fatalf("Usage: %s <abcfile>", os.Args[0])
    }
    path := os.Args[1]
    data, err := os.ReadFile(path)
    if err != nil {
        log.Fatalf("Failed to read ABC file: %v", err)
    }

    abcFile, err := abc.ParseABC(data)
    if err != nil {
        log.Fatalf("Failed to parse ABC file: %v", err)
    }

    fmt.Printf("Parsed ABC file: %d methods, %d instances, %d classes, %d scripts, %d method bodies\n",
        len(abcFile.Methods), len(abcFile.Instances), len(abcFile.Classes), len(abcFile.Scripts), len(abcFile.MethodBodies))
}
