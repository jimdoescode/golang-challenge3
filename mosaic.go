package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"math"
	"os"

	// Packages image/gif and image/jpeg are not used explicitly in the code
	// below, but are imported for their initialization side-effect, which
	// allows image.Decode to understand GIF and JPEG formatted images.
	_ "image/gif"
	_ "image/jpeg"
)

type ImageTile struct {
	Image        *image.Image
	AverageColor LCH
}

type ImageBlock struct {
	Image  *image.Image
	Bounds image.Rectangle
}

// Scans a directory and loads tile images from that directory. If the images
// don't match the specified tile dimensions then the image is scaled to fit.
func getTiles(dir *os.File, tw, th int) ([]ImageTile, error) {

	files, err := dir.Readdir(0)
	if err != nil {
		return []ImageTile{}, err
	}

	count := len(files)
	tiles := make([]ImageTile, 0, count)

	itchan := make(chan *ImageTile)

	for _, file := range files {

		go func(fi os.FileInfo, dir string) {

			if fi.IsDir() {
				itchan <- nil
				return
			}

			f, err := os.Open(dir + "/" + fi.Name())
			i, _, err := image.Decode(f)
			f.Close()

			if err != nil {
				itchan <- nil
				return
			}

			if i.Bounds().Dx() != tw || i.Bounds().Dy() != th {
				i = ScaleImage(i, tw, th)
			}

			c := AverageImageColor(i)

			itchan <- &ImageTile{&i, c}

		}(file, dir.Name())
	}

	for i := 0; i < count; i++ {
		itptr := <-itchan
		if itptr != nil {
			tiles = append(tiles, *itptr)
		}
	}

	return tiles, nil
}

// Breaks a source image up into blocks of the specified size. It then
// averages the colors of that image block and searches the ImageTile
// slice for an ImageTile that is the closest to the average block color.
func getBlocks(i image.Image, bw, bh int, tiles []ImageTile) []ImageBlock {

	ib := i.Bounds()
	bct := int(math.Ceil(float64(ib.Dx())/float64(bw)) * math.Ceil(float64(ib.Dy())/float64(bh)))

	blocks := make([]ImageBlock, 0, bct)
	bchan := make(chan ImageBlock, bct)

	for by := ib.Min.Y; by < ib.Max.Y; by += bh {
		for bx := ib.Min.X; bx < ib.Max.X; bx += bw {

			go func(i image.Image, b image.Rectangle, t []ImageTile) {
				sub := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
				draw.Draw(sub, sub.Bounds(), i, b.Min, draw.Src)

				c := AverageImageColor(sub)
				best := tiles[0]

				for _, tile := range tiles {
					if c.Distance(tile.AverageColor) < c.Distance(best.AverageColor) {
						best = tile
					}
				}

				bchan <- ImageBlock{best.Image, b}

			}(i, image.Rect(bx, by, bx+bw, by+bh), tiles)
		}
	}

	for ct := 0; ct < bct; ct++ {
		block := <-bchan
		blocks = append(blocks, block)
	}

	return blocks
}

func writeImage(name string, m image.Image) error {

	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, m)
}

func main() {

	var tw, th int

	flag.IntVar(&tw, "w", 60, "The width of mosaic tiles")
	flag.IntVar(&th, "h", 60, "The height of mosaic tiles")
	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <src> <dir>\n", os.Args[0])
		flag.PrintDefaults()
		return
	}

	if tw < 1 || th < 1 {
		log.Fatal("Invalid tile dimensions specified.")
	}

	file, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal("Could not open src image!")
	}
	defer file.Close()

	src, _, err := image.Decode(file)
	if err != nil {
		log.Fatal("Could not decode source image!")
	}

	dir, err := os.Open(flag.Arg(1))
	if err != nil {
		log.Fatal("Could not open tile directory!")
	}
	defer dir.Close()

	info, err := dir.Stat()
	if err != nil || !info.IsDir() {
		log.Fatal("Tile directory is not a directory!")
	}

	tiles, err := getTiles(dir, tw, th)
	if err != nil || len(tiles) == 0 {
		log.Fatal("Could not load tile images!")
	}

	blocks := getBlocks(src, tw, th, tiles)

	dst := image.NewRGBA(src.Bounds())

	for _, block := range blocks {
		draw.Draw(dst, block.Bounds, *block.Image, image.ZP, draw.Src)
	}

	err = writeImage("mosaic.png", dst)
	if err != nil {
		log.Fatal(err)
	}
}
