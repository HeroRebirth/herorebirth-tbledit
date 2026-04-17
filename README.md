### Hero Rebirth TBL Editor
An open source TBL Editor for Hero Client, built on top of the syntaxgame foundation with ongoing improvements and refinements. The editor is inteded for the [Hero Rebirth Server](https://github.com/HeroRebirth/herorebirth-server).

### Installation
```go
go build -o tbledit
```

### Commands and examples

**Export a single `.tbl` to Excel**
```
tbledit export -i tb_cashshop.tbl -o cashshop.xlsx
```

**Import an Excel file back to `.tbl`**
```
tbledit import -i cashshop.xlsx -o tb_cashshop.tbl
```

**Export all `.tbl` files in a directory to CSV** (required before `analyze` and `seed`)
```
tbledit export -d data
```
Optional: specify a custom output folder (default is `./tmp`)
```
tbledit export -d data --tmp ./tmp
```

**Analyze exported CSVs against `map.json`** — shows sample values per mapped SQL column so you can verify mappings before seeding
```
tbledit analyze
```
Optional flags (defaults shown):
```
tbledit analyze --tmp ./tmp --map map.json
```

**Generate the SQL seed file** from exported CSVs — writes `INSERT ... ON DUPLICATE KEY UPDATE` statements for all mapped tables
```
tbledit seed
```
Optional flags (defaults shown):
```
tbledit seed --tmp ./tmp --map map.json -o herorebirth_seed.sql
```

**Full pipeline** (run from the project root)
```
tbledit export -d data
tbledit analyze
tbledit seed
```

## License

See [LICENSE](LICENSE) file.

### Credits
- Syntaxgame — original foundation
- All contributors who have helped improve and maintain this project