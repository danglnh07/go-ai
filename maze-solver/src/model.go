package src

import (
	"fmt"
	"math"
	"strings"
)

// Constant definitions
type Algo string

type Action string

const (
	BFS      Algo = "bfs"
	DFS      Algo = "dfs"
	GBFS     Algo = "gbfs"
	ASTAR    Algo = "astar"
	DIJKSTRA Algo = "dijkstra"

	UP    Action = "up"
	DOWN  Action = "down"
	LEFT  Action = "left"
	RIGHT Action = "right"
	NONE  Action = "none"
)

func IsAlgo(algo string) bool {
	a := Algo(algo)
	return a == BFS || a == DFS || a == GBFS || a == ASTAR || a == DIJKSTRA
}

// The Coordinate struct
type Point struct {
	Row int
	Col int
}

// Square in the maze, which can be either empty (can move to) and wall (cannot move to)
type Square struct {
	Coordinate Point
	IsWall     bool
	Cost       int // The cost to go pass this square (in a maze context, it maybe a an obstacle or something)
}

// Node use for Graph algorithm
type Node struct {
	Index  int // This index is used for priority queue implementation, has nothing to do with the algorithm itself
	Square Square
	Parent *Node
	Action Action
	Cost   int // This cost is used for for calculation in the algorithm, it may change depend on which algo you use
}

// The Manhattan Distance, which simply is the sum of total columns and rows you need to go to read the destination
func (node *Node) ManhattanDistance(dest Point) int {
	return Abs(dest.Col-node.Square.Coordinate.Col) + Abs(dest.Row-node.Square.Coordinate.Row)
}

// The Eudiclian Distance
func (node *Node) EuclidianDistance(dest Point) float64 {
	col2 := math.Pow(float64(dest.Col-node.Square.Coordinate.Col), 2)
	row2 := math.Pow(float64(dest.Row-node.Square.Coordinate.Row), 2)
	return math.Sqrt(col2 + row2)
}

// Solution
type Solution struct {
	Actions []Action
	Path    []Point
}

func (s *Solution) String() string {
	var builder strings.Builder

	// Handle empty path (start = goal)
	if len(s.Path) == 0 || len(s.Actions) == 0 {
		return "Start and goal are the same; no moves required."
	}

	// Assume the first point in Path is reached after the first action
	// The start point is not in Path, so we need the starting coordinate
	// Since Path is built by backtracking from goal, we need the parent of the first Path point
	// However, we don't have the start point directly, so we'll describe from the first move

	for i := 0; i < len(s.Path); i++ {
		action := s.Actions[i]
		if action == NONE {
			continue // Skip NONE action (initial state)
		}
		coord := s.Path[i]
		if i == 0 {
			// First step: imply starting from the previous point
			fmt.Fprintf(&builder, "Move %s to (%d, %d)", action, coord.Row, coord.Col)
		} else {
			// Subsequent steps
			fmt.Fprintf(&builder, ", move %s to (%d, %d)", action, coord.Row, coord.Col)
		}
	}

	// If no valid actions were found, return a message
	if builder.Len() == 0 {
		return "No valid moves in the solution."
	}

	return fmt.Sprintf("Start, %s, reach goal.", builder.String())
}

// Maze struct
type Maze struct {
	Height         int
	Width          int
	Start          Point
	Goal           Point
	Squares        [][]Square // All the squares information in the maze
	CurrentNode    *Node      // The current place we are in
	Solution       Solution   // Maze's solution
	Explored       []Point    // Squares (more specifically, empty square), that we have visited
	ExperimentPath []Point    // The actual path that solver has taken, including incorrect path. Use solely for animation
	Steps          int        // Number of step we have made
	SearchType     Algo       // Which algorithm being used to solve this particular maze
}

// Parse the string maze into Maze struct.
// The structure should be a 2D array, where the start point is 'A', goal is 'B', wall is '#' and empty squares as empty (' ').
func (m *Maze) Load(maze string) error {
	data := strings.TrimSpace(string(maze))
	lines := strings.Split(data, "\n")
	if !strings.Contains(data, "A") || !strings.Contains(data, "B") {
		return fmt.Errorf("need both starting and ending position for the maze")
	}

	// Get the width and height of the maze
	m.Height = len(lines)
	m.Width = len(lines[0])

	// Get maze information (start, goal, squares coordinates)
	var squares [][]Square

	for i, row := range lines {
		var cols []Square

		for j, letter := range row {
			var square Square

			// Check if the letter is valid
			if letter != 'A' && letter != 'B' && letter != ' ' && letter != '#' && !('1' <= letter && letter <= '9') {
				return fmt.Errorf("invalid character")
			}

			square.Coordinate.Row = i
			square.Coordinate.Col = j

			switch {
			case letter == 'A':
				m.Start = Point{Row: i, Col: j}
				square.IsWall = false
				square.Cost = 1
			case letter == 'B':
				m.Goal = Point{Row: i, Col: j}
				square.IsWall = false
				square.Cost = 1
			case letter == ' ':
				square.IsWall = false
				square.Cost = 1
			case letter == '#':
				square.IsWall = true
			case '2' <= letter && letter <= '9':
				square.IsWall = false
				square.Cost = int(letter - '0')
			}

			cols = append(cols, square)
		}

		squares = append(squares, cols)
	}

	m.Squares = squares

	return nil
}

// Get the total of empty squares in the maze
func (maze *Maze) GetEmptySquares() int {
	empty := 0
	for _, row := range maze.Squares {
		for _, sq := range row {
			if !sq.IsWall {
				empty++
			}
		}
	}

	return empty
}

// Universal interface for maze-solver
type Solver interface {
	Add(node *Node)
	ContainsSquare(node *Node) bool
	IsEmpty() bool
	Remove() *Node
	GetNeighbor(node *Node) []*Node
	Solve()
}
