# GoAVM2 Full (Transpiled features)

This repository contains a Go implementation of an AVM2-like runtime, ported conceptually from the Mariana C# project and extended with a pure-Go interpreter design.

**This release includes the features implemented during our porting work so far.** It is intended to be a working, extendable codebase — not a complete spec implementation of AVM2. Use it as a foundation for further porting and tests.

## Key implemented features in this release

- ABC parsing (partial): reads header, string constant pool, and simple method bodies (code bytes).
- VM interpreter loop: supports many AVM2-like opcodes in a simplified form:
  - Stack ops: pushbyte, pushstring, pushint (as byte), pop
  - Arithmetic: add
  - Flow: returnvalue
  - Property ops: getproperty, setproperty, initproperty, getlex
  - Function & class ops: newfunction, newclass (create class object), constructprop (instantiate class), callprop (call method with args)
  - Scope ops: pushscope, popscope
  - Loop ops: hasnext, hasnext2 (simple property iteration)
- Object model:
  - ASObject with dynamic properties and prototype chain
  - ASClass with Traits and Constructor (ASFunction)
  - ASFunction representing bytecode-backed or native functions (supports Native implementations)
  - Trait and Namespace types (simplified resolution rules)
- Native classes & methods: `Math` (sin), basic `Array` push behavior (simplified), Date.now native.
- SWF DoABC extractor for FWS/CWS (zlib) files (extracts DoABC tags and returns abc bytes).
- Examples and demos:
  - `examples/simple.abc` — toy ABC returning "Hello, AVM2!"
  - `examples/simple_math.abc` — toy ABC computing 5+7
  - `cmd/main.go` demonstrates loading ABCs, running methods, and creating/calling a simple class created programmatically.
- Project layout follows Mariana-like structure adapted to Go packages:
  - `avm2/abc`, `avm2/vm`, `avm2/objects`, `avm2/natives`, `avm2/swf`, `cmd`

## Quick start

```bash
go run ./cmd
```

This will execute a few demos:
- Run `examples/simple.abc` and print a string.
- Run `examples/simple_math.abc` and print numeric addition result.
- Create a `Person` class programmatically, instantiate, set property and call a method implemented in Go to demonstrate class behavior.

## Notes & Roadmap

- The ABC parser is intentionally small; full ABC (traits, multinames, namespaces, exception tables) is planned.
- Opcode semantics are simplified; consult Mariana for precise AVM2 semantics when expanding.
- The SWF loader supports FWS and CWS; ZWS (LZMA) is not implemented.
- Future work: full ABC parsing, trait import from real ABC files, method linking, more native methods, tests converting Mariana tests to Go.

---
Licensed MIT (adaptation of porting work)
