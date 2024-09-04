# Movie Buff Database Golang Server

This is a very basic Golang server for the Movie Buff Database!

This is a toy project meant to help me accumulate experience with programming on the web application side of things, as opposed to C++ game engine/graphics stuff where I usually spend my time. The stack is chosen pretty arbitrarily, just some things I was curious about and hadn't ever used before. The goal is to iterate on this (as time permits) until I end up with a robust full-featured movie database website with a rich user interface.

I'm not a front-end whiz by any stretch, and before I dive into the TypeScript world I wanted to try some server-side rendering stuff in a new language. Forgive any Go noob-isms, I'm still learning :)

In its current state, we have a bunch of ugly placeholder landing pages and a single viewable/editable actor, whose page can be viewed by entering this in your browser with the server running locally on port `8080`:
~~~~
http://localhost:8080/actors/view/Tom+Cruise
http://localhost:8080/actors/edit/Tom+Cruise
~~~~

Current tech stack:
1. Golang backend with `templ` for HTML templating
   a. for releases, I'll use `goreleaser` 
2. I'll try to sprinkle in htmx
3. TailwindCSS most likely (and a component library, probably `daisy`)
4. PostgreSQL to mess around with databases

### Running the server

To start a server instance, navigate to the `server` directory and run this command:

~~~~
go run server.go
~~~~

### Go Dependencies

* Postgres driver: `pq`
  * to install: `go get github.com/lib/pq`
* `templ` library
  * to install: `go install github.com/a-h/templ/cmd/templ@latest`

### Janky Initial DB setup

* For now, you just need to be able to run PostgreSQL; it's not hard to start up and create a db
  * as the app matures I'll be looking into different ways to ensure backups and shit like that are happening;
  * since this is a toy app it's probably okay to just make sure we have code for building up the DB based on whatever test data we end up using; 
  * we also don't have users really, I'll aim to have a simple `postgres` user that's expected or something

* Currently you can connect to the database using these credentials (note the lack of security! don't ever use this application to put anything sensitive on the db)
  
~~~~Go
  // golang database connection params
	const (
		host = "localhost"
		port = 5432
		user = "postgres"
		password = "admin"
		sslmode = "disable"
	)
~~~~

* at this stage, the database is basically unused, but future updates will add more site features that will make use of it 
  * although at the end of the day, this project is intended for me to practice building little CRUD apps

## TailwindCSS 

To use the TailwindCSS CLI and have it watch and rebuild our stylesheets, just run this command in a powershell instance:

~~~~
npx tailwindcss -c ./tailwind.config.js -i ./input.css -o ./assets/css/styles.css --minify --watch
~~~~

While this process is running, `assets/css/styles.css` will be up-to-date with any changes you make to the code. There is a `tailwind.config.js` file that you can use to modify your tailwind setup to add/customize various themes and plugins.