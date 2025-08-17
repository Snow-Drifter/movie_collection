package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

type Movie struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Format  string `json:"format,omitempty"`
	Edition string `json:"edition,omitempty"`
}

func InitDB() (*Database, error) {
	db, err := sql.Open("sqlite3", "./database/movies.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	createMovieTableSQL := `
	CREATE TABLE IF NOT EXISTS movie (
		movie_id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		format TEXT NOT NULL DEFAULT 'hd'
	);`

	_, err = db.Exec(createMovieTableSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %v", err)
	}

	createMovieEditionTableSQL := `
	CREATE TABLE IF NOT EXISTS movie_edition (
		movie_edition_id INTEGER PRIMARY KEY AUTOINCREMENT,
		movie_id INTEGER NOT NULL,
		edition TEXT NOT NULL,
		FOREIGN KEY (movie_id) REFERENCES movie(movie_id)
	);`

	_, err = db.Exec(createMovieEditionTableSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %v", err)
	}

	database := &Database{db: db}

	err = database.seedInitialData()
	if err != nil {
		log.Printf("Warning: Could not populate database with initial data: %v", err)
	}

	return database, nil
}

func (d *Database) seedInitialData() error {
	var count int

	err := d.db.QueryRow("SELECT COUNT(*) FROM movie").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		var movies []Movie

		jsonData, err := os.ReadFile("./database/movies_owned.json")
		if err != nil {
			return fmt.Errorf("failed to read movies.json: %v", err)
		}

		err = json.Unmarshal(jsonData, &movies)
		if err != nil {
			return fmt.Errorf("failed to unmarshal movies: %v", err)
		}

		for _, movie := range movies {
			var movieID int

			if movie.Format == "" {
				movie.Format = "hd"
			}

			err := d.db.QueryRow("INSERT INTO movie (name, format) VALUES (?, ?) RETURNING movie_id", movie.Name, movie.Format).Scan(&movieID)

			if err != nil {
				return fmt.Errorf("failed to insert movie %s: %v", movie.Name, err)
			}

			if movie.Edition != "" {
				_, err := d.db.Exec("INSERT INTO movie_edition (movie_id, edition) VALUES (?, ?)", movieID, movie.Edition)
				if err != nil {
					return fmt.Errorf("failed to insert movie edition %s: %v", movie.Edition, err)
				}
			}
		}

		log.Println("Seeded database with initial movie data")
	}

	return nil
}

func (d *Database) GetMoviesByName(name string) ([]Movie, error) {
	query := "SELECT movie_id, name, format FROM movie WHERE name LIKE ?"
	rows, err := d.db.Query(query, "%"+name+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []Movie
	for rows.Next() {
		var movie Movie
		err := rows.Scan(&movie.ID, &movie.Name, &movie.Format)
		if err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}

	return movies, nil
}

func (d *Database) GetAllMovies() ([]Movie, error) {
	query := "SELECT movie_id, name, format FROM movie"
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []Movie
	for rows.Next() {
		var movie Movie
		err := rows.Scan(&movie.ID, &movie.Name, &movie.Format)
		if err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}

	return movies, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func GetMoviesByName(name string, db *Database) ([]Movie, error) {
	if name == "" {
		return db.GetAllMovies()
	}
	return db.GetMoviesByName(name)
}
