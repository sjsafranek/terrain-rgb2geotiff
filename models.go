package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/ryankurte/go-mapbox/lib"
)

const (
	DEFAULT_ZOOM int = 10
	ZOOM_MAX     int = 15
	ZOOM_MIN     int = 1
)

var (
	ErrorViewNotSet error = errors.New("View not set")
)

func NewTerrainMap(token string) (*TerrainMap, error) {
	mb, err := mapbox.NewMapbox(MAPBOX_TOKEN)
	if nil != err {
		return &TerrainMap{}, err
	}

	return &TerrainMap{MapBox: mb}, err
}

type MapView struct {
	zoom   int
	minlat float64
	maxlat float64
	minlng float64
	maxlng float64
}

func (self *MapView) Zoom() int {
	return self.zoom
}

func (self *MapView) MinLat() float64 {
	return self.minlat
}

func (self *MapView) MaxLat() float64 {
	return self.maxlat
}

func (self *MapView) MinLng() float64 {
	return self.minlng
}

func (self *MapView) MaxLng() float64 {
	return self.maxlng
}

func (self *MapView) GetTiles() []*TerrainTile {
	tiles := []*TerrainTile{}

	maxlat := self.MaxLat()
	maxlng := self.MaxLng()
	minlat := self.MinLat()
	minlng := self.MinLng()

	// z := self.GetZoom()
	z := self.Zoom()

	// upper right
	ur_tile_x, ur_tile_y := deg2num(maxlat, maxlng, z)

	// lower left
	ll_tile_x, ll_tile_y := deg2num(minlat, minlng, z)

	// Add buffer to make sure output image
	// fills specified height and width.
	for x := ll_tile_x - 1; x < ur_tile_x+1; x++ {
		if x < 0 {
			x = 0
		}
		// Add buffer to make sure output image
		// fills specified height and width.
		for y := ur_tile_y - 1; y < ll_tile_y+1; y++ {
			if y < 0 {
				y = 0
			}
			// tiles = append(tiles, &TerrainTile{maps: self.MapBox.Maps, x: uint64(x), y: uint64(y), z: uint64(z)})
			tiles = append(tiles, &TerrainTile{x: uint64(x), y: uint64(y), z: uint64(z)})
		}
	}

	return tiles
}

type TerrainMap struct {
	MapBox *mapbox.Mapbox
	view   *MapView
	// zoom   int
	// minlat float64
	// maxlat float64
	// minlng float64
	// maxlng float64

	// View TerrainView
	directory string
}

func (self *TerrainMap) Destroy() error {
	directory := self.getDirectory()
	// check if in temp directory
	if strings.HasPrefix(directory, os.TempDir()) {
		// remove artificts
		self.directory = ""
		return os.RemoveAll(directory)
	}
	return nil
}

// func (self *TerrainMap) SetZoom(zoom int) error {
// 	if ZOOM_MIN > zoom {
// 		return fmt.Errorf("Must supply a map zoom (%v to %v)", ZOOM_MIN, ZOOM_MAX)
// 	}
//
// 	// https://docs.mapbox.com/help/troubleshooting/access-elevation-data/
// 	// max is zoom 15
// 	if zoom > ZOOM_MAX {
// 		log.Printf("Mapbox Terrain-RGB tiles have a max zoom of %v\n", ZOOM_MAX)
// 		log.Println("	See https://docs.mapbox.com/help/troubleshooting/access-elevation-data/ for more details")
// 		zoom = ZOOM_MAX
// 	}
//
// 	self.zoom = zoom
// 	return nil
// }
//
// func (self *TerrainMap) GetZoom() int {
// 	if ZOOM_MAX < self.zoom || ZOOM_MIN > self.zoom {
// 		return DEFAULT_ZOOM
// 	}
// 	return self.zoom
// }

func (self *TerrainMap) SetView(minlat, maxlat, minlng, maxlng float64, zoom int) {
	self.view = &MapView{minlat: minlat, maxlat: maxlat, minlng: minlng, maxlng: maxlng, zoom: zoom}
}

// getTilesFromMapView returns tile xyz for bounding box and zoom
// func (self *TerrainMap) getTilesFromMapView(minlat, maxlat, minlng, maxlng float64) ([]*TerrainTile, error) {
// func (self *TerrainMap) getTilesFromMapView() ([]*TerrainTile, error) {
// 	tiles := []*TerrainTile{}
//
// 	if nil == self.view {
// 		return tiles, ErrorViewNotSet
// 	}
//
// 	maxlat := self.view.MaxLat()
// 	maxlng := self.view.MaxLng()
// 	minlat := self.view.MinLat()
// 	minlng := self.view.MinLng()
//
// 	// z := self.GetZoom()
// 	z := self.view.Zoom()
//
// 	// upper right
// 	ur_tile_x, ur_tile_y := deg2num(maxlat, maxlng, z)
//
// 	// lower left
// 	ll_tile_x, ll_tile_y := deg2num(minlat, minlng, z)
//
// 	// Add buffer to make sure output image
// 	// fills specified height and width.
// 	for x := ll_tile_x - 1; x < ur_tile_x+1; x++ {
// 		if x < 0 {
// 			x = 0
// 		}
// 		// Add buffer to make sure output image
// 		// fills specified height and width.
// 		for y := ur_tile_y - 1; y < ll_tile_y+1; y++ {
// 			if y < 0 {
// 				y = 0
// 			}
// 			tiles = append(tiles, &TerrainTile{maps: self.MapBox.Maps, x: uint64(x), y: uint64(y), z: uint64(z)})
// 		}
// 	}
//
// 	return tiles, nil
// }
func (self *TerrainMap) getTilesFromMapView() ([]*TerrainTile, error) {
	if nil == self.view {
		return []*TerrainTile{}, ErrorViewNotSet
	}
	return self.view.GetTiles(), nil
}

