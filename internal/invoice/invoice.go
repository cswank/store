package invoice

import (
	"github.com/jung-kurt/gofpdf"
)

func New() {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Hello, world")
	err := pdf.OutputFileAndClose("/tmp/hello.pdf")
}
