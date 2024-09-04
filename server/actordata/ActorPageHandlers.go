package actordata

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"

	"database/sql"
)

type ActorEnvironment struct {
	db *sql.DB
	// any other global contextual data we want to give our handlers access to
}

// ActorEnvironment ctor
func NewActorEnvironment(db *sql.DB) *ActorEnvironment {
	env := new(ActorEnvironment)
	env.db = db
	return env
}

/**
* This will be the data structure corresponding to a page on a particular
* actor, so that movie buffs will be able to quickly find the movies they've been
* in (and whatever else is relevant, such as actors and directors they've worked with, etc)
 */
type ActorPage struct {
	Env ActorEnvironment // TODO - evaluate whether this needs to be here

	ActorName string

	MovieCount int

	Biography []byte // this is in byte slice form for the io libraries we're gonna use
}

func obtainActorNameFromURL(urlName string) string {
	actorNameSplit := strings.Split(urlName, "+")
	return strings.Join(actorNameSplit, " ")
}

// TODO - this should use the database going forward
func load(name string, env ActorEnvironment) (*ActorPage, error) {
	// here we should be using the name to make a simple query
	// for the table of actors I'm guessing, then we just build the object up
	resultRows, error := env.db.Query("SELECT actor_name, biography, movie_count FROM actordata WHERE actor_name = $1", name)

	if error != nil {
		log.Printf("Unable to obtain actor's data; actor name: %s", name)
		return nil, error
	}

	// a given actor should only exist once so maybe we can just scan the row now
	var actorName *string
	var biography *[]byte
	var movieCount *int

	resultRows.Next()
	scanError := resultRows.Scan(&actorName, &biography, &movieCount)
	if scanError != nil {
		return nil, scanError
	}

	actorEnvironment := &ActorEnvironment{db: env.db}

	return &ActorPage{Env: *actorEnvironment, ActorName: *actorName, Biography: *biography, MovieCount: *movieCount}, nil
}

/**
* Actor landing page handler
 */
type ActorLandingPageHandler struct {
	env ActorEnvironment
}

func NewActorLandingPageHandler(env ActorEnvironment) *ActorLandingPageHandler {
	handler := new(ActorLandingPageHandler)
	handler.env = env
	return handler
}

func (handler ActorLandingPageHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	actorLandingPageComponent := ActorLandingPage()
	actorLandingPageComponent.Render(context.Background(), responseWriter)
}

/**
* View actor handler
 */
type ViewActorPageHandler struct {
	env ActorEnvironment
}

func NewViewActorHandler(env ActorEnvironment) *ViewActorPageHandler {
	handler := new(ViewActorPageHandler)
	handler.env = env
	return handler
}

func (handler ViewActorPageHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	// this will handle the route '/view/actor/{actor name goes here, + to delineate spaces}'
	actorName := obtainActorNameFromURL(request.URL.Path[len("/actors/view/"):])

	actorPageFile, error := load(actorName, handler.env) // TODO - better path handling for actor test data
	if error != nil {
		actorPageFile = &ActorPage{ActorName: actorName, Biography: []byte("Unknown"), MovieCount: 0}
	}
	actorPageComponent := ActorPageTemplate(actorPageFile.ActorName, string(actorPageFile.Biography), actorPageFile.MovieCount)

	actorPageComponent.Render(context.Background(), responseWriter)
}

/**
* Edit actor handler
 */
type EditActorPageHandler struct {
	Env ActorEnvironment
}

func (handler EditActorPageHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	// this will handle the route '/edit/actor/{actor name goes here}'
	actorName := obtainActorNameFromURL(request.URL.Path[len("/actors/edit/"):])
	actorPageFile, error := load(actorName, handler.Env)
	if error != nil {
		actorPageFile = &ActorPage{ActorName: actorName, Biography: []byte("Unknown"), MovieCount: 0}
	}

	editActorPageComponent := EditActorPageTemplate(actorPageFile.ActorName, string(actorPageFile.Biography), actorPageFile.MovieCount)
	editActorPageComponent.Render(context.Background(), responseWriter)
}

/**
* Save actor handler
 */
type SaveActorHandler struct {
	Env ActorEnvironment
}

func (page *ActorPage) save(env ActorEnvironment) error {
	// TODO - save the actor in the database
	saveQuery := `INSERT INTO actordata (actor_name, biography, movie_count) VALUES ($1, $2, $3) ON CONFLICT (actor_name) DO UPDATE SET actor_name = EXCLUDED.actor_name, biography = EXCLUDED.biography, movie_count = EXCLUDED.movie_count`
	_, upsertActorError := env.db.Exec(saveQuery, page.ActorName, page.Biography, page.MovieCount)
	return upsertActorError
}

func (handler SaveActorHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// get requested actor name from the request
	actorName := obtainActorNameFromURL(request.URL.Path[len("/actors/save/"):])

	// obtain request data (DISABLING NAME CHANGES FOR NOW -> TODO until we add proper update SQL stuff)
	// actorName := request.FormValue("name") // in case the save request includes a name change
	// if (actorName == "") {
	// 	actorName = request.URL.Path[len("/save/actor/"):]
	// }
	biography := request.FormValue("biography") // obtain string corresponding to the body
	var movieCount int
	var atoiError error
	if request.FormValue("moviecount") != "" {
		movieCount, atoiError = strconv.Atoi(request.FormValue("moviecount"))
		if atoiError != nil {
			movieCount = 0
			// TODO - log / actual error handling
			log.Print("Warning: strconv.Atoi() failed on MovieCount field when saving actor ", actorName)
		}
	} else {
		movieCount = 0
		// TODO - read movie count from database (or refactor this so that it gets passed in)
	}
	// build actor page struct from request data
	actorPage := &ActorPage{ActorName: actorName, Biography: []byte(biography), MovieCount: movieCount}

	// save the file component
	saveError := actorPage.save(handler.Env)
	if saveError != nil {
		log.Print("Warning: unable to save actor ", actorName, "; ", saveError.Error())
	}

	// redirect to view the page that just got saved
	newActorName := strings.Join(strings.Split(actorName, " "), "+")
	http.Redirect(writer, request, "/actors/view/"+newActorName, http.StatusFound)
}
