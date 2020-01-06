package main

import (
	"errors"
	"fmt"
)

// https://docs.mapbox.com/help/troubleshooting/access-elevation-data/
// max is zoom 15
const (
	DEFAULT_ZOOM int = 10
	ZOOM_MAX     int = 15
	ZOOM_MIN     int = 1
	MAX_TILES    int = 100
)

var (
	ErrorViewNotSet     error = errors.New("View not set")
	ErrorZoomOutOfRange error = fmt.Errorf("Must supply a map zoom (%v to %v)", ZOOM_MIN, ZOOM_MAX)
)
