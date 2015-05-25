package main

import (
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
	// Packages image/gif and image/jpeg are not used explicitly in the code
	// below, but are imported for their initialization side-effect, which
	// allows image.Decode to understand GIF and JPEG formatted images.
	_ "image/gif"
	_ "image/jpeg"
)

type ImageTile struct {
	Image        image.Image
	AverageColor LCH
}

type ImageBlock struct {
	Image  image.Image
	Bounds image.Rectangle
}

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

			c := LCHModel.Convert(ScaleImage(i, 1, 1).At(0, 0)).(LCH)

			itchan <- &ImageTile{i, c}

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

func getBlocks(i image.Image, ibx int, iby int, tiles []ImageTile) []ImageBlock {

	bct := ibx * iby
	ib := i.Bounds()
	bw := ib.Dx() / ibx
	bh := ib.Dy() / iby

	blocks := make([]ImageBlock, 0, bct)
	bchan := make(chan ImageBlock)

	for by := ib.Min.Y; by < ib.Max.Y; by += bh {
		for bx := ib.Min.X; bx < ib.Max.X; bx += bw {

			go func(i image.Image, b image.Rectangle, t []ImageTile) {
				sub := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
				draw.Draw(sub, sub.Bounds(), i, b.Min, draw.Src)

				c := LCHModel.Convert(ScaleImage(sub, 1, 1).At(0, 0)).(LCH)
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

	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <src> <tile dir>", os.Args[0])
	}

	tx, ty := 20, 20 //desired tile count along the x and y of the image

	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal("Could not open src image!")
	}
	defer file.Close()

	src, _, err := image.Decode(file)
	if err != nil {
		log.Fatal("Could not decode source image!")
	}

	dir, err := os.Open(os.Args[2])
	if err != nil {
		log.Fatal("Could not open tile directory!")
	}
	defer dir.Close()

	info, err := dir.Stat()
	if err != nil || !info.IsDir() {
		log.Fatal("Tile directory is not a directory!")
	}

	tiles, err := getTiles(dir, src.Bounds().Dx()/tx, src.Bounds().Dy()/ty)
	if err != nil || len(tiles) == 0 {
		log.Fatal("Could not load tile images!")
	}

	blocks := getBlocks(src, tx, ty, tiles)

	dst := image.NewRGBA(src.Bounds())

	for _, block := range blocks {
		draw.Draw(dst, block.Bounds, block.Image, image.ZP, draw.Src)
	}

	err = writeImage("mosaic.png", dst)
	if err != nil {
		log.Fatal(err)
	}
}
