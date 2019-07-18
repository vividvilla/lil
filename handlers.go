package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
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

// RedirectResponse represents all redirect urls.
type RedirectResponse struct {
	ID   string `json:"id"`
	URL  string `json:"url"`
	Page string `json:"page"`
}

// Params represents input params used to create new short url.
type Params struct {
	URL    string         `json:"url"`
	OGTags []*store.OGTag `json:"og_tags"`
	Title  string         `json:"title"`
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

// redirectResponse returns a RedirectResponse for given id
func redirectResponse(id string) RedirectResponse {
	return RedirectResponse{
		ID:   id,
		URL:  fmt.Sprintf("%s/%s", baseURL, id),
		Page: fmt.Sprintf("%s/%s/%s", baseURL, pageRedirectPrefix, id),
	}
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
func createShortURL(params *Params) (string, error) {
	// To avoid collision check if random generated string is already
	// stored. If stored then generate new.
	var (
		err error
		id  string
	)
	for {
		// Generate random string.
		id, err = generateRandomString(shortURLLength)
		if err != nil {
			return id, err
		}
		// Try getting the url for new generated id.
		// If its present then generate new id and check.
		_, _, err := str.Get(id)
		if err == store.ErrNotFound {
			break
		} else if err != nil {
			return id, err
		}
	}

	err = str.Set(id, params.URL, &store.Meta{
		OGTags: params.OGTags,
		Title:  params.Title,
	})
	return id, err
}

func handleWelcome(w http.ResponseWriter, r *http.Request) {
	sendJSONResp("welcome", nil, http.StatusOK, w)
	return
}

// handleRedirect handles short url to actual url redirect.
func handleRedirect(w http.ResponseWriter, r *http.Request) {
	uri := chi.URLParam(r, "uri")
	url, _, err := str.Get(uri)
	if err == store.ErrNotFound {
		sendJSONResp(nil, fmt.Errorf("Not found"), http.StatusNotFound, w)
		return
	} else if err != nil {
		log.Printf("error getting short url: %v", err)
		sendGeneralError(w)
		return
	}
	http.Redirect(w, r, string(url), 301)
}

func handlePageRedirect(w http.ResponseWriter, r *http.Request) {
	uri := chi.URLParam(r, "uri")
	url, meta, err := str.Get(uri)
	if err == store.ErrNotFound {
		sendJSONResp(nil, fmt.Errorf("Not found"), http.StatusNotFound, w)
		return
	} else if err != nil {
		log.Printf("error getting short url: %v", err)
		sendGeneralError(w)
		return
	}

	params := Params{
		URL: url,
	}
	if meta != nil {
		params.Title = meta.Title
		params.OGTags = meta.OGTags
	}
	var tplBody bytes.Buffer
	if err := redirectTpl.Execute(&tplBody, &params); err != nil {
		log.Printf("error executing redirect template: %v", err)
		sendJSONResp(nil, fmt.Errorf("Couldn't redirect"), http.StatusInternalServerError, w)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(tplBody.Bytes())
}

// handleCreate creates a new short url. Accepts one post params `url`
// which is url which has to be shortened.
func handleCreate(w http.ResponseWriter, r *http.Request) {
	params := &Params{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(params)
	if err != nil {
		sendJSONResp(nil, errors.New("Invalid input"), http.StatusBadRequest, w)
		return
	}
	// Validate params
	if params.URL == "" {
		sendJSONResp(nil, fmt.Errorf("Invalid url"), http.StatusBadRequest, w)
		return
	}
	// Create short url
	id, err := createShortURL(params)
	if err != nil {
		// Send log.Printf("error creating url: %v", err)
		sendGeneralError(w)
	} else {
		sendJSONResp(redirectResponse(id), nil, http.StatusOK, w)
	}
}

func handleGetRedirects(w http.ResponseWriter, r *http.Request) {
	uri := chi.URLParam(r, "uri")
	_, _, err := str.Get(uri)
	if err == store.ErrNotFound {
		sendJSONResp(nil, fmt.Errorf("Not found"), http.StatusNotFound, w)
		return
	} else if err != nil {
		log.Printf("error getting short url: %v", err)
		sendGeneralError(w)
		return
	}
	sendJSONResp(redirectResponse(uri), nil, http.StatusOK, w)
}

// handleDelete deletes the short url.
func handleDelete(w http.ResponseWriter, r *http.Request) {
	uri := chi.URLParam(r, "uri")
	err := str.Del(uri)
	if err == store.ErrNotFound {
		sendJSONResp(nil, fmt.Errorf("Not found"), http.StatusNotFound, w)
		return
	}
	sendJSONResp(true, nil, http.StatusOK, w)
}
