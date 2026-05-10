# Changelog — gqlkit-ts

All notable changes to the [`gqlkit-ts`](https://www.npmjs.com/package/gqlkit-ts) npm package. Format: [Keep a Changelog](https://keepachangelog.com/), versioning: [SemVer](https://semver.org/).

## [0.2.0] — 2026-05-10

### Added
- `batch(client, builders, options?)` — merges multiple generated builders into a single GraphQL operation with aliased root fields. One HTTP round trip instead of N. Result is keyed by alias, each value typed from that builder's `execute()` return type via the `BatchResult<T>` mapped type.
- `BatchableBuilder<TResult>`, `BatchResult<T>`, `BatchOptions`, and `OpFragment` exports for callers building custom batch helpers on top of the runtime.
- `BaseBuilder.getOpFragment(alias)` — emits an alias-namespaced fragment that `batch()` splices into a merged operation. Argument names are prefixed with the alias (`$open_filter` instead of `$filter`) so two builders sharing an argument name coexist without colliding.

### Changed
- Generated TypeScript SDKs (via `gqlkit generate-ts`) now include a `getOpFragment(alias)` forwarder on every operation builder, so they satisfy `BatchableBuilder` structurally without inheriting from `BaseBuilder` directly.

### Fixed
- Self-referencing object fields (e.g. `Item.parent: Item`) no longer generate as scalar leaves — they get a proper selector callback (`parent<U>(...)`) like every other object field. Affects schemas with tree / graph types. Generator-side fix in `gqlkit@v0.8.0`; this release ships the runtime side.

## [0.1.0] — earlier

### Added
- Initial release. Lightweight, zero-dependency runtime library consumed by SDKs generated via `gqlkit generate-ts`.
- `GraphQLClient` — HTTP client with bearer-token auth, custom headers, custom-fetch injection.
- `GraphQLErrors` — structured error class wrapping the server's GraphQL error array.
- `FieldSelection` + `BaseBuilder` — query-assembly primitives that generated SDKs extend.
- Build pipeline: `tsc` → `dist/` (CommonJS + `.d.ts`).

[0.2.0]: https://github.com/khanakia/gqlkit/releases/tag/gqlkit-ts%40v0.2.0
