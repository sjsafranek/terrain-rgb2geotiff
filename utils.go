package main

import (
	"math"
)

// degTorad converts degree to radians.
func degTorad(deg float64) float64 {
	return deg * math.Pi / 180
}

// deg2num converts latlng to tile number
func deg2num(latDeg float64, lonDeg float64, zoom int) (int, int) {
	lat_rad := degTorad(latDeg)
	n := math.Pow(2.0, float64(zoom))
	xtile := int((lonDeg + 180.0) / 360.0 * n)
	ytile := int((1.0 - math.Log(math.Tan(lat_rad)+(1/math.Cos(lat_rad)))/math.Pi) / 2.0 * n)
	return xtile, ytile
}
