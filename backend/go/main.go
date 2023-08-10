package main

import (
	"examples/database"
	"examples/database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type server struct {
	// Specify dependencies here
	// loggers, external API clients, etc
	testDependency string
	// We'll want an implementation of a logger, we'll use standard package for this
	logger log.Logger
	// We'll also have a database dependency
	db database.Storer
}

func main() {
	// Retrieve any needed values from environment variables and include them in the server struct, and also validate them, or check if they're missing
	testEnvVar := os.Getenv("TEST_ENVIRONMENT_VARIABLE")
	if testEnvVar == "" {
		// Here we use panic(), which will stop further execution. You should never use a panic intentionally after service initialization.
		panic("TEST_ENVIRONMENT_VARIABLE is required for this service to run")
	}

	// We'll also need to initialize our Database connection, we'll be using the SQL implementation of our Storer interface
	db, err := sql.NewSQLDB(os.Getenv("DATABASE_URL"))
	if err != nil {
		// Again we'll use a panic here, as if we cannot connect to our database, this API won't be able to function
		panic(fmt.Sprintf("Error connecting to database: %v", err))
	}

	// Combine all our connections and settings into a single struct that we can use to make handler methods on
	s := server{
		// Init our logger with standard package, we'll just output to console using os.Stdout
		logger:         *log.New(os.Stdout, "logger: ", log.Lshortfile),
		testDependency: testEnvVar,
		db:             db,
	}

	// Create a GoRoutine that can run in the background for any async tasks
	go func() {
		// Using an open ended for loop can be dangerous, but this case it is perfect, so long as we include a time.Sleep
		for {
			// Specify the time interval this background task should run at (In our case, 10 minutes)
			time.Sleep(time.Minute * 10)
			// Here I'll need to keep our database clean of expired login sessions
			count, err := s.db.ClearExpiredSessions()
			if err != nil {
				s.logger.Println("ERROR: Unable to clear expired login sessions: %v", err)
				// We'll skip to next loop iteration
				continue
			}
			// If we didn't encounter an error, operation was successful, let's still log it:
			s.logger.Println("INFO: Cleared %d expired login sessions", count)
		}
	}() // Adding "()" immediately after this anonymous goroutine starts it.

	// Cross Origin Resource Sharing (CORS)
	// This allows a frontend to communicate with a backend that is hosted at a different URL.
	//
	// By default, if you have a frontend hosted at https://myCoolWebsite.com, and you try to make an API call
	// to your API hosted at https://myAwesomeAPI.com, you'll encounter CORS errors.
	//
	// Most modern web browsers (Chrome, Firefox, Safari, Edge, Opera, etc) accomplish this by performing a
	// "pre-flight" request using the OPTIONS http verb to check CORS options.
	//
	// Note that API testing tools like Postman (allows you to make requests to your backend) will not send a
	// pre-flight request, and will never encounter CORS errors, so be sure to test with a frontend before ever
	// pushing something straight to production.
	//
	// Here's we'll use a Middleware function that only uses standard library
	// Middleware allows us to wrap a Handler function, it is perfect for performing actions such as authentication checks, or
	// in this case handling CORS configuration.
	cors := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Here we can specify what Origins are allowed. (Example: An Origin could be our frontend hosted at https://myCoolWebsite.com")
			// For testing purposes, we'll use the wildcard "*" to allow any Origin. This SHOULD NOT be present in a production-ready service!
			w.Header().Set("Access-Control-Allow-Origin", "*")
			// Here we specify allowed headers, including any custom headers you may wish to be included in a request
			w.Header().Set("Access-Control-Allow-Headers", strings.Join([]string{"Content-Type", "Authorization"}, ","))
			// Here you'll specify what HTTP methods (verbs) your API allows.
			// Note that http.MethodOptions may need to be explicity allowed for CORS pre-flight requests.
			w.Header().Set("Access-Control-Allow-Methods", strings.Join([]string{http.MethodGet, http.MethodOptions}, ","))

			// Typically in an API, the http.MethodOptions verb will only be used for CORS, so we'll explicitly return nothing in
			// the event of an OPTIONS call. Otherwise we'll serve the wrapped handler
			if r.Method == http.MethodOptions {
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	// Set up a Router, I'll use Gorilla Mux, although you can use standard library Mux, or other routers such as Chi
	// In this case since we're doing a RESTful API, GorillaMux allows us to easily use parameters included in the path
	// (such as "/users/{username}", we'll be able to easily retrieve the username)
	router := mux.NewRouter()
	// GorillaMux also gives us a handy Use method, which is perfect for Middleware! Any request that is handled by this
	// router, will execute any middleware before the actual endpoint.
	// Apply any Middleware you need to your Router with the Use method (In this case we'll use our CORS middleware)
	router.Use(cors)

	// Standard Health Check endpoint that just returns a 200 status and empty response body, useful for simply checking if your API is running
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})

	// Set up a Login endpoint to authenticate with your service
	router.HandleFunc("/login/", s.login).Methods(http.MethodPost)

	// We'll make a Subrouter to secure our endpoints
	loggedin := router.PathPrefix("").Subrouter()
	// Apply our Middleware that checks authentication for every endpoint
	// loggedin.Use(s.auth) // TODO (IME): Need to create an example implmentation of this.

	// Hook up our endpoints
	// We'll need a logout endpoint
	loggedin.HandleFunc("/logout/", s.logout).Methods(http.MethodPost)

	// Here's an example of a typical REST style API
	// Users API
	loggedin.HandleFunc("/users/", s.userInfoSelf).Methods(http.MethodGet)
	// TODO (IME): Finish creating a REST API
	// loggedin.HandleFunc("/users/{username}", s.userInfoOther).Methods(http.MethodGet)
	// loggedin.HandleFunc("/users/{username}/exists", s.userExists).Methods(http.MethodGet)
	// loggedin.HandleFunc("/users/search/{name}", s.userSearch).Methods(http.MethodGet)
	// loggedin.HandleFunc("/users/password", s.resetPasswordSelf).Methods(http.MethodPut)
	// loggedin.HandleFunc("/users/{username}/password", s.resetPasswordOther).Methods(http.MethodPut)
	// loggedin.HandleFunc("/users/{username}/email", s.sendInstallLinks).Methods(http.MethodPost)
	// loggedin.HandleFunc("/users/", s.userAdd).Methods(http.MethodPost)
	// loggedin.HandleFunc("/users/{username}", s.userRemove).Methods(http.MethodDelete)
	// loggedin.HandleFunc("/users/{username}/dealership/{cid}", s.userAddToDealership).Methods(http.MethodPut)
	// loggedin.HandleFunc("/users/{username}/dealership/{cid}", s.userRemoveFromDealership).Methods(http.MethodDelete)
	// loggedin.HandleFunc("/users/{username}/email", s.userEmail).Methods(http.MethodPut)
	// loggedin.HandleFunc("/users/{username}/enabled", s.userEnable).Methods(http.MethodPut)
	// loggedin.HandleFunc("/users/{username}/enabled", s.userDisable).Methods(http.MethodDelete)
	// loggedin.HandleFunc("/users/{username}/lock", s.userUnlock).Methods(http.MethodDelete)
	// loggedin.HandleFunc("/users/{username}/admin", s.userPromote).Methods(http.MethodPut)
	// loggedin.HandleFunc("/users/{username}/admin", s.userDemote).Methods(http.MethodDelete)

	// Start the webserver
	// Allow environment to set the port
	port := ":" + os.Getenv("PORT")
	if port == ":" {
		// Default to port 8080 if no port is specified
		port = ":8080"
	}
	// It's a good idea to wrap your http.ListenAndServe call in a Fatal or Critical logger call, as when ListenAndServe
	// returns, it means your API is no longer running!
	s.logger.Fatalln(http.ListenAndServe(port, router))
}
