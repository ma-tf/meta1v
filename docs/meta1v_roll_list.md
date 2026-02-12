## meta1v roll list

Display roll information in human-readable format

### Synopsis

Display film roll information including film ID, title, load date, frame count, 
ISO, and user-provided remarks.

```
meta1v roll list <filename> [flags]
```

### Examples

```
  # Display roll information
  meta1v roll list data.efd

  # Using the short alias
  meta1v roll ls data.efd

  # With strict mode
  meta1v r ls data.efd --strict
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

* [meta1v roll](meta1v_roll.md)	 - List or export roll information from EFD files

