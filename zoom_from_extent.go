package main

import (
	"math"
)

// getZoomLevelFromBbox adapted from
// https://stackoverflow.com/questions/10620515/how-do-i-determine-the-zoom-level-of-a-latlngbounds-before-using-map-fitbounds
func getZoomLevelFromBbox(minlat, maxlat, minlng, maxlng float64, mapWidthPx int, mapHeightPx int) int {
	latFraction := (latRad(maxlat) - latRad(minlat)) / math.Pi
	lngDiff := maxlng - minlng
	lngFraction := lngDiff
	if lngDiff < 0 {
		lngFraction = lngDiff + 360
	}
	lngFraction = lngFraction / 360
	latZoom := zoom(mapHeightPx, TILE_SIZE, latFraction)
	lngZoom := zoom(mapWidthPx, TILE_SIZE, lngFraction)
	if latZoom < lngZoom {
		return int(latZoom)
	}
	return int(lngZoom)
}

func latRad(lat float64) float64 {
	sin := math.Sin(lat * math.Pi / 180)
	radX2 := math.Log((1+sin)/(1-sin)) / 2
	return math.Max(math.Min(radX2, math.Pi), -math.Pi) / 2
}

func zoom(mapPx int, worldPx int, fraction float64) float64 {
	return math.Floor(math.Log(float64(mapPx)/float64(worldPx)/fraction) / math.Ln2)
}
