# mprotc

`mprotc` is a compiler for mprot to translate schema definitions into source code.

## Usage

```
Synopsis:
  mprotc <language> [options] [schema-file ...]
  mprotc help <language>

Description:
  The mprotc tool compiles schema definition files into source code of the specified programming language.

  The following options are available:
    --out <path>
        Specify the output path for the generated code. The default is the current directory.
    --deprecated
        Include the deprecated fields in the generated code.
    --package <package name>
        Override the schema's package name.
```
