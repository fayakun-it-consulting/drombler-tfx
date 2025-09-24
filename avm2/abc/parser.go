package abc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

type ABCFile struct {
	MinorVersion uint16
	MajorVersion uint16
	CP           ConstantPool
	Methods      []MethodInfo
	Meta         []MetadataInfo
	Instances    []InstanceInfo
	Classes      []ClassInfo
	Scripts      []ScriptInfo
	MethodBodies []MethodBody
}

type ConstantPool struct {
	Ints       []int32
	Uints      []uint32
	Doubles    []float64
	Strings    []string
	Namespaces []*Namespace
	NSets      [][]*Namespace
	Multinames []Multiname
}

type NamespaceKind uint8

const (
	NS_Public NamespaceKind = iota + 1
	NS_Protected
	NS_Explicit
	NS_StaticProtected
	NS_Private = 5
)

type Namespace struct {
	Kind      NamespaceKind
	NameIndex uint32
}

type MultinameKind uint8

type Multiname struct {
	Kind              MultinameKind
	NameIndex         uint32
	NamespaceIndex    uint32
	NamespaceSetIndex uint32
}

type MethodInfo struct {
	ParamCount    uint32
	ReturnType    uint32
	ParamTypes    []uint32
	NameIndex     uint32
	Flags         uint8
	OptionalCount uint32
	OptionalKinds []uint8
	OptionalVals  []uint32
	ParamNames    []uint32
}

type MetadataInfo struct {
	NameIndex uint32
}

type TraitKind uint8

const (
	Trait_Slot TraitKind = iota
	Trait_Method
	Trait_Getter
	Trait_Setter
	Trait_Class
	Trait_Function
)

type TraitInfo struct {
	NameIndex uint32
	Kind      uint8
	Tag       TraitKind

	SlotID   uint32
	SlotType uint32
	VIndex   uint32
	VKind    uint8

	DispID uint32
	Method uint32

	ClassSlot    uint32
	ClassIndex   uint32
	FunctionSlot uint32
	FunctionIndex uint32

	Metadata []uint32
}

type InstanceInfo struct {
	NameIndex      uint32
	SuperNameIndex uint32
	Flags          uint8
	ProtectedNS    uint32
	InterfaceCount uint32
	Interfaces     []uint32
	Initializer    uint32
	Traits         []TraitInfo
}

type ClassInfo struct {
	Initializer uint32
	Traits      []TraitInfo
}

type ScriptInfo struct {
	InitMethod uint32
	Traits     []TraitInfo
}

type ExceptionInfo struct {
	From    uint32
	To      uint32
	Target  uint32
	TypeName uint32
	VName   uint32
}

type MethodBody struct {
	Method         uint32
	MaxStack       uint32
	LocalCount     uint32
	InitScopeDepth uint32
	Code           []byte
	Exceptions     []ExceptionInfo
	Traits         []TraitInfo
}

// helpers
func readU8(r *bytes.Reader) (uint8, error) {
	b, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	return b, nil
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
		if shift > 35 {
			return 0, fmt.Errorf("u30 overflow")
		}
	}
	return result, nil
}

func readS32(r *bytes.Reader) (int32, error) {
	u, err := readU30(r)
	if err != nil {
		return 0, err
	}
	return int32(u), nil
}

