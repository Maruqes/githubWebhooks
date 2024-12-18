package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"
)

var Secret string
var Port string

type Repo struct {
	Name string
	Path string
}

var repos []Repo

type PushEvent struct {
	Ref        string `json:"ref"`
	Before     string `json:"before"`
	After      string `json:"after"`
	Repository struct {
		ID       int    `json:"id"`
		NodeID   string `json:"node_id"`
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		Private  bool   `json:"private"`
		Owner    struct {
			Name      string `json:"name"`
			Email     string `json:"email"`
			Login     string `json:"login"`
			ID        int    `json:"id"`
			NodeID    string `json:"node_id"`
			AvatarURL string `json:"avatar_url"`
			URL       string `json:"url"`
			HtmlURL   string `json:"html_url"`
		} `json:"owner"`
		HtmlURL string `json:"html_url"`
		URL     string `json:"url"`
	} `json:"repository"`
	Pusher struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"pusher"`
	HeadCommit struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		Author  struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Username string `json:"username"`
		} `json:"author"`
		Committer struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Username string `json:"username"`
		} `json:"committer"`
		Modified []string `json:"modified"`
	} `json:"head_commit"`
}

func initInit() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
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

	for i := 0; i < 100; i++ {
		repo := os.Getenv(fmt.Sprintf("REPO_PATH%d", i))
		if repo == "" {
			break
		}
		//only get last folder
		repoName := repo[strings.LastIndex(repo, "/")+1:]
		repos = append(repos, Repo{
			Name: repoName,
			Path: repo,
		})

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
	// Read the request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Check Content-Type and decode payload if necessary
	var payloadBytes []byte
	contentType := r.Header.Get("Content-Type")
	if contentType == "application/x-www-form-urlencoded" {
		// Parse URL-encoded payload
		values, err := url.ParseQuery(string(bodyBytes))
		if err != nil {
			http.Error(w, "Invalid URL-encoded payload", http.StatusBadRequest)
			return
		}
		payload := values.Get("payload")
		payloadBytes = []byte(payload)
	} else if contentType == "application/json" {
		payloadBytes = bodyBytes
	} else {
		http.Error(w, "Unsupported Content-Type", http.StatusBadRequest)
		return
	}

	// Validate the signature
	signature := r.Header.Get("X-Hub-Signature-256")
	verifySignature(payloadBytes, signature)

	// Parse the JSON payload
	var payload PushEvent
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		fmt.Printf("Error parsing JSON: %v\n", err)
		return
	}

	// Check if the push is to the main branch
	if payload.Ref == "refs/heads/main" {
		fmt.Println("Received push to main branch. Pulling changes...")
		if err := pullChanges(payload.Repository.Name); err != nil {
			http.Error(w, "Failed to pull changes", http.StatusInternalServerError)
			fmt.Printf("Error pulling changes: %v\n", err)
			return
		}
		fmt.Println("Repository updated successfully!")
	}

	w.WriteHeader(http.StatusOK)
}

func verifySignature(payloadBody []byte, receivedSignature string) error {
	// Get the secret token from the environment
	secretToken := os.Getenv("SECRET_TOKEN")
	if secretToken == "" {
		return fmt.Errorf("SECRET_TOKEN is not set in the environment")
	}

	// Generate the expected signature
	mac := hmac.New(sha256.New, []byte(secretToken))
	mac.Write(payloadBody)
	expectedSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	// Securely compare the received signature with the expected signature
	if !secureCompare(expectedSignature, receivedSignature) {
		return fmt.Errorf("signatures didn't match")
	}

	return nil
}

func secureCompare(a, b string) bool {
	// Securely compare two strings to avoid timing attacks
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}

	return result == 0
}

func pullChanges(repoName string) error {
	for _, repo := range repos {
		if repoName == repo.Name {
			cmd := exec.Command("git", "-C", repo.Path, "pull", "origin", "main")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		}
	}
	return fmt.Errorf("Repository not found")
}
