package main

import (
	"fmt"
	"io/ioutil"
	"sync"
)

func tileWorker(queue chan *TerrainTile, directory string, workwg *sync.WaitGroup) {
	for tile := range queue {

		// create temp file
		// basename := fmt.Sprintf("%v_%v_%v_*.csv", tile.X(), tile.Y(), tile.Z())
		basename := fmt.Sprintf("%v_%v_%v_*.xyz", tile.X(), tile.Y(), tile.Z())
		fh, err := ioutil.TempFile(directory, basename)
		if nil != err {
			panic(err)
		}
		defer fh.Close()

		// write file to xyz file
		err = tile.ToXYZ(fh)
		if nil != err {
			panic(err)
		}
		workwg.Done()
	}
}
