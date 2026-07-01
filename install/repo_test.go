package install

import (
	"encoding/json"
	"testing"
)

func TestTemplateBody(t *testing.T) {
	tests := []struct {
		name    string
		owner   string
		repo    string
		private bool
	}{
		{"private", "acme", ".octomerge", true},
		{"public", "beta", ".octomerge", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, err := templateBody(tt.owner, tt.repo, tt.private)
			if err != nil {
				t.Fatalf("templateBody(%q, %q, %v) error = %v", tt.owner, tt.repo, tt.private, err)
			}
			var got struct {
				Owner       string `json:"owner"`
				Name        string `json:"name"`
				Private     bool   `json:"private"`
				Description string `json:"description"`
			}
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatalf("unmarshal body: %v", err)
			}
			if got.Owner != tt.owner {
				t.Errorf("owner = %q, want %q", got.Owner, tt.owner)
			}
			if got.Name != tt.repo {
				t.Errorf("name = %q, want %q", got.Name, tt.repo)
			}
			if got.Private != tt.private {
				t.Errorf("private = %v, want %v", got.Private, tt.private)
			}
			if got.Description == "" {
				t.Error("description = \"\", want non-empty")
			}
		})
	}
}

func TestVisibilityLabel(t *testing.T) {
	if got := visibilityLabel(true); got != "private" {
		t.Errorf("visibilityLabel(true) = %q, want %q", got, "private")
	}
	if got := visibilityLabel(false); got != "public" {
		t.Errorf("visibilityLabel(false) = %q, want %q", got, "public")
	}
}
