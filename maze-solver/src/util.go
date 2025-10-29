package src

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

var (
	// Logger
	LOGGER = slog.New(slog.NewTextHandler(os.Stdout, nil))

	// GIF definitions
	cellSize    = 20
	borderWidth = 2
	palette     = color.Palette{
		color.White,                    // 0: empty/background
		color.Black,                    // 1: wall
		color.RGBA{0, 255, 0, 255},     // 2: start (green)
		color.RGBA{255, 0, 0, 255},     // 3: goal (red)
		color.RGBA{128, 128, 128, 255}, // 4: visited (gray)
		color.RGBA{255, 255, 0, 255},   // 5: cursor (yellow)
		color.RGBA{255, 0, 255, 255},   // 6: solution path (magenta)
		color.RGBA{0, 0, 255, 255},     // 7: border (blue)
		color.RGBA{255, 165, 0, 255},   // 8: weighted squares (orange)
	}
)

// Get neighbor of the current node, which is needed for all algorithms to work
func GetNeighbors(node *Node, width, height int, squares [][]Square) []*Node {
	// Get nodes in order: left (row, col - 1), top (row - 1, col), right (row, col + 1), bottom (row + 1, col)
	// The rol and col start with index 0
	row, col := node.Square.Coordinate.Row, node.Square.Coordinate.Col
	neighbors := []*Node{}

	// Get left node
	if node.Square.Coordinate.Col > 0 && !squares[row][col-1].IsWall {
		neighbors = append(neighbors, &Node{
			Square: squares[row][col-1],
			Action: LEFT,
			Parent: node,
		})
	}

	// Get top node
	if node.Square.Coordinate.Row > 0 && !squares[row-1][col].IsWall {
		neighbors = append(neighbors, &Node{
			Square: squares[row-1][col],
			Action: UP,
			Parent: node,
		})

	}

	// Get right node
	if node.Square.Coordinate.Col < width-1 && !squares[row][col+1].IsWall {
		neighbors = append(neighbors, &Node{
			Square: squares[row][col+1],
			Action: RIGHT,
			Parent: node,
		})
	}

	// Get bottom node
	if node.Square.Coordinate.Row < height-1 && !squares[row+1][col].IsWall {
		neighbors = append(neighbors, &Node{
			Square: squares[row+1][col],
			Action: DOWN,
			Parent: node,
		})
	}

	return neighbors

}

