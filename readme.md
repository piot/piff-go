# Piff

A stream friendly chunk based format. Inspired by [RIFF](https://en.wikipedia.org/wiki/Resource_Interchange_File_Format).

Every chunk has an four octet identifier and an octet length. It is up to the application to define the identifiers and their usage.

## Install

```shell
go install ./...
```

## Usage

```shell
piff-view some_file.piff
```
