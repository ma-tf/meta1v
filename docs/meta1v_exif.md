## meta1v exif

Write EXIF metadata from EFD file to target image file

### Synopsis

Extract exposure metadata (Tv, Av, ISO, exposure compensation) from a specific 
frame in an EFD file and write it as EXIF data to a target image file.

```
meta1v exif <efd_file> <frame_number> <target_file> [flags]
```

### Examples

```
  # Write EXIF from frame 1 to an image file
  meta1v exif data.efd 1 image.jpg

  # Write EXIF with strict mode enabled
  meta1v exif data.efd 12 photo.jpg --strict
```

### Options

```
  -h, --help   help for exif
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.meta1v/config)
  -s, --strict          enable strict mode (fail on unknown metadata values)
```

### SEE ALSO

* [meta1v](meta1v.md)	 - Provides a way to interact with Canon's EFD files.

