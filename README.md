# sql2diagram

## Generate diagram from sql dump

```bash
go run . schema.sql > schema.svg
```

## Example

```bash
make run-example
```

```bash
go run github.com/golang-cz/sql2diagram/cmd/sql2diagram ./_example/schema.sql > ./_example/schema.svg
```

![Database Schema](./_example/schema.svg)
