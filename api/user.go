package api

import (
	"encoding/json"
	"log"
	"net/http"
	"path"
)

type UserInfo struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Token string `json:"token"`
}

func AddUser(w http.ResponseWriter, r *http.Request) {
	// parse the request body
	decoder := json.NewDecoder(r.Body)
	var userInfo UserInfo
	err := decoder.Decode(&userInfo)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}
	// bypass the request to HandlerChannel
	HandlerChannel <- HandlerMessage{
		Type:  TypeAddUser,
		Value: []interface{}{userInfo},
	}
	// receive a message from the other end of HandlerChannel
	msg, ok := <-HandlerChannel
	// check the channel status
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// respond with the result
	switch msg.Type {
	case TypeUserAdded: // respond with the added UserInfo
		userInfo, ok = msg.Value[0].(UserInfo)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		log.Printf("[HandlerChannel] UserAdded (name, email, token) = %v, %v, %v", userInfo.Name, userInfo.Email, userInfo.Token)
		js, err := json.Marshal(userInfo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(js)
	case TypeUserExisted: // there's a user in the system already
		log.Printf("[HandlerChannel] UserExisted, not replacing with (name, email) = %v, %v", userInfo.Name, userInfo.Email)
		w.WriteHeader(http.StatusConflict)
	default: // something went wrong
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	// bypass the request to HandlerChannel
	HandlerChannel <- HandlerMessage{
		Type: TypeGetUser,
	}
	// receive a message from the other end of HandlerChannel
	msg, ok := <-HandlerChannel
	// check the channel status
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// respond with the result
	switch msg.Type {
	case TypeCurrentUser: // respond with the current UserInfo
		userInfo, ok := msg.Value[0].(UserInfo)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
		}
		userInfo.Token = "" // remove the token
		log.Printf("[HandlerChannel] CurrentUser (name, email) = %v, %v", userInfo.Name, userInfo.Email)
		js, err := json.Marshal(userInfo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(js)
	default: // something went wrong
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func RemoveUser(w http.ResponseWriter, r *http.Request) {
	// get the token from the path
	token := path.Base(r.URL.Path)
	HandlerChannel <- HandlerMessage{
		Type:  TypeDeleteUser,
		Value: []interface{}{token},
	}
	// receive a message from the other end of HandlerChannel
	msg, ok := <-HandlerChannel
	// check the channel status
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// respond with the result
	switch msg.Type {
	case TypeUserDeleted: // the user removed
		log.Printf("[HandlerChannel] UserDeleted with token = %v", token)
		w.WriteHeader(http.StatusNoContent)
	case TypeUserNotFound: // no user with the token
		log.Printf("[HandlerChannel] UserNotfound with token = %v", token)
		w.WriteHeader(http.StatusNotFound)
	default: // something went wrong
		w.WriteHeader(http.StatusInternalServerError)
	}
}
