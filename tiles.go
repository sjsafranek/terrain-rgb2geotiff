package main

// GetTileNames
func GetTileNames(minlat, maxlat, minlng, maxlng float64, z int) []xyz {
	tiles := []xyz{}

	// upper right
	ur_tile_x, ur_tile_y := deg2num(maxlat, maxlng, z)
	// lower left
	ll_tile_x, ll_tile_y := deg2num(minlat, minlng, z)

	// Add buffer to make sure output image
	// fills specified height and width.
	for x := ll_tile_x - 1; x < ur_tile_x+1; x++ {
		// for x := ll_tile_x - 2; x < ur_tile_x+2; x++ {
		if x < 0 {
			x = 0
		}
		// Add buffer to make sure output image
		// fills specified height and width.
		for y := ur_tile_y - 1; y < ll_tile_y+1; y++ {
			// for y := ur_tile_y - 2; y < ll_tile_y+2; y++ {
			if y < 0 {
				y = 0
			}
			tiles = append(tiles, xyz{uint64(x), uint64(y), uint64(z)})
		}
	}

	return tiles
}
