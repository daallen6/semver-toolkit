# semver-toolkit

A small Go library for parsing and comparing Semantic Versioning
2.0.0 strings (https://semver.org), plus a thin CLI for the parts of
that work that come up in shell scripts and CI: is this string a
valid version, which of these two is newer, put this list in order.

Most of the semver packages I've run into either pull in a pile of
dependencies or don't expose a stable comparison you can sort with
directly. This one is stdlib only and small enough to read in one
sitting.

## Library

```go
import "github.com/daallen6/semver-toolkit"

v1, err := semver.Parse("1.2.3-rc.1+build.5")
v2, err := semver.Parse("1.2.3")

if semver.Less(v1, v2) {
    fmt.Println(v1, "is older than", v2)
}

fmt.Println(v1.String()) // "1.2.3-rc.1+build.5"
```

`Version` is a plain struct (`Major`, `Minor`, `Patch`, `Prerelease`,
`Build`), so it sorts easily with `sort.Slice` using `Compare` or
`Less` directly.

## CLI

Build it with `go build ./cmd/semver`.

Validate a file of versions, one per line:

```
$ semver validate versions.txt
OK      1.0.0
OK      2.1.0-beta.2
INVALID 1.0
```

Or pipe them in — every subcommand that reads versions falls back to
stdin when no files are given:

```
$ git tag | semver sort | tail -1
```

Compare two versions directly:

```
$ semver compare 1.2.0 1.10.0
<
```

Check versions against a range constraint:

```
$ semver check ">=1.0.0 <2.0.0" versions.txt
MATCH   1.5.0
NOMATCH 2.1.0
```

A range is one or more comparators (`=`, `!=`, `>`, `>=`, `<`, `<=`,
`^`, `~`) separated by spaces, all of which must match; several such
groups can be joined with `||`, any one of which may match. `^1.2.3`
allows changes that don't touch the leftmost nonzero component;
`~1.2.3` allows patch-level changes only.

Exit status is 0 when all input was valid (or, for `compare`, always
0 once both versions parse; for `check`, when every version was both
valid and matched), and 1 otherwise, so `validate` and `check` are
safe to use as CI gates.

## Status

Early. The parser and comparator follow the semver 2.0.0 spec,
including prerelease precedence, and range constraints (`^1.2.3`,
`>=1.0.0 <2.0.0`) are supported by both the library (`ParseRange`)
and the CLI (`check`).

## License

MIT, see LICENSE.
