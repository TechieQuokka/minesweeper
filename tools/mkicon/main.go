// Command mkicon writes the app icon as PNG files, e.g. for desktop entries
// or packaging:
//
//	go run ./tools/mkicon -out assets
package main

import (
	"flag"
	"fmt"
	"image/png"
	"os"
	"path/filepath"

	"github.com/TechieQuokka/minesweeper/internal/ui"
)

func main() {
	out := flag.String("out", "assets", "output directory")
	flag.Parse()

	if err := os.MkdirAll(*out, 0o755); err != nil {
		fail(err)
	}
	for _, size := range []int{16, 32, 48, 64, 128, 256, 512} {
		name := filepath.Join(*out, fmt.Sprintf("icon-%d.png", size))
		if err := writePNG(name, size); err != nil {
			fail(err)
		}
		fmt.Println(name)
	}
}

func writePNG(name string, size int) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	if err := png.Encode(f, ui.Icon(size)); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "mkicon: %v\n", err)
	os.Exit(1)
}
