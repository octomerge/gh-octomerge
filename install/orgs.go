package install

import (
	"fmt"

	"github.com/cli/go-gh/v2/pkg/api"
)

// Org is the slice of a GitHub organization the picker needs: its login and the
// numeric account ID, which the deep-link install URL passes as
// suggested_target_id to pre-select the org.
type Org struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
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

// LookupOrgID resolves an organization login to its numeric account ID via
// GET /orgs/{org}. The picker already knows the ID for orgs it listed; this
// covers the paths where it doesn't — manual entry and the --org flag — so the
// install URL can still pre-select the org via suggested_target_id.
func LookupOrgID(login string) (int64, error) {
	client, err := api.DefaultRESTClient()
	if err != nil {
		return 0, fmt.Errorf("creating REST client: %w", err)
	}
	var org Org
	if err := client.Get("orgs/"+login, &org); err != nil {
		return 0, fmt.Errorf("looking up organization %q: %w", login, err)
	}
	return org.ID, nil
}
