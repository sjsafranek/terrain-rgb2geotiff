package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"image"
	"image/draw"
	"image/jpeg"
	"image/png"

	"crypto/tls"
	"net/http"

	"flag"
)

var (
	TILELAYER_URL string
	SAVEFILE      string
	MIN_LAT       float64
	MAX_LAT       float64
	MIN_LNG       float64
	MAX_LNG       float64
	ZOOM          int
	HEIGHT        int
	WIDTH         int
	NUM_WORKERS   int
	TILE_SIZE     int = 256
	NUM_ROWS      int
	NUM_COLS      int
	output        *image.RGBA64
	workwg        sync.WaitGroup
	queue         chan Tile
	MAX_X         int
	MIN_X         int
	MAX_Y         int
	MIN_Y         int
)

var tr = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

var client = &http.Client{
	Timeout:   time.Second * 60,
	Transport: tr,
}

type Tile struct {
	X     int
	Y     int
	Z     int
	Px    int
	Py    int
	Url   string
	Image image.Image
}

func Worker(n int) {
	for tile := range queue {
		start_time := time.Now()
		data := GetTilePngBytesFromUrl(tile.Url)
		tile.Image = BytesToPngImage(data)
		drawTile(output, &tile)
		fmt.Println(n, tile.Z, tile.X, tile.Y, time.Since(start_time))
		workwg.Done()
	}
}

func init() {
	// flag.StringVar(&TILELAYER_URL, "u", "https://a.tile.openstreetmap.org/{z}/{x}/{y}.png", "tile layer url")
	flag.StringVar(&TILELAYER_URL, "u", "http://services.arcgisonline.com/ArcGIS/rest/services/World_Topo_Map/MapServer/tile/{z}/{y}/{x}.png", "tile layer url")
	flag.StringVar(&SAVEFILE, "o", "output.png", "save png file")
	flag.Float64Var(&MIN_LAT, "minlat", -85, "min latitude")
	flag.Float64Var(&MAX_LAT, "maxlat", 85, "max latitude")
	flag.Float64Var(&MIN_LNG, "minlng", -175, "min longitude")
	flag.Float64Var(&MAX_LNG, "maxlng", 175, "max longitude")
	flag.IntVar(&ZOOM, "z", -1, "zoom. This will be automatically calculated if not provided.")
	flag.IntVar(&HEIGHT, "height", 1080, "Image height")
	flag.IntVar(&WIDTH, "width", 1920, "Image height")
	flag.IntVar(&NUM_WORKERS, "w", runtime.NumCPU(), "Number of workers")
	flag.Parse()

	// Calculate zoom if not specified
	if -1 == ZOOM {
		ZOOM = getZoomLevelFromBbox(MIN_LAT, MAX_LAT, MIN_LNG, MAX_LNG, HEIGHT, WIDTH)
	}

	queue = make(chan Tile, NUM_WORKERS*2)
}

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

// xyz
type xyz struct {
	x int
	y int
	z int
}

// GetTileNames
func GetTileNames(minlat, maxlat, minlng, maxlng float64, z int) []xyz {
	tiles := []xyz{}

	// upper right
	ur_tile_x, ur_tile_y := deg2num(maxlat, maxlng, z)
	// lower left
	ll_tile_x, ll_tile_y := deg2num(minlat, minlng, z)

	// Add buffer to make sure output image
	// fills specified height and width.
	for x := ll_tile_x - 2; x < ur_tile_x+2; x++ {
		if x < 0 {
			x = 0
		}
		NUM_ROWS++
		NUM_COLS = 0
		// Add buffer to make sure output image
		// fills specified height and width.
		for y := ur_tile_y - 2; y < ll_tile_y+2; y++ {
			if y < 0 {
				y = 0
			}
			NUM_COLS++
			tiles = append(tiles, xyz{x, y, z})
		}
	}

	for i := range tiles {
		if MAX_X < tiles[i].x {
			MAX_X = tiles[i].x
		}
		if MAX_Y < tiles[i].y {
			MAX_Y = tiles[i].y
		}
	}

	newRect := image.Rectangle{
		Min: image.Point{X: 0, Y: 0},
		Max: image.Point{X: NUM_ROWS * TILE_SIZE, Y: NUM_COLS * TILE_SIZE},
	}
	output = image.NewRGBA64(newRect)

	return tiles
}

