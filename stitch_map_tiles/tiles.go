package main

import (
	"image"
)

// GetTileNames
func GetTileNames(minlat, maxlat, minlng, maxlng float64, z int) []xyz {
	tiles := []xyz{}

	// upper right
	ur_tile_x, ur_tile_y := deg2num(maxlat, maxlng, z)
	// lower left
	ll_tile_x, ll_tile_y := deg2num(minlat, minlng, z)

	// Add buffer to make sure output image
	// fills specified height and width.
	for x := ll_tile_x - 2; x < ur_tile_x+2; x++ {
		if x < 0 {
			x = 0
		}
		NUM_ROWS++
		NUM_COLS = 0
		// Add buffer to make sure output image
		// fills specified height and width.
		for y := ur_tile_y - 2; y < ll_tile_y+2; y++ {
			if y < 0 {
				y = 0
			}
			NUM_COLS++
			tiles = append(tiles, xyz{x, y, z})
		}
	}

	for i := range tiles {
		if MAX_X < tiles[i].x {
			MAX_X = tiles[i].x
		}
		if MAX_Y < tiles[i].y {
			MAX_Y = tiles[i].y
		}
	}

	newRect := image.Rectangle{
		Min: image.Point{X: 0, Y: 0},
		Max: image.Point{X: NUM_ROWS * TILE_SIZE, Y: NUM_COLS * TILE_SIZE},
	}
	output = image.NewRGBA64(newRect)

	return tiles
}
