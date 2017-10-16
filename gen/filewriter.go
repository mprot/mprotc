package gen

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Printer defines an interface for printing lines of code into a memory buffer.
type Printer interface {
	Println(args ...interface{}) // always appends newline
}

// FileWriter manages the printer for the requested files. The base name of the
// target file cannot be changed. Printers are distinguished by the file extension
// of the target file.
type FileWriter struct {
	baseName string
	printers map[string]*printer // filename => printer
}

// NewFileWriter creates a new file writer from the output directory and the target's
// filename. The base name for each file results from outputDir and targetName, where
// the file extension of targetName will be trimmed.
func NewFileWriter(outputDir string, targetName string) *FileWriter {
	ext := filepath.Ext(targetName)
	targetName = targetName[:len(targetName)-len(ext)]

	return &FileWriter{
		baseName: filepath.Join(outputDir, targetName),
		printers: make(map[string]*printer),
	}
}

// Printer returns a printer for the provided file extension. For each requested
// file extension a printer will be created.
func (w *FileWriter) Printer(fileExt string) Printer {
	filename := w.baseName + fileExt
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
		if err := p.writeFile(filename); err != nil {
			return err
		}
	}
	return nil
}

type printer struct {
	bytes.Buffer
}

func (p *printer) Println(args ...interface{}) {
	fmt.Fprint(p, args...)
	p.WriteByte('\n')
}

func (p *printer) writeFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	n, err := f.Write(p.Bytes())
	switch {
	case err != nil:
		return err
	case n != p.Len():
		return io.ErrShortWrite
	default:
		return nil
	}
}
