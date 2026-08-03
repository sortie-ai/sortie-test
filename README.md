# sortie-test
 
Test fixtures repository for [Sortie](https://github.com/sortie-ai/sortie) integration and E2E tests. Not a real project.
 
Do not use for production work. Issues and labels in this repository are created and modified by automated tests.
 
## Layout
 
- `e2e/` holds markers written by automated adapter tests. Do not edit by hand.
- `internal/api/` is a small HTTP service used as a working target for agent runs.
 
## Conventions
 
When changing `internal/api`, follow the patterns already in the package:
 
- Every failing response uses the envelope in `errors.go`: `{"error": {"code": ..., "message": ...}}`. Write it with `WriteError`. Never write a bare status code or a plain string.
- Error codes are `lower_snake_case` and name the cause, not the HTTP status.
- Routes are registered in `NewRouter` with method patterns, for example `GET /items`.
- Successful collection responses are wrapped in a named struct, never a bare slice or map.
- Handler tests are table-driven: a `tests` slice of anonymous structs and one `t.Run` per case.
 
Run `go test ./...` before opening a pull request.
