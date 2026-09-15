# Staticcheck

A Dagger module for linting Go projects with [Staticcheck](https://staticcheck.dev/).
Built with the Dang SDK.

## Usage

Install this module in a Go project's Dagger workspace:

```sh
dagger install github.com/dagger/staticcheck
dagger check staticcheck:lint-all
```

For local development, run it against a Go workspace without installing it:

```sh
dagger -W /path/to/go-project check -m /path/to/staticcheck
```

The module discovers `go.mod` files at or below the working directory, including
the enclosing module when called from a subdirectory. It runs `staticcheck ./...`
in each selected module and fails if any module has diagnostics. Test files are
analyzed too; tests and generators are not executed.

## Settings

```toml
[modules.staticcheck]
source = "github.com/dagger/staticcheck"
settings.version = "v0.8.1"
settings.goVersion = "1.26"
settings.lint = ["**", "!examples"]
settings.includeExtraFiles = ["shared-assets/**"]
```

- `version`: Staticcheck release for `go install` (default `v0.8.1`).
- `goVersion`: Go image version (default `1.26`). Choose a toolchain supported by
  the selected Staticcheck release and new enough for the project.
- `base`: Optional Go container, mutually exclusive with `goVersion`. The default
  Alpine image includes a C/C++ toolchain. Custom bases must supply their own Go
  toolchain and any native dependencies.
- `lint`: Select module roots with paths or `path/**`; `!path` excludes a subtree.
  Exclusions win regardless of order. With no positive patterns, all modules
  are included. `*` and `**` select all modules; `!**` skips all. Other glob forms
  are not supported.
- `includeExtraFiles`: Extra workspace-root-relative file patterns to mount.

`staticcheck.conf` files are mounted with their directory structure intact, so
Staticcheck's normal configuration inheritance applies. Go sources, native source
files, embedded assets, and local `go.mod` replacement dependencies within the
workspace are included automatically. Module/build/analysis caches are reused.

For programmatic use, `modules` returns discovered `GoModule` objects and
`module(path: ...)` finds an enclosing module. Each exposes `lint`, `skipLint`,
`source`, and its include patterns for inspection.

## Development

```sh
dagger check -m .dagger/modules/e2e
```

The checks cover discovery, lint diagnostics (including test files), exclusions,
configuration inheritance, embedded assets, local replacements, Cgo, custom base
containers, and the source-discovery helper's unit tests.
