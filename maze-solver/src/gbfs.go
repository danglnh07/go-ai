package src

import (
	"container/heap"
	"slices"
)

// Greedy Best First Search implementation
type GBFSSolver struct {
	Frontier PriorityQueue
	Maze     *Maze
}

// GBFS Solver constructor
func NewGBFSSolver(maze *Maze) Solver {
	return &GBFSSolver{
		Frontier: make(PriorityQueue, 0),
		Maze:     maze,
	}
}

// Add node into Frontier
func (gbfs *GBFSSolver) Add(node *Node) {
	gbfs.Frontier.Push(node)
	heap.Init(&gbfs.Frontier)
}

// Check if a node exists in Frontier
func (gbfs *GBFSSolver) ContainsSquare(node *Node) bool {
	for _, f := range gbfs.Frontier {
		if f.Square.Coordinate == node.Square.Coordinate {
			return true
		}
	}

	return false
}

// Check if Frontier is empty
func (gbfs *GBFSSolver) IsEmpty() bool {
	return len(gbfs.Frontier) == 0
}

// Remove a node from Frontier
func (gbfs *GBFSSolver) Remove() *Node {
	// Just like with Dijkstra, we also use priority queue here
	if len(gbfs.Frontier) > 0 {
		return heap.Pop(&gbfs.Frontier).(*Node)
	}

	return nil
}

// Get list of neighbors of a node
func (gbfs *GBFSSolver) GetNeighbor(node *Node) []*Node {
	return GetNeighbors(node, gbfs.Maze.Width, gbfs.Maze.Height, gbfs.Maze.Squares)
}

// Solve maze using GBFS
func (gbfs *GBFSSolver) Solve() {
	// Create the start node, add it to the frontier slice, and set the current node to start
	start := Node{
		Square: Square{
			Coordinate: gbfs.Maze.Start,
			IsWall:     false,
			Cost:       1,
		}, Parent: nil,
		Action: NONE,
	}
	gbfs.Add(&start)
	gbfs.Maze.CurrentNode = &start

	// Whenever current node change, we record it into the ExpirementPath slice
	gbfs.Maze.ExperimentPath = append(gbfs.Maze.ExperimentPath, gbfs.Maze.CurrentNode.Square.Coordinate)

	// Make an infinite loop until we found the solution, or stop because we explored all squares without finding a solution
	for {
		// If frontier is empty (which should mean that we have explored every path possible), return
		if gbfs.IsEmpty() {
			return
		}

		// Get the current node (by pulling the node from the frontier)
		current := gbfs.Remove()
		if current == nil {
			// If current == nil -> len(frontier) = 0 -> return
			return
		}

		gbfs.Maze.CurrentNode = current
		gbfs.Maze.ExperimentPath = append(gbfs.Maze.ExperimentPath, gbfs.Maze.CurrentNode.Square.Coordinate)

		//If the current node is the goal
		if gbfs.Maze.Goal == current.Square.Coordinate {
			// Build the solution
			var (
				actions []Action
				path    []Point
			)

			// Backtracking
			for {
				if current.Parent != nil {
					// Append to the start of the slice since we are backtracking
					actions = append([]Action{current.Action}, actions...)
					path = append([]Point{current.Square.Coordinate}, path...)

					// Set the current node to its parent (backtrack)
					current = current.Parent
				} else {
					// If we reach the solution without passing any square -> Start = Goal, then stop here
					break
				}
			}

			gbfs.Maze.Solution = Solution{
				Actions: actions,
				Path:    path,
			}

			// Add the current node as explored
			gbfs.Maze.Explored = append(gbfs.Maze.Explored, current.Square.Coordinate)
			return
		}

		// If we haven't found the solution yet
		gbfs.Maze.Explored = append(gbfs.Maze.Explored, current.Square.Coordinate)

		// Loop through the neighbors of the current node
		for _, neighbor := range gbfs.GetNeighbor(current) {
			// 1. Add neighbor into frontier. Neighbor should only be added if they are not already exists in the frontier
			// and we havent's explored it.
			// 2. Greedy Best First Search, is almost similar to how Dijkstra works, except on how it calculate the cost.
			// In GBFS, we we assume that the closest neighbor to the goal the local optimal point
			if !gbfs.ContainsSquare(neighbor) && !slices.Contains(gbfs.Maze.Explored, neighbor.Square.Coordinate) {
				// Calculate the Manhattan cost first before adding to the Frontier
				neighbor.Cost = neighbor.ManhattanDistance(gbfs.Maze.Goal)
				gbfs.Add(neighbor)
			}
		}
	}
}
