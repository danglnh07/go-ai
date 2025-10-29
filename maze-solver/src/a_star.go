package src

import (
	"container/heap"
	"slices"
)

// A* implementation
type AStarSolver struct {
	Frontier PriorityQueue
	Maze     *Maze
}

// A* Solver constructor
func NewAStarSolver(maze *Maze) Solver {
	return &AStarSolver{
		Frontier: make(PriorityQueue, 0),
		Maze:     maze,
	}
}

// Add a node into Frontier
func (astar *AStarSolver) Add(node *Node) {
	astar.Frontier.Push(node)
	heap.Init(&astar.Frontier)
}

// Check if a node exists in Frontier
func (astar *AStarSolver) ContainsSquare(node *Node) bool {
	for _, f := range astar.Frontier {
		if f.Square.Coordinate == node.Square.Coordinate {
			return true
		}
	}

	return false
}

// Check if Frontier is empty
func (astar *AStarSolver) IsEmpty() bool {
	return len(astar.Frontier) == 0
}

// Remove a node from Frontier
func (astar *AStarSolver) Remove() *Node {
	if len(astar.Frontier) > 0 {
		return heap.Pop(&astar.Frontier).(*Node)
	}

	return nil
}

// Get list of neighbors of a node
func (astar *AStarSolver) GetNeighbor(node *Node) []*Node {
	return GetNeighbors(node, astar.Maze.Width, astar.Maze.Height, astar.Maze.Squares)
}

// Solve maze using A*
func (astar *AStarSolver) Solve() {
	// Create the start node, add it to the frontier slice, and set the current node to start
	start := Node{
		Square: Square{
			Coordinate: astar.Maze.Start,
			IsWall:     false,
			Cost:       1,
		}, Parent: nil,
		Action: NONE,
	}
	astar.Add(&start)
	astar.Maze.CurrentNode = &start

	// Whenever current node change, we record it into the ExpirementPath slice
	astar.Maze.ExperimentPath = append(astar.Maze.ExperimentPath, astar.Maze.CurrentNode.Square.Coordinate)

	// Make an infinite loop until we found the solution, or stop because we explored all squares without finding a solution
	for {

		// If frontier is empty (which should mean that we have explored every path possible), return
		if astar.IsEmpty() {
			return
		}

		// Get the current node (by pulling the node from the frontier)
		current := astar.Remove()
		if current == nil {
			// If current == nil -> len(frontier) = 0 -> return
			return
		}

		astar.Maze.CurrentNode = current
		astar.Maze.ExperimentPath = append(astar.Maze.ExperimentPath, astar.Maze.CurrentNode.Square.Coordinate)

		//If the current node is the goal
		if astar.Maze.Goal == current.Square.Coordinate {
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

			astar.Maze.Solution = Solution{
				Actions: actions,
				Path:    path,
			}

			// Add the current node as explored
			astar.Maze.Explored = append(astar.Maze.Explored, current.Square.Coordinate)
			return
		}

		// If we haven't found the solution yet
		astar.Maze.Explored = append(astar.Maze.Explored, current.Square.Coordinate)

		// Loop through the neighbors of the current node
		for _, neighbor := range astar.GetNeighbor(current) {
			// 1. Add neighbor into frontier. Neighbor should only be added if they are not already exists in the frontier
			// and we havent's explored it.
			// 2. A*, is the combination of Dijkstra and GBFS works, its cost calculation basically the cost from the current node
			// to the start node + the estimate cost from current node to the goal
			if !astar.ContainsSquare(neighbor) && !slices.Contains(astar.Maze.Explored, neighbor.Square.Coordinate) {
				// Calculate the cost first before adding to the Frontier
				neighbor.Cost = current.Cost + neighbor.Square.Cost + int(neighbor.EuclidianDistance(astar.Maze.Goal))
				astar.Add(neighbor)
			}
		}

	}
}
