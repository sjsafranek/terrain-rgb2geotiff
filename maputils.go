package main

import (
	"math"
)

const earthRadius = float64(6371)

/*
 * The haversine formula will calculate the spherical distance as the crow flies
 * between lat and lon for two given points in km
 */
func Haversine(lonFrom float64, latFrom float64, lonTo float64, latTo float64) (distance float64) {
	var deltaLat = (latTo - latFrom) * (math.Pi / 180)
	var deltaLon = (lonTo - lonFrom) * (math.Pi / 180)

	var a = math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(latFrom*(math.Pi/180))*math.Cos(latTo*(math.Pi/180))*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	var c = 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance = earthRadius * c

	return
}

// https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames#Go
func tile2lon(x uint64, z uint64) float64 {
	return float64(x)/math.Pow(2.0, float64(z))*360.0 - 180
}

func tile2lat(y uint64, z uint64) float64 {
	n := math.Pi - (2.0*math.Pi*float64(y))/math.Pow(2.0, float64(z))
	return math.Atan(math.Sinh(n)) * (180 / math.Pi)
}

//.end

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
