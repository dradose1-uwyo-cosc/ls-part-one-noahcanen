# COSC 3750 - Homework 04 (`gols`)

This project implements a simplified Unix-like `ls` clone with the required flags:

- `-a` include dot entries and hidden files
- `-l` long listing format
- `-n` numeric uid/gid (implies `-l`)
- `-h` human-readable size (base 1024, meaningful with `-l`)
- `-R` recursive traversal

## Run

```bash
go run .
go run . -l
go run . -l -a -R
```

## Build

```bash
go build -o gols .
```
