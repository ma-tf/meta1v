## meta1v roll export

Export roll information to CSV format

### Synopsis

Export film roll information to CSV format, including film ID, title, load date, 
frame count, ISO, and user-provided remarks. Output can be directed to stdout or saved 
to a specified file.

```
meta1v roll export <efd_file> [target_file] [flags]
```

### Examples

```
  # Export roll data to stdout
  meta1v roll export data.efd

  # Export to a file
  meta1v roll export data.efd output.csv

  # Overwrite existing file
  meta1v r export data.efd output.csv --force
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

* [meta1v roll](meta1v_roll.md)	 - List or export roll information from EFD files

