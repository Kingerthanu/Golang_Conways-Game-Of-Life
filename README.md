# Golang_Conways-Game-Of-Life

Program Is A Simple Representation Of Conway's Game Of Life Utilizing The Terminal For Front-End. The Program Will Generate A Random Number Of 
Bloom Patterns (As Outlined Through Constants) On The Grid And Will Show The Grid In The Terminal. This Program Is Quite Simplistic, But Was To Help Learn 
Golang And How To Handle Goroutines Compared To Other Means Of Mulithreading Like What I've Done In C++.

-----

<img src="https://github.com/user-attachments/assets/6ab13164-1e5f-4b16-b91c-383ecca48976" alt="Cornstarch <3" width="55" height="49"> <img src="https://github.com/user-attachments/assets/6ab13164-1e5f-4b16-b91c-383ecca48976" alt="Cornstarch <3" width="55" height="49"> <img src="https://github.com/user-attachments/assets/6ab13164-1e5f-4b16-b91c-383ecca48976" alt="Cornstarch <3" width="55" height="49"> <img src="https://github.com/user-attachments/assets/6ab13164-1e5f-4b16-b91c-383ecca48976" alt="Cornstarch <3" width="55" height="49">


**The Breakdown:**

The Process Starts By Initializing 2 Matrixes -- _old_matrix_ & _new_matrix_ -- These Two Matrixes Basically Work As Front And Back Buffers If You Think Of It In Rendering Means. The _old_matrix_ Holds The Old Generation State, While The _new_matrix_ Holds The New Generation State. Both Of These Matrixes Are Allocated With The Same Size, As Assumed By The Constant Variables Defined At The Top Of The Main **_conways.go_** File (_MATRIX_ROWS_ & _MATRIX_COLS_). After Allocation, The _old_matrix_ Is Initialized With The Dead Icon Which Is Also Defined In The Top Constants. While, Technically We Don't Need To Do This -- As The First Iteration Will Have The Default Value Of 0 -- It Is Better To Set It Up So We Don't Have 2 Iterations Of Setup (Default Is 0, First Scan Sets All To Dead, THEN We Finally Do Our Logic).

After Initialization To Default State, We Utilize Our `giveLife(...)`. This Function Randomly Makes From [0, _GENERATION_BLOOM_AMOUNT_) Sets Of Bloom Patterns -- These Bloom Patterns Are A Pattern Seen In Conway's Game Of Life And Compound Over Generations To Make Nuanced Patterns That Keep Growing Outwards. These Are Created Randomly Around The Predefined Grid As Expressed Through Our Constant Of _MATRIX_ROWS_ & _MATRIX_COLS_.

After This, We Print This Initial State Before Moving Forward In Our Process.

The Main Interest In This Program Is With Our Go-Routines. This Is Done By Defining Through Our Constants The Amount Of Worker Threads We Wish To Utilize (_WORKER_COUNT_). Each Thread Is Divied Out Its Focused Scope Of Amount Of Rows To Utilize Through _rowsPerWorker_.

In Our Main Loop We Have A Lazy Loop (Which We Break Out With **CTRL + C** [Win-OS] Or **CMND + C** [Mac-OS]) This Loop Sets Up A Barrier-Like Synchronizer, Utilizing A **GoLang Wait Group** Which Makes Each Worker Work On It's Own Group Of Elements In The _old_matrix_. Each Worker Will Call `checkNeighbors(...)` Which Will Check All 8 Surrounding Neighbors Of Each Cell, Determining If It Should Be Alive Or Dead Based On Their Alive Neighbors.

In `checkNeighbors(...)` It Will Save Away The _aliveCount_ Of Neighboring Cells. The Neighboring Cells Will Be Saved In Our Global _CELL_NEIGHBORS_ Constant. For A Given Cell/Element's Neighbor We Will Tally The Amount Which Match As Alive -- As Defined In _RENDER_ALIVE_ICON_ -- After Getting This Tally, We Utilize The Rules Of Conway's Game Of Life To Determine If A Dead Cell Becomes Alive, An Alive Cell Stays Alive Or Dies.

After Each Worker Checks Their Given Range Of Cells, We -- As The Main Thread -- Wait Till All Them Conclude, Then We Print Results Of Our New, Updated Matrix (_new_matrix_) Then We Swap Both, Finally Just Waiting If Necessary To Ensure Proper Framerating.

<img src="https://github.com/user-attachments/assets/33ca5408-42e7-4b88-8ef7-d23816253092" alt="Cornstarch <3" width="55" height="49"> <img src="https://github.com/user-attachments/assets/33ca5408-42e7-4b88-8ef7-d23816253092" alt="Cornstarch <3" width="55" height="49"> <img src="https://github.com/user-attachments/assets/33ca5408-42e7-4b88-8ef7-d23816253092" alt="Cornstarch <3" width="55" height="49"> <img src="https://github.com/user-attachments/assets/33ca5408-42e7-4b88-8ef7-d23816253092" alt="Cornstarch <3" width="55" height="49">

----

<img src="https://github.com/user-attachments/assets/31491277-77c4-45b9-9ea9-a2a46467ca9a" alt="Cornstarch <3" width="55" height="49"> <img src="https://github.com/user-attachments/assets/31491277-77c4-45b9-9ea9-a2a46467ca9a" alt="Cornstarch <3" width="55" height="49"> <img src="https://github.com/user-attachments/assets/31491277-77c4-45b9-9ea9-a2a46467ca9a" alt="Cornstarch <3" width="55" height="49"> <img src="https://github.com/user-attachments/assets/31491277-77c4-45b9-9ea9-a2a46467ca9a" alt="Cornstarch <3" width="55" height="49">

**Features:**

<img src="https://github.com/user-attachments/assets/7414c500-119e-4732-9850-04924f722e38" alt="Cornstarch <3" width="55" height="49"> <img src="https://github.com/user-attachments/assets/7414c500-119e-4732-9850-04924f722e38" alt="Cornstarch <3" width="55" height="49"> <img src="https://github.com/user-attachments/assets/7414c500-119e-4732-9850-04924f722e38" alt="Cornstarch <3" width="55" height="49"> <img src="https://github.com/user-attachments/assets/7414c500-119e-4732-9850-04924f722e38" alt="Cornstarch <3" width="55" height="49">
