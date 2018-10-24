package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
)

// GetTilePngBytesFromUrl requests map tile png from url.
func GetTilePngBytesFromUrl(tile_url string) []byte {
	// Just a simple GET request to the image URL
	// We get back a *Response, and an error
	res, err := client.Get(tile_url)
	if err != nil {
		fmt.Printf("Error http.Get -> %v\n", err)
		return []byte("")
	}

	// We read all the bytes of the image
	// Types: data []byte
	data, err := ioutil.ReadAll(res.Body)

	// You have to manually close the body, check docs
	// This is required if you want to use things like
	// Keep-Alive and other HTTP sorcery.
	defer res.Body.Close()

	if err != nil {
		fmt.Printf("Error ioutil.ReadAll -> %v\n", err)
		return []byte("")
	}

	return data
}

// BytesToPngImage converts bytes to png image struct.
func BytesToPngImage(b []byte) image.Image {
	img, err := png.Decode(bytes.NewReader(b))
	if nil != err {
		img, err = jpeg.Decode(bytes.NewReader(b))
		if nil != err {
			fmt.Println("************")
			fmt.Println(string(b))
			fmt.Println(b)
			fmt.Println("************")
			panic(err)
		}
	}
	return img
}

// clipImage clips image to specified height and width
func clipImage(inImage image.Image) image.Image {
	newRect := image.Rectangle{
		Min: image.Point{X: 0, Y: 0},
		Max: image.Point{X: WIDTH, Y: HEIGHT},
	}
	outImage := image.NewRGBA(newRect)

	var x int
	var y int
	imgX := inImage.Bounds().Max.X
	imgY := inImage.Bounds().Max.Y

	bounds := outImage.Bounds()

	if imgX >= WIDTH {
		x = (imgX - WIDTH) / 2
	} else {
		x = 0
		bounds.Min.X = (WIDTH - imgX) / 2
		bounds.Max.X = imgX + (WIDTH-imgX)/2
	}

	if imgY >= HEIGHT {
		y = (imgY - HEIGHT) / 2
	} else {
		y = 0
		bounds.Min.Y = (HEIGHT - imgY) / 2
		bounds.Max.Y = (HEIGHT - imgY) / 2
		bounds.Max.Y = imgY + (HEIGHT-imgY)/2
	}

	draw.Draw(outImage, bounds, inImage, image.Point{x, y}, draw.Over)
	return outImage
}

/* Draws a tile's image on a target canvas according to pixel position
information in the Tile. */
func drawTile(target *image.RGBA64, tile *Tile) {
	if nil == tile.Image {
		panic("Error with tile image")
	}
	tile_x := tile.Px
	tile_y := tile.Py
	rect := image.Rect(tile_x, tile_y, tile_x+TILE_SIZE, tile_y+TILE_SIZE)
	draw.Draw(target, rect, tile.Image, image.Point{0, 0}, draw.Over)
}

// savePng writes png image to file
func savePng(filename string, img image.Image) {
	// Create a new file and write to it.
	out, err := os.Create(filename)
	if err != nil {
		panic(err)
		os.Exit(1)
	}
	err = png.Encode(out, img)
	if err != nil {
		panic(err)
		os.Exit(1)
	}
}
