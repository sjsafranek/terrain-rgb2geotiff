package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"strings"
	"os"

	"github.com/ryankurte/go-mapbox/lib"
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
	if nil != err {
		return &TerrainMap{}, err
	}

	return &TerrainMap{MapBox: mb, zoom: DEFAULT_ZOOM}, err
}

type TerrainMap struct {
	MapBox    *mapbox.Mapbox
	zoom      int
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

func (self *TerrainMap) getDirectory() string {
	if "" == self.directory {
		// directory, err := os.Getwd()
		directory, _ := ioutil.TempDir("", "terrain-rgb")
		self.directory = directory
	}
	return self.directory
}

func (self TerrainMap) Render(out_file string) error {
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
func (self *TerrainMap) FetchTiles(minLat, maxLat, minLng, maxLng float64) error {
	log.Printf("Fetch tiles")

	directory := self.getDirectory()
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
		go terrainWorker(self.MapBox, queue, directory, &workwg)
	}

	log.Println("Requesting tiles")
	for _, v := range tiles {
		workwg.Add(1)
		queue <- v
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

# build tiff from each file
echo "Building tif files from csv map tiles"
for FILE in $DIRECTORY/*.csv; do
	GEOTIFF="${FILE%%.*}.tif"
	XYZ="${FILE%%.*}.xyz"

    echo "Building $XYZ from $FILE"
    $(echo head -n 1 $FILE) >  "$XYZ"; \
        tail -n +2 $FILE | sort -n -t ',' -k2 -k1 >> "$XYZ";

    echo "Building $GEOTIFF from $XYZ"
    gdal_translate "$XYZ" "$GEOTIFF"
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
