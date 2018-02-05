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

## Supported Languages
* [Golang](gen/golang/README.md):
```
mprotc go [options] [schema-file ...]

Additional Options:
  --scoped-enums
      Scope the enumerators of the generated enums, i.e. prefix the enumerator names with the enum name.
      The default is false.
```
* [JavaScript/TypeScript](gen/js/README.md):
```
mprotc js [options] [schema-file ...]

Additional Options:
  --typedecls
      Generate type declarations in a separate .d.ts file. This flag should be used to generate TypeScript
      instead of JavaScript. The default is false.
```
