package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os/exec"
)

func main() {
    log.Println("Initializing")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// return ok
		w.WriteHeader(http.StatusOK)
		w.Write([]byte{})

		// get params
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		query, err := url.QueryUnescape(string(body))
		if err != nil {
			log.Fatal(err)
		}
		params, err := url.ParseQuery(query)
		if err != nil {
			log.Fatal(err)
		}
		for k, v := range params {
			fmt.Println(k, v)
		}

		/*
			if params["channel_id"][0] != "CFD5676RL" {
				return
			}
		*/

		var action string

		switch r.URL.Path {
		case "/leubot/tool/restart":
			action = "Restarting leubot.service"
		}

		// notify before the action
		reply := map[string]string{
			"response_type": "ephemeral",
			"text":          fmt.Sprintf("%s has been requested, wait for a while...", action),
		}
		jsonBytes, _ := json.Marshal(reply)
		if err != nil {
			log.Fatal(err)
		}
		resp, err := http.Post(params["response_url"][0], "application/json", bytes.NewReader(jsonBytes))
		if err != nil {
			log.Print(err)
            return
		}
		log.Println(resp)

		// do stuff
		go func() {
			out, err := exec.Command("sudo", "/bin/systemctl", "restart", "leubot.service").Output()

			if err != nil {
				log.Print(err)
                return
			}
			log.Printf("%s\n", out)
			// notify after the action as a public message
			reply = map[string]string{
				"response_type": "in_channel",
				"text":          fmt.Sprintf("%s [triggered by<@%s>] has been completed", action, params["user_id"][0]),
			}
			jsonBytes, _ = json.Marshal(reply)
			if err != nil {
				log.Print(err)
                return
			}
			resp, err = http.Post(params["response_url"][0], "application/json", bytes.NewReader(jsonBytes))
			if err != nil {
				log.Print(err)
                return
			}
			log.Println(resp)
		}()
	})

    log.Println("Starting")
	http.ListenAndServe("0.0.0.0:30002", nil)
}
