# sdkexample_gqlgenapi

Example SDK module generated from the local `gqlgenapi` GraphQL server using `gqlsdl` + `gqlcustom`.

## Flow

1. Start the `gqlgenapi` server (gqlgen-based API).
2. Use `gqlsdl` to fetch the schema from `http://localhost:8081/query` into `cmd/generate/schema.graphql`.
3. Run the generator to produce the SDK in `./sdk`.
4. Run the sample to execute a `ping` query against `gqlgenapi`.

See root `Taskfile.yml` for one-liner tasks.

