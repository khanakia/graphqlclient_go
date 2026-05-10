# Changelog — gqlkit-sdl (CLI)

All notable changes to the `gqlkit-sdl` binary. Format: [Keep a Changelog](https://keepachangelog.com/), versioning: [SemVer](https://semver.org/).

Tagged as `gqlkit-sdl@vX.Y.Z`. See releases at <https://github.com/khanakia/gqlkit/releases>.

## [0.8.0] — 2026-03-31

### Added
- Query / mutation filtering on `fetch` — `--include-query` / `--exclude-query` (and the mutation duals) let you scope the fetched SDL to a subset of operations. Useful when introspecting large schemas where only a handful of root fields are relevant to your client.
- Unused-type removal pass — types unreachable from the included operations are dropped from the output. Produces a smaller, focused SDL file.

## [0.7.0] — 2026-03-22

### Added
- `--format json` flag — emits the introspection result as raw JSON instead of SDL. Useful for piping into other tooling that expects the GraphQL introspection response shape.

## [0.6.0] — 2026-03-18

### Added
- `--package` flag accepts a full Go import path (paired release with `gqlkit@v0.6.0`).

## [0.5.0] — 2026-03-18

### Added
- Auto-detect SDK import path from `go.mod` (paired release with `gqlkit@v0.5.0`).

## [0.4.0] — 2026-03-18

### Changed
- Generated `graphqlclient` is bundled into the SDK; full GitHub module paths in generated imports (paired release with `gqlkit@v0.4.0`).

## [0.3.0] — 2026-03-18

### Added
- `-o` / `-c` shorthand flags; `--config` is optional (paired release with `gqlkit@v0.3.0`).

## [0.2.0] — 2026-03-18

### Added
- TypeScript codegen support, custom scalar examples, AI-friendly docs (paired release with `gqlkit@v0.2.0`).

## [0.1.0] — 2026-03-17

### Added
- Initial release. Cobra-based `gqlkit-sdl fetch` command — performs GraphQL introspection against a running endpoint and writes SDL to disk.
- `--debug-curl` flag for troubleshooting introspection requests.
- GoReleaser-driven release workflow tagged as `gqlkit-sdl@vX.Y.Z`.

[0.8.0]: https://github.com/khanakia/gqlkit/releases/tag/gqlkit-sdl%40v0.8.0
[0.7.0]: https://github.com/khanakia/gqlkit/releases/tag/gqlkit-sdl%40v0.7.0
[0.6.0]: https://github.com/khanakia/gqlkit/releases/tag/gqlkit-sdl%40v0.6.0
[0.5.0]: https://github.com/khanakia/gqlkit/releases/tag/gqlkit-sdl%40v0.5.0
[0.4.0]: https://github.com/khanakia/gqlkit/releases/tag/gqlkit-sdl%40v0.4.0
[0.3.0]: https://github.com/khanakia/gqlkit/releases/tag/gqlkit-sdl%40v0.3.0
[0.2.0]: https://github.com/khanakia/gqlkit/releases/tag/gqlkit-sdl%40v0.2.0
[0.1.0]: https://github.com/khanakia/gqlkit/releases/tag/gqlkit-sdl%40v0.1.0
