package main

import (
    "bytes"
    "encoding/binary"
    "fmt"
    "os"
)

// ABCFile represents a parsed ABC file (simplified)
type ABCFile struct {
    Strings []string
    Methods []MethodInfo
}

// MethodInfo holds a single method body
type MethodInfo struct {
    MaxStack  uint8
    Code      []byte
}

func readU30(r *bytes.Reader) (uint32, error) {
    var result uint32
    var shift uint
    for {
        b, err := r.ReadByte()
        if err != nil {
            return 0, err
        }
        result |= uint32(b&0x7F) << shift
        if (b & 0x80) == 0 {
            break
        }
        shift += 7
    }
    return result, nil
}

func parseABC(data []byte) (*ABCFile, error) {
    r := bytes.NewReader(data)
    abc := &ABCFile{}

    // Skip minor_version, major_version
    var minor, major uint16
    binary.Read(r, binary.LittleEndian, &minor)
    binary.Read(r, binary.LittleEndian, &major)

    // Parse constant pool: just strings for now
    strCount, _ := readU30(r)
    abc.Strings = make([]string, strCount)
    for i := uint32(1); i < strCount; i++ { // index 0 is empty
        strlen, _ := readU30(r)
        buf := make([]byte, strlen)
        r.Read(buf)
        abc.Strings[i] = string(buf)
    }

    // Methods count
    methodCount, _ := readU30(r)
    abc.Methods = make([]MethodInfo, methodCount)

    // For simplicity assume each method has only code length + code
    for i := uint32(0); i < methodCount; i++ {
        codeLen, _ := readU30(r)
        code := make([]byte, codeLen)
        r.Read(code)
        abc.Methods[i] = MethodInfo{
            MaxStack:  4,
            Code:      code,
        }
    }

    return abc, nil
}

// Helper: coerce interface{} into float64 for arithmetic
func toNumber(v interface{}) float64 {
    switch t := v.(type) {
    case int:
        return float64(t)
    case float64:
        return t
    case string:
        // naive conversion
        var f float64
        fmt.Sscanf(t, "%f", &f)
        return f
    default:
        return 0
    }
}

func runMethod(abc *ABCFile, m MethodInfo) interface{} {
    stack := []interface{}{}
    pc := 0

    for pc < len(m.Code) {
        opcode := m.Code[pc]
        pc++
        switch opcode {
        case 0x24: // pushbyte
            val := m.Code[pc]
            pc++
            stack = append(stack, int(val))
        case 0x2C: // pushstring
            idx := m.Code[pc]
            pc++
            stack = append(stack, abc.Strings[idx])
        case 0x2A: // add (note: simplified handling)
            if len(stack) < 2 {
                panic("stack underflow on add")
            }
            b := stack[len(stack)-1]
            a := stack[len(stack)-2]
            stack = stack[:len(stack)-2]
            res := toNumber(a) + toNumber(b)
            // push back as float64 to preserve numeric behavior
            stack = append(stack, res)
        case 0x48: // returnvalue
            if len(stack) == 0 {
                return nil
            }
            return stack[len(stack)-1]
        default:
            panic(fmt.Sprintf("unhandled opcode 0x%X at pc %d", opcode, pc-1))
        }
    }
    return nil
}

func main() {
    data, err := os.ReadFile("simple_math.abc")
    if err != nil {
        panic(err)
    }

    abc, err := parseABC(data)
    if err != nil {
        panic(err)
    }

    result := runMethod(abc, abc.Methods[0])
    fmt.Printf("Result: %v\n", result)
}
