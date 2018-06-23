package gen

import (
	"path/filepath"
	"sort"
)

// FileWriter manages the printer for the requested files. The base name of the
// target file cannot be changed. Printers are distinguished by the file extension
// of the target file.
type FileWriter struct {
	rootDir  string
	printers map[string]*printer // relative filename => printer
}

// NewFileWriter creates a new file writer from a root directory. The given
// root directory acts as a working directory and is used as the base for
// all generated files.
func NewFileWriter(rootDir string) (*FileWriter, error) {
	root, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, err
	}

	return &FileWriter{
		rootDir:  root,
		printers: make(map[string]*printer),
	}, nil
}

// Printer returns a printer for the provided mprot file. The file extension of
// filename will be replaced with fileExt.
//
// For each requested file, a printer will be registered.
func (w *FileWriter) Printer(filename string, fileExt string) Printer {
	filename = filename[:len(filename)-len(filepath.Ext(filename))] + fileExt
	filename = filepath.Clean(filepath.Join(w.rootDir, filename))

	p := w.printers[filename]
	if p == nil {
		p = &printer{}
		w.printers[filename] = p
	}
	return p
}

// Flush writes the buffered code into the corresponding files. The target filename
// is the base name with the respective file extension appended.
func (w *FileWriter) Flush() error {
	for filename, p := range w.printers {
		if err := p.writeToFile(filename); err != nil {
			return err
		}
	}
	return nil
}

// WalkFiles walks all the files registered in the file writer and calls
// iter for each of these files.
func (w *FileWriter) WalkFiles(iter func(filename string)) {
	filenames := make([]string, 0, len(w.printers))
	for filename := range w.printers {
		filenames = append(filenames, filename)
	}
	sort.Strings(filenames)

	for _, filename := range filenames {
		iter(filename)
	}
}
