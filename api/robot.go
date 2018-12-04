package api

import (
	"encoding/json"
	"log"
	"net/http"
)

type RobotCommand struct {
	Token string `json:"token"`
	Value uint16 `json:"value"`
}

func PutElbow(w http.ResponseWriter, r *http.Request) {
	// parse the request body
	decoder := json.NewDecoder(r.Body)
	var robotCommand RobotCommand
	err := decoder.Decode(&robotCommand)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}
	// bypass the request to HandlerChannel
	HandlerChannel <- HandlerMessage{
		Type:  TypePutElbow,
		Value: []interface{}{robotCommand},
	}
	// receive a message from the other end of HandlerChannel
	msg, ok := <-HandlerChannel
	// check the channel status
	if !ok {
		w.WriteHeader(http.StatusInternalServerError) // 500
		return
	}
	// respond with the result
	switch msg.Type {
	case TypeActionPerformed: // the requested action is performed
		log.Printf("[HandlerChannel] ElbowRotation: %v", robotCommand.Value)
		w.WriteHeader(http.StatusAccepted) // 202
	case TypeInvalidCommand: // the invalid value provided
		log.Printf("[HandlerChannel] InvalidCommand: %v", robotCommand.Value)
		w.WriteHeader(http.StatusBadRequest) // 400
	case TypeInvalidToken: // the invalid token provided
		log.Printf("[HandlerChannel] InvalidToken: %v", robotCommand.Token)
		w.WriteHeader(http.StatusUnauthorized) // 401
	default: // something went wrong
		w.WriteHeader(http.StatusInternalServerError) // 500
	}
}

func PutWristAngle(w http.ResponseWriter, r *http.Request) {
	// parse the request body
	decoder := json.NewDecoder(r.Body)
	var robotCommand RobotCommand
	err := decoder.Decode(&robotCommand)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}
	// bypass the request to HandlerChannel
	HandlerChannel <- HandlerMessage{
		Type:  TypePutWristAngle,
		Value: []interface{}{robotCommand},
	}
	// receive a message from the other end of HandlerChannel
	msg, ok := <-HandlerChannel
	// check the channel status
	if !ok {
		w.WriteHeader(http.StatusInternalServerError) // 500
		return
	}
	// respond with the result
	switch msg.Type {
	case TypeActionPerformed: // the requested action is performed
		log.Printf("[HandlerChannel] WristAngle: %v", robotCommand.Value)
		w.WriteHeader(http.StatusAccepted) // 202
	case TypeInvalidCommand: // the invalid value provided
		log.Printf("[HandlerChannel] InvalidCommand: %v", robotCommand.Value)
		w.WriteHeader(http.StatusBadRequest) // 400
	case TypeInvalidToken: // the invalid token provided
		log.Printf("[HandlerChannel] InvalidToken: %v", robotCommand.Token)
		w.WriteHeader(http.StatusUnauthorized) // 401
	default: // something went wrong
		w.WriteHeader(http.StatusInternalServerError) // 500
	}
}

func PutWristRotation(w http.ResponseWriter, r *http.Request) {
	// parse the request body
	decoder := json.NewDecoder(r.Body)
	var robotCommand RobotCommand
	err := decoder.Decode(&robotCommand)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}
	// bypass the request to HandlerChannel
	HandlerChannel <- HandlerMessage{
		Type:  TypePutWristRotation,
		Value: []interface{}{robotCommand},
	}
	// receive a message from the other end of HandlerChannel
	msg, ok := <-HandlerChannel
	// check the channel status
	if !ok {
		w.WriteHeader(http.StatusInternalServerError) // 500
		return
	}
	// respond with the result
	switch msg.Type {
	case TypeActionPerformed: // the requested action is performed
		log.Printf("[HandlerChannel] WristRotation: %v", robotCommand.Value)
		w.WriteHeader(http.StatusAccepted) // 202
	case TypeInvalidCommand: // the invalid value provided
		log.Printf("[HandlerChannel] InvalidCommand: %v", robotCommand.Value)
		w.WriteHeader(http.StatusBadRequest) // 400
	case TypeInvalidToken: // the invalid token provided
		log.Printf("[HandlerChannel] InvalidToken: %v", robotCommand.Token)
		w.WriteHeader(http.StatusUnauthorized) // 401
	default: // something went wrong
		w.WriteHeader(http.StatusInternalServerError) // 500
	}
}

func PutGripper(w http.ResponseWriter, r *http.Request) {
	// parse the request body
	decoder := json.NewDecoder(r.Body)
	var robotCommand RobotCommand
	err := decoder.Decode(&robotCommand)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}
	// bypass the request to HandlerChannel
	HandlerChannel <- HandlerMessage{
		Type:  TypePutGripper,
		Value: []interface{}{robotCommand},
	}
	// receive a message from the other end of HandlerChannel
	msg, ok := <-HandlerChannel
	// check the channel status
	if !ok {
		w.WriteHeader(http.StatusInternalServerError) // 500
		return
	}
	// respond with the result
	switch msg.Type {
	case TypeActionPerformed: // the requested action is performed
		log.Printf("[HandlerChannel] Gripper: %v", robotCommand.Value)
		w.WriteHeader(http.StatusAccepted) // 202
	case TypeInvalidCommand: // the invalid value provided
		log.Printf("[HandlerChannel] InvalidCommand: %v", robotCommand.Value)
		w.WriteHeader(http.StatusBadRequest) // 400
	case TypeInvalidToken: // the invalid token provided
		log.Printf("[HandlerChannel] InvalidToken: %v", robotCommand.Token)
		w.WriteHeader(http.StatusUnauthorized) // 401
	default: // something went wrong
		w.WriteHeader(http.StatusInternalServerError) // 500
	}
}

func PutReset(w http.ResponseWriter, r *http.Request) {
	// parse the request body
	decoder := json.NewDecoder(r.Body)
	var robotCommand RobotCommand
	err := decoder.Decode(&robotCommand)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}
	// bypass the request to HandlerChannel
	HandlerChannel <- HandlerMessage{
		Type:  TypePutReset,
		Value: []interface{}{robotCommand},
	}
	// receive a message from the other end of HandlerChannel
	msg, ok := <-HandlerChannel
	// check the channel status
	if !ok {
		w.WriteHeader(http.StatusInternalServerError) // 500
		return
	}
	// respond with the result
	switch msg.Type {
	case TypeActionPerformed: // the requested action is performed
		log.Println("[HandlerChannel] Reset")
		w.WriteHeader(http.StatusAccepted) // 202
	case TypeInvalidToken: // the invalid token provided
		log.Printf("[HandlerChannel] InvalidToken: %v", robotCommand.Token)
		w.WriteHeader(http.StatusUnauthorized) // 401
	default: // something went wrong
		w.WriteHeader(http.StatusInternalServerError) // 500
	}

}
