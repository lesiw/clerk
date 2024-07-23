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
package clerk

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type ClerkFS []fs.FS

func (cfs *ClerkFS) Add(fsys fs.FS) error {
	err := fs.WalkDir(fsys, ".",
		func(p string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				return nil
			}
			if cfs.fileExists(p) {
				return fmt.Errorf("bad fs: file '%s' conflicts", p)
			}
			return nil
		},
	)
	if err != nil {
		return err
	}
	*cfs = append(*cfs, fsys)
	return nil
}

func (cfs *ClerkFS) Apply(dir string) error {
	sums, err := loadSums(filepath.Join(dir, "clerk.sum"))
	rmlist := make(map[string]bool, len(sums))
	for path := range sums {
		rmlist[path] = true
	}
	if err != nil {
		return fmt.Errorf("failed to load clerk.sum: %w", err)
	}
	for _, a := range *cfs {
		err := fs.WalkDir(a, ".", func(
			path string, d fs.DirEntry, err error,
		) error {
			if d.IsDir() {
				return nil
			}
			delete(rmlist, path)

			realpath := filepath.Join(dir, path)
			dir := filepath.Dir(realpath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to make directory '%s': %w",
					path, err)
			}

			if !bytes.Equal(fileHash(realpath), sums[path]) {
				if !confirm("File '%s' changed. Overwrite?", path) {
					return nil
				}
			}

			dst, err := os.Create(realpath)
			if err != nil {
				return fmt.Errorf("failed to create '%s': %w", realpath, err)
			}

			src, err := a.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open '%s': %w", path, err)
			}

			if _, err = io.Copy(dst, src); err != nil {
				return fmt.Errorf("failed to copy '%s' -> '%s': %w",
					path, realpath, err)
			}

			sums[path] = fileHash(path)
			return nil
		})
		if err != nil {
			return err
		}
	}
	for path := range rmlist {
		realpath := filepath.Join(dir, path)
		if !bytes.Equal(fileHash(realpath), sums[path]) {
			if !confirm("File '%s' changed. Delete?", path) {
				continue
			}
		}
		if err := os.Remove(realpath); err != nil {
			return fmt.Errorf("failed to remove '%s': %w", realpath, err)
		}
	}
	return sums.Save(filepath.Join(dir, "clerk.sum"))
}

func (cfs *ClerkFS) fileExists(path string) (found bool) {
	for _, f := range *cfs {
		fs.WalkDir(f, ".", func(p string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				return nil
			}
			if path == p {
				found = true
				return fs.SkipAll
			}
			return nil
		})
	}
	return
}

func confirm(format string, a ...any) bool {
	fmt.Printf(format, a...)
	fmt.Print(" [y/N] ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil || input == "" {
		return false
	}
	return []rune(strings.ToLower(input))[0] == 'y'
}
