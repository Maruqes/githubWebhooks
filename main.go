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

	"github.com/joho/godotenv"
)

var RepoPath string
var Secret string
var Port string

// Payload represents the structure of the GitHub webhook payload
type Payload struct {
	Ref string `json:"ref"`
}

func initInit() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		os.Exit(1)
	}

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

	Port = os.Getenv("PORT")
	if Port == "" {
		Port = "8080" // Default to port 8080 if not specified
	}
}

func main() {
	initInit()
	http.HandleFunc("/webhook", handleWebhook)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	fmt.Printf("Listening on port %s...\n", Port)
	if err := http.ListenAndServe(":"+Port, nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	// Read the request body once
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Validate the signature
	signature := r.Header.Get("X-Hub-Signature-256")
	if !validateSignature(bodyBytes, signature) {
		http.Error(w, "Invalid signature", http.StatusForbidden)
		return
	}

	// Parse the payload
	var payload Payload
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
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

func validateSignature(body []byte, signature string) bool {
	mac := hmac.New(sha256.New, []byte(Secret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func pullChanges() error {
	cmd := exec.Command("git", "-C", RepoPath, "pull", "origin", "main")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
