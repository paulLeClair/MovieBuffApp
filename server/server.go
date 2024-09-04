package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"database/sql"

	"github.com/lib/pq"

	"server/actordata"
	server "server/homepage"
	"server/moviedata"
	"server/oscars"
)

// NOTE -> ultra basic unsecured DB connection (DON'T PUT ANYTHING SENSITIVE HERE)
func connectToPostgresTestDatabase() *sql.DB {
	const (
		host     = "localhost"
		port     = 5432
		user     = "postgres"
		password = "admin"
		sslmode  = "disable"
	)
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=%s",
		host, port, user, password, sslmode)

	database, error := sql.Open("postgres", connectionString)
	if error != nil {
		// TODO - better error handling lol
		log.Fatal("Unable to form initial connection with postgresql!")
	}

	_, createDbError := database.Exec(`CREATE DATABASE moviebuffdb`)
	pqError := createDbError.(*pq.Error)
	if createDbError != nil && pqError.Code != "42P04" { // error code 42P04: db already exists, which is okay
		log.Print("Create database returned an error: ", createDbError.Error())
	}

	database, error = sql.Open("postgres", connectionString+" dbname=moviebuffdb")
	if error != nil {
		log.Fatal("Unable to reconnect to postgresql with dbname specified!")
	}

	return database
}

/*
* This is gonna be janky while the project is in its infancy;
* It will check every time the server starts that all the data
* is good to go. For now that just means that the few test files I'll write
* are being shoved in there.
* Basically I'm just going to iterate over all the text files
* in the assets/temp_test_files/actors directory (in this project)
* and add them to the table if they're not already there.
 */
func validateDatabaseState(db *sql.DB) {
	// I guess for now I'll just check if our singular
	// "actors" relation exists... if not, then it'll be created
	_, error := db.Exec(`CREATE TABLE IF NOT EXISTS actordata (
		actor_name varchar(50) NOT NULL,
		biography varchar(500) NOT NULL,
		movie_count integer NOT NULL DEFAULT '0',
		PRIMARY KEY (actor_name)
	)`)

	if error != nil {
		log.Fatal("Failure during database validation: unable to create actordata table")
	}

	// now we need to iterate over the files...
	testFilesDir := "assets/temp_test_files/actors/"
	testFiles, readTestFilesError := os.ReadDir(testFilesDir)

	if readTestFilesError != nil {
		log.Fatal("Failure during database validation: unable to obtain list of test files")
	}

	for _, file := range testFiles {
		log.Print("Processing file: ", file.Name())
		if file.IsDir() {
			log.Print("Warning: directory ", file.Name(), " has been skipped inside ", testFilesDir)
			continue
		}

		// now we want to read in the fields from each file, using a simple CSV file format for now
		testFilePath := testFilesDir + file.Name()
		openedFile, openFileError := os.Open(testFilePath)
		if openFileError != nil {
			log.Fatal("Unable to open file ", testFilePath)
		}
		newReader := csv.NewReader(openedFile)
		newReader.Comma = '|'
		csvLines, csvError := newReader.ReadAll()
		if csvError != nil {
			log.Fatal("Unable to read file ", testFilePath)
		}

		var actorsFromFile []actordata.ActorPage
		for _, line := range csvLines {
			actorName := line[0]
			biography := []byte(line[1])
			movieCount, atoiError := strconv.Atoi(line[2])
			if atoiError != nil {
				log.Fatal("Unable to convert csv line to integer!")
			}

			actorPageEnv := actordata.NewActorEnvironment(db)
			actorPage := &actordata.ActorPage{Env: *actorPageEnv, ActorName: actorName, Biography: biography, MovieCount: movieCount}
			actorsFromFile = append(actorsFromFile, *actorPage)
		}

		// NOTE -> this naive implementation results in the data being overwritten every time the server starts;
		// (TODO -> obtain actor data in a completely different way (likely a free third-party API that we'll sample from))
		for _, actor := range actorsFromFile {
			query := `INSERT INTO actordata (actor_name, biography, movie_count) VALUES ($1, $2, $3) ON CONFLICT (actor_name) DO UPDATE SET biography = EXCLUDED.biography, movie_count = EXCLUDED.movie_count`
			insertResult, insertError := db.Exec(query, actor.ActorName, actor.Biography, actor.MovieCount)
			if insertError != nil {
				log.Fatal("Unable to insert actor ", actor.ActorName)
			}

			rowsAffected, rowsAffectedError := insertResult.RowsAffected()
			if rowsAffectedError != nil {
				log.Fatal("Unable to determine rows affected by insert of row for actor ", actor.ActorName)
			}
			if rowsAffected != 1 {
				log.Print("Warning: unexpected number of rows were affected by insertion of row for actor ", actor.ActorName, "; Rows Affected: ", rowsAffected)
			}
		}
	}
}

func staticHandler(responseWriter http.ResponseWriter, request *http.Request) {
	path := request.URL.Path
	if strings.HasSuffix(path, ".ico") {
		responseWriter.Header().Set("Content-Type", "image/x-icon")
	}
	if strings.HasSuffix(path, ".css") {
		responseWriter.Header().Set("Content-Type", "text/css")
	}
	if strings.HasSuffix(path, ".js") {
		responseWriter.Header().Set("Content-Type", "text/javascript")
	}
	filePath := path[len("/static/"):]
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Warning! Unable to read file %s", filePath)
		http.Error(responseWriter, "Unable to read static file", http.StatusInternalServerError)
		return
	}
	responseWriter.Write(fileData)
}

func placeholderHomepageHandler(responseWriter http.ResponseWriter, request *http.Request) {
	homepageComponent := server.HomePageTemplate()
	homepageComponent.Render(context.Background(), responseWriter)
}

func setupActorHandlers(database *sql.DB) {
	environment := actordata.NewActorEnvironment(database)

	landingPageHandler := actordata.NewActorLandingPageHandler(*environment)
	http.Handle("/actors", landingPageHandler)

	viewActorHandler := actordata.NewViewActorHandler(*environment)
	http.Handle("/actors/view/", viewActorHandler)

	editActorHandler := &actordata.EditActorPageHandler{Env: *environment}
	http.Handle("/actors/edit/", editActorHandler)

	saveActorHandler := &actordata.SaveActorHandler{Env: *environment}
	http.Handle("/actors/save/", saveActorHandler)
}

func main() {
	// set up DB connection
	database := connectToPostgresTestDatabase()
	if database == nil {
		log.Fatal("Database connection failed.")
	}
	validateDatabaseState(database)

	http.HandleFunc("/static/", staticHandler)

	http.HandleFunc("/", placeholderHomepageHandler)

	setupActorHandlers(database)

	// set up oscars handler (content forthcoming!)
	oscarsHandler := oscars.NewOscarsHandler()
	http.Handle("/oscars", oscarsHandler)

	// set up movie handlers (content forthcoming!)
	http.HandleFunc("/movies", moviedata.LandingPageHandler)

	// TODO:
	// http.HandleFunc("/movies/view/", moviedata.ViewMoviePageHandler)
	// http.HandleFunc("/movies/edit/", moviedata.EditMoviePageHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
