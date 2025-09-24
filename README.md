# Go AVM2 - Step 2 (Full Parser Integrated)

This project is a step in the process of porting **Mariana (C# AVM2)** into Go.

## Features implemented in this step

- Project structure mirroring Mariana:
  - `avm2/abc`: full ABC parser
  - `avm2/objects`: ActionScript objects, classes, functions
  - `avm2/vm`: VM runtime that loads ABCs
  - `avm2/astypes`: stubs for ActionScript types (Number, String, Boolean)
  - `avm2/swf`: placeholder for SWF container parsing
- Can load and parse real `.abc` files (constant pool, multinames, methods, traits, instances, classes, scripts, method bodies)
- VM integrates parsed classes and methods into runtime objects
- Simple demo runner prints summary of parsed ABC contents

## Run

```bash
go run ./cmd ./examples/simple.abc
```

This will parse the example file and print basic information.

## Next Steps

- Extend VM to execute parsed method bodies
- Implement more opcodes in the interpreter
- Implement AS3 built-in classes and types
