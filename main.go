package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "fga"
)

var (
	db  *sql.DB
	err error
)

func main() {
	connectionQuery := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err = sql.Open("postgres", connectionQuery)
	if err != nil {
		panic(err)
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Terkoneksi ke database postgres: %s", dbname)

	router := gin.Default()
	router.GET("/books", GetAllBooks)
	router.POST("/books", CreateBook)
	router.PUT("/books/:id", UpdateBook)
	router.GET("/books/:id", GetBookByID)
	router.DELETE("/books/:id", DeleteBook)

	router.Run(":8080")
}

type Book struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	Description string `json:"description"`
}

func GetAllBooks(c *gin.Context) {
	var books []Book
	getAllBooksQuery := `select * from books`

	rows, err := db.Query(getAllBooksQuery)
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var book Book

		err = rows.Scan(&book.ID, &book.Title, &book.Author, &book.Description)
		if err != nil {
			panic(err)
		}

		books = append(books, book)
	}

	c.JSON(http.StatusOK, gin.H{
		"books": books,
	})
}

func CreateBook(c *gin.Context) {
	var newBook Book
	var insertedBook Book

	if err := c.ShouldBindJSON(&newBook); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	insertBookQuery := `insert into books (title, author, description)
						values ($1, $2, $3)
						returning *`

	err = db.QueryRow(insertBookQuery, newBook.Title, newBook.Author, newBook.Description).Scan(&insertedBook.ID, &insertedBook.Title, &insertedBook.Author, &insertedBook.Description)

	if err != nil {
		panic(err)
	}

	c.JSON(http.StatusCreated, gin.H{
		"book": insertedBook,
	})

}

func UpdateBook(c *gin.Context) {
	bookID := c.Param("id")
	var updateBook Book

	if err := c.ShouldBindJSON(&updateBook); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	getBookQuery := `update books
					set title = $2, author = $3, description = $4
					where id = $1;`

	res, err := db.Exec(getBookQuery, bookID, updateBook.Title, updateBook.Author, updateBook.Description)
	if err != nil {
		panic(err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		panic(err)
	}

	if count < 1 {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"message": fmt.Sprintf("Book with id %v is not found", bookID),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("%v row affected. Book with id %v has been succesfully updated", count, bookID),
	})

}

func GetBookByID(c *gin.Context) {
	bookID := c.Param("id")
	var book Book

	getBookQuery := `select * from books where id = $1;`

	err = db.QueryRow(getBookQuery, bookID).Scan(&book.ID, &book.Title, &book.Author, &book.Description)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"message": fmt.Sprintf("Book with id %v is not found", bookID),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"book": book,
	})
}

func DeleteBook(c *gin.Context) {
	bookID := c.Param("id")
	deleteBookQuery := `delete from books where id = $1;`

	res, err := db.Exec(deleteBookQuery, bookID)
	if err != nil {
		panic(err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		panic(err)
	}

	if count < 1 {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"message": fmt.Sprintf("Book with id %v is not found", bookID),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("%v row affected. Book with id %v has been successfully deleted", count, bookID),
	})
}
