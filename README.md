# meta1v

[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/ma-tf/meta1v)](https://pkg.go.dev/github.com/ma-tf/meta1v)
[![Go Report Card](https://goreportcard.com/badge/github.com/ma-tf/meta1v)](https://goreportcard.com/report/github.com/ma-tf/meta1v)
![Codecov](https://img.shields.io/codecov/c/github/ma-tf/meta1v)
[![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/ma-tf/meta1v/ci.yml)](https://github.com/ma-tf/meta1v/actions)
[![GitHub Releases](https://img.shields.io/github/v/release/ma-tf/meta1v)](https://github.com/ma-tf/meta1v/releases/latest)
[![GitHub License](https://img.shields.io/github/license/ma-tf/meta1v)](https://github.com/ma-tf/meta1v/blob/master/COPYING)

meta1v is a command-line tool for viewing and manipulating metadata for Canon EOS-1V files of the EFD format.

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