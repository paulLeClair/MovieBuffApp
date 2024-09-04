package oscars

import (
	"context"
	"net/http"
)

// TODO - database integration etc for the oscars

type OscarsHandler struct {
	// TODO - likely we'll want to hook in the database via some kind of "environment" type
}

func NewOscarsHandler() *OscarsHandler {
	handler := new(OscarsHandler)
	// TODO - setup in the future; for now just placeholder
	return handler
}

func (handler OscarsHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	oscarHomepageComponent := OscarsHomepage()
	oscarHomepageComponent.Render(context.Background(), responseWriter)
}
