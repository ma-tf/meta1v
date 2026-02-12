## meta1v customfunctions list

Display custom function settings in human-readable format

### Synopsis

Display a table of custom function settings used by the frames.

For the meaning of each custom function and its respective value, refer to the 
Canon EOS-1V manual.

```
meta1v customfunctions list <filename> [flags]
```

### Examples

```
  # Display custom functions
  meta1v customfunctions list data.efd

  # Using the short alias
  meta1v customfunctions ls data.efd

  # With strict mode
  meta1v cf ls data.efd --strict
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

* [meta1v customfunctions](meta1v_customfunctions.md)	 - List or export custom function settings from EFD files