func (self *TerrainMap) getDirectory() string {
	if "" == self.directory {
		// directory, err := os.Getwd()
		directory, _ := ioutil.TempDir("", "terrain-rgb")
		self.directory = directory
	}
	return self.directory
}

func (self TerrainMap) Render(out_file string) error {

	if !strings.Contains(out_file, ".tif") {
		out_file += ".tif"
	}

	log.Printf("Rendering to GeoTIFF: %v", out_file)
	directory := self.getDirectory()
	return createAndExecuteScript(directory, "merge_geotiffs_*.sh", fmt.Sprintf(`#!/bin/bash

DIRECTORY="%v"
OUT_FILE="%v"

echo "Merging tif files to $OUT_FILE"
gdalwarp --config GDAL_CACHEMAX 3000 -wm 3000 $DIRECTORY/*.tif $OUT_FILE
	`, directory, out_file))
}

// FetchTiles
// func (self *TerrainMap) FetchTiles(minLat, maxLat, minLng, maxLng float64) error {
func (self *TerrainMap) FetchTiles() error {
	log.Printf("Fetch tiles")

	directory := self.getDirectory()
	// zoom := self.GetZoom()
	// maxlat := self.view.MaxLat()
	// maxlng := self.view.MaxLng()
	// minlat := self.view.MinLat()
	// minlng := self.view.MinLng()

	// tiles := self.getTilesFromMapView(minLat, maxLat, minLng, maxLng)
	tiles, err := self.getTilesFromMapView()
	if nil != err {
		return err
	}

	// log.Printf(`Parameters:
	// extent:	[%v, %v, %v, %v]
	// zoom:	%v
	// tiles:	%v`, minLat, maxLat, minLng, maxLng, zoom, len(tiles))

	if 100 < len(tiles) {
		return errors.New("Too many map tiles. Please raise map zoom or change bounds")
	}

	var workwg sync.WaitGroup
	queue := make(chan *TerrainTile, numWorkers*2)

	log.Println("Spawning workers")
	for i := 0; i < numWorkers; i++ {
		go tileWorker(queue, directory, &workwg)
	}

	log.Println("Requesting tiles")
	for _, tile := range tiles {
		// HACK...
		tile.maps = self.MapBox.Maps
		workwg.Add(1)
		queue <- tile
	}

	close(queue)

	workwg.Wait()

	return self.tiles2Rasters()
}

func (self TerrainMap) tiles2Rasters() error {
	log.Printf("Converting tiles to geotiffs")
	directory := self.getDirectory()
	return createAndExecuteScript(directory, "tiles_to_geotiffs_*.sh", fmt.Sprintf(`#!/bin/bash

DIRECTORY="%v"

# build xyz from each file
for FILE in $DIRECTORY/*.csv; do
	XYZ="${FILE%%.*}.xyz"
    echo "Building $XYZ from $FILE"
    $(echo head -n 1 $FILE) >  "$XYZ"; \
        tail -n +2 $FILE | sort -n -t ',' -k2 -k1 >> "$XYZ";
done

# build geotiff from each file
echo "Building tif files from csv map tiles"
for FILE in $DIRECTORY/*.xyz; do
	GEOTIFF="${FILE%%.*}.tif"
    echo "Building $GEOTIFF from $FILE"
    gdal_translate "$FILE" "$GEOTIFF"
done

	`, directory))

}

func (self TerrainMap) Rasters2pgsql(dbname, dbuser, dbpass, dbtable, dbhost string, dbport int64) error {
	log.Printf("Importing geotiffs to pgsql")
	directory := self.getDirectory()
	return createAndExecuteScript(directory, "rasters_to_pgsql_*.sh", fmt.Sprintf(`#!/bin/bash

DIRECTORY="%v"
DBTABLE="%v"
DBNAME="%v"
DBUSER="%v"
DBPASS="%v"
DBHOST="%v"
DBPORT=%v

# cleanup table
echo "DROP TABLE '"$DBTABLE"'" > "$DIRECTORY/import_to_pgsql.sql"

# import
raster2pgsql -d -I -C -M -F -t 256x256 -s 4326 $DIRECTORY/*.tif $DBTABLE >> "$DIRECTORY/import_to_pgsql.sql"

echo "Import to PostGreSQL"
PGPASSWORD=$DBPASS psql -U $DBUSER -d $DBNAME -h $DBHOST -p $DBPORT -f "$DIRECTORY/import_to_pgsql.sql"
	`, directory, dbtable, dbname, dbuser, dbpass, dbhost, dbport))
}

//.end
