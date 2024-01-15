package schema

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type fileset map[string]struct{}

func (fs fileset) add(filename string) {
	fs[filename] = struct{}{}
}

func (fs fileset) size() int {
	return len(fs)
}

func (fs fileset) filenames() []string {
	res := make([]string, 0, len(fs))
	for f := range fs {
		res = append(res, f)
	}
	sort.Strings(res)
	return res
}

func glob(root string, patterns []string) (fileset, error) {
	fset := make(fileset)
	for _, pattern := range patterns {
		if filepath.Separator != '/' {
			pattern = strings.ReplaceAll(pattern, string(filepath.Separator), "/")
		}
		components := strings.Split(pattern, "/")
		if err := doGlob(components, 0, root, fset); err != nil {
			return nil, err
		}
	}
	return fset, nil
}

func doGlob(components []string, idx int, rootPath string, fset fileset) error {
	if idx == len(components) {
		return nil
	}

	isLast := idx == len(components)-1
	switch components[idx] {
	case "":
		return nil

	case "*":
		if isLast {
			files, err := readFiles(rootPath)
			if err != nil {
				return err
			}

			for _, f := range files {
				fset.add(filepath.Join(rootPath, f.Name()))
			}
			return nil
		} else {
			// walk over all subdirs
			subdirs, err := readSubdirs(rootPath)
			if err != nil {
				return err
			}

			for _, subdir := range subdirs {
				err = doGlob(components, idx+1, filepath.Join(rootPath, subdir.Name()), fset)
				if err != nil {
					return err
				}
			}
			return nil
		}

	case "**":
		if !isLast {
			subdirs, err := readSubdirs(rootPath)
			if err != nil {
				return err
			}

			err = doGlob(components, idx+1, rootPath, fset)
			if err != nil {
				return err
			}

			for _, subdir := range subdirs {
				subdirPath := filepath.Join(rootPath, subdir.Name())
				err = doGlob(components, idx, subdirPath, fset)
				if err != nil {
					return err
				}

				err = doGlob(components, idx+1, subdirPath, fset)
				if err != nil {
					return err
				}
			}
		}
		return nil

	default:
		rx, err := globRegexp(components[idx])
		if err != nil {
			return err
		}

		if isLast {
			files, err := readFiles(rootPath)
			if err != nil {
				return err
			}

			for _, f := range files {
				if name := f.Name(); rx.MatchString(name) {
					fset.add(filepath.Join(rootPath, name))
				}
			}
		} else {
			subdirs, err := readSubdirs(rootPath)
			if err != nil {
				return err
			}

			for _, subdir := range subdirs {
				if name := subdir.Name(); rx.MatchString(name) {
					err = doGlob(components, idx+1, filepath.Join(rootPath, name), fset)
					if err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
}

func globRegexp(s string) (*regexp.Regexp, error) {
	var rx strings.Builder
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '*':
			if i+1 < len(s) && s[i+1] == '*' {
				return nil, errorString("invalid glob token '**'")
			}
			rx.WriteString(".*")

		case '?':
			rx.WriteByte('.')

		case '[':
			rx.WriteByte('[')
			if i++; i < len(s) {
				if s[i] == '!' {
					rx.WriteByte('^')
				} else {
					rx.WriteByte(s[i])
				}
				for i++; i < len(s) && s[i] != ']'; i++ {
					rx.WriteByte(s[i])
				}
				if i < len(s) {
					rx.WriteByte(']')
				}
			}

		case '.':
			rx.WriteString("\\.")

		default:
			rx.WriteByte(s[i])
		}
	}

	res, err := regexp.Compile(rx.String())
	if err != nil {
		return nil, errorf("invalid glob token '%s'", s)
	}
	return res, nil
}

func readSubdirs(dirname string) ([]os.DirEntry, error) {
	return readdir(dirname, func(entry os.DirEntry) bool {
		return entry.IsDir()
	})
}

func readFiles(dirname string) ([]os.DirEntry, error) {
	return readdir(dirname, func(info os.DirEntry) bool {
		return !info.IsDir()
	})
}

func readdir(dirname string, filter func(os.DirEntry) bool) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	res := make([]os.DirEntry, 0, len(entries))
	for _, entry := range entries {
		if filter(entry) {
			res = append(res, entry)
		}
	}
	return res, nil
}
