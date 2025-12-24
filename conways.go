// Save In Our Main Package
package main

// Load In Our Necessary Libraries
import (
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"encoding/json"

	"github.com/gorilla/websocket"
)

// Explicitly Call To Utilize Our encoding/json Package
var _ = json.Marshal

// Hold Some Constants For Our Board State
const (
	MATRIX_ROWS = 300
	MATRIX_COLS = 300

	RENDER_FPS        = 50
	RENDER_DEAD_ICON  = '.'
	RENDER_ALIVE_ICON = 'X'

	GENERATION_BLOOM_AMOUNT = 100
)

// Create A Struct For Points On The Grid
type Point struct {
	x, y int
}

// Game State For Our WebSocket
type GameState struct {
	Matrix     [][]string `json:"matrix"`     // Save Our Matrix As String To Avoid Byte Encoding Differences Between Web & Go (ie 64, 32 bit)
	Generation int        `json:"generation"` // Save Away Our Current Generation Count
	Stats      Stats      `json:"stats"`      // Save Away The Statistics Of Our Generation For Client/HTML
}

// Statistics About Our Current Grid State
type Stats struct {
	AliveCells int `json:"aliveCells"`
	DeadCells  int `json:"deadCells"`
	TotalCells int `json:"totalCells"`
}

// WebSocket Upgrader (Converts Our HTTP To A WebSocket Handshake)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Save Away A Mapping Of All Connection To Clients State And If WebSocket Is Actively Communicating
var clients = make(map[*websocket.Conn]bool)

// Ensure Locking To Make Sure No Data Races Between Goroutines
var clientsMutex sync.Mutex

// Create A Broadcast Of Our Current Game State To Our Network Broker
var broadcast = make(chan GameState)

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
Desc: Function Will Help Establish A Proper WebSocket Connection With The
Provided Client. It Will Set Up The Given WebSocket, Ensure Proper Logging
Of Their Connection, And Persist The Connection Until The User Concludes.

Preconditions:

	1.) w Is The Channel In Which We Will Interface Back To The Client With
	2.) r Is The Request Metadata (IE GET, POST, etc.) Of The Given w Channel

Postconditions:

	1.) Will Upgrade The w Response To A WebSocket Pipeline For Continual Updates
	2.) Will Register The Client In Our Internal Registry (clients)
	3.) Will Properly Checkout The Client In clients When Disconnected
	4.) Will Lazily Loop, Pinging The WebSocket For Status Until User Concludes
*/
func handleWebSocket(w http.ResponseWriter, r *http.Request) {

	// Establish A WebSocket Connection
	conn, err := upgrader.Upgrade(w, r, nil)

	// Ensure Our Connection Is Established Before Proceeding...
	if err != nil {
		log.Printf("WebSocket Upgrade Error: %v", err)
		return
	}

	// Ensure Our Connection Concludes After Scope
	defer conn.Close()

	// Register Our Client
	clientsMutex.Lock()
	clients[conn] = true
	clientCount := len(clients)
	clientsMutex.Unlock()
	log.Printf("Client Connected. Total Clients: %d", clientCount)

	// Remove Our Client When Function Returns
	defer func() {
		clientsMutex.Lock()
		delete(clients, conn)
		clientCount := len(clients)
		clientsMutex.Unlock()
		log.Printf("Client Disconnected. Total Clients: %d", clientCount)
	}()

	// Keep Connection Alive And Handle Client Requests
	for {

		// We Don't Really Care About Intake, So Simply Check Just Error Code
		_, _, err := conn.ReadMessage()

		// If Landed On Error, Print State
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket Error: %v", err)
			}
			break
		}

	}

}

