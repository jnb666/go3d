// ungamma utility converts an image to linear RGB color space
package main

import (
	"flag"
	"fmt"
	"github.com/jnb666/go3d/glu"
	"image/jpeg"
	"image/png"
	"os"
	"path"
	"strings"
)

func main() {
	var format string
	flag.StringVar(&format, "f", "png", "output format: png or jpeg")
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Println("Usage: ungamma [-f format] files...")
		os.Exit(1)
	}
	for _, infile := range flag.Args() {
		convert(format, infile)
	}
}

func convert(format, inFile string) {
	base := strings.Split(path.Base(inFile), ".")[0]
	outFile := fmt.Sprintf("%s_rgb.%s", base, format)
	out, err := os.Create(outFile)
	if err != nil {
		fmt.Println("error opening output", err)
		os.Exit(1)
	}
	defer out.Close()
	img, err := glu.PNGImage(inFile)
	if err != nil {
		fmt.Printf("error loading PNG image from %s : %s\n", inFile, err)
		os.Exit(1)
	}
	dst := glu.ToNRGBA(img, true)
	switch format {
	case "png":
		fmt.Printf("encoding %s to %s in PNG format\n", inFile, outFile)
		err = png.Encode(out, dst)
	case "jpg", "jpeg":
		fmt.Printf("encoding %s to %s in JPEG format\n", inFile, outFile)
		err = jpeg.Encode(out, dst, nil)
	default:
		fmt.Println("unknown output format", format)
		os.Exit(1)
	}
	if err != nil {
		fmt.Println("error encoding output", err)
		os.Exit(1)
	}
}
