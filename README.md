# meta1v

[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/ma-tf/meta1v)](https://pkg.go.dev/github.com/ma-tf/meta1v)
[![Go Report Card](https://goreportcard.com/badge/github.com/ma-tf/meta1v)](https://goreportcard.com/report/github.com/ma-tf/meta1v)
![Codecov](https://img.shields.io/codecov/c/github/ma-tf/meta1v)
[![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/ma-tf/meta1v/ci.yml)](https://github.com/ma-tf/meta1v/actions)
[![GitHub Releases](https://img.shields.io/github/v/release/ma-tf/meta1v)](https://github.com/ma-tf/meta1v/releases/latest)
[![GitHub License](https://img.shields.io/github/license/ma-tf/meta1v)](https://github.com/ma-tf/meta1v/blob/master/COPYING)

meta1v is a command-line tool for viewing and manipulating metadata for Canon EOS-1V files of the EFD format.

## Installation

### From releases

Download the latest release for your platform from the [releases page](https://github.com/ma-tf/meta1v/releases/latest).

Extract the archive and optionally add the binary to your PATH:

```bash
tar -xzf meta1v_*.tar.gz
sudo mv meta1v /usr/local/bin/
```

### From source

Requires Go 1.25.7 or later:

```bash
go install github.com/ma-tf/meta1v@latest
```

### Install man pages

Man pages are included in the release archives. To install them:

```bash
# After extracting the release archive
sudo cp man/*.1 /usr/local/share/man/man1/
sudo mandb  # Update man database (Linux)
man meta1v
```

## Quick Start

View roll information from an EFD file:
```bash
meta1v roll list data.efd
```

Export frame data to CSV:
```bash
meta1v frame export data.efd output.csv
```

Write EXIF metadata to an image:
```bash
meta1v exif data.efd 1 image.jpg
```

## Documentation

- **[CLI Reference](docs/meta1v.md)** - Complete command reference
- **Man pages** - Use `man meta1v` for offline reference (after installation)

### Available Commands

- `roll` - List or export roll information from EFD files
- `frame` - List or export frame information from EFD files
- `exif` - Write EXIF metadata from EFD file to target image file
- `customfunctions` - List or export custom function settings from EFD files
- `focusingpoints` - Display autofocus point grids from EFD files
- `thumbnail` - Display embedded thumbnail images from EFD files

Run `meta1v --help` for detailed usage information, or see the [complete CLI reference](docs/cli/meta1v.md).

## Configuration

meta1v can be configured via:
- **Config file**: `$HOME/.meta1v/config.yaml` or `./config.yaml`
- **Environment variables**: Prefix with `META1V_` (e.g., `META1V_LOG_LEVEL=debug`)
- **Command-line flags**: `--strict`, `--config`, etc.

Example configuration file (`~/.meta1v/config.yaml`):

```yaml
log:
  level: info
strict: false
timeout: 3m
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `log.level` | string | `warn` | Log level: `debug`, `info`, `warn`, `error` |
| `strict` | boolean | `false` | Enable strict mode (fail on unknown metadata values) |
| `timeout` | duration | `3m` | Command execution timeout |

### Global Flags

- `--config` - Specify custom config file path
- `-s, --strict` - Enable strict mode
- `-h, --help` - Display help for any command

## Licence

Copyright (C) 2026  Matt F

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.

## Development Prerequisites

To contribute to this project, you'll need:
- Go 1.25.7 or later
- [go-licenses](https://github.com/google/go-licenses) for dependency license tracking:
    go install github.com/google/go-licenses/v2@latest

The pre-commit hook automatically generates a NOTICE file with dependency licenses.