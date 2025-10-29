package main

import (
	"flag"
	"fmt"
	"maze-solver/src"
	"os"
	"sync"
	"time"
)

func Solve(solver src.Solver, maze *src.Maze) {
	now := time.Now()
	solver.Solve()
	elapsed := time.Since(now)

	src.LOGGER.Info("Maze solving complete", "algo", maze.SearchType, "second(s)", elapsed.Seconds())
	src.LOGGER.Info("Path length", "algo", maze.SearchType, "val", len(maze.Solution.Path))
	explored := len(maze.Explored)
	coverage := float32(explored) / float32(maze.GetEmptySquares())
	src.LOGGER.Info("Total node explored", "algo", maze.SearchType, "nodes", explored, "coverage", fmt.Sprintf("%.2f%%", coverage))
	fmt.Println("Solution: ")
	fmt.Println(maze.Solution)
}

func SolveWithAlgo(maze *src.Maze) {
	// Create solver based on algo
	var solver src.Solver
	switch maze.SearchType {
	case src.DFS:
		solver = src.NewDFSSolver(maze)
	case src.BFS:
		solver = src.NewBFSSolver(maze)
	case src.DIJKSTRA:
		solver = src.NewDijkstraSolver(maze)
	case src.GBFS:
		solver = src.NewGBFSSolver(maze)
	case src.ASTAR:
		solver = src.NewAStarSolver(maze)
	}

	// Solve
	Solve(solver, maze)
}

func Output(input string, searchType src.Algo, maze *src.Maze) error {
	src.LOGGER.Info("Start creating GIF result. This can take time depend on how large the maze")

	// Create the result image
	img, err := src.CreateSolutionImage(maze)
	if err != nil {
		return err
	}

	output := src.CreateResultFilename(".", input, string(searchType), "png")
	if err = os.WriteFile(output, img.Bytes(), 0644); err != nil {
		return err
	}

	// Create the GIF file
	buf, err := src.CreateGIF(maze)
	if err != nil {
		return err
	}

	output = src.CreateResultFilename(".", input, string(searchType), "gif")
	if err = os.WriteFile(output, buf.Bytes(), 0644); err != nil {
		return err
	}

	src.LOGGER.Info("Create result (image, GIF) successfully", "path", output)
	return nil
}

func SolveAllAlgo(input string) {
	algos := []src.Algo{
		src.DFS, src.BFS, src.DIJKSTRA, src.GBFS, src.ASTAR,
	}

	// Read input from file system
	data, err := src.ReadFile(input)
	if err != nil {
		src.LOGGER.Error("failed to read data from file", "error", err)
		return
	}

	// Run the maze solving in concurrency
	wg := sync.WaitGroup{}

	for _, algo := range algos {
		wg.Add(1)
		go func(mazeInput string, searchType src.Algo) {
			defer wg.Done()

			// Load the maze

			maze := src.Maze{SearchType: searchType}
			if err := maze.Load(mazeInput); err != nil {
				src.LOGGER.Error("Failed to load maze", "algo", searchType, "error", err)
				return
			}

			// Solve maze
			SolveWithAlgo(&maze)

			// Create the result image
			output := src.CreateResultFilename(".", input, string(searchType), "png")
			src.LOGGER.Info("Start creating image result. This can take time depend on how large the maze")
			img, err := src.CreateSolutionImage(&maze)
			if err != nil {
				return
			}

			if err = os.WriteFile(output, img.Bytes(), 0644); err != nil {
				return
			}

			// Output GIF
			src.LOGGER.Info("Start creating GIF result. This can take time depend on how large the maze")

			// Create the GIF file
			buf, err := src.CreateGIF(&maze)
			if err != nil {
				src.LOGGER.Error("Failed to create GIF", "algo", searchType, "error", err)
				return
			}

			// Write to file system
			output = src.CreateResultFilename(".", input, string(searchType), "gif")
			if err = os.WriteFile(output, buf.Bytes(), 0644); err != nil {
				src.LOGGER.Error("Failed to write GIF result to file system", "algo", searchType, "error", err)
			}

			src.LOGGER.Info("Create GIF successfully", "path", output)
		}(data, algo)
	}

	wg.Wait()
	src.LOGGER.Info("All algos complete")
}

func main() {
	// Get the parameters
	var input, searchType string
	flag.StringVar(&input, "maze", "mazes/maze.txt", "The maze input file")
	flag.StringVar(&searchType, "search", "", "The search algorithm") // If empty, solve the maze with all algorithms
	flag.Parse()

	// Check for searchType value
	switch searchType {
	case "":
		SolveAllAlgo(input)
	default:
		if !src.IsAlgo(searchType) {
			src.LOGGER.Warn("Unsupported algorithm")
			return
		}
		// Read input from file system
		data, err := src.ReadFile(input)
		if err != nil {
			src.LOGGER.Error("failed to read data from file", "error", err)
			return
		}

		algo := src.Algo(searchType)
		maze := src.Maze{SearchType: algo}
		if err := maze.Load(data); err != nil {
			src.LOGGER.Error("Failed to load maze", "error", err)
			return
		}

		SolveWithAlgo(&maze)

		fmt.Print("Do you want to ouput GIF (y/n): ")
		var confirm string
		fmt.Scanln(&confirm)

		if confirm == "y" {
			if err := Output(input, maze.SearchType, &maze); err != nil {
				src.LOGGER.Error("Failed to output results", "error", err)
				return
			}
		}
	}
}
