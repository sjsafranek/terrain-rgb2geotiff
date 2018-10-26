package main

import (
	"errors"
	// "fmt"
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
	MapBox *mapbox.Mapbox
	zoom   int
}

func (self *TerrainMap) SetZoom(zoom int) {
	self.zoom = zoom
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
	shell.RunScript("/bin/sh", "./build_tiff.sh", directory, outFile)

	// files, _ := utils.FilesInDirectory(directory)
	// for _, file := range files {
	// 	err := self.csv2geotiff(file)
	// 	if nil != err {
	// 		log.Println(err)
	// 	}
	// }

	// shell.RunScript("bash", "-c", fmt.Sprintf(`
	// 	'gdalwarp --config GDAL_CACHEMAX 3000 -wm 3000 %v/*.tif %v'
	// `, directory, outFile))
}

/*
func (self *TerrainMap) csv2geotiff(csvfile string) error {
	xyzfile := strings.Replace(csvfile, ".csv", ".xyz", -1)
	tiffile := strings.Replace(csvfile, ".csv", ".tif", -1)
	log.Printf("Converting %v to %v", csvfile, tiffile)

	shell.RunScript(`$(echo head -n 1 `+csvfile+`)`, ">", xyzfile)
	shell.RunScript("tail", "-n", "+2", csvfile, "|", "sort", "-n", "-t", "','", "-k2", "-k1", ">>", xyzfile)

	// cmd := fmt.Sprintf(`'$(echo head -n 1 "%v") >  "%v"; tail -n +2 "%v" | sort -n -t ',' -k2 -k1 >> "%v";'`, csvfile, xyzfile, csvfile, xyzfile)
	// shell.RunScript("bash", "-c", cmd)
	// fmt.Println("bash", "-c", cmd)
	// out, err := exec.Command("bash", "-c", cmd).Output()
	// log.Println(out)
	// log.Println(err)

	// shell.RunScript("bash", "-c", fmt.Sprintf("gdal_translate %v %v", xyzfile, tiffile))

	return nil
}
*/
