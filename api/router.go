package api

import (
	"fmt"
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

var routes = Routes{
	/*
		Route{
			"Index",
			"GET",
			"/52-5226/api/leubot/1.0.0/",
			Index,
		},
	*/
	Route{
		"AddUser",
		strings.ToUpper("Post"),
		"/52-5226/api/leubot/1.0.0/user",
		AddUser,
	},
	Route{
		"GetUser",
		strings.ToUpper("Get"),
		"/52-5226/api/leubot/1.0.0/user",
		GetUser,
	},
	Route{
		"RemoveUser",
		strings.ToUpper("Delete"),
		"/52-5226/api/leubot/1.0.0/user/{token}",
		RemoveUser,
	},
	Route{
		"PutElbow",
		strings.ToUpper("Put"),
		"/52-5226/api/leubot/1.0.0/elbow",
		PutElbow,
	},
	Route{
		"PutWristAngle",
		strings.ToUpper("Put"),
		"/52-5226/api/leubot/1.0.0/wrist/angle",
		PutWristAngle,
	},
	Route{
		"PutWristRotation",
		strings.ToUpper("Put"),
		"/52-5226/api/leubot/1.0.0/wrist/rotation",
		PutWristRotation,
	},
	Route{
		"PutGripper",
		strings.ToUpper("Put"),
		"/52-5226/api/leubot/1.0.0/gripper",
		PutGripper,
	},
	Route{
		"PutReset",
		strings.ToUpper("Put"),
		"/52-5226/api/leubot/1.0.0/reset",
		PutReset,
	},
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
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
