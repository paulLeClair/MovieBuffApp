package moviedata

import (
	"context"
	"net/http"
)

/**
* This will contain all the relevant information for a particular movie! It's movie time!
* To simplify things, I'll get the actor stuff going first and then come back to this
 */
type MoviePage struct {
}

// for now, I'll just add a simple function handler for the landing page
func LandingPageHandler(responseWriter http.ResponseWriter, request *http.Request) {
	movieLandingPageComponent := MovieLandingPage()
	movieLandingPageComponent.Render(context.Background(), responseWriter)
}

func ViewMoviePageHandler(responseWriter http.ResponseWriter, request *http.Request) {
	// TODO
}

func EditMoviePageHandler(responseWriter http.ResponseWriter, request *http.Request) {
	// TODO
}
