[![GoDoc](https://godoc.org/github.com/sjsafranek/mapbox-terrain-rgb2geotiff?status.png)](https://godoc.org/github.com/sjsafranek/mapbox-terrain-rgb2geotiff)
[![Go Report Card](https://goreportcard.com/badge/github.com/sjsafranek/mapbox-terrain-rgb2geotiff)](https://goreportcard.com/report/github.com/sjsafranek/mapbox-terrain-rgb2geotiff)
[![License MIT License](https://img.shields.io/github/license/mashape/apistatus.svg)](http://sjsafranek.github.io/mapbox-terrain-rgb2geotiff/)

# Terrain-RGB 2 GeoTIFF

Fetches Mapbox Terrain-RGB tiles for a geographical extent and converts them into a GeoTiff.

Rewrite of https://github.com/sjsafranek/geotiff_elevation_generator

Uses GDAL to build GeoTIFF.

## Usage
```bash
$ go run *.go \
    -token <mapbox_token> \
    -zoom 13 \
    -minlng -77.004897 \
    -minlat -12.028719 \
    -maxlng -76.965650 \
    -maxlat -11.982242 \
    -o out.tif
```




# TODO
Move build_tiff.sh into go program
