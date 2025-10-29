package src

import (
	"slices"
)

// Maze-solver using DFS
type DFSSolver struct {
	Frontier []*Node
	Maze     *Maze
}

// Constructor of DFS Solver
func NewDFSSolver(maze *Maze) Solver {
	return &DFSSolver{
		Frontier: make([]*Node, 0),
		Maze:     maze,
	}
}

// Add node into the Frontier slice
func (dfs *DFSSolver) Add(node *Node) {
	// Use LIFO since this is DFS
	dfs.Frontier = append(dfs.Frontier, node)
}

// Check if the Frontier contain a node that has the same coordinate as 'node'
func (dfs *DFSSolver) ContainsSquare(node *Node) bool {
	for _, f := range dfs.Frontier {
		if f.Square.Coordinate == node.Square.Coordinate {
			return true
		}
	}

	return false
}

// Check if Frontier is empty
func (dfs *DFSSolver) IsEmpty() bool {
	return len(dfs.Frontier) == 0
}

// Remove the node out of Frontier
func (dfs *DFSSolver) Remove() *Node {
	if dfs.IsEmpty() {
		return nil
	}

	// Since this is LIFO, we remove the last element
	node := dfs.Frontier[len(dfs.Frontier)-1]
	dfs.Frontier = dfs.Frontier[0 : len(dfs.Frontier)-1]
	return node
}

// Get the list of neighbors of the current node
func (dfs *DFSSolver) GetNeighbor(node *Node) []*Node {
	return GetNeighbors(node, dfs.Maze.Width, dfs.Maze.Height, dfs.Maze.Squares)
}

// Solve maze
func (dfs *DFSSolver) Solve() {
	// Create the start node, add it to the frontier slice, and set the current node to start
	start := Node{
		Square: Square{
			Coordinate: dfs.Maze.Start,
			IsWall:     false,
			Cost:       1,
		},
		Parent: nil,
		Action: NONE,
	}
	dfs.Add(&start)
	dfs.Maze.CurrentNode = &start

	// Whenever current node change, we record it into the ExpirementPath slice
	dfs.Maze.ExperimentPath = append(dfs.Maze.ExperimentPath, dfs.Maze.CurrentNode.Square.Coordinate)

	// Make an infinite loop until we found the solution, or stop because we explored all squares without finding a solution
	for {
		// If frontier is empty (which should mean that we have explored every path possible), return
		if dfs.IsEmpty() {
			return
		}

		// Get the current node (by pulling the node from the frontier)
		current := dfs.Remove()
		if current == nil {
			// If current == nil -> len(frontier) = 0 -> return
			return
		}

		dfs.Maze.CurrentNode = current
		dfs.Maze.ExperimentPath = append(dfs.Maze.ExperimentPath, dfs.Maze.CurrentNode.Square.Coordinate)

		//If the current node is the goal
		if dfs.Maze.Goal == current.Square.Coordinate {
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

			dfs.Maze.Solution = Solution{
				Actions: actions,
				Path:    path,
			}

			// Add the current node as explored
			dfs.Maze.Explored = append(dfs.Maze.Explored, current.Square.Coordinate)
			return
		}

		// If we haven't found the solution yet
		dfs.Maze.Explored = append(dfs.Maze.Explored, current.Square.Coordinate)

		// Loop through the neighbors of the current node
		hasNewNeighbor := false
		for _, neighbor := range dfs.GetNeighbor(current) {
			// Add neighbor into frontier. Neighbor should only be added if they are not already exists in the frontier
			// and we havent's explored it.
			// In DFS, we only add the first unvisited neighbor and immediately move on the next step (backtrack/going deeper)
			if !dfs.ContainsSquare(neighbor) && !slices.Contains(dfs.Maze.Explored, neighbor.Square.Coordinate) {
				dfs.Add(neighbor)
				hasNewNeighbor = true
				break
			}
		}

		// If we go into a state that their is no new square to explored (no neighbor that get add to frontier)
		// We have to backtrack to a place that has new path to move
		for !hasNewNeighbor {
			current = current.Parent
			dfs.Maze.ExperimentPath = append(dfs.Maze.ExperimentPath, current.Square.Coordinate)
			for _, neighbor := range dfs.GetNeighbor(current) {
				if !dfs.ContainsSquare(neighbor) && !slices.Contains(dfs.Maze.Explored, neighbor.Square.Coordinate) {
					dfs.Add(neighbor)
					hasNewNeighbor = true
					break // Found new neighbor, no need to check more
				}
			}
		}
	}
}
