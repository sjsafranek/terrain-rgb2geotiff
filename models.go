package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/ryankurte/go-mapbox/lib"
	"github.com/sjsafranek/goutils/shell"
)

// xyz
type xyz struct {
	x uint64
	y uint64
	z uint64
}

func NewTerrainMap(token string) (*TerrainMap, error) {
	mb, err := mapbox.NewMapbox(MAPBOX_TOKEN)
	return &TerrainMap{MapBox: mb}, err
}

type TerrainMap struct {
	MapBox *mapbox.Mapbox
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
	shell.RunScript("./build_tiff.sh", directory, outFile)
}
