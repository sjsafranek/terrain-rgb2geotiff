package main

import (
	"math"
)

// degTorad converts degree to radians.
func degTorad(deg float64) float64 {
	return deg * math.Pi / 180
}

// deg2num converts latlng to tile number
func deg2num(lat_deg float64, lon_deg float64, zoom int) (int, int) {
	lat_rad := degTorad(lat_deg)
	n := math.Pow(2.0, float64(zoom))
	xtile := int((lon_deg + 180.0) / 360.0 * n)
	ytile := int((1.0 - math.Log(math.Tan(lat_rad)+(1/math.Cos(lat_rad)))/math.Pi) / 2.0 * n)
	return xtile, ytile
}