// GetTilePngBytesFromUrl requests map tile png from url.
func GetTilePngBytesFromUrl(tile_url string) []byte {
	// Just a simple GET request to the image URL
	// We get back a *Response, and an error
	res, err := client.Get(tile_url)
	if err != nil {
		fmt.Printf("Error http.Get -> %v\n", err)
		return []byte("")
	}

	// We read all the bytes of the image
	// Types: data []byte
	data, err := ioutil.ReadAll(res.Body)

	// You have to manually close the body, check docs
	// This is required if you want to use things like
	// Keep-Alive and other HTTP sorcery.
	defer res.Body.Close()

	if err != nil {
		fmt.Printf("Error ioutil.ReadAll -> %v\n", err)
		return []byte("")
	}

	return data
}

// BytesToPngImage converts bytes to png image struct.
func BytesToPngImage(b []byte) image.Image {
	img, err := png.Decode(bytes.NewReader(b))
	if nil != err {
		img, err = jpeg.Decode(bytes.NewReader(b))
		if nil != err {
			fmt.Println("************")
			fmt.Println(string(b))
			fmt.Println(b)
			fmt.Println("************")
			panic(err)
		}
	}
	return img
}

// clipImage clips image to specified height and width
func clipImage(inImage image.Image) image.Image {
	newRect := image.Rectangle{
		Min: image.Point{X: 0, Y: 0},
		Max: image.Point{X: WIDTH, Y: HEIGHT},
	}
	outImage := image.NewRGBA(newRect)

	var x int
	var y int
	imgX := inImage.Bounds().Max.X
	imgY := inImage.Bounds().Max.Y

	bounds := outImage.Bounds()

	if imgX >= WIDTH {
		x = (imgX - WIDTH) / 2
	} else {
		x = 0
		bounds.Min.X = (WIDTH - imgX) / 2
		bounds.Max.X = imgX + (WIDTH-imgX)/2
	}

	if imgY >= HEIGHT {
		y = (imgY - HEIGHT) / 2
	} else {
		y = 0
		bounds.Min.Y = (HEIGHT - imgY) / 2
		bounds.Max.Y = (HEIGHT - imgY) / 2
		bounds.Max.Y = imgY + (HEIGHT-imgY)/2
	}

	draw.Draw(outImage, bounds, inImage, image.Point{x, y}, draw.Over)
	return outImage
}

/* Draws a tile's image on a target canvas according to pixel position
information in the Tile. */
func drawTile(target *image.RGBA64, tile *Tile) {
	if nil == tile.Image {
		panic("Error with tile image")
	}
	tile_x := tile.Px
	tile_y := tile.Py
	rect := image.Rect(tile_x, tile_y, tile_x+TILE_SIZE, tile_y+TILE_SIZE)
	draw.Draw(target, rect, tile.Image, image.Point{0, 0}, draw.Over)
}

// savePng writes png image to file
func savePng(filename string, img image.Image) {
	// Create a new file and write to it.
	out, err := os.Create(filename)
	if err != nil {
		panic(err)
		os.Exit(1)
	}
	err = png.Encode(out, img)
	if err != nil {
		panic(err)
		os.Exit(1)
	}
}

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

func main() {

	tiles := GetTileNames(MIN_LAT, MAX_LAT, MIN_LNG, MAX_LNG, ZOOM)

	cooked_tiles := 0

	start_time := time.Now()

	fmt.Println("Requesting tiles", time.Since(start_time))

	for i := 0; i < NUM_WORKERS; i++ {
		go Worker(i)
	}

	for _, v := range tiles {

		tile_url := fmt.Sprintf("/%v/%v/%v.png", v.z, v.x, v.y)
		basemap_url := TILELAYER_URL + tile_url
		if strings.Contains(TILELAYER_URL, "{z}") {
			basemap_url = TILELAYER_URL
			basemap_url = strings.Replace(basemap_url, "{z}", fmt.Sprintf("%v", v.z), 1)
			basemap_url = strings.Replace(basemap_url, "{y}", fmt.Sprintf("%v", v.y), 1)
			basemap_url = strings.Replace(basemap_url, "{x}", fmt.Sprintf("%v", v.x), 1)
		}

		workwg.Add(1)
		cooked_tiles++
		queue <- Tile{
			X:   v.x,
			Y:   v.y,
			Z:   v.z,
			Url: basemap_url,
			Px:  (NUM_ROWS - (MAX_X - v.x + 1)) * TILE_SIZE,
			Py:  (NUM_COLS - (MAX_Y - v.y + 1)) * TILE_SIZE,
		}
	}

	close(queue)

	workwg.Wait()

	fmt.Println("Finished recieving tiles", cooked_tiles, time.Since(start_time))

	savePng("./"+SAVEFILE, clipImage(output))

	fmt.Println("Finished merging tiles", cooked_tiles, time.Since(start_time))
}
