# sql2diagram

## Installation
- When integrating this library https://github.com/pganalyze/pg_query_go using Go modules, and using a vendor/ directory, you will need to explicitly copy over some of the C build files, since Go does not copy files in subfolders without .go files whilst vendoring.
The best way to do so is to use modvendor, and vendor your modules like this:
```
go mod vendor
go get -u github.com/goware/modvendor
modvendor -copy="**/*.c **/*.h **/*.proto" -v
```