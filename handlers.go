package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"gitlab.zerodha.tech/commons/lil/store"
)

// Response represents response struct.
type Response struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

// generateRandomString generates a random string of given length n.
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

// sendJSONResp sends data and error as a JSON HTTP response.
func sendJSONResp(data interface{}, err error, code int, w http.ResponseWriter) {
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
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(r)
}

// sendGeneralError sends 500 internal server error, used to send unknown server errors to user.
func sendGeneralError(w http.ResponseWriter) {
	sendJSONResp(nil, fmt.Errorf("Something went wrong"), http.StatusInternalServerError, w)
}

// createShortURL takes full url and creates a random short url.
// Only the random URI is returned.
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
	sendJSONResp("welcome", nil, http.StatusOK, w)
	return
}

// handleRedirect handles short url to actual url redirect.
func handleRedirect(w http.ResponseWriter, r *http.Request) {
	uri := chi.URLParam(r, "uri")
	url, err := str.GetFullURL(uri)
	if err == store.ErrNotFound {
		sendJSONResp(nil, fmt.Errorf("Not found"), http.StatusNotFound, w)
		return
	}
	http.Redirect(w, r, string(url), 301)
}

// handleCreate creates a new short url. Accepts one post params `url`
// which is url which has to be shortened.
func handleCreate(w http.ResponseWriter, r *http.Request) {
	url := r.PostFormValue("url")
	// Validate params
	if url == "" {
		sendJSONResp(nil, fmt.Errorf("Invalid url"), http.StatusBadRequest, w)
		return
	}
	// Create short url
	sURI, err := createShortURL(url)
	if err != nil {
		// Send log.Printf("error creating url: %v", err)
		sendGeneralError(w)
	} else {
		sendJSONResp(fmt.Sprintf("%s/%s", baseURL, sURI), nil, http.StatusOK, w)
	}
}

// handleDelete deletes the short url.
func handleDelete(w http.ResponseWriter, r *http.Request) {
	uri := chi.URLParam(r, "uri")
	err := str.Delete(uri)
	if err == store.ErrNotFound {
		sendJSONResp(nil, fmt.Errorf("Not found"), http.StatusNotFound, w)
		return
	}
	sendJSONResp(true, nil, http.StatusOK, w)
}
