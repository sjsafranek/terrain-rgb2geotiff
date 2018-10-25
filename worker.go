package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ryankurte/go-mapbox/lib/base"
	"github.com/ryankurte/go-mapbox/lib/maps"
)

func Worker(n int) {
	for xyz := range queue {
		// fetch tile
		highDPI := false
		log.Println("Fetch tile", xyz)
		tile, err := mapBox.Maps.GetTile(maps.MapIDTerrainRGB, xyz.x, xyz.y, xyz.z, maps.MapFormatPngRaw, highDPI)
		if nil != err {
			// panic(err)
			log.Println(err)
			workwg.Done()
			continue
		}

		// log.Println("Parsing tile", xyz)
		fileHandler, err := os.Create(fmt.Sprintf("tmp/%v_%v_%v.csv", xyz.x, xyz.y, xyz.z))
		if nil != err {
			panic(err)
		}
		defer fileHandler.Close()

		fmt.Fprintf(fileHandler, "x,y,z\n")

		for x := 0; x < tile.Bounds().Max.X; x++ {
			for y := 0; y < tile.Bounds().Max.Y; y++ {

				loc, err := tile.PixelToLocation(float64(x), float64(y))
				if nil != err {
					panic(err)
				}

				ll := base.Location{Latitude: loc.Latitude, Longitude: loc.Longitude}

				elevation, err := tile.GetAltitude(ll)
				if nil != err {
					panic(err)
				}

				line := fmt.Sprintf("%v,%v,%v\n", loc.Longitude, loc.Latitude, elevation)
				fmt.Fprintf(fileHandler, line)

			}
		}

		workwg.Done()
	}
}
