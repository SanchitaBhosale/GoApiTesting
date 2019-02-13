package main

import (
	// Import the gorilla/mux library we just installed
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// The new router function creates the router and
// returns it to us. We can now use this function
// to instantiate and test the router outside of the main function
func newRouter() *mux.Router {
	// Declare a new router
	r := mux.NewRouter()

	// This is where the router is useful, it allows us to declare methods that
	// this path will be valid for
	// Define Route: `GET /hello`
	r.HandleFunc("/hello", handler).Methods("GET")

	// Declare the static file directory and point it to the
	// directory we just made
	staticFileDirectory := http.Dir("./assets/")
	// Declare the handler, that routes requests to their respective filename.
	// The fileserver is wrapped in the `stripPrefix` method, because we want to
	// remove the "/assets/" prefix when looking for files.
	// For example, if we type "/assets/index.html" in our browser, the file server
	// will look for only "index.html" inside the directory declared above.
	// If we did not strip the prefix, the file server would look for
	// "./assets/assets/index.html", and yield an error
	staticFileHandler := http.StripPrefix("/assets/", http.FileServer(staticFileDirectory))
	// The "PathPrefix" method acts as a matcher, and matches all routes starting
	// with "/assets/", instead of the absolute route itself
	r.PathPrefix("/assets/").Handler(staticFileHandler).Methods("GET")

	// These lines are added inside the newRouter() function before returning r
	r.HandleFunc("/bird", getBirdHandler).Methods("GET")
	r.HandleFunc("/bird", createBirdHandler).Methods("POST")
	return r
}

func main() {

	// The router is now formed by calling the `newRouter` constructor function
	// that we defined above. The rest of the code stays the same
	r := newRouter()
	// We can then pass our router (after declaring all our routes) to this method
	// (where previously, we were leaving the second argument as nil)
	http.ListenAndServe(":8080", r)
}

// Handler functions are responsible for exposing the business logic i.e.
// serving to client
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}

type Bird struct {
	Species     string `json:"species"`
	Description string `json:"description"`
}

var birds []Bird

func getBirdHandler(w http.ResponseWriter, r *http.Request) {
	// To test the GET call set birds to some initial value
	birds = []Bird{{"Chimni", "Found in India"}}
	/*
		The list of birds is now taken from the store instead of the package level  `birds` variable we had earlier
		The `store` variable is the package level variable that we defined in
		`store.go`, and is initialized during the initialization phase of the
		application
	*/
	// birds, err := store.GetBirds()

	//Convert the "birds" variable to json
	birdListBytes, err := json.Marshal(birds)

	// If there is an error, print it to the console, and return a server
	// error response to the user
	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// If all goes well, write the JSON list of birds to the response
	w.Write(birdListBytes)
}

func createBirdHandler(w http.ResponseWriter, r *http.Request) {
	// Create a new instance of Bird
	bird := Bird{}

	// We send all our data as HTML form data
	// the `ParseForm` method of the request, parses the
	// form values
	err := r.ParseForm()

	// In case of any error, we respond with an error to the user
	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get the information about the bird from the form info
	bird.Species = r.Form.Get("species")
	bird.Description = r.Form.Get("description")

	// Append our existing list of birds with a new entry
	birds = append(birds, bird)

	// The only change we made here is to use the `CreateBird` method instead of
	// appending to the `bird` variable like we did earlier
	// err = store.CreateBird(&bird)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	//Finally, we redirect the user to the original HTMl page
	// (located at `/assets/`), using the http libraries `Redirect` method
	http.Redirect(w, r, "/assets/", http.StatusFound)
}

// Our store will have two methods, to add a new bird,
// and to get all existing birds
// Each method returns an error, in case something goes wrong
type Store interface {
	CreateBird(bird *Bird) error
	GetBirds() ([]*Bird, error)
}

// The store variable is a package level variable that will be available for
// use throughout our application code
var store Store

// The `dbStore` struct will implement the `Store` interface
// It also takes the sql DB connection object, which represents
// the database connection.
type dbStore struct {
	db *sql.DB
}

func (store *dbStore) CreateBird(bird *Bird) error {
	// 'Bird' is a simple struct which has "species" and "description" attributes
	// THe first underscore means that we don't care about what's returned from
	// this insert query. We just want to know if it was inserted correctly,
	// and the error will be populated if it wasn't
	_, err := store.db.Query("INSERT INTO birds(species, description) VALUES ($1,$2)", bird.Species, bird.Description)
	return err
}

func (store *dbStore) GetBirds() ([]*Bird, error) {
	// Query the database for all birds, and return the result to the
	// `rows` object
	rows, err := store.db.Query("SELECT species, description from birds")
	// We return incase of an error, and defer the closing of the row structure
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Create the data structure that is returned from the function.
	// By default, this will be an empty array of birds
	birds := []*Bird{}
	for rows.Next() {
		// For each row returned by the table, create a pointer to a bird,
		bird := &Bird{}
		// Populate the `Species` and `Description` attributes of the bird,
		// and return incase of an error
		if err := rows.Scan(&bird.Species, &bird.Description); err != nil {
			return nil, err
		}
		// Finally, append the result to the returned array, and repeat for
		// the next row
		birds = append(birds, bird)
	}
	return birds, nil
}

/*
We will need to call the InitStore method to initialize the store. This will
typically be done at the beginning of our application (in this case, when the server starts up)
This can also be used to set up the store as a mock, which we will be observing
later on
*/
func InitStore(s Store) {
	store = s
}
