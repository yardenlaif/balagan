# Balagan
## A source code obfuscator for go

**This project is not ready for production use. Run with caution. Be sure to test the obfuscated code. Open an issue if you encounter a new one.**

# How to use
```sh
./balagan <source directory> <target directory>
```
**Arguments**
---
| Argument | Type | Description |
---
| source directory | Required | Directory with code to obfuscate |
---
| target directory | Required | Directory to write obfuscated code to |
---

# About
Balagan will recurse over a directory, obfuscating any go code in its path. It will maintain directory structure, filenames and public names.
