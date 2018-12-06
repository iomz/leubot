package api

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var HandlerChannel chan HandlerMessage

// TODO: feed from main?
var APIProto = "https://"
var APIHost = "interactions.iit.unisg.ch"
var APIBaseURL = "/52-5226/api/leubot/1.0.0"

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
