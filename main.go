package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
)

type Todo struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

var db *sql.DB

func main() {
	// Buat folder 'data' jika belum ada
	os.MkdirAll("data", os.ModePerm)

	var err error
	db, err = sql.Open("sqlite3", "file:data/database.db?cache=shared&mode=rwc")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Buat tabel jika belum ada
	createTable()

	e := echo.New()

	e.GET("/todos", getTodos)
	e.POST("/todos", addTodo)
	e.GET("/todos/:id", getTodo)                      // Get todo by ID
	e.PUT("/todos/:id", updateTodo)                   // Update todo
	e.DELETE("/todos/:id", deleteTodo)                // Delete todo
	e.PATCH("/todos/:id/completed", updateTodoStatus) // Update completed status

	log.Println("Server is running on port 8080...")
	log.Fatal(e.Start(":8080"))
}

func createTable() {
	sqlStmt := `CREATE TABLE IF NOT EXISTS todos (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT,
        description TEXT,
        completed BOOLEAN DEFAULT FALSE
    );`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
}

func getTodos(c echo.Context) error {
	rows, err := db.Query("SELECT id, name, description, completed FROM todos")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		if err := rows.Scan(&todo.ID, &todo.Name, &todo.Description, &todo.Completed); err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		todos = append(todos, todo)
	}
	return c.JSON(http.StatusOK, todos)
}

func addTodo(c echo.Context) error {
	var todo Todo
	if err := c.Bind(&todo); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Eksekusi INSERT dan ambil ID yang baru dibuat
	result, err := db.Exec("INSERT INTO todos (name, description, completed) VALUES (?, ?, ?)", todo.Name, todo.Description, todo.Completed)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Ambil ID yang baru dibuat
	id, err := result.LastInsertId()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Set ID pada todo sebelum mengembalikannya
	todo.ID = int(id)

	return c.JSON(http.StatusCreated, todo)
}

func getTodo(c echo.Context) error {
	id := c.Param("id")
	var todo Todo
	err := db.QueryRow("SELECT id, name, description, completed FROM todos WHERE id = ?", id).Scan(&todo.ID, &todo.Name, &todo.Description, &todo.Completed)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, echo.Map{"message": "Todo not found"})
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, todo)
}

func updateTodo(c echo.Context) error {
	id := c.Param("id")
	var todo Todo

	// Bind data dari request
	if err := c.Bind(&todo); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Konversi ID dari string ke integer
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid ID")
	}

	// Eksekusi UPDATE
	_, err = db.Exec("UPDATE todos SET name = ?, description = ?, completed = ? WHERE id = ?", todo.Name, todo.Description, todo.Completed, idInt)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Set ID pada todo sebelum mengembalikannya
	todo.ID = idInt

	return c.JSON(http.StatusOK, todo)
}

func deleteTodo(c echo.Context) error {
	id := c.Param("id")
	_, err := db.Exec("DELETE FROM todos WHERE id = ?", id)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func updateTodoStatus(c echo.Context) error {
	id := c.Param("id")
	completed := c.QueryParam("completed") // Ambil status dari query parameter
	status := completed == "true"

	_, err := db.Exec("UPDATE todos SET completed = ? WHERE id = ?", status, id)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"id": id, "completed": status})
}
