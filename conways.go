// Save In Our Main Package
package main

// Load In Our Necessary Libraries
import (
	"math/rand"
	"strings"
	"sync"
	"time"
)

// Hold Some Constants For Our Board State
const (
	MATRIX_ROWS = 300
	MATRIX_COLS = 300

	RENDER_FPS        = 40
	RENDER_DEAD_ICON  = '.'
	RENDER_ALIVE_ICON = 'X'

	GENERATION_BLOOM_AMOUNT = 100
)

// Create A Struct For Points On The Grid
type Point struct {
	x, y int
}

// Hold Away Considered Neighbors
var CELL_NEIGHBORS = []Point{
	{-1, 0},
	{1, 0},
	{0, -1},
	{0, 1},
	{-1, -1},
	{-1, 1},
	{1, 1},
	{1, -1},
}

/*
Desc: Utilized For Rendering A Given 2D Matrix To The Console.

Preconditions:

	1.) Expected To Hold More Than 1 Element

Postconditions:

	1.) Will Move The Cursor To The Top To Help With Terminal Jittering
	2.) Will Output matrix's Contents To The Terminal With One Space Between Elements
*/
func print_matrix(matrix *[][]byte) {

	// Create A String Builder To Allow Us A 1-Call To Print
	var sb strings.Builder

	// Reserve The Amount For Our Given matrix's Size
	sb.Grow(MATRIX_ROWS*(MATRIX_COLS*2+1) + 10)

	// Move The Cursor To The Top Of The Matrix
	sb.WriteString("\033[H")

	// Go Through Each Element And Add It's Contents
	for rowIndx := range *matrix {

		for colIndx := range (*matrix)[rowIndx] {
			sb.WriteByte((*matrix)[rowIndx][colIndx])
			sb.WriteByte(' ')
		}

		sb.WriteByte('\n')
	}

	// Now Print The Build String
	print(sb.String())

}

/*
Desc: Will Look At cell's Point And Tally It's Neighbors, Determining If cell Will Become
Dead Or Alive Based On Conway's Game Of Life's Logic Based On The State Of The old_matrix,
Updating new_matrix With The New State Of cell.

Preconditions:

	1.) cell Is Assumed In Range Of old_matrix & new_matrix
	2.) alive Is The Current State Of cell (Alive Or Dead)
	3.) old_matrix & new_matrix Are The Same Size

Postconditions:

	1.) Will Update cell In new_matrix With The New State Of cell
*/
func checkNeighbors(cell Point, alive bool, old_matrix *[][]byte, new_matrix *[][]byte) {

	// Save Away The Alive Count Of Neighbors
	var aliveCount uint = 0

	// Go Through Each Neighbors Difference From This cell
	for _, dPoint := range CELL_NEIGHBORS {

		// Get The New Coordinate
		nx, ny := cell.x+dPoint.x, cell.y+dPoint.y

		// Ensure It's In Range Of Our Matrix
		if -1 < nx && nx < MATRIX_ROWS && -1 < ny && ny < MATRIX_COLS {

			// If It's Alive, Add It To The Tally For cell
			if (*old_matrix)[nx][ny] == RENDER_ALIVE_ICON {
				aliveCount++
			}

		}

	}

	// If The Node Was Alive, Look If Overpopulated Or Such
	if alive {

		// If Underpopulated, Die
		if aliveCount != 2 && aliveCount != 3 {
			(*new_matrix)[cell.x][cell.y] = RENDER_DEAD_ICON
		} else {
			(*new_matrix)[cell.x][cell.y] = RENDER_ALIVE_ICON
		}

	} else {
		// If Dead And Has Neighbors, Reproduce At This Node
		if aliveCount == 3 {
			(*new_matrix)[cell.x][cell.y] = RENDER_ALIVE_ICON
		} else {
			(*new_matrix)[cell.x][cell.y] = RENDER_DEAD_ICON
		}
	}

}

