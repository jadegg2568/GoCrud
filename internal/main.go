package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const dbAddr = "user=user123 password=secretpassword host=localhost port=5432 dbname=gin_db sslmode=disable"

type User struct {
	Id        int       `db:"id" json:"-"` // hidden
	Uuid      uuid.UUID `db:"uuid" json:"uuid"`
	Name      string    `db:"name" json:"name"`
	Password  string    `db:"password" json:"-"` // hidden
	Title     string    `db:"title" json:"title"`
	PhotoURL  string    `db:"photo_url" json:"photo_url"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type CreateUserDto struct {
	Name     string `json:"name" binding:"required"`
	Title    string `json:"title" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UpdateUserDto struct {
	Uuid     string  `db:"uuid" json:"-"`
	Name     *string `db:"name" json:"name,omitempty"`
	Title    *string `db:"title" json:"title,omitempty"`
	PhotoURL *string `db:"photo_url" json:"photo_url,omitempty"`
}

func main() {
	db, err := sqlx.Connect("postgres", dbAddr)
	if err != nil {
		panic("Failed connecting to database: " + err.Error())
	}
	defer db.Close()

	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		uuid UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
		name TEXT NOT NULL,
		title TEXT NOT NULL,
		photo_url TEXT NOT NULL DEFAULT '',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);`

	_, err = db.Exec(schema)
	if err != nil {
		panic("Failed to create table: " + err.Error())
	}

	r := gin.Default()

	// === TEST ===
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Main Page"})
	})

	// === CREATE ===
	r.POST("/users", func(c *gin.Context) {
		var dto CreateUserDto
		if err := c.ShouldBindJSON(&dto); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input: " + err.Error()})
			return
		}

		// Use QueryRowx + RETURNING
		query := `INSERT INTO users (name, title, photo_url) 
				  VALUES ($1, $2, $3) 
				  RETURNING id, uuid, name, title, photo_url, created_at`

		var createdUser User
		err := db.QueryRowx(query, dto.Name, dto.Title, "").StructScan(&createdUser)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to insert user: " + err.Error()})
			return
		}

		c.JSON(201, createdUser)
	})

	// === UPDATE (PATCH) ===
	r.PATCH("/users/:uuid", func(c *gin.Context) {
		var dto UpdateUserDto
		if err := c.ShouldBindJSON(&dto); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input: " + err.Error()})
			return
		}
		dto.Uuid = c.Param("uuid")

		// update only non-nil fields
		query := `
		UPDATE users 
		SET 
			name = COALESCE(:name, name),
			title = COALESCE(:title, title),
			photo_url = COALESCE(:photo_url, photo_url)
		WHERE uuid = :uuid`

		_, err = db.NamedExec(query, dto)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to update user: " + err.Error()})
			return
		}
		c.Status(204) // No Content
	})

	// === GET BY UUID ===
	r.GET("/users/:uuid", func(c *gin.Context) {
		uuidParam := c.Param("uuid")

		query := `SELECT id, uuid, name, title, photo_url, created_at FROM users WHERE uuid = $1`

		var user User
		err := db.Get(&user, query, uuidParam)
		if err != nil {
			c.JSON(404, gin.H{"error": "User not found"})
			return
		}

		c.JSON(200, user)
	})

	// === GET BY NAME (QUERY PARAM) ===
	r.GET("/users", func(c *gin.Context) {
		name := c.Query("name")

		query := `SELECT id, uuid, name, title, photo_url, created_at FROM users WHERE name = $1`

		var user User
		err := db.Get(&user, query, name)
		if err != nil {
			c.JSON(404, gin.H{"error": "User not found"})
			return
		}

		c.JSON(200, user)
	})

	// === DELETE ===
	r.DELETE("/users/:uuid", func(c *gin.Context) {
		uuidParam := c.Param("uuid")

		query := `DELETE FROM users WHERE uuid = $1`

		_, err := db.Exec(query, uuidParam)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to delete user: " + err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "User deleted successfully"})
	})

	// === GET ALL ===
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