/*
Desc: Works As A Goroutine In Which Will Block And Wait Until Status Updates
In Which Will Be Posted To All clients We Have Active. The Message Will Be
Formatted In JSON Format.

Preconditions:

	1.) broadcast Is Formatted In JSON
	2.) Thread Will Be Withheld For Broadcasting

Postconditions:

	1.) Will Send Over gameState To All clients (Via JSON)
	2.) Will Lock On clientsMutex During Transmission To Clients
*/
func handleBroadcast() {

	// Keep Broadcasting Gamestate To Clients
	for {

		// Receive From broadcast Then Assign To Our gameState
		gameState := <-broadcast

		// Lock Our Mutex
		clientsMutex.Lock()

		// Go Through Each Client And Send Over JSON
		for client := range clients {
			err := client.WriteJSON(gameState)

			// If There's An Error, Post And Close
			if err != nil {
				log.Printf("WebSocket Write Error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}

		// Unlock After Posting To Each
		clientsMutex.Unlock()

	}

}

/*
Desc: Utilizing matrix, Will Go Through And Tally All Alive & Dead
Cells Of The matrix; Based Upon Our Constants Of The Alive Icon
If Not RENDER_ALIVE_ICON, Its Assumed As Dead.

Preconditions:

	1.) matrix Is Assumed To Utilize Solely RENDER_ALIVE_ICON & RENDER_DEAD_ICON

Postconditions:

	1.) Will Traverse matrix Tallying States
	2.) Returns Stats Which Explains Whats In matrix
*/
func calculateStats(matrix *[][]byte) Stats {

	// Save Our Tally For All Cells
	alive := 0
	dead := 0

	// Go Through Each Cell
	for rowIndx := range *matrix {

		for colIndx := range (*matrix)[rowIndx] {

			// If Alive, Tally, Else Tally Dead
			if (*matrix)[rowIndx][colIndx] == RENDER_ALIVE_ICON {
				alive++
			} else {
				dead++
			}
		}

	}

	// Return Our Stats Struct Of Our Status
	return Stats{
		AliveCells: alive,
		DeadCells:  dead,
		TotalCells: alive + dead,
	}

}

// Broadcast Game State To All Connected Clients
// Broadcast Game State To All Connected Clients
func broadcastGameState(matrix *[][]byte, generation int) {

	// Quickly Lock And Tally clients Count
	clientsMutex.Lock()
	clientCount := len(clients)
	clientsMutex.Unlock()

	// If No clients Listening, Skip
	if clientCount == 0 {
		log.Printf("No clients connected, skipping broadcast")
		return
	}

	// Compute Our Statistics On The Grid
	stats := calculateStats(matrix)

	// Convert From [][]byte -> [][]string
	stringMatrix := make([][]string, len(*matrix))
	for i := range *matrix {
		stringMatrix[i] = make([]string, len((*matrix)[i]))
		for j := range (*matrix)[i] {
			stringMatrix[i][j] = string((*matrix)[i][j])
		}
	}

	// Create Our gameState JSON Struct
	gameState := GameState{
		Matrix:     stringMatrix, // Send as strings instead of bytes
		Generation: generation,
		Stats:      stats,
	}

	// Send Over Our gameState To Our broadcast Channel, Else Wait
	select {
	case broadcast <- gameState:
	default:
		log.Printf("Broadcast Channel Is Full, Skipping Update")
	}
}

/*
Desc: Provide A Client With Our Stored Webpage.
We Ensure Our Request Is Valid Before Posting The Given Page.

Preconditions:

	1.) w Is The Route We Send Our File Through
	2.) r Is The Given Request Metadataconst

Postconditions:

	1.) Will Serve An Error Code If The Client Requests A Invalid Page
	2.) Will Serve The Client Our Static Main Page If The Pathing Is Valid
*/
func serverHomePage(w http.ResponseWriter, r *http.Request) {

	// If Is Something Other Than A Request To Home Page
	if r.URL.Path != "/" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// Else If Its A Request To Our Home Page, Return It
	http.ServeFile(w, r, "static/index.html")
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

	// Establish Our Handlers For Page Routing
	http.HandleFunc("/", serverHomePage)
	http.HandleFunc("/ws", handleWebSocket)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Start Broadcast Handler
	go handleBroadcast()

	// Start HTTP Server
	go func() {
		log.Println("Server Starting On :8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

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
	generation := 0

	broadcastGameState(&old_matrix, generation)

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
		//print_matrix(&new_matrix)
		old_matrix, new_matrix = new_matrix, old_matrix
		generation++

		// Broadcast Our Updated State
		broadcastGameState(&old_matrix, generation)

		// Sleep For The Amount Of Milliseconds Of RENDER_FPS
		time.Sleep(RENDER_FPS * time.Millisecond)

	}

}
