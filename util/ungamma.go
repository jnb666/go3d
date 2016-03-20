// ungamma utility converts an image to linear RGB color space
package main

import (
	"flag"
	"fmt"
	"github.com/jnb666/go3d/glu"
	"image"
	"image/draw"
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

// get input image
func getImage(file string) (img image.Image) {
	r, err := os.Open(file)
	if err != nil {
		fmt.Println("error opening input", err)
		os.Exit(1)
	}
	defer r.Close()
	switch glu.GetFormat(file) {
	case glu.PngFormat:
		img, err = png.Decode(r)
	case glu.JpegFormat:
		img, err = jpeg.Decode(r)
	default:
		fmt.Println("unsupported image format", file)
	}
	if err != nil {
		fmt.Println("error decoding input", err)
	}
	return img
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
	img := &glu.ToNRGBA{getImage(inFile)}
	dst := image.NewNRGBA(img.Bounds())
	draw.Draw(dst, img.Bounds(), img, image.ZP, draw.Src)
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
