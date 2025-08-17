package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"movies/database"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Template struct {
	tmpl *template.Template
}

func newTemplate() *Template {
	return &Template{
		tmpl: template.Must(template.ParseGlob("views/*.html")),
	}
}

func (t *Template) Render(w io.Writer, name string, data any, c echo.Context) error {
	return t.tmpl.ExecuteTemplate(w, name, data)
}

func main() {
	e := echo.New()

	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	e.Renderer = newTemplate()
	e.Use(middleware.Logger())

	e.GET("/", func(c echo.Context) error {
		movies, err := db.GetAllMovies()
		if err != nil {
			log.Printf("Error getting all movies: %v", err)
			movies = []database.Movie{}
		}

		movieData := struct {
			Movies []database.Movie
		}{
			Movies: movies,
		}

		return c.Render(200, "index.html", movieData)
	})

	e.POST("/movies", func(c echo.Context) error {
		name := c.FormValue("name")

		movieList, err := database.GetMoviesByName(name, db)
		if err != nil {
			log.Printf("Error getting movies: %v", err)
			return c.HTML(500, "<p>Error retrieving movies</p>")
		}

		movies := ""
		for _, movie := range movieList {
			movies += fmt.Sprintf("<p>Name: %s</p><p>Format: %s</p>", movie.Name, movie.Format)
		}

		return c.HTML(200, movies)
	})

	log.Println("Server starting on :8080")
	e.Logger.Fatal(e.Start(":8080"))
}
