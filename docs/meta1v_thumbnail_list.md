## meta1v thumbnail list

Display embedded thumbnails as ASCII art

### Synopsis

Display embedded thumbnail images as ASCII art, including the file path and 
rendered ASCII representation.

```
meta1v thumbnail list <filename> [flags]
```

### Examples

```
  # Display thumbnail information
  meta1v thumbnail list data.efd

  # Using the short alias
  meta1v thumbnail ls data.efd

  # With strict mode
  meta1v t ls data.efd --strict
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

* [meta1v thumbnail](meta1v_thumbnail.md)	 - Display embedded thumbnail images from EFD files

