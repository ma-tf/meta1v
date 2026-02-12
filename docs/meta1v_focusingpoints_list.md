## meta1v focusingpoints list

Display autofocus point grids in human-readable format

### Synopsis

Display rendered grids of autofocus points used when capturing each photograph.

For setting autofocus points on the camera, refer to the Canon EOS-1V manual.

```
meta1v focusingpoints list <filename> [flags]
```

### Examples

```
  # Display focusing points information
  meta1v focusingpoints list data.efd

  # Using the short alias
  meta1v focusingpoints ls data.efd

  # With strict mode
  meta1v fp ls data.efd --strict
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

* [meta1v focusingpoints](meta1v_focusingpoints.md)	 - Display autofocus point grids from EFD files

