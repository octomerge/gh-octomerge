package install

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/api"
)

// generatedRepo is the slice of a repository we need from GitHub: its full name
// (owner/name) and browser URL.
type generatedRepo struct {
	FullName string `json:"full_name"`
	HTMLURL  string `json:"html_url"`
}

// RepoExists reports whether owner/name is a repository the authenticated user
// can see. A 404 means it does not exist (or is not visible), which the caller
// treats as free to create; any other error is surfaced. Like the rest of the
// package it reuses gh's credentials via go-gh's default REST client.
func RepoExists(owner, name string) (bool, error) {
	client, err := api.DefaultRESTClient()
	if err != nil {
		return false, fmt.Errorf("creating REST client: %w", err)
	}
	var repo generatedRepo
	if err := client.Get(fmt.Sprintf("repos/%s/%s", owner, name), &repo); err != nil {
		var httpErr *api.HTTPError
		if errors.As(err, &httpErr) && httpErr.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("checking repository %s/%s: %w", owner, name, err)
	}
	return true, nil
}

// CreateFromTemplate generates owner/name from the octomerge template repository
// via POST /repos/octomerge/octomerge/generate, honoring the requested visibility.
func CreateFromTemplate(owner, name string, private bool) (generatedRepo, error) {
	client, err := api.DefaultRESTClient()
	if err != nil {
		return generatedRepo{}, fmt.Errorf("creating REST client: %w", err)
	}
	body, err := templateBody(owner, name, private)
	if err != nil {
		return generatedRepo{}, err
	}
	var repo generatedRepo
	path := fmt.Sprintf("repos/%s/%s/generate", templateOwner, templateRepo)
	if err := client.Post(path, bytes.NewReader(body), &repo); err != nil {
		return generatedRepo{}, fmt.Errorf("creating %s/%s from template: %w", owner, name, err)
	}
	return repo, nil
}

// templateBody builds the JSON request body for the generate endpoint. It is
// factored out from CreateFromTemplate so it can be unit-tested without a network
// call.
func templateBody(owner, name string, private bool) ([]byte, error) {
	payload := struct {
		Owner       string `json:"owner"`
		Name        string `json:"name"`
		Private     bool   `json:"private"`
		Description string `json:"description"`
	}{
		Owner:       owner,
		Name:        name,
		Private:     private,
		Description: "octomerge configuration",
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encoding template request: %w", err)
	}
	return b, nil
}

// visibilityLabel returns the human word for a repository's visibility.
func visibilityLabel(private bool) string {
	if private {
		return "private"
	}
	return "public"
}
