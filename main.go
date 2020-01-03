package main

import (
	"errors"
	"flag"
	// "io/ioutil"
	"log"
	"os"
	"runtime"
	"time"
)

const (
	// DEFAULT_MAPBOX_TOKEN default mapbox access token
	// DEFAULT_MAPBOX_TOKEN string = ""
	// DEFAULT_OUT_FILE default tif file to generate
	DEFAULT_OUT_FILE          string = ""
	PROJECT                   string = "Terrain RGB Sticher"
	VERSION                   string = "0.1.0"
	DEFAULT_DATABASE_ENGINE   string = "postgres"
	DEFAULT_DATABASE_NAME     string = "geodev"
	DEFAULT_DATABASE_PASSWORD string = "dev"
	DEFAULT_DATABASE_USERNAME string = "geodev"
	DEFAULT_DATABASE_HOST     string = "localhost"
	DEFAULT_DATABASE_PORT     int64  = 5432
	DEFAULT_DATABASE_TABLE    string = "terrain"
)

var (
	// OUT_FILE tif file to generate
	OUT_FILE string = DEFAULT_OUT_FILE
	// DB_TABLE to import geotiffs into
	// MAPBOX_TOKEN mapbox access token
	// MAPBOX_TOKEN string = DEFAULT_MAPBOX_TOKEN
	MAPBOX_TOKEN string = os.Getenv("MAPBOX_TOKEN")
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

	//
	DATABASE_ENGINE   string = DEFAULT_DATABASE_ENGINE
	DATABASE_NAME     string = DEFAULT_DATABASE_NAME
	DATABASE_PASSWORD string = DEFAULT_DATABASE_PASSWORD
	DATABASE_USERNAME string = DEFAULT_DATABASE_USERNAME
	DATABASE_HOST     string = DEFAULT_DATABASE_HOST
	DATABASE_PORT     int64  = DEFAULT_DATABASE_PORT
	DATABASE_TABLE    string = DEFAULT_DATABASE_TABLE
)

func init() {
	// flag.StringVar(&MAPBOX_TOKEN, "token", DEFAULT_MAPBOX_TOKEN, "Mapbox access token")
	flag.StringVar(&MAPBOX_TOKEN, "token", MAPBOX_TOKEN, "Mapbox access token")
	flag.StringVar(&OUT_FILE, "o", DEFAULT_OUT_FILE, "Out file")
	flag.StringVar(&DATABASE_TABLE, "table", DEFAULT_DATABASE_TABLE, "Database table")
	flag.Float64Var(&MIN_LAT, "minlat", -85, "min latitude")
	flag.Float64Var(&MAX_LAT, "maxlat", 85, "max latitude")
	flag.Float64Var(&MIN_LNG, "minlng", -175, "min longitude")
	flag.Float64Var(&MAX_LNG, "maxlng", 175, "max longitude")
	flag.IntVar(&ZOOM, "zoom", DEFAULT_ZOOM, "zoom. This will be automatically calculated if not provided.")
	flag.StringVar(&DATABASE_HOST, "dbhost", DEFAULT_DATABASE_HOST, "database host")
	flag.StringVar(&DATABASE_NAME, "dbname", DEFAULT_DATABASE_NAME, "database name")
	flag.StringVar(&DATABASE_PASSWORD, "dbpass", DEFAULT_DATABASE_PASSWORD, "database password")
	flag.StringVar(&DATABASE_USERNAME, "dbuser", DEFAULT_DATABASE_USERNAME, "database username")
	flag.Int64Var(&DATABASE_PORT, "dbport", DEFAULT_DATABASE_PORT, "Database port")
	flag.IntVar(&numWorkers, "w", runtime.NumCPU(), "Number of workers")

	flag.Parse()

	// Calculate zoom if not specified
	if 1 > ZOOM {
		panic(errors.New("Must supply a map zoom"))
	}

	// If MAPBOX_TOKEN is not defined get from os environmental variables
	// if "" == MAPBOX_TOKEN {
	// 	MAPBOX_TOKEN = os.Getenv("MAPBOX_TOKEN")
	// }

	if "" == MAPBOX_TOKEN {
		panic(errors.New("Must supply a MAPBOX_TOKEN"))
	}
	//.end
}

func main() {

	tmap, err := NewTerrainMap(MAPBOX_TOKEN)
	if nil != err {
		panic(err)
	}

	startTime := time.Now()

	tmap.SetView(MIN_LAT, MAX_LAT, MIN_LNG, MAX_LNG, ZOOM)
	// tmap.SetZoom(ZOOM)
	// tmap.FetchTiles(MIN_LAT, MAX_LAT, MIN_LNG, MAX_LNG)
	tmap.FetchTiles()
	if "" != OUT_FILE {
		tmap.Render(OUT_FILE)
	} else {
		tmap.Rasters2pgsql(DATABASE_NAME, DATABASE_USERNAME, DATABASE_PASSWORD, DATABASE_TABLE, DATABASE_HOST, DATABASE_PORT)
	}
	tmap.Destroy()

	log.Println("Runtime:", time.Since(startTime))
}
