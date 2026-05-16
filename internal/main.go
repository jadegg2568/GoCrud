package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	DBUser     = "user123"
	DBPassword = "secretpassword"
	DBHost     = "localhost"
	DBPort     = 5432
	DBName     = "gin_db"
	DBSSL      = "disable"
)

type User struct {
	Id        int       `json:"id" db:"id"`
	Uuid      uuid.UUID `json:"uuid" db:"uuid"`
	Name      string    `json:"name" db:"name"`
	Title     string    `json:"title" db:"title"`
	PhotoURL  string    `json:"photo_url" db:"photo_url"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func main() {
	r := gin.Default()

	// connection address
	connectAddr := fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=%s",
		DBUser, DBPassword, DBHost, DBPort, DBName, DBSSL)

	// connecting
	db, err := sqlx.Connect("postgres", connectAddr)
	if err != nil {
		panic("Failed connecting to database")
	}
	defer db.Close()

	// create table
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		uuid UUID NOT NULL UNIQUE,
		name TEXT NOT NULL,
		title TEXT NOT NULL,
		photo_url TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);`

	_, err = db.Exec(schema)
	if err != nil {
		panic("Failed to create table: " + err.Error())
	}

	// handlers
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Main Page"})
	})
	r.GET("/add", func(c *gin.Context) {
		exampleUser := User{
			Id:        1,
			Uuid:      uuid.New(),
			Name:      "george",
			Title:     "George",
			PhotoURL:  "someobjectstorageurl",
			CreatedAt: time.Now(),
		}

		query := `INSERT INTO users (id, uuid, name, title, photo_url, created_at) 
		          VALUES (:id, :uuid, :name, :title, :photo_url, :created_at)`

		_, err = db.NamedExec(query, exampleUser)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to insert user: " + err.Error()})
			return
		}

		c.JSON(201, exampleUser)
	})
	r.GET("/all", func(c *gin.Context) {
		query := `SELECT id, uuid, name, title, photo_url, created_at FROM users`

		var users []User
		err := db.Select(&users, query)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to get users: " + err.Error()})
			return
		}
		c.JSON(200, users)
	})

	r.Run(":8080")
}
