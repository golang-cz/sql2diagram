# sql2diagram

![Database Schema](./_example/schema.svg)

## Generate diagram from sql dump

```bash
go run . -schema schema.sql > schema.svg
```

## Example

```bash
make run-example
```

```bash
go run . -schema ./_example/schema.sql > ./_example/schema.svg
```
