package src

import "slices"

// BFS implementation
type BFSSolver struct {
	Frontier []*Node
	Maze     *Maze
}

// Constructor of BFS solver
func NewBFSSolver(maze *Maze) Solver {
	return &BFSSolver{
		Frontier: make([]*Node, 0),
		Maze:     maze,
	}
}

// Add node into the Frontier slice
func (bfs *BFSSolver) Add(node *Node) {
	// Since this is BFS, we use FIFO
	bfs.Frontier = append(bfs.Frontier, node)
}

// Check if the Frontier containt a node that has the same coordinate as 'node'
func (bfs *BFSSolver) ContainsSquare(node *Node) bool {
	for _, f := range bfs.Frontier {
		if f.Square.Coordinate == node.Square.Coordinate {
			return true
		}
	}

	return false
}

// Check if the Frontier is empty
func (bfs *BFSSolver) IsEmpty() bool {
	return len(bfs.Frontier) == 0
}

// Remove the node out of Frontier
func (bfs *BFSSolver) Remove() *Node {
	if bfs.IsEmpty() {
		return nil
	}

	// Since this is FIFO, we pull out the first element
	node := bfs.Frontier[0]
	bfs.Frontier = bfs.Frontier[1:]
	return node
}

// Get the list of neighbors of the current node
func (bfs *BFSSolver) GetNeighbor(node *Node) []*Node {
	return GetNeighbors(node, bfs.Maze.Width, bfs.Maze.Height, bfs.Maze.Squares)
}

// Solve maze
func (bfs *BFSSolver) Solve() {
	// Create the start node, add it to the frontier slice, and set the current node to start
	start := Node{
		Square: Square{
			Coordinate: bfs.Maze.Start,
			IsWall:     false,
			Cost:       1,
		},
		Parent: nil,
		Action: NONE,
	}
	bfs.Add(&start)
	bfs.Maze.CurrentNode = &start

	// Whenever current node change, we record it into the ExpirementPath slice
	bfs.Maze.ExperimentPath = append(bfs.Maze.ExperimentPath, bfs.Maze.CurrentNode.Square.Coordinate)

	// Make an infinite loop until we found the solution, or stop because we explored all squares without finding a solution
	for {
		// If frontier is empty (which should mean that we have explored every path possible), return
		if bfs.IsEmpty() {
			return
		}

		// Get the current node (by pulling the node from the frontier)
		current := bfs.Remove()
		if current == nil {
			// If current == nil -> len(frontier) = 0 -> return
			return
		}

		bfs.Maze.CurrentNode = current
		bfs.Maze.ExperimentPath = append(bfs.Maze.ExperimentPath, bfs.Maze.CurrentNode.Square.Coordinate)

		//If the current node is the goal
		if bfs.Maze.Goal == current.Square.Coordinate {
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

			bfs.Maze.Solution = Solution{
				Actions: actions,
				Path:    path,
			}

			// Add the current node as explored
			bfs.Maze.Explored = append(bfs.Maze.Explored, current.Square.Coordinate)
			return
		}

		// If we haven't found the solution yet
		bfs.Maze.Explored = append(bfs.Maze.Explored, current.Square.Coordinate)

		// Loop through the neighbors of the current node
		for _, neighbor := range bfs.GetNeighbor(current) {
			// Add neighbor into frontier. Neighbor should only be added if they are not already exists in the frontier
			// and we havent's explored it.
			// Unlike with DFS, in BFS, we will add all the neighbors into Frontier before moving to the next step
			// (backtrack/going deeper)
			if !bfs.ContainsSquare(neighbor) && !slices.Contains(bfs.Maze.Explored, neighbor.Square.Coordinate) {
				bfs.Add(neighbor)
			}
		}
	}
}
