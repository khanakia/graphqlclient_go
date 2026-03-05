# gqlgenapi – Test GraphQL API (gqlgen)

This module provides a small GraphQL API built with [`github.com/99designs/gqlgen`](https://github.com/99designs/gqlgen).  
It is intended as a **local test server** for exercising the SDK/builder in this workspace.

## Schema

The schema is based on the gqlgen getting-started example, with an extra `ping` field:

- **Query**
  - `ping: String!` – simple healthcheck used by the SDK example.
  - `todos: [Todo!]!` – sample list query.
- **Mutation**
  - `createTodo(input: NewTodo!): Todo!`

You can inspect the full schema via GraphQL introspection or in `graph/schema.graphqls`.

## Running the server

From the workspace root:

```bash
task gqlgenapi:run
```

This starts the server on `http://localhost:8081` with:

- GraphQL endpoint: `http://localhost:8081/query`
- GraphQL Playground UI: `http://localhost:8081/`

## Regenerating gqlgen code

If you change the schema in `graph/schema.graphqls`, regenerate the gqlgen artifacts:

```bash
task gqlgenapi:generate
```

> Note: The generate task runs with `GOWORK=off` because gqlgen currently works more reliably outside of Go workspace mode.

