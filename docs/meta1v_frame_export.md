## meta1v frame export

Export frame information to CSV format

### Synopsis

Export detailed frame information to CSV format, including frame number, exposure 
settings (Tv, Av, ISO), exposure compensation, and user-provided remarks. Output can be 
directed to stdout or saved to a specified file.

```
meta1v frame export <efd_file> [target_file] [flags]
```

### Examples

```
  # Export frame data to stdout
  meta1v frame export data.efd

  # Export to a file
  meta1v frame export data.efd output.csv

  # Overwrite existing file
  meta1v f export data.efd output.csv --force
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

* [meta1v frame](meta1v_frame.md)	 - List or export frame information from EFD files

