package main

import (
	"image"
)

type Tile struct {
	X     int
	Y     int
	Z     int
	Px    int
	Py    int
	Url   string
	Image image.Image
}

// xyz
type xyz struct {
	x int
	y int
	z int
}
