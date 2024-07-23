# Balagan
## A source code obfuscator for go

Balagan will recurse over a directory, obfuscating any go code in its path. It will maintain directory structure, filenames and interfaces. Any non-go files in the source directory will be copied as-is to the target directory.

**Run with caution. May take a few minutes to obfuscate. Be sure to test the obfuscated code. Open an issue if you encounter one.**

# Installation
```sh
go install github.com/yardenlaif/balagan@latest
```

# How to use
```sh
balagan -s <source> -t <target> [-i <ignore1> [<ignore2>]]
```
**Arguments**
| Flag | Name | Type | Description |
| ---- | ---- | ---- | ----------- |
| -s | source | Required | Directory with code to obfuscate |
| -t | target | Required | Directory to write obfuscated code to |
| -i | ignore | Optional | Directory (one or more) to ignore. These directories will not be copied or obfuscated to the target directory |

# Support
There are cases where Balagan can't obfuscate but can keep the integrity of the program, and there are cases where Balagan will break your code. See below what can be obfuscated and what is supported:

| Desc | Obfuscate | Supported | Notes |
| ---- | --------- | --------- | ----- |
| Function names | ✔ | ✔ | `main` and `init` will not be obfuscated |
| Variable names | ✔ | ✔ | |
| Method names | ✔ | ✔ | If the receiver implements any interface, all its methods won't be obfuscated |
| Struct names | ✔ | ✔ | |
| Struct field names | ✔ | ✔ |
| Interface names | ✔ | ✔ | |
| Interface method names | ✗ | ✔ | |
| Type switch assignments | ✗ | ✔ | |
| Package names | ✗ | ✔ | |
| Strings | ✗ | ✔ | |
| Comments | ✔ | ✔ | All comments will be removed |
| Filenames | ✗ | ✔ | |
| Build tag directives | ✗ | ✗ | Only files included in the default build tags will be obfuscated and copied to the target directory, but build tag directives will remain in the obfuscated files |
| Linkname directives | ✗ | ✗ | |
| Go assembly | ✗ | ✗ | |
