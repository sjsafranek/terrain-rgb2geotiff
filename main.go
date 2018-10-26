package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"runtime"
	"time"
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
	// numWorkers number of workers
	numWorkers int
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
}

func main() {
	// If MAPBOX_TOKEN is not defined get from os environmental variables
	if "" == MAPBOX_TOKEN {
		MAPBOX_TOKEN = os.Getenv("MAPBOX_TOKEN")
	}

	if "" == MAPBOX_TOKEN {
		panic(errors.New("Must supply a MAPBOX_TOKEN"))
	}

	//
	tmap, err := NewTerrainMap(MAPBOX_TOKEN)
	if nil != err {
		panic(err)
	}

	startTime := time.Now()
	tmap.Render(MIN_LAT, MAX_LAT, MIN_LNG, MAX_LNG, ZOOM, OUT_FILE)
	log.Println("Runtime:", time.Since(startTime))
}
