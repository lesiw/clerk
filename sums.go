package clerk

import (
	"bufio"
	"cmp"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

type sums map[string][]byte

func loadSums(path string) (sums, error) {
	s := make(map[string][]byte)
	file, err := os.Open(path)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return s, nil
	} else if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(file)
	for line := 0; scanner.Scan(); line++ {
		path, sum, ok := strings.Cut(scanner.Text(), " ")
		if !ok {
			return nil, fmt.Errorf("bad format: line %d", line)
		}
		var buf []byte
		buf, err = hex.DecodeString(sum)
		if err != nil {
			return nil, fmt.Errorf("bad hash: line %d: %w", line, err)
		}
		s[path] = buf
	}
	return s, nil
}

func (s sums) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create '%s': %w", path, err)
	}
	for _, k := range sort(keys(s)) {
		if len(s[k]) == 0 {
			return fmt.Errorf("bad hash: empty hash for '%s'", k)
		}
		fmt.Fprintf(file, "%s %s\n", k, hex.EncodeToString(s[k]))
	}
	return nil
}

func fileHash(path string) []byte {
	file, err := os.Open(path)
	if err != nil {
		return []byte{}
	}
	hash := sha1.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return []byte{}
	}
	return hash.Sum(nil)
}

func sort[S ~[]E, E cmp.Ordered](x S) []E {
	slices.Sort(x)
	return x
}

func keys[M ~map[K]V, K comparable, V any](x M) []K {
	ret := make([]K, 0, len(x))
	for k := range x {
		ret = append(ret, k)
	}
	return ret
}
