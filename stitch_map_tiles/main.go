package main

import (
	"flag"
	"fmt"
	"image"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	TILELAYER_URL string
	SAVEFILE      string
	MIN_LAT       float64
	MAX_LAT       float64
	MIN_LNG       float64
	MAX_LNG       float64
	ZOOM          int
	HEIGHT        int
	WIDTH         int
	NUM_WORKERS   int
	TILE_SIZE     int = 256
	NUM_ROWS      int
	NUM_COLS      int
	output        *image.RGBA64
	workwg        sync.WaitGroup
	queue         chan Tile
	MAX_X         int
	MIN_X         int
	MAX_Y         int
	MIN_Y         int
)

func Worker(n int) {
	for tile := range queue {
		start_time := time.Now()
		data := GetTilePngBytesFromUrl(tile.Url)
		tile.Image = BytesToPngImage(data)
		drawTile(output, &tile)
		fmt.Println(n, tile.Z, tile.X, tile.Y, time.Since(start_time))
		workwg.Done()
	}
}

func init() {
	// flag.StringVar(&TILELAYER_URL, "u", "https://a.tile.openstreetmap.org/{z}/{x}/{y}.png", "tile layer url")
	flag.StringVar(&TILELAYER_URL, "u", "http://services.arcgisonline.com/ArcGIS/rest/services/World_Topo_Map/MapServer/tile/{z}/{y}/{x}.png", "tile layer url")
	flag.StringVar(&SAVEFILE, "o", "output.png", "save png file")
	flag.Float64Var(&MIN_LAT, "minlat", -85, "min latitude")
	flag.Float64Var(&MAX_LAT, "maxlat", 85, "max latitude")
	flag.Float64Var(&MIN_LNG, "minlng", -175, "min longitude")
	flag.Float64Var(&MAX_LNG, "maxlng", 175, "max longitude")
	flag.IntVar(&ZOOM, "z", -1, "zoom. This will be automatically calculated if not provided.")
	flag.IntVar(&HEIGHT, "height", 1080, "Image height")
	flag.IntVar(&WIDTH, "width", 1920, "Image height")
	flag.IntVar(&NUM_WORKERS, "w", runtime.NumCPU(), "Number of workers")
	flag.Parse()

	// Calculate zoom if not specified
	if -1 == ZOOM {
		ZOOM = getZoomLevelFromBbox(MIN_LAT, MAX_LAT, MIN_LNG, MAX_LNG, HEIGHT, WIDTH)
	}

	queue = make(chan Tile, NUM_WORKERS*2)
}

func main() {

	tiles := GetTileNames(MIN_LAT, MAX_LAT, MIN_LNG, MAX_LNG, ZOOM)

	cooked_tiles := 0

	start_time := time.Now()

	fmt.Println("Requesting tiles", time.Since(start_time))

	for i := 0; i < NUM_WORKERS; i++ {
		go Worker(i)
	}

	for _, v := range tiles {

		tile_url := fmt.Sprintf("/%v/%v/%v.png", v.z, v.x, v.y)
		basemap_url := TILELAYER_URL + tile_url
		if strings.Contains(TILELAYER_URL, "{z}") {
			basemap_url = TILELAYER_URL
			basemap_url = strings.Replace(basemap_url, "{z}", fmt.Sprintf("%v", v.z), 1)
			basemap_url = strings.Replace(basemap_url, "{y}", fmt.Sprintf("%v", v.y), 1)
			basemap_url = strings.Replace(basemap_url, "{x}", fmt.Sprintf("%v", v.x), 1)
		}

		workwg.Add(1)
		cooked_tiles++
		queue <- Tile{
			X:   v.x,
			Y:   v.y,
			Z:   v.z,
			Url: basemap_url,
			Px:  (NUM_ROWS - (MAX_X - v.x + 1)) * TILE_SIZE,
			Py:  (NUM_COLS - (MAX_Y - v.y + 1)) * TILE_SIZE,
		}
	}

	close(queue)

	workwg.Wait()

	fmt.Println("Finished recieving tiles", cooked_tiles, time.Since(start_time))

	savePng("./"+SAVEFILE, clipImage(output))

	fmt.Println("Finished merging tiles", cooked_tiles, time.Since(start_time))
}
