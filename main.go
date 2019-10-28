package main

import (
	"encoding/json"
	"fmt"
	"github.com/beefsack/go-astar"
	"github.com/gin-gonic/gin"
	"math"
	"strings"
)

type InputMessage struct {
	WorldInput string
}

type Coordinate struct {
	X, Y int
}

type OutputMessage struct {
	Path []Coordinate
}

var world World

func main() {
		r := gin.Default()
		r.POST("/createWorld", createWorld)
	    r.GET("/getPath", getPath)
		r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

}

func createWorld (context *gin.Context) {
	// Validate the request.
	request := context.Request

	defer request.Body.Close()
	decoder := json.NewDecoder(request.Body)

	var msg InputMessage
	// Validate we got a device resource sucessfully
	if err := decoder.Decode(&msg); err != nil {
		fmt.Printf("Error")
	}

	fmt.Printf("InputMessage is \n%s\n", msg.WorldInput)
	////////////////////
	world = ParseWorld(msg.WorldInput)
	renderedWorld := world.RenderPath([]astar.Pather{})
	fmt.Printf("Input world\n%s", renderedWorld)

	context.JSON(200, gin.H{
		"world": renderedWorld,
	})
}

func getPath (context *gin.Context) {
	fmt.Printf("Input world\n%s", world.RenderPath([]astar.Pather{}))
	path, distance, found := astar.Path(world.From(), world.To())
	fmt.Printf("Distance is %d\n", int(math.Floor(distance)))
	pathCoords := make([]Coordinate, 0)
	response := OutputMessage{
		Path: pathCoords,
	}
	if !found {
		fmt.Printf("Could not find a path")
	} else {
		worldSizeX := len(world)
		worldSizeY := len(world[0])
		fmt.Printf("World dimensions: %d, %d", worldSizeX, worldSizeY)
		fmt.Printf("Resulting path\n")
		initLoc := true
		var currentX, currentY int
		for _, p := range path {
			pT := p.(*Tile)
			if !initLoc {
				nextX := worldSizeX - pT.X - 1
				nextY := worldSizeY - pT.Y - 1
				coordinate := Coordinate{X:nextX - currentX, Y: nextY - currentY}
				//coordinate := Coordinate{X: worldSizeX - pT.X - 1, Y: worldSizeY - pT.Y - 1}
				//coordinate := Coordinate{X: pT.X, Y: pT.Y}
				fmt.Printf("coordinate is %+v", coordinate)
				response.Path = append(response.Path, coordinate)
				fmt.Printf(fmt.Sprintf("%+v", coordinate))
				currentX = nextX
				currentY = nextY
			} else {
				initLoc = false
				currentX = worldSizeX - pT.X - 1
				currentY = worldSizeY - pT.Y - 1
			}
		}
	}
	//////////////////

	context.JSON(200, response)
}
// Kind* constants refer to tile kinds for input and output.
const (
	// KindPlain (.) is a plain tile with a movement cost of 1.
	KindPlain = iota
	// KindRiver (~) is a river tile with a movement cost of 2.
	KindRiver
	// KindMountain (M) is a mountain tile with a movement cost of 3.
	KindMountain
	// KindBlocker (X) is a tile which blocks movement.
	KindBlocker
	// KindFrom (F) is a tile which marks where the path should be calculated
	// from.
	KindFrom
	// KindTo (T) is a tile which marks the goal of the path.
	KindTo
	// KindPath (●) is a tile to represent where the path is in the output.
	KindPath
)

// KindRunes map tile kinds to output runes.
var KindRunes = map[int]rune{
	KindPlain:    '.',
	KindRiver:    '~',
	KindMountain: 'M',
	KindBlocker:  'X',
	KindFrom:     'F',
	KindTo:       'T',
	KindPath:     '●',
}

// RuneKinds map input runes to tile kinds.
var RuneKinds = map[rune]int{
	'.': KindPlain,
	'~': KindRiver,
	'M': KindMountain,
	'X': KindBlocker,
	'F': KindFrom,
	'T': KindTo,
}

