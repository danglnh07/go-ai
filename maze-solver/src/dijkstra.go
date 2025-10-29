package src

import (
	"container/heap"
	"slices"
)

// Dijkstra implementation
type DijkstraSolver struct {
	Frontier PriorityQueue
	Maze     *Maze
}

// Constructor of DijkstraSolver
func NewDijkstraSolver(maze *Maze) Solver {
	return &DijkstraSolver{
		Frontier: make([]*Node, 0),
		Maze:     maze,
	}
}

// Add node into Frontier
func (d *DijkstraSolver) Add(node *Node) {
	d.Frontier.Push(node)
	heap.Init(&d.Frontier)
	// d.Frontier = append(d.Frontier, node)
}

// Check if a node exists in Frontier
func (d *DijkstraSolver) ContainsSquare(node *Node) bool {
	for _, f := range d.Frontier {
		if f.Square.Coordinate == node.Square.Coordinate {
			return true
		}
	}

	return false
}

// Check if Frontier is empty
func (d *DijkstraSolver) IsEmpty() bool {
	return len(d.Frontier) == 0
}

// Remove a node from Frontier
func (d *DijkstraSolver) Remove() *Node {
	// For Dijkstra, we would want to take the node which the smallest distance to the start node.
	// Since we always pull the smallest node, the order does not matter
	// sort.Slice(d.Frontier, func(i, j int) bool {
	// 	return d.Frontier[i].Cost < d.Frontier[j].Cost
	// })
	// node := d.Frontier[0]
	// d.Frontier = d.Frontier[1:]
	// return node

	if len(d.Frontier) > 0 {
		return heap.Pop(&d.Frontier).(*Node)
	}

	return nil
}

// Get list of neighbors of a node
func (d *DijkstraSolver) GetNeighbor(node *Node) []*Node {
	return GetNeighbors(node, d.Maze.Width, d.Maze.Height, d.Maze.Squares)
}

// Solve maze using Dijkstra
func (d *DijkstraSolver) Solve() {
	// Create the start node, add it to the frontier slice, and set the current node to start
	start := Node{
		Square: Square{
			Coordinate: d.Maze.Start,
			IsWall:     false,
			Cost:       1,
		}, Parent: nil,
		Action: NONE,
	}
	d.Add(&start)
	d.Maze.CurrentNode = &start

	// Whenever current node change, we record it into the ExpirementPath slice
	d.Maze.ExperimentPath = append(d.Maze.ExperimentPath, d.Maze.CurrentNode.Square.Coordinate)

	// Make an infinite loop until we found the solution, or stop because we explored all squares without finding a solution
	for {
		// If frontier is empty (which should mean that we have explored every path possible), return
		if d.IsEmpty() {
			return
		}

		// Get the current node (by pulling the node from the frontier)
		current := d.Remove()
		if current == nil {
			// If current == nil -> len(frontier) = 0 -> return
			return
		}

		d.Maze.CurrentNode = current
		d.Maze.ExperimentPath = append(d.Maze.ExperimentPath, d.Maze.CurrentNode.Square.Coordinate)

		//If the current node is the goal
		if d.Maze.Goal == current.Square.Coordinate {
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

			d.Maze.Solution = Solution{
				Actions: actions,
				Path:    path,
			}

			// Add the current node as explored
			d.Maze.Explored = append(d.Maze.Explored, current.Square.Coordinate)
			return
		}

		// If we haven't found the solution yet
		d.Maze.Explored = append(d.Maze.Explored, current.Square.Coordinate)

		// Loop through the neighbors of the current node
		for _, neighbor := range d.GetNeighbor(current) {
			// 1. Add neighbor into frontier. Neighbor should only be added if they are not already exists in the frontier
			// and we havent's explored it.
			// 2. Unlike both DFS and BFS, Dijkstra care about cost, so we have to calculate it before adding to Frontier.
			// Unlike normal Dijkstra, this maze is a positive node-weighted graph, so the node we pick is likely to be optimal,
			// no need to update the cost. For example:
			// 2.1. In an edge-weighted graph: A -> B take 10 cost. A -> C take 2, C -> B take 5. Then A -> C -> B is the optimal path.
			// In the case that B get added first (cost = 10), we have to update its cost later (cost = 2 + 5 = 7)
			// 2.2. In node-weighted graph, since the cost always positive, there is no way that A + B > A + B + C, so updating is
			// unnecessary. It would be a different problem if the node's weight can be negative though.
			if !d.ContainsSquare(neighbor) && !slices.Contains(d.Maze.Explored, neighbor.Square.Coordinate) {
				// Calculate the Manhattan cost first before adding to the Frontier
				neighbor.Cost = current.Cost + neighbor.Square.Cost
				d.Add(neighbor)
			}
		}
	}
}
