# V2 Domain Package

`internal/core/domain` owns value types and invariants that are independent of persistence, transport protocols, collectors, and presentation adapters.

Rules:

- Constructors and parsers enforce invariants before values enter application commands.
- Struct fields remain opaque where an unconstrained Go zero value would be ambiguous.
- Text and JSON decoding are strict and do not silently accept unknown values.
- This package must not import GORM, SQLite drivers, MCP/JSON-RPC packages, CLI packages, source SDKs, or adapter packages.
- Evidence payload schemas are introduced in Slice 1.3; persistence remains deferred to later slices.
