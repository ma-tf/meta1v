## meta1v customfunctions export

Export custom function settings to CSV format

### Synopsis

Export custom function settings to CSV format, including function ID, name, 
description, parameters, and user-provided remarks. Output can be directed to stdout or 
saved to a specified file.

```
meta1v customfunctions export <efd_file> [target_file] [flags]
```

### Examples

```
  # Export custom functions to stdout
  meta1v customfunctions export data.efd

  # Export to a file
  meta1v customfunctions export data.efd output.csv

  # Overwrite existing file
  meta1v cf export data.efd output.csv --force
```

### Options

```
  -F, --force   overwrite output file if it exists
  -h, --help    help for export
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.meta1v/config)
  -s, --strict          enable strict mode (fail on unknown metadata values)
```

### SEE ALSO

* [meta1v customfunctions](meta1v_customfunctions.md)	 - List or export custom function settings from EFD files

