package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	// "os/exec"
	// "strings"
	"sync"

	"github.com/ryankurte/go-mapbox/lib"
	// "github.com/sjsafranek/goutils"
	"github.com/sjsafranek/goutils/shell"
)

const (
	DEFAULT_ZOOM int = 10
	ZOOM_MAX     int = 15
	ZOOM_MIN     int = 1
)

// xyz
type xyz struct {
	x uint64
	y uint64
	z uint64
}

func NewTerrainMap(token string) (*TerrainMap, error) {
	mb, err := mapbox.NewMapbox(MAPBOX_TOKEN)
	return &TerrainMap{MapBox: mb, zoom: 2}, err
}

type TerrainMap struct {
	MapBox    *mapbox.Mapbox
	zoom      int
	directory string
}

func (self *TerrainMap) SetZoom(zoom int) error {
	if ZOOM_MIN > zoom {
		return fmt.Errorf("Must supply a map zoom (%v to %v)", ZOOM_MIN, ZOOM_MAX)
	}

	// https://docs.mapbox.com/help/troubleshooting/access-elevation-data/
	// max is zoom 15
	if zoom > ZOOM_MAX {
		log.Printf("Mapbox Terrain-RGB tiles have a max zoom of %v\n", ZOOM_MAX)
		log.Println("	See https://docs.mapbox.com/help/troubleshooting/access-elevation-data/ for more details")
		zoom = ZOOM_MAX
	}

	self.zoom = zoom
	return nil
}

func (self *TerrainMap) GetZoom() int {
	if ZOOM_MAX < self.zoom || ZOOM_MIN > self.zoom {
		return DEFAULT_ZOOM
	}
	return self.zoom
}

func (self *TerrainMap) SetDirectory(directory string) {
	self.directory = directory
}

func (self *TerrainMap) Render(minLat, maxLat, minLng, maxLng float64, zoom int, outFile string) {
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
		go terrainWorker(self.MapBox, queue, directory, &workwg)
	}

	log.Println("Requesting tiles")
	for _, v := range tiles {
		workwg.Add(1)
		queue <- v
	}

	close(queue)

	workwg.Wait()

	log.Println("Building GeoTIFF")
	err = self.buildGeoTIFF(directory, outFile)
	if nil != err {
		log.Fatal(err)
	}
}

func (self TerrainMap) buildGeoTIFF(directory, outFile string) error {
	// bash script contents
	script := `
#!/bin/bash

DIRECTORY=$1
OUT_FILE=$2

# build tiff from each file
echo "Building tif files from csv map tiles"
for FILE in $DIRECTORY/*.csv; do
    GEOTIFF="${FILE%.*}.tif"
    XYZ="${FILE%.*}.xyz"

    echo "Building $XYZ from $FILE"
    $(echo head -n 1 $FILE) >  "$XYZ"; \
        tail -n +2 $FILE | sort -n -t ',' -k2 -k1 >> "$XYZ";

    echo "Building $GEOTIFF from $XYZ"
    gdal_translate "$XYZ" "$GEOTIFF"
done

echo "Merging tif files to $OUT_FILE"
gdalwarp --config GDAL_CACHEMAX 3000 -wm 3000 $DIRECTORY/*.tif $OUT_FILE
	`

	// write to bash script
	fh, err := ioutil.TempFile("", "build_geotiff.*.sh")
	if nil != err {
		return err
	}
	fmt.Fprintf(fh, script)
	fh.Close()
	defer os.Remove(fh.Name())

	// execute bash script
	log.Println(directory, outFile)
	shell.RunScript("/bin/sh", fh.Name(), directory, outFile)

	return nil
}

// RenderTiles
func (self *TerrainMap) FetchTiles(minLat, maxLat, minLng, maxLng float64) error {

	if "" == self.directory {
		path, _ := os.Getwd()
		self.directory = path
	}

	zoom := self.GetZoom()

	tiles := GetTileNamesFromMapView(minLat, maxLat, minLng, maxLng, zoom)

	log.Printf(`Parameters:
	extent:	[%v, %v, %v, %v]
	zoom:	%v
	tiles:	%v`, minLat, maxLat, minLng, maxLng, zoom, len(tiles))

	if 100 < len(tiles) {
		return errors.New("Too many map tiles. Please raise map zoom or change bounds")
	}

	var workwg sync.WaitGroup
	queue := make(chan xyz, numWorkers*2)

	log.Println("Spawning workers")
	for i := 0; i < numWorkers; i++ {
		go terrainWorker(self.MapBox, queue, self.directory, &workwg)
	}

	log.Println("Requesting tiles")
	for _, v := range tiles {
		workwg.Add(1)
		queue <- v
	}

	close(queue)

	workwg.Wait()

	return nil
}

func (self TerrainMap) BuildRasters() error {

	if "" == self.directory {
		path, _ := os.Getwd()
		self.directory = path
	}

	// bash script contents
	script := `
#!/bin/bash

DIRECTORY="` + self.directory + `"

# build tiff from each file
echo "Building tif files from csv map tiles"
for FILE in $DIRECTORY/*.csv; do
	GEOTIFF="${FILE%.*}.tif"
	XYZ="${FILE%.*}.xyz"
	#GEOTIFF="${GEOTIFF/(MISSING)/}"
	#XYZ="${XYZ/(MISSING)/}"

    echo "Building $XYZ from $FILE"
    $(echo head -n 1 $FILE) >  "$XYZ"; \
        tail -n +2 $FILE | sort -n -t ',' -k2 -k1 >> "$XYZ";

    echo "Building $GEOTIFF from $XYZ"
    gdal_translate "$XYZ" "$GEOTIFF"
done
	`

	log.Println(script)

	// write to bash script
	fh, err := ioutil.TempFile("", "build_geotiff.*.sh")
	if nil != err {
		return err
	}
	fmt.Fprintf(fh, script)
	fh.Close()
	defer os.Remove(fh.Name())

	// execute bash script
	results, err := shell.RunScript("/bin/sh", fh.Name())
	log.Println(err)
	log.Println(results)

	return nil
}

func (self TerrainMap) Rasters2pgsql(pg_table string) error {

	if "" == self.directory {
		path, _ := os.Getwd()
		self.directory = path
	}

	// bash script contents
	script := `
#!/bin/bash

DIRECTORY="` + self.directory + `"
DBTABLE="` + pg_table + `"
DBHOST="localhost"
DBPORT=5432

# cleanup
rm tmp.sql
psql -h $DBHOST -p $DBPORT -c 'DROP TABLE '"$DBTABLE"';'

# import
raster2pgsql -d -I -C -M -F -t 256x256 -s 4326 $DIRECTORY/*.tif $DBTABLE > "tmp.sql"

echo "Import to PostGreSQL"
psql -h $DBHOST -p $DBPORT -f tmp.sql

# cleanup
rm tmp.sql
	`

	log.Println(script)

	// write to bash script
	fh, err := ioutil.TempFile("", "build_geotiff.*.sh")
	if nil != err {
		return err
	}
	fmt.Fprintf(fh, script)
	fh.Close()
	defer os.Remove(fh.Name())

	// execute bash script
	results, err := shell.RunScript("/bin/sh", fh.Name())
	log.Println(err)
	log.Println(results)

	return nil
}

//.end
