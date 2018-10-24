package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/ryankurte/go-mapbox/lib"
	"github.com/ryankurte/go-mapbox/lib/base"
	"github.com/ryankurte/go-mapbox/lib/maps"
)

const (
	DEFAULT_OUT_FILE     string = "out.tiff"
	DEFAULT_ACCESS_TOKEN string = ""
	DEFAULT_ZOOM         int    = -1
	DEFAULT_EXTENT       string = ""
)

var (
	OUT_FILE     string = DEFAULT_OUT_FILE
	ACCESS_TOKEN string = DEFAULT_ACCESS_TOKEN
	MIN_LAT      float64
	MAX_LAT      float64
	MIN_LNG      float64
	MAX_LNG      float64
	ZOOM         int = DEFAULT_ZOOM
	HEIGHT       int
	WIDTH        int
	NUM_WORKERS  int
	TILE_SIZE    int = 256
	workwg       sync.WaitGroup
	queue        chan xyz
	mapBox       *mapbox.Mapbox
)

func Worker(n int) {
	for xyz := range queue {
		// fetch tile
		highDPI := false
		tile, err := mapBox.Maps.GetTile(maps.MapIDTerrainRGB, xyz.x, xyz.y, xyz.z, maps.MapFormatPngRaw, highDPI)
		if nil != err {
			panic(err)
		}

		fileHandler, err := os.Create(fmt.Sprintf("tmp/%v_%v_%v.csv", xyz.x, xyz.y, xyz.z))
		if nil != err {
			panic(err)
		}
		defer fileHandler.Close()

		fmt.Fprintf(fileHandler, "x,y,z\n")

		for x := 0; x < tile.Bounds().Max.X; x++ {
			for y := 0; y < tile.Bounds().Max.Y; y++ {

				loc, err := tile.PixelToLocation(float64(x), float64(y))
				if nil != err {
					panic(err)
				}

				ll := base.Location{Latitude: loc.Latitude, Longitude: loc.Longitude}

				elevation, err := tile.GetAltitude(ll)
				if nil != err {
					panic(err)
				}

				line := fmt.Sprintf("%v,%v,%v\n", loc.Longitude, loc.Latitude, elevation)
				fmt.Fprintf(fileHandler, line)

			}
		}

		workwg.Done()
	}
}

func XYZ2GeoTIFF() {
	cmd := exec.Command("./build_tiff.sh", OUT_FILE)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if nil != err {
		log.Println(out.String())
		panic(err)
	}
	log.Println(out.String())
}

func init() {
	flag.StringVar(&ACCESS_TOKEN, "token", DEFAULT_ACCESS_TOKEN, "Mapbox access token")
	flag.StringVar(&OUT_FILE, "out_file", DEFAULT_OUT_FILE, "Out file")
	flag.Float64Var(&MIN_LAT, "minlat", -85, "min latitude")
	flag.Float64Var(&MAX_LAT, "maxlat", 85, "max latitude")
	flag.Float64Var(&MIN_LNG, "minlng", -175, "min longitude")
	flag.Float64Var(&MAX_LNG, "maxlng", 175, "max longitude")
	flag.IntVar(&ZOOM, "zoom", DEFAULT_ZOOM, "zoom. This will be automatically calculated if not provided.")
	flag.IntVar(&HEIGHT, "height", 1080, "Image height")
	flag.IntVar(&WIDTH, "width", 1920, "Image height")
	flag.IntVar(&NUM_WORKERS, "w", runtime.NumCPU(), "Number of workers")
	flag.Parse()

	// Calculate zoom if not specified
	if -1 == ZOOM {
		ZOOM = getZoomLevelFromBbox(MIN_LAT, MAX_LAT, MIN_LNG, MAX_LNG, HEIGHT, WIDTH)
	}

	mb, err := mapbox.NewMapbox(ACCESS_TOKEN)
	if nil != err {
		panic(err)
	}
	mapBox = mb

	queue = make(chan xyz, NUM_WORKERS*2)
}

func main() {

	tiles := GetTileNames(MIN_LAT, MAX_LAT, MIN_LNG, MAX_LNG, ZOOM)

	cooked_tiles := 0

	start_time := time.Now()

	log.Println("Requesting tiles", time.Since(start_time))

	for i := 0; i < NUM_WORKERS; i++ {
		go Worker(i)
	}

	for _, v := range tiles {
		workwg.Add(1)
		cooked_tiles++
		queue <- v
	}

	close(queue)

	workwg.Wait()

	log.Println("Finished recieving tiles", cooked_tiles, time.Since(start_time))

	log.Println("Building GeoTIFF")

	// savePng("./"+SAVEFILE, clipImage(output))
	XYZ2GeoTIFF()

	log.Println("GeoTIFF complete", cooked_tiles, time.Since(start_time))
}
