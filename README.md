# Balagan
## A source code obfuscator for go

**Run with caution. Be sure to test the obfuscated code. Open an issue if you encounter one.**

# How to use
```sh
./balagan -s <source> -t <target> [-i <ignore1> [<ignore2>]]
```
**Arguments**
| Flag | Name | Type | Description |
| ---- | ---- | ---- | ----------- |
| -s | source | Required | Directory with code to obfuscate |
| -t | target | Required | Directory to write obfuscated code to |
| -i | ignore | Optional | Directory (one or more) to ignore. These directories will not be copied or obfuscated to the target directory |

# About
Balagan will recurse over a directory, obfuscating any go code in its path. It will maintain directory structure, filenames and interfaces.

# Unsupported
Your code will not be successfully obfuscated if it includes:

- Go assembly
- Linkname directives
