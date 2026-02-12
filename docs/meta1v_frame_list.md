## meta1v frame list

Display frame information in human-readable format

### Synopsis

Display detailed information about frames including exposure settings (Tv, Av, ISO), 
exposure compensation, focus points, custom functions, and user-provided remarks.

```
meta1v frame list <filename> [flags]
```

### Examples

```
  # Display frame information
  meta1v frame list data.efd

  # Using the short alias
  meta1v frame ls data.efd

  # With strict mode
  meta1v f ls data.efd --strict
```

### Options

```
  -h, --help   help for list
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.meta1v/config)
  -s, --strict          enable strict mode (fail on unknown metadata values)
```

### SEE ALSO

* [meta1v frame](meta1v_frame.md)	 - List or export frame information from EFD files

