# lesiw.io/clerk

[![Go Reference](https://pkg.go.dev/badge/lesiw.io/clerk.svg)](https://pkg.go.dev/lesiw.io/clerk)

```go
// Package clerk provides a utility for updating a directory without
// disturbing the rest of the directory's contents. Create a [ClerkFS] and
// [ClerkFS.Add] one or more [fs.FS] to it, then [ClerkFS.Apply] the combined
// result to a directory.
//
// A [ClerkFS] writes checksums of the files it created to a clerk.sum file in
// the target directory.
//
// If a [ClerkFS] detects that a file has changed since it was last written, it
// will prompt the user before overwriting or deleting it; otherwise, it will
// continue to silently manage the file.
```
