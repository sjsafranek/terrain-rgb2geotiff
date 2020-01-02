package main

import (
	"fmt"
	// "io/ioutil"
	"log"
	"math"
	"os"

	"github.com/ryankurte/go-mapbox/lib/base"
	"github.com/ryankurte/go-mapbox/lib/maps"
)

// degTorad converts degree to radians.
func degTorad(deg float64) float64 {
	return deg * math.Pi / 180
}

// deg2num converts latlng to tile number
func deg2num(latDeg float64, lonDeg float64, zoom int) (int, int) {
	latRad := degTorad(latDeg)
	n := math.Pow(2.0, float64(zoom))
	xtile := int((lonDeg + 180.0) / 360.0 * n)
	ytile := int((1.0 - math.Log(math.Tan(latRad)+(1/math.Cos(latRad)))/math.Pi) / 2.0 * n)
	return xtile, ytile
}

type TerrainTile struct {
	maps *maps.Maps
	tile *maps.Tile
	x    uint64
	y    uint64
	z    uint64
}

func (self *TerrainTile) X() uint64 {
	return self.x
}

func (self *TerrainTile) Y() uint64 {
	return self.y
}

func (self *TerrainTile) Z() uint64 {
	return self.z
}

func (self *TerrainTile) Fetch() error {
	log.Println("Fetch tile", self.x, self.y, self.z)
	highDPI := false
	tile, err := self.maps.GetTile(maps.MapIDTerrainRGB, self.x, self.y, self.z, maps.MapFormatPngRaw, highDPI)
	if nil != err {
		return err
	}
	self.tile = tile
	return nil
}

func (self *TerrainTile) GetTile() *maps.Tile {
	return self.tile
}

func (self *TerrainTile) Write(fh *os.File) error {

	if nil == self.tile {
		err := self.Fetch()
		if nil != err {
			return err
		}
	}

	fmt.Fprintf(fh, "x,y,z\n")

	self.forEach(func(longitude, latitude, elevation float64) {
		line := fmt.Sprintf("%v,%v,%v\n", longitude, latitude, elevation)
		fmt.Fprintf(fh, line)
	})

	// // y axis needs to be sorted for xyz files
	// for y := 0; y < self.tile.Bounds().Max.Y; y++ {
	// 	for x := 0; x < self.tile.Bounds().Max.X; x++ {
	//
	// 		loc, err := self.tile.PixelToLocation(float64(x), float64(y))
	// 		if nil != err {
	// 			log.Println(err)
	// 			continue
	// 		}
	//
	// 		ll := base.Location{Latitude: loc.Latitude, Longitude: loc.Longitude}
	//
	// 		elevation, err := self.tile.GetAltitude(ll)
	// 		if nil != err {
	// 			log.Println(err)
	// 			continue
	// 		}
	//
	// 		line := fmt.Sprintf("%v,%v,%v\n", loc.Longitude, loc.Latitude, elevation)
	// 		fmt.Fprintf(fh, line)
	//
	// 	}
	// }

	return nil
}

func (self *TerrainTile) forEach(clbk func(float64, float64, float64)) error {
	// y axis needs to be sorted for xyz files
	for y := 0; y < self.tile.Bounds().Max.Y; y++ {
		for x := 0; x < self.tile.Bounds().Max.X; x++ {

			loc, err := self.tile.PixelToLocation(float64(x), float64(y))
			if nil != err {
				log.Println(err)
				continue
			}

			ll := base.Location{Latitude: loc.Latitude, Longitude: loc.Longitude}

			elevation, err := self.tile.GetAltitude(ll)
			if nil != err {
				log.Println(err)
				continue
			}

			clbk(loc.Longitude, loc.Latitude, elevation)
		}
	}
	return nil
}
