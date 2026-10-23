# PotluckDB for Go

This repository holds the Go modules of PotluckDB 0.8.0-dev.1 (SQLite
3.54.1). PotluckDB's release writes it; it is never edited by hand. Its
source is the `go/` folder of
[potluckdb/potluckdb](https://github.com/potluckdb/potluckdb/tree/main/go).

**To use it, see [sqlite3/README.md](sqlite3/README.md).** In short: import
`github.com/potluckdb/potluckdb-go/sqlite3` instead of
`github.com/mattn/go-sqlite3`.

| Module | What |
| --- | --- |
| `github.com/potluckdb/potluckdb-go/sqlite3` | The database/sql driver, a drop-in replacement for go-sqlite3. The one to import. |
| `github.com/potluckdb/potluckdb-go/lib/linux-arm64-gnu` | The SQLite library for Linux arm64, glibc 2.28 or later |
| `github.com/potluckdb/potluckdb-go/lib/linux-arm64-musl` | The SQLite library for Linux arm64, musl (Alpine) |
| `github.com/potluckdb/potluckdb-go/lib/linux-x64-gnu` | The SQLite library for Linux x64, glibc 2.28 or later |
| `github.com/potluckdb/potluckdb-go/lib/linux-x64-musl` | The SQLite library for Linux x64, musl (Alpine) |
| `github.com/potluckdb/potluckdb-go/lib/macos-arm64` | The SQLite library for macOS 11 or later, Apple silicon |
| `github.com/potluckdb/potluckdb-go/lib/macos-x64` | The SQLite library for macOS 11 or later, Intel |

Each module is tagged `<folder>/v0.8.0-dev.1`, for example
`sqlite3/v0.8.0-dev.1`. The driver needs every library module of its own
version; Go downloads only the one for the platform it builds for.

Not in this version:

- windows-arm64: potluckdb-sqlite-0.8.0-dev.1-windows-arm64.zip has no static library that cgo can link (lib/static/libsqlite3.a)
- windows-x64: potluckdb-sqlite-0.8.0-dev.1-windows-x64.zip has no static library that cgo can link (lib/static/libsqlite3.a)

## Licences

- `sqlite3/`: Apache-2.0 (`LICENSE`). Parts come from
  go-sqlite3, under the MIT licence (`THIRD_PARTY_NOTICES.md`).
- `lib/*/`: the PotluckDB library, under the PotluckDB binary licence
  (`LICENSE.txt` in each). The SQLite source code compiled into it is in
  the public domain (`THIRD_PARTY_NOTICES.txt`).
