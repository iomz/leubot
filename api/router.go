package api

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// Route is a basic pattern of the rounting
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// Routes contain the Route
type Routes []Route

// HandlerChannel is used to communicate
// between the router and other application logic
var HandlerChannel chan HandlerMessage

// TODO: feed from main?

// APIProto for API access protocol
var APIProto = "https://"

// APIHost is the hostname
var APIHost = "api.interactions.ics.unisg.ch"

// APIBaseURL is the path to the API
var APIBaseURL = "/leubot"

// TODO: construct after/within NewRouter()?
var routes = Routes{
	Route{
		"AddUser",
		strings.ToUpper("Post"),
		APIBaseURL + "/user",
		AddUser,
	},
	Route{
		"GetUser",
		strings.ToUpper("Get"),
		APIBaseURL + "/user",
		GetUser,
	},
	Route{
		"RemoveUser",
		strings.ToUpper("Delete"),
		APIBaseURL + "/user/{token}",
		RemoveUser,
	},
	Route{
		"PutBase",
		strings.ToUpper("Put"),
		APIBaseURL + "/base",
		PutBase,
	},
	Route{
		"PutShoulder",
		strings.ToUpper("Put"),
		APIBaseURL + "/shoulder",
		PutShoulder,
	},
	Route{
		"PutElbow",
		strings.ToUpper("Put"),
		APIBaseURL + "/elbow",
		PutElbow,
	},
	Route{
		"PutWristAngle",
		strings.ToUpper("Put"),
		APIBaseURL + "/wrist/angle",
		PutWristAngle,
	},
	Route{
		"PutWristRotation",
		strings.ToUpper("Put"),
		APIBaseURL + "/wrist/rotation",
		PutWristRotation,
	},
	Route{
		"PutGripper",
		strings.ToUpper("Put"),
		APIBaseURL + "/gripper",
		PutGripper,
	},
	Route{
		"PutReset",
		strings.ToUpper("Put"),
		APIBaseURL + "/reset",
		PutReset,
	},
}

// Logger handles the logging in the router
func Logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		log.Printf(
			"%s %s %s %s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
	})
}

// NewRouter creats a new instance of Router
func NewRouter(hmc chan HandlerMessage) *mux.Router {
	HandlerChannel = hmc
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}