// Create GIF animation for maze solving
func CreateGIF(m *Maze) (*bytes.Buffer, error) {
	// Define the width and height of the maze image
	width := m.Width*cellSize + 2*borderWidth
	height := m.Height*cellSize + 2*borderWidth

	// Create GIF
	g := &gif.GIF{
		LoopCount: 0, // Infinite loop
	}

	// Use a map to track visited points progressively
	visited := make(map[Point]bool)

	// Loop through every square the solver/cursor has moved
	for i := 0; i < len(m.ExperimentPath); i++ {
		current := m.ExperimentPath[i]

		// Mark as visited if not already (first appearance)
		visited[current] = true

		// Create image
		img := image.NewPaletted(image.Rect(0, 0, width, height), palette)

		// Draw background (white)
		draw.Draw(img, img.Bounds(), &image.Uniform{palette[0]}, image.Point{}, draw.Src)

		// Draw border (blue)
		borderRect := image.Rect(borderWidth, borderWidth, width-borderWidth, height-borderWidth)
		draw.Draw(img, borderRect, &image.Uniform{palette[7]}, image.Point{}, draw.Over)

		// Draw base maze (empty white, walls black)
		for row := 0; row < m.Height; row++ {
			for col := 0; col < m.Width; col++ {
				// Square image
				rect := image.Rect(
					col*cellSize+borderWidth,
					row*cellSize+borderWidth,
					(col+1)*cellSize+borderWidth,
					(row+1)*cellSize+borderWidth,
				)

				// Check if this is a wall or empty square
				colIdx := 0 // empty
				if m.Squares[row][col].IsWall {
					colIdx = 1 // wall
				} else if m.Squares[row][col].Cost > 1 {
					colIdx = 8 // weighted square (orange)
				}

				// Draw square
				draw.Draw(img, rect, &image.Uniform{palette[colIdx]}, image.Point{}, draw.Src)

				// Draw cost text for weighted squares (Cost > 1)
				if m.Squares[row][col].Cost > 1 && !m.Squares[row][col].IsWall {
					// Center the text in the cell
					x := col*cellSize + borderWidth + cellSize/4
					y := row*cellSize + borderWidth + cellSize/2
					point := fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 64)}
					drawer := &font.Drawer{
						Dst:  img,
						Src:  image.NewUniform(color.Black),
						Face: basicfont.Face7x13,
						Dot:  point,
					}
					drawer.DrawString(fmt.Sprintf("%d", m.Squares[row][col].Cost))
				}
			}
		}

		// Draw visited (full path taken so far, unique points)
		for p := range visited {
			rect := image.Rect(
				p.Col*cellSize+borderWidth,
				p.Row*cellSize+borderWidth,
				(p.Col+1)*cellSize+borderWidth,
				(p.Row+1)*cellSize+borderWidth,
			)
			draw.Draw(img, rect, &image.Uniform{palette[4]}, image.Point{}, draw.Over)
		}

		// Draw cursor (solver position)
		rect := image.Rect(
			current.Col*cellSize+borderWidth,
			current.Row*cellSize+borderWidth,
			(current.Col+1)*cellSize+borderWidth,
			(current.Row+1)*cellSize+borderWidth,
		)
		draw.Draw(img, rect, &image.Uniform{palette[5]}, image.Point{}, draw.Over)

		// Draw start
		startRect := image.Rect(
			m.Start.Col*cellSize+borderWidth,
			m.Start.Row*cellSize+borderWidth,
			(m.Start.Col+1)*cellSize+borderWidth,
			(m.Start.Row+1)*cellSize+borderWidth,
		)
		draw.Draw(img, startRect, &image.Uniform{palette[2]}, image.Point{}, draw.Over)

		// Draw goal
		goalRect := image.Rect(
			m.Goal.Col*cellSize+borderWidth,
			m.Goal.Row*cellSize+borderWidth,
			(m.Goal.Col+1)*cellSize+borderWidth,
			(m.Goal.Row+1)*cellSize+borderWidth,
		)
		draw.Draw(img, goalRect, &image.Uniform{palette[3]}, image.Point{}, draw.Over)

		g.Image = append(g.Image, img)
		g.Delay = append(g.Delay, 20) // 0.2 seconds per frame
		g.Disposal = append(g.Disposal, gif.DisposalBackground)
	}

	// If solution found, add a final frame with solution path highlighted (no cursor)
	if len(m.Solution.Path) > 0 {
		img := image.NewPaletted(image.Rect(0, 0, width, height), palette)

		// Draw background (white)
		draw.Draw(img, img.Bounds(), &image.Uniform{palette[0]}, image.Point{}, draw.Src)

		// Draw border (blue)
		borderRect := image.Rect(borderWidth, borderWidth, width-borderWidth, height-borderWidth)
		draw.Draw(img, borderRect, &image.Uniform{palette[7]}, image.Point{}, draw.Over)

		// Draw base maze
		for row := 0; row < m.Height; row++ {
			for col := 0; col < m.Width; col++ {
				rect := image.Rect(
					col*cellSize+borderWidth,
					row*cellSize+borderWidth,
					(col+1)*cellSize+borderWidth,
					(row+1)*cellSize+borderWidth,
				)
				colIdx := 0 // empty
				if m.Squares[row][col].IsWall {
					colIdx = 1 // wall
				}
				draw.Draw(img, rect, &image.Uniform{palette[colIdx]}, image.Point{}, draw.Src)
			}
		}

		// Draw all visited (full exploration)
		for p := range visited {
			rect := image.Rect(
				p.Col*cellSize+borderWidth,
				p.Row*cellSize+borderWidth,
				(p.Col+1)*cellSize+borderWidth,
				(p.Row+1)*cellSize+borderWidth,
			)
			draw.Draw(img, rect, &image.Uniform{palette[4]}, image.Point{}, draw.Over)
		}

		// Draw solution path (magenta)
		for _, p := range m.Solution.Path {
			rect := image.Rect(
				p.Col*cellSize+borderWidth,
				p.Row*cellSize+borderWidth,
				(p.Col+1)*cellSize+borderWidth,
				(p.Row+1)*cellSize+borderWidth,
			)
			draw.Draw(img, rect, &image.Uniform{palette[6]}, image.Point{}, draw.Over)
		}

		// Draw start and goal on top
		startRect := image.Rect(
			m.Start.Col*cellSize+borderWidth,
			m.Start.Row*cellSize+borderWidth,
			(m.Start.Col+1)*cellSize+borderWidth,
			(m.Start.Row+1)*cellSize+borderWidth,
		)
		draw.Draw(img, startRect, &image.Uniform{palette[2]}, image.Point{}, draw.Over)

		goalRect := image.Rect(
			m.Goal.Col*cellSize+borderWidth,
			m.Goal.Row*cellSize+borderWidth,
			(m.Goal.Col+1)*cellSize+borderWidth,
			(m.Goal.Row+1)*cellSize+borderWidth,
		)
		draw.Draw(img, goalRect, &image.Uniform{palette[3]}, image.Point{}, draw.Over)

		g.Image = append(g.Image, img)
		g.Delay = append(g.Delay, 300) // 1 second for final frame
		g.Disposal = append(g.Disposal, gif.DisposalBackground)
	}

	buf := new(bytes.Buffer)
	if err := gif.EncodeAll(buf, g); err != nil {
		return nil, err
	}

	return buf, nil
}

