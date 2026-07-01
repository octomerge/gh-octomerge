package install

import (
	"fmt"

	"github.com/cli/go-gh/v2/pkg/api"
)

// Org is the slice of a GitHub organization the picker needs.
type Org struct {
	Login string `json:"login"`
}

// ListUserOrgs returns the organizations the authenticated gh user belongs to.
// It reuses gh's credentials and host through go-gh's default REST client, so
// there is no token handling here. Note that GET /user/orgs only returns orgs
// whose membership is visible; the form's manual-entry fallback covers the rest.
func ListUserOrgs() ([]Org, error) {
	client, err := api.DefaultRESTClient()
	if err != nil {
		return nil, fmt.Errorf("creating REST client: %w", err)
	}
	var orgs []Org
	if err := client.Get("user/orgs?per_page=100", &orgs); err != nil {
		return nil, fmt.Errorf("listing organizations: %w", err)
	}
	return orgs, nil
}
