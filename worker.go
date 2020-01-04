package main

import (
	"fmt"
	"io/ioutil"
	"sync"
	"log"
)

func tileWorkerXYZ(queue chan *TerrainTile, directory string, workwg *sync.WaitGroup) {
	for tile := range queue {

		log.Println(tile.X(), tile.Y(), tile.Z())

		// create temp file
		basename := fmt.Sprintf("%v_%v_%v_*.xyz", tile.X(), tile.Y(), tile.Z())
		fh, err := ioutil.TempFile(directory, basename)
		if nil != err {
			panic(err)
		}
		defer fh.Close()

		// write file to xyz file
		err = tile.WriteXYZ(fh)
		if nil != err {
			panic(err)
		}
		workwg.Done()
	}
}

func tileWorkerGeoTiff(queue chan *TerrainTile, directory string, workwg *sync.WaitGroup) {
	for tile := range queue {

		log.Println(tile.X(), tile.Y(), tile.Z())

		filename := fmt.Sprintf("%v/%v_%v_%v.tif", directory, tile.X(), tile.Y(), tile.Z())

		// write file to geotiff file
		err := tile.WriteGeoTiff(filename)
		if nil != err {
			panic(err)
		}
		workwg.Done()
	}
}
