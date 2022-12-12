# Organize & Deduplicate Files

Organize your photos, pictures, etc into a folder ensuring that duplicates are removed.

![logo](./logo.svg)

## Usage

```
$> organize-dup-files [OPTIONS]

Application Options:
  -s, --src=                 the source folder/file (default: .)
  -d, --dst=                 the destination folder (default: .)
  -x, --exclude=             exclude file/folder
  -e, --ext=                 a list of file extensions to consider
      --preserve-file-names  if provided, preserve the source filename (default truncates/clean them)

Help Options:
  -h, --help                 Show this help message
```

The output is in the form of a shell script so it's easy to edit and examine.

Enjoy!