func readString(r *bytes.Reader) (string, error) {
	length, err := readU30(r)
	if err != nil {
		return "", err
	}
	if length == 0 {
		return "", nil
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

func readDouble(r *bytes.Reader) (float64, error) {
	var bits uint64
	if err := binary.Read(r, binary.LittleEndian, &bits); err != nil {
		return 0, err
	}
	return math.Float64frombits(bits), nil
}

func ParseABC(data []byte) (*ABCFile, error) {
	r := bytes.NewReader(data)
	var minor, major uint16
	if err := binary.Read(r, binary.LittleEndian, &minor); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &major); err != nil {
		return nil, err
	}
	af := &ABCFile{MinorVersion: minor, MajorVersion: major}

	intCount, err := readU30(r)
	if err != nil { return nil, err }
	if intCount == 0 { intCount = 1 }
	af.CP.Ints = make([]int32, intCount)
	for i := uint32(1); i < intCount; i++ {
		v, err := readS32(r)
		if err != nil { return nil, err }
		af.CP.Ints[i] = v
	}

	uintCount, err := readU30(r)
	if err != nil { return nil, err }
	if uintCount == 0 { uintCount = 1 }
	af.CP.Uints = make([]uint32, uintCount)
	for i := uint32(1); i < uintCount; i++ {
		v, err := readU30(r)
		if err != nil { return nil, err }
		af.CP.Uints[i] = v
	}

	doubleCount, err := readU30(r)
	if err != nil { return nil, err }
	if doubleCount == 0 { doubleCount = 1 }
	af.CP.Doubles = make([]float64, doubleCount)
	for i := uint32(1); i < doubleCount; i++ {
		f, err := readDouble(r)
		if err != nil { return nil, err }
		af.CP.Doubles[i] = f
	}

	strCount, err := readU30(r)
	if err != nil { return nil, err }
	if strCount == 0 { strCount = 1 }
	af.CP.Strings = make([]string, strCount)
	for i := uint32(1); i < strCount; i++ {
		s, err := readString(r)
		if err != nil { return nil, err }
		af.CP.Strings[i] = s
	}

	nsCount, err := readU30(r)
	if err != nil { return nil, err }
	if nsCount == 0 { nsCount = 1 }
	af.CP.Namespaces = make([]*Namespace, nsCount)
	for i := uint32(1); i < nsCount; i++ {
		kind, err := readU8(r)
		if err != nil { return nil, err }
		nameIdx, err := readU30(r)
		if err != nil { return nil, err }
		af.CP.Namespaces[i] = &Namespace{Kind: NamespaceKind(kind), NameIndex: nameIdx}
	}

	nssetCount, err := readU30(r)
	if err != nil { return nil, err }
	if nssetCount == 0 { nssetCount = 1 }
	af.CP.NSets = make([][]*Namespace, nssetCount)
	for i := uint32(1); i < nssetCount; i++ {
		count, err := readU30(r)
		if err != nil { return nil, err }
		set := make([]*Namespace, count)
		for j := uint32(0); j < count; j++ {
			nsIdx, err := readU30(r)
			if err != nil { return nil, err }
			if nsIdx < uint32(len(af.CP.Namespaces)) {
				set[j] = af.CP.Namespaces[nsIdx]
			} else {
				set[j] = nil
			}
		}
		af.CP.NSets[i] = set
	}

	mnCount, err := readU30(r)
	if err != nil { return nil, err }
	if mnCount == 0 { mnCount = 1 }
	af.CP.Multinames = make([]Multiname, mnCount)
	for i := uint32(1); i < mnCount; i++ {
		kindByte, err := readU8(r)
		if err != nil { return nil, err }
		kind := MultinameKind(kindByte)
		mn := Multiname{Kind: kind}
		switch kind {
		case 0x07, 0x0D:
			nIdx, err := readU30(r); if err != nil { return nil, err }
			nsIdx, err := readU30(r); if err != nil { return nil, err }
			mn.NameIndex = nIdx; mn.NamespaceIndex = nsIdx
		case 0x0F, 0x10:
		case 0x09, 0x0E:
			nameIdx, err := readU30(r); if err != nil { return nil, err }
			nsSetIdx, err := readU30(r); if err != nil { return nil, err }
			mn.NameIndex = nameIdx; mn.NamespaceSetIndex = nsSetIdx
		case 0x1B, 0x1C, 0x1D:
		default:
		}
		af.CP.Multinames[i] = mn
	}

	methodCount, err := readU30(r); if err != nil { return nil, err }
	af.Methods = make([]MethodInfo, methodCount)
	for i := uint32(0); i < methodCount; i++ {
		paramCount, err := readU30(r); if err != nil { return nil, err }
		retType, err := readU30(r); if err != nil { return nil, err }
		paramTypes := make([]uint32, paramCount)
		for j := uint32(0); j < paramCount; j++ {
			pt, err := readU30(r); if err != nil { return nil, err }
			paramTypes[j] = pt
		}
		nameIdx, err := readU30(r); if err != nil { return nil, err }
		flags, err := readU8(r); if err != nil { return nil, err }
		mi := MethodInfo{ParamCount: paramCount, ReturnType: retType, ParamTypes: paramTypes, NameIndex: nameIdx, Flags: flags}
		if flags&0x08 != 0 {
			optCount, err := readU30(r); if err != nil { return nil, err }
			mi.OptionalCount = optCount
			mi.OptionalKinds = make([]uint8, optCount)
			mi.OptionalVals = make([]uint32, optCount)
			for k := uint32(0); k < optCount; k++ {
				val, err := readU30(r); if err != nil { return nil, err }
				kind, err := readU8(r); if err != nil { return nil, err }
				mi.OptionalVals[k] = val; mi.OptionalKinds[k] = kind
			}
		}
		if flags&0x80 != 0 {
			names := make([]uint32, paramCount)
			for k := uint32(0); k < paramCount; k++ {
				n, err := readU30(r); if err != nil { return nil, err }
				names[k] = n
			}
			mi.ParamNames = names
		}
		af.Methods[i] = mi
	}

	metaCount, err := readU30(r); if err != nil { return nil, err }
	af.Meta = make([]MetadataInfo, metaCount)
	for i := uint32(0); i < metaCount; i++ {
		nameIdx, err := readU30(r); if err != nil { return nil, err }
		kcount, err := readU30(r); if err != nil { return nil, err }
		for j := uint32(0); j < kcount; j++ {
			_, err := readU30(r); if err != nil { return nil, err }
			_, err = readU30(r); if err != nil { return nil, err }
		}
		af.Meta[i] = MetadataInfo{NameIndex: nameIdx}
	}

	instCount, err := readU30(r); if err != nil { return nil, err }
	af.Instances = make([]InstanceInfo, instCount)
	for i := uint32(0); i < instCount; i++ {
		nameIdx, err := readU30(r); if err != nil { return nil, err }
		superIdx, err := readU30(r); if err != nil { return nil, err }
		flags, err := readU8(r); if err != nil { return nil, err }
		inst := InstanceInfo{NameIndex: nameIdx, SuperNameIndex: superIdx, Flags: flags}
		if flags&0x08 != 0 {
			pns, err := readU30(r); if err != nil { return nil, err }
			inst.ProtectedNS = pns
		}
		ifCount, err := readU30(r); if err != nil { return nil, err }
		inst.InterfaceCount = ifCount
		inst.Interfaces = make([]uint32, ifCount)
		for j := uint32(0); j < ifCount; j++ {
			ii, err := readU30(r); if err != nil { return nil, err }
			inst.Interfaces[j] = ii
		}
		initIdx, err := readU30(r); if err != nil { return nil, err }
		inst.Initializer = initIdx
		tcount, err := readU30(r); if err != nil { return nil, err }
		inst.Traits = make([]TraitInfo, tcount)
		for t := uint32(0); t < tcount; t++ {
			tr, err := parseTrait(r); if err != nil { return nil, err }
			inst.Traits[t] = tr
		}
		af.Instances[i] = inst
	}

	classCount := instCount
	af.Classes = make([]ClassInfo, classCount)
	for i := uint32(0); i < classCount; i++ {
		cinit, err := readU30(r); if err != nil { return nil, err }
		tcount, err := readU30(r); if err != nil { return nil, err }
		traits := make([]TraitInfo, tcount)
		for t := uint32(0); t < tcount; t++ {
			tr, err := parseTrait(r); if err != nil { return nil, err }
			traits[t] = tr
		}
		af.Classes[i] = ClassInfo{Initializer: cinit, Traits: traits}
	}

	scriptCount, err := readU30(r); if err != nil { return nil, err }
	af.Scripts = make([]ScriptInfo, scriptCount)
	for i := uint32(0); i < scriptCount; i++ {
		initIdx, err := readU30(r); if err != nil { return nil, err }
		tcount, err := readU30(r); if err != nil { return nil, err }
		traits := make([]TraitInfo, tcount)
		for t := uint32(0); t < tcount; t++ {
			tr, err := parseTrait(r); if err != nil { return nil, err }
			traits[t] = tr
		}
		af.Scripts[i] = ScriptInfo{InitMethod: initIdx, Traits: traits}
	}

	bodyCount, err := readU30(r); if err != nil { return nil, err }
	af.MethodBodies = make([]MethodBody, bodyCount)
	for i := uint32(0); i < bodyCount; i++ {
		methodIdx, err := readU30(r); if err != nil { return nil, err }
		maxStack, err := readU30(r); if err != nil { return nil, err }
		localCount, err := readU30(r); if err != nil { return nil, err }
		initScopeDepth, err := readU30(r); if err != nil { return nil, err }
		codeLen, err := readU30(r); if err != nil { return nil, err }
		code := make([]byte, codeLen)
		if codeLen > 0 {
			if _, err := io.ReadFull(r, code); err != nil { return nil, err }
		}
		exCount, err := readU30(r); if err != nil { return nil, err }
		exs := make([]ExceptionInfo, exCount)
		for e := uint32(0); e < exCount; e++ {
			from, err := readU30(r); if err != nil { return nil, err }
			to, err := readU30(r); if err != nil { return nil, err }
			target, err := readU30(r); if err != nil { return nil, err }
			typeName, err := readU30(r); if err != nil { return nil, err }
			vname, err := readU30(r); if err != nil { return nil, err }
			exs[e] = ExceptionInfo{From: from, To: to, Target: target, TypeName: typeName, VName: vname}
		}
		tbCount, err := readU30(r); if err != nil { return nil, err }
		tlist := make([]TraitInfo, tbCount)
		for t := uint32(0); t < tbCount; t++ {
			tr, err := parseTrait(r); if err != nil { return nil, err }
			tlist[t] = tr
		}
		af.MethodBodies[i] = MethodBody{Method: methodIdx, MaxStack: maxStack, LocalCount: localCount, InitScopeDepth: initScopeDepth, Code: code, Exceptions: exs, Traits: tlist}
	}

	return af, nil
}

func parseTrait(r *bytes.Reader) (TraitInfo, error) {
	var tr TraitInfo
	nameIdx, err := readU30(r)
	if err != nil { return tr, err }
	k, err := readU8(r)
	if err != nil { return tr, err }
	tr.NameIndex = nameIdx
	tr.Kind = k
	tag := TraitKind(k & 0x0F)
	tr.Tag = tag
	switch tag {
	case Trait_Slot:
		slotID, err := readU30(r)
		if err != nil { return tr, err }
		typeName, err := readU30(r)
		if err != nil { return tr, err }
		vindex, err := readU30(r)
		if err != nil { return tr, err }
		var vkind uint8
		if vindex != 0 {
			vkind, err = readU8(r)
			if err != nil { return tr, err }
		}
		tr.SlotID = slotID; tr.SlotType = typeName; tr.VIndex = vindex; tr.VKind = vkind
	case Trait_Method, Trait_Getter, Trait_Setter:
		dispID, err := readU30(r); if err != nil { return tr, err }
		methodIdx, err := readU30(r); if err != nil { return tr, err }
		tr.DispID = dispID; tr.Method = methodIdx
	case Trait_Class:
		slotID, err := readU30(r); if err != nil { return tr, err }
		classIdx, err := readU30(r); if err != nil { return tr, err }
		tr.ClassSlot = slotID; tr.ClassIndex = classIdx
	case Trait_Function:
		slotID, err := readU30(r); if err != nil { return tr, err }
		funcIdx, err := readU30(r); if err != nil { return tr, err }
		tr.FunctionSlot = slotID; tr.FunctionIndex = funcIdx
	default:
	}
	if (tr.Kind & 0x40) != 0 {
		mdCount, err := readU30(r)
		if err != nil { return tr, err }
		tr.Metadata = make([]uint32, mdCount)
		for i := uint32(0); i < mdCount; i++ {
			v, err := readU30(r)
			if err != nil { return tr, err }
			tr.Metadata[i] = v
		}
	}
	return tr, nil
}