// KindCosts map tile kinds to movement costs.
var KindCosts = map[int]float64{
	KindPlain:    1.0,
	KindFrom:     1.0,
	KindTo:       1.0,
	KindRiver:    2.0,
	KindMountain: 3.0,
}

// A Tile is a tile in a grid which implements Pather.
type Tile struct {
	// Kind is the kind of tile, potentially affecting movement.
	Kind int
	// X and Y are the coordinates of the tile.
	X, Y int
	// W is a reference to the World that the tile is a part of.
	W World
}

// PathNeighbors returns the neighbors of the tile, excluding blockers and
// tiles off the edge of the board.
func (t *Tile) PathNeighbors() []astar.Pather {
	neighbors := []astar.Pather{}
	for _, offset := range [][]int{
		{-1, 0},
		{1, 0},
		{0, -1},
		{0, 1},
	} {
		if n := t.W.Tile(t.X+offset[0], t.Y+offset[1]); n != nil &&
			n.Kind != KindBlocker {
			neighbors = append(neighbors, n)
		}
	}
	return neighbors
}

// PathNeighborCost returns the movement cost of the directly neighboring tile.
func (t *Tile) PathNeighborCost(to astar.Pather) float64 {
	//toT := to.(*Tile)
	//return KindCosts[toT.Kind]
	return 1.0
}

// PathEstimatedCost uses Manhattan distance to estimate orthogonal distance
// between non-adjacent nodes.
func (t *Tile) PathEstimatedCost(to astar.Pather) float64 {
	toT := to.(*Tile)
	absX := toT.X - t.X
	if absX < 0 {
		absX = -absX
	}
	absY := toT.Y - t.Y
	if absY < 0 {
		absY = -absY
	}
	return float64(absX + absY)
}

// World is a two dimensional map of Tiles.
type World map[int]map[int]*Tile

// Tile gets the tile at the given coordinates in the world.
func (w World) Tile(x, y int) *Tile {
	if w[x] == nil {
		return nil
	}
	return w[x][y]
}

// SetTile sets a tile at the given coordinates in the world.
func (w World) SetTile(t *Tile, x, y int) {
	if w[x] == nil {
		w[x] = map[int]*Tile{}
	}
	w[x][y] = t
	t.X = x
	t.Y = y
	t.W = w
}

// FirstOfKind gets the first tile on the board of a kind, used to get the from
// and to tiles as there should only be one of each.
func (w World) FirstOfKind(kind int) *Tile {
	for _, row := range w {
		for _, t := range row {
			if t.Kind == kind {
				return t
			}
		}
	}
	return nil
}

// From gets the from tile from the world.
func (w World) From() *Tile {
	return w.FirstOfKind(KindFrom)
}

// To gets the to tile from the world.
func (w World) To() *Tile {
	return w.FirstOfKind(KindTo)
}

// RenderPath renders a path on top of a world.
func (w World) RenderPath(path []astar.Pather) string {
	width := len(w)
	if width == 0 {
		return ""
	}
	height := len(w[0])
	pathLocs := map[string]bool{}
	for _, p := range path {
		pT := p.(*Tile)
		pathLocs[fmt.Sprintf("%d,%d", pT.X, pT.Y)] = true
	}
	rows := make([]string, height)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			t := w.Tile(x, y)
			r := ' '
			if pathLocs[fmt.Sprintf("%d,%d", x, y)] {
				r = KindRunes[KindPath]
			} else if t != nil {
				r = KindRunes[t.Kind]
			}
			rows[y] += string(r)
		}
	}
	return strings.Join(rows, "\n")
}

// ParseWorld parses a textual representation of a world into a world map.
func ParseWorld(input string) World {
	w := World{}
	for y, row := range strings.Split(strings.TrimSpace(input), "\n") {
		for x, raw := range row {
			kind, ok := RuneKinds[raw]
			if !ok {
				kind = KindBlocker
			}
			w.SetTile(&Tile{
				Kind: kind,
			}, x, y)
		}
	}
	return w
}