/*
Desc: Will Randomly Populate matrix With Bloom Pattern Of Conway's Game Of Life,
Creating GENERATION_BLOOM_AMOUNT Of Blooms On The Matrix.

Preconditions:

	1.) matrix's Size Is The Same As MATRIX_ROWS & MATRIX_COLS

Postconditions:

	1.) Will Randomly Pick A Number From [0, GENERATION_BLOOM_AMOUNT] And
	Generate That Many Bloom Patterns
*/
func giveLife(matrix *[][]byte) {

	// Save Away The Random Amount Of Blooms To Make
	var max int = rand.Intn(GENERATION_BLOOM_AMOUNT)

	// Save Away The Row And Column We Will Add From
	var baseRow, baseCol int

	// Go Through Each Bloom
	for indx := 0; indx < max; indx++ {

		// Grab The Random Row And Column To Add To (In Range)
		baseRow = rand.Intn(MATRIX_ROWS-10) + 4
		baseCol = rand.Intn(MATRIX_COLS-10) + 4

		// Place Bloom Pattern At Random Position
		(*matrix)[baseRow][baseCol+1] = RENDER_ALIVE_ICON   // Top center
		(*matrix)[baseRow+1][baseCol+3] = RENDER_ALIVE_ICON // Middle right
		(*matrix)[baseRow+2][baseCol] = RENDER_ALIVE_ICON   // Bottom left
		(*matrix)[baseRow+2][baseCol+1] = RENDER_ALIVE_ICON // Bottom center-left
		(*matrix)[baseRow+2][baseCol+4] = RENDER_ALIVE_ICON // Bottom right+1
		(*matrix)[baseRow+2][baseCol+5] = RENDER_ALIVE_ICON // Bottom right+2
		(*matrix)[baseRow+2][baseCol+6] = RENDER_ALIVE_ICON // Bottom right+3
	}

}

func main() {

	// Create Our Matrixes
	old_matrix := [][]byte{}
	old_matrix = make([][]byte, MATRIX_ROWS)
	for indx := range old_matrix {
		old_matrix[indx] = make([]byte, MATRIX_COLS)
	}
	new_matrix := make([][]byte, MATRIX_ROWS)
	for rowIndx := range new_matrix {
		new_matrix[rowIndx] = make([]byte, MATRIX_COLS)
	}
	// Set To Dead Initially
	for rowIndx := range old_matrix {
		for colIndx := range old_matrix[rowIndx] {
			old_matrix[rowIndx][colIndx] = RENDER_DEAD_ICON
		}
	}

	// Populate Our matrix
	giveLife(&old_matrix)

	// Print It Now
	print_matrix(&old_matrix)

	// Set Our Worker Count As Well As Rows Each Will Be Relegated
	numWorkers := 8
	rowsPerWorker := MATRIX_ROWS / numWorkers

	// Lazy Loop Through
	for {

		// Create Our Barrier Wait Group
		var wg sync.WaitGroup
		for worker := 0; worker < numWorkers; worker++ {

			// Add This Worker
			wg.Add(1)

			// Get The Rows Each Is Relegated
			startRow := worker * rowsPerWorker
			endRow := startRow + rowsPerWorker

			// If Is Last Worker, Use All The Remainder
			if worker == numWorkers-1 {
				endRow = MATRIX_ROWS
			}

			// Set Our Function For Our Workers
			go func(start, end int) {

				// Defer To Wait Group That This Is Done
				defer wg.Done()

				// Go Through Each In Divied Range
				for rowIndx := start; rowIndx < end; rowIndx++ {

					for colIndx := range old_matrix[rowIndx] {

						checkNeighbors(Point{rowIndx, colIndx}, old_matrix[rowIndx][colIndx] == RENDER_ALIVE_ICON, &old_matrix, &new_matrix)

					}

				}
			}(startRow, endRow)

		}

		// Wait Until All Workers Conclude
		wg.Wait()

		// Then Print And Swap
		print_matrix(&new_matrix)
		old_matrix, new_matrix = new_matrix, old_matrix

		// Sleep For The Amount Of Milliseconds Of RENDER_FPS
		time.Sleep(RENDER_FPS * time.Millisecond)

	}

}
