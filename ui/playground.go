package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Example struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type runRequest struct {
	Code string `json:"code"`
}

type runResponse struct {
	Output string `json:"output,omitempty"`
	Error  string `json:"error,omitempty"`
}

func main() {
	if err := ensureBinary(); err != nil {
		log.Fatalf("failed to build borno binary: %v", err)
	}

	http.Handle("/", http.FileServer(http.Dir("web")))
	http.HandleFunc("/examples", handleExamples)
	http.HandleFunc("/run", handleRun)

	addr := ":3000"
	log.Println("Playground listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func ensureBinary() error {
	if _, err := os.Stat("borno"); os.IsNotExist(err) {
		cmd := exec.Command("go", "build", "-o", "borno")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	return nil
}

func handleExamples(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir("example")
	if err != nil {
		http.Error(w, "could not read examples", http.StatusInternalServerError)
		return
	}

	var examples []Example
	for _, f := range files {
		if filepath.Ext(f.Name()) != ".bn" {
			continue
		}
		b, err := os.ReadFile(filepath.Join("example", f.Name()))
		if err != nil {
			continue
		}
		examples = append(examples, Example{Name: f.Name(), Code: string(b)})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"examples": examples})
}

func handleRun(w http.ResponseWriter, r *http.Request) {
	var req runRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	tmp, err := os.CreateTemp("", "*.bn")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(req.Code); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmp.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "./borno", tmp.Name())
	out, err := cmd.CombinedOutput()
	resp := runResponse{}
	if err != nil {
		resp.Error = string(out)
	} else {
		resp.Output = string(out)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
