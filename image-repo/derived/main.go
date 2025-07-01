package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/fermyon/spin/sdk/go/v2/sqlite"
	"github.com/gorilla/mux"
)

func getDb() *sql.DB {
	db := sqlite.Open("default")
	db.Exec("'PRAGMA foreign_keys = ON;")

	return db
}

func errorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]string{
		"reason": message,
	}

	encoder := json.NewEncoder(w)
	encoder.Encode(response)
}

func parseDimension(dim string) (int, error) {
	if dim == "" {
		return 0, nil
	}

	var value int
	_, err := fmt.Sscanf(dim, "%d", &value)
	if err != nil || value < 1 || value > 50000 {
		return 0, fmt.Errorf("must be a positive integer between 1 and 50000")
	}

	return value, nil
}

func parseType(typeStr string) (string, error) {
	if typeStr == "" {
		return "image/png", nil
	}

	if typeStr == "image/png" || typeStr == "image/jpeg" {
		return typeStr, nil
	}

	return "", fmt.Errorf("must be either empty, \"image/png\" or \"image/jpeg\"")
}

func getQuery(r *http.Request) (width int, height int, typ string, err error) {
	width, err = parseDimension(r.URL.Query().Get("width"))
	if err != nil {
		return
	}

	height, err = parseDimension(r.URL.Query().Get("height"))
	if err != nil {
		return
	}

	typ, err = parseType(r.URL.Query().Get("type"))

	return
}

func calculateTransform(width int, height int, typ string) string {
	var transformParts []string

	if width > 0 {
		transformParts = append(transformParts, fmt.Sprintf("width=%d", width))
	}

	if height > 0 {
		transformParts = append(transformParts, fmt.Sprintf("height=%d", height))
	}

	transformParts = append(transformParts, fmt.Sprintf("type=%s", typ))

	return strings.Join(transformParts, ";")
}

func updateDerivativeLastAccess(db *sql.DB, name string, transform string) {
	db.Exec("UPDATE derived SET last_access = ? WHERE original_name = ? AND transformation = ?",
		time.Now().Unix(), name, transform)

}

func getDerivative(db *sql.DB, name string, transform string) ([]byte, error) {
	var data []byte
	err := db.QueryRow("SELECT data FROM derived WHERE original_name = ? AND transformation = ?", name, transform).Scan(&data)
	if err != nil {
		return nil, err // Return the original error to check for sql.ErrNoRows
	}
	return data, nil
}

func getOriginal(db *sql.DB, name string) ([]byte, error) {
	var data []byte
	err := db.QueryRow("SELECT data FROM originals WHERE name = ?", name).Scan(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func saveDerivative(db *sql.DB, name string, transform string, data []byte) error {
	now := time.Now().Unix()
	_, err := db.Exec(
		"INSERT INTO derived (original_name, transformation, created, last_access, data) VALUES (?, ?, ?, ?, ?)",
		name, transform, now, now, data,
	)
	return err
}

func buildTransformURL(width int, height int, typ string) string {
	transformURL := "/transform"

	queryParams := ""
	if width > 0 {
		queryParams += fmt.Sprintf("width=%d&", width)
	}
	if height > 0 {
		queryParams += fmt.Sprintf("height=%d&", height)
	}
	queryParams += fmt.Sprintf("type=%s", typ)

	if queryParams != "" {
		transformURL += "?" + queryParams
	}

	return transformURL
}

func transformImage(originalData []byte, transformURL string) ([]byte, int, string, error) {
	resp, err := spinhttp.Post(transformURL, "image/png", bytes.NewReader(originalData))
	if err != nil {
		return nil, 0, "", fmt.Errorf("Error transforming image: %s", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, "", fmt.Errorf("Error reading response: %s", err)
	}

	return body, resp.StatusCode, resp.Header.Get("Content-Type"), nil
}

func handleDerived(w http.ResponseWriter, r *http.Request) {
	width, height, typ, err := getQuery(r)
	if err != nil {
		errorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	name := mux.Vars(r)["name"]
	transform := calculateTransform(width, height, typ)

	db := getDb()
	defer db.Close()

	updateDerivativeLastAccess(db, name, transform)

	data, err := getDerivative(db, name, transform)
	if err == nil {
		w.Header().Set("Content-Type", typ)
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	} else if err != sql.ErrNoRows {
		errorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	originalData, err := getOriginal(db, name)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Image not found")
		return
	} else if err != nil {
		errorResponse(w, fmt.Sprintf("Database error: %s", err), http.StatusInternalServerError)
		return
	}

	transformURL := buildTransformURL(width, height, typ)

	transformedData, statusCode, contentType, err := transformImage(originalData, transformURL)
	if err != nil {
		errorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if statusCode != http.StatusOK {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(statusCode)
		w.Write(transformedData)
		return
	}

	err = saveDerivative(db, name, transform, transformedData)
	if err != nil {
		fmt.Printf("Error saving transformed image: %s\n", err)
	}

	w.Header().Set("Content-Type", typ)
	w.WriteHeader(http.StatusOK)
	w.Write(transformedData)
}

func init() {
	router := mux.NewRouter()

	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		router.HandleFunc("/derived/{name}", handleDerived).Methods("GET")

		router.ServeHTTP(w, r)
	})
}
