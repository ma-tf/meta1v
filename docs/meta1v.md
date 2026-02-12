## meta1v

Provides a way to interact with Canon's EFD files.

### Synopsis

meta1v is a command line tool to interact with Canon's EFD files.

You can print out information to stdout about the film roll, including focus
points, custom functions, roll information, thumbnail previews, and more.

### Examples

```
  # View roll information
  meta1v roll list data.efd

  # Export frame data to CSV
  meta1v frame export data.efd output.csv

  # Write EXIF metadata to an image
  meta1v exif data.efd 1 image.jpg
```

### Options

```
      --config string   config file (default is $HOME/.meta1v/config)
  -h, --help            help for meta1v
  -s, --strict          enable strict mode (fail on unknown metadata values)
```

### SEE ALSO

* [meta1v customfunctions](meta1v_customfunctions.md)	 - List or export custom function settings from EFD files
* [meta1v exif](meta1v_exif.md)	 - Write EXIF metadata from EFD file to target image file
* [meta1v focusingpoints](meta1v_focusingpoints.md)	 - Display autofocus point grids from EFD files
* [meta1v frame](meta1v_frame.md)	 - List or export frame information from EFD files
* [meta1v roll](meta1v_roll.md)	 - List or export roll information from EFD files
* [meta1v thumbnail](meta1v_thumbnail.md)	 - Display embedded thumbnail images from EFD files
* [meta1v version](meta1v_version.md)	 - Print version information

