package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"gitlab.zerodha.tech/commons/lil/store"
)

// Response represents response struct.
type Response struct {
	Data  interface{} `json:"data"`
	Error interface{} `json:"error,omitempty"`
}

func generateRandomString(n int) (string, error) {
	const dictionary = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	var bytes = make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}

	return string(bytes), nil
}

func sendResp(data interface{}, err error, code int, w http.ResponseWriter) {
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}
	resp := Response{
		Data:  data,
		Error: errMsg,
	}
	r, errJSON := json.Marshal(resp)
	if errJSON != nil {
		log.Printf("error marshalling response: %v", errJSON)
	}
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	w.Write(r)
}

func sendGeneralError(w http.ResponseWriter) {
	sendResp(nil, fmt.Errorf("Something went wrong."), http.StatusInternalServerError, w)
}

func createShortURL(url string) (string, error) {
	sURI, err := str.GetShortURL(url)
	if err == store.ErrNotFound {
		// To avoid collision check if random generated string is already
		// stored. If stored then generate new.
		for {
			// Generate random string.
			sURI, err = generateRandomString(shortURLLength)
			if err != nil {
				return sURI, err
			}
			// Try getting the full url for short url to see if it exists.
			_, err := str.GetFullURL(sURI)
			if err == store.ErrNotFound {
				err = str.Set(sURI, url)
				return sURI, err
			} else if err != nil {
				return sURI, err
			}
		}
	}
	return sURI, err
}

func handleWelcome(w http.ResponseWriter, r *http.Request) {
	sendResp("welcome", nil, http.StatusOK, w)
}

func handleAll(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Welcome handler
		if r.URL.Path == "/" {
			handleWelcome(w, r)
			return
		}
		handleRedirect(w, r)
	} else if r.Method == http.MethodDelete {
		handleDelete(w, r)
	} else {
		sendResp(nil, fmt.Errorf("Method not allowed"), http.StatusMethodNotAllowed, w)
	}
}

func handleRedirect(w http.ResponseWriter, r *http.Request) {
	url, err := str.GetFullURL(r.URL.Path[1:])
	if err == store.ErrNotFound {
		sendResp(nil, fmt.Errorf("Not found"), http.StatusNotFound, w)
		return
	}
	http.Redirect(w, r, string(url), 301)
}

func handleCreate(w http.ResponseWriter, r *http.Request) {
	url := r.PostFormValue("url")

	// Validate params
	if url == "" {
		sendResp(nil, fmt.Errorf("Invalid url"), http.StatusBadRequest, w)
		return
	}

	sURI, err := createShortURL(url)
	if err != nil {
		log.Printf("error creating url: %v", err)
		sendGeneralError(w)
	} else {
		sendResp(fmt.Sprintf("%s/%s", baseURL, sURI), nil, http.StatusOK, w)
	}
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	err := str.Delete(r.URL.Path[1:])
	if err == store.ErrNotFound {
		sendResp(nil, fmt.Errorf("Not found"), http.StatusNotFound, w)
		return
	}
	sendResp(true, nil, http.StatusOK, w)
}
