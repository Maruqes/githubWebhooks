package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

var RepoPath string
var Secret string

func init() {
	RepoPath = os.Getenv("REPO_PATH")
	if RepoPath == "" {
		fmt.Println("REPO_PATH environment variable is required")
		os.Exit(1)
	}

	Secret = os.Getenv("SECRET")
	if Secret == "" {
		fmt.Println("SECRET environment variable is required")
		os.Exit(1)
	}
}

// Payload represents the structure of the GitHub webhook payload
type Payload struct {
	Ref string `json:"ref"`
}

func main() {
	http.HandleFunc("/webhook", handleWebhook)

	fmt.Println("Listening on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	// Validate the request
	signature := r.Header.Get("X-Hub-Signature-256")
	if !validateSignature(r.Body, signature) {
		http.Error(w, "Invalid signature", http.StatusForbidden)
		return
	}

	// Parse the payload
	var payload Payload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Check if the push is to the main branch
	if payload.Ref == "refs/heads/main" {
		fmt.Println("Received push to main branch. Pulling changes...")
		if err := pullChanges(); err != nil {
			http.Error(w, "Failed to pull changes", http.StatusInternalServerError)
			fmt.Printf("Error pulling changes: %v\n", err)
			return
		}
		fmt.Println("Repository updated successfully!")
	}

	w.WriteHeader(http.StatusOK)
}

func validateSignature(body io.Reader, signature string) bool {
	mac := hmac.New(sha256.New, []byte(Secret))
	data, _ := io.ReadAll(body)
	mac.Write(data)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func pullChanges() error {
	cmd := exec.Command("git", "-C", RepoPath, "pull", "origin", "main")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
