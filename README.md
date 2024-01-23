# mprotc
`mprotc` is a compiler for [mprot](https://github.com/mprot/mprot) to translate schema definitions into source code.

## Usage

```
Synopsis:
  mprotc <language> [options] [schema-file ...]
  mprotc help <language>

Description:
  The mprotc tool compiles schema definition files into source code of the specified programming language.

  The following options are available:
    --root <path>
        Specify the root path for the schema files. The given schema files are interpreted relative to this
        directory. The default is the current directory.
    --out <path>
        Specify the output path for the generated code. The default is the current directory.
    --deprecated
        Include the deprecated fields in the generated code.
    --dryrun
        Print the names of generated files only instead of writing the files.
```

## Supported Languages
* [Golang](internal/gen/golang/README.md):
```
mprotc go [options] [schema-file ...]

Additional Options:
  --import-root
      Import root path for all schema imports. The default is the path specified by the output directory.
  --scoped-enums
      Scope the enumerators of the generated enums, i.e. prefix the enumerator names with the enum name.
      The default is false.
  --unwrap-union
      Unwrap the union types of the generated struct fields, i.e. use an empty interface as the field type.
      The default is false.
```
* [JavaScript/TypeScript](internal/gen/js/README.md):
```
mprotc js [options] [schema-file ...]

Additional Options:
  --typedecls
      Generate type declarations in a separate .d.ts file. This flag should be used to generate TypeScript
      instead of JavaScript. The default is false.
```
