package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/ryankurte/go-mapbox/lib"
	"github.com/sjsafranek/goutils/shell"
)

const (
	// DEFAULT_OUT_FILE default tif file to generate
	DEFAULT_OUT_FILE string = "out.tif"
	// DEFAULT_MAPBOX_TOKEN default mapbox access token
	DEFAULT_MAPBOX_TOKEN string = ""
	// DEFAULT_ZOOM default map zoom level
	DEFAULT_ZOOM int = -1
)

var (
	// OUT_FILE tif file to generate
	OUT_FILE string = DEFAULT_OUT_FILE
	// MAPBOX_TOKEN mapbox access token
	MAPBOX_TOKEN string = DEFAULT_MAPBOX_TOKEN
	// MIN_LAT min latitude
	MIN_LAT float64
	// MAX_LAT max latitude
	MAX_LAT float64
	// MIN_LNG min longitude
	MIN_LNG float64
	// MAX_LNG max longitude
	MAX_LNG float64
	// ZOOM map zoom level
	ZOOM int = DEFAULT_ZOOM

	numWorkers int
	mapBox     *mapbox.Mapbox
)

func init() {
	flag.StringVar(&MAPBOX_TOKEN, "token", DEFAULT_MAPBOX_TOKEN, "Mapbox access token")
	flag.StringVar(&OUT_FILE, "o", DEFAULT_OUT_FILE, "Out file")
	flag.Float64Var(&MIN_LAT, "minlat", -85, "min latitude")
	flag.Float64Var(&MAX_LAT, "maxlat", 85, "max latitude")
	flag.Float64Var(&MIN_LNG, "minlng", -175, "min longitude")
	flag.Float64Var(&MAX_LNG, "maxlng", 175, "max longitude")
	flag.IntVar(&ZOOM, "zoom", DEFAULT_ZOOM, "zoom. This will be automatically calculated if not provided.")
	flag.IntVar(&numWorkers, "w", runtime.NumCPU(), "Number of workers")
	flag.Parse()

	// Calculate zoom if not specified
	if -1 == ZOOM {
		panic(errors.New("Must supply a map zoom"))
	}

	if "" == MAPBOX_TOKEN {
		MAPBOX_TOKEN = os.Getenv("MAPBOX_TOKEN")
	}
	if "" == MAPBOX_TOKEN {
		panic(errors.New("Must supply a MAPBOX_TOKEN"))
	}

	mb, err := mapbox.NewMapbox(MAPBOX_TOKEN)
	if nil != err {
		panic(err)
	}
	mapBox = mb
}

func BuildGeoTIFFFromExtent(minLat, maxLat, minLng, maxLng float64, zoom int, outFile string) {
	tiles := GetTileNamesFromMapView(minLat, maxLat, minLng, maxLng, zoom)

	log.Printf(`Parameters:
	extent:	[%v, %v, %v, %v]
	zoom:	%v
	tiles:	%v`, minLat, maxLat, minLng, maxLng, zoom, len(tiles))

	if 100 < len(tiles) {
		panic(errors.New("Too many map tiles. Please raise map zoom or change bounds"))
	}

	// create temp directroy
	directory, err := ioutil.TempDir(os.TempDir(), "terrain-rgb")
	if nil != err {
		panic(err)
	}
	defer os.RemoveAll(directory)
	//.end

	var workwg sync.WaitGroup
	queue := make(chan xyz, numWorkers*2)

	log.Println("Spawning workers")
	for i := 0; i < numWorkers; i++ {
		go worker(queue, directory, &workwg)
	}

	log.Println("Requesting tiles")
	for _, v := range tiles {
		workwg.Add(1)
		queue <- v
	}

	close(queue)

	workwg.Wait()

	log.Println("Building GeoTIFF")
	shell.RunScript("./build_tiff.sh", directory, outFile)
}

func main() {
	startTime := time.Now()
	BuildGeoTIFFFromExtent(MIN_LAT, MAX_LAT, MIN_LNG, MAX_LNG, ZOOM, OUT_FILE)
	log.Println("Runtime:", time.Since(startTime))
}