func CreateSolutionImage(m *Maze) (*bytes.Buffer, error) {
	// Define the width and height of the maze image
	width := m.Width*cellSize + 2*borderWidth
	height := m.Height*cellSize + 2*borderWidth

	// Create image
	img := image.NewPaletted(image.Rect(0, 0, width, height), palette)

	// Draw background (white)
	draw.Draw(img, img.Bounds(), &image.Uniform{palette[0]}, image.Point{}, draw.Src)

	// Draw border (blue)
	borderRect := image.Rect(borderWidth, borderWidth, width-borderWidth, height-borderWidth)
	draw.Draw(img, borderRect, &image.Uniform{palette[7]}, image.Point{}, draw.Over)

	// Draw base maze (empty white, walls black, weighted orange)
	for row := 0; row < m.Height; row++ {
		for col := 0; col < m.Width; col++ {
			rect := image.Rect(
				col*cellSize+borderWidth,
				row*cellSize+borderWidth,
				(col+1)*cellSize+borderWidth,
				(row+1)*cellSize+borderWidth,
			)
			colIdx := 0 // empty
			if m.Squares[row][col].IsWall {
				colIdx = 1 // wall
			} else if m.Squares[row][col].Cost > 1 {
				colIdx = 8 // weighted square (orange)
			}
			draw.Draw(img, rect, &image.Uniform{palette[colIdx]}, image.Point{}, draw.Src)

			// Draw cost text for weighted squares (Cost > 1)
			if m.Squares[row][col].Cost > 1 && !m.Squares[row][col].IsWall {
				x := col*cellSize + borderWidth + cellSize/4
				y := row*cellSize + borderWidth + cellSize/2
				point := fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 64)}
				drawer := &font.Drawer{
					Dst:  img,
					Src:  image.NewUniform(color.Black),
					Face: basicfont.Face7x13,
					Dot:  point,
				}
				drawer.DrawString(fmt.Sprintf("%d", m.Squares[row][col].Cost))
			}
		}
	}

	// Draw visited squares (gray)
	for _, p := range m.Explored {
		rect := image.Rect(
			p.Col*cellSize+borderWidth,
			p.Row*cellSize+borderWidth,
			(p.Col+1)*cellSize+borderWidth,
			(p.Row+1)*cellSize+borderWidth,
		)
		draw.Draw(img, rect, &image.Uniform{palette[4]}, image.Point{}, draw.Over)
	}

	// Draw solution path (magenta)
	for _, p := range m.Solution.Path {
		rect := image.Rect(
			p.Col*cellSize+borderWidth,
			p.Row*cellSize+borderWidth,
			(p.Col+1)*cellSize+borderWidth,
			(p.Row+1)*cellSize+borderWidth,
		)
		draw.Draw(img, rect, &image.Uniform{palette[6]}, image.Point{}, draw.Over)
	}

	// Draw start (green)
	startRect := image.Rect(
		m.Start.Col*cellSize+borderWidth,
		m.Start.Row*cellSize+borderWidth,
		(m.Start.Col+1)*cellSize+borderWidth,
		(m.Start.Row+1)*cellSize+borderWidth,
	)
	draw.Draw(img, startRect, &image.Uniform{palette[2]}, image.Point{}, draw.Over)

	// Draw goal (red)
	goalRect := image.Rect(
		m.Goal.Col*cellSize+borderWidth,
		m.Goal.Row*cellSize+borderWidth,
		(m.Goal.Col+1)*cellSize+borderWidth,
		(m.Goal.Row+1)*cellSize+borderWidth,
	)
	draw.Draw(img, goalRect, &image.Uniform{palette[3]}, image.Point{}, draw.Over)

	// Draw the weighted squares
	for row := 0; row < m.Height; row++ {
		for col := 0; col < m.Width; col++ {
			// Only draw if this is weighted
			if m.Squares[row][col].Cost > 1 {
				rect := image.Rect(
					col*cellSize+borderWidth,
					row*cellSize+borderWidth,
					(col+1)*cellSize+borderWidth,
					(row+1)*cellSize+borderWidth,
				)

				colIdx := 8 // weighted square (orange)
				draw.Draw(img, rect, &image.Uniform{palette[colIdx]}, image.Point{}, draw.Src)

				// Draw cost text for weighted squares (Cost > 1)
				if m.Squares[row][col].Cost > 1 && !m.Squares[row][col].IsWall {
					x := col*cellSize + borderWidth + cellSize/4
					y := row*cellSize + borderWidth + cellSize/2
					point := fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 64)}
					drawer := &font.Drawer{
						Dst:  img,
						Src:  image.NewUniform(color.Black),
						Face: basicfont.Face7x13,
						Dot:  point,
					}
					drawer.DrawString(fmt.Sprintf("%d", m.Squares[row][col].Cost))
				}
			}
		}
	}

	// Encode as PNG
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %v", err)
	}

	return buf, nil
}

func CreateResultFilename(dir, input, algo, ext string) string {
	return filepath.Join(dir, fmt.Sprintf("%s_%s.%s", input, algo, ext))
}

func Abs(a int) int {
	if a < 0 {
		return -a
	}

	return a
}

func ReadFile(input string) (string, error) {
	data, err := os.ReadFile(input)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}
