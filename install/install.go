// Package install holds the behavior of the extension: discovering the user's
// organizations, rendering the TUI form, and opening the octomerge GitHub App
// install page. It is deliberately decoupled from the command layer (cmd) and
// from TUI rendering so each piece stays testable in isolation.
package install

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/cli/go-gh/v2/pkg/browser"
)

// appPage is the octomerge GitHub App's public base URL on github.com. The
// install flow lives beneath it; see InstallURL.
const appPage = "https://github.com/apps/octomerge"

// Options configure a single run. They come from command flags; anything left
// unset is gathered interactively via the TUI form.
type Options struct {
	Org         string
	AutoConfirm bool
}

// InstallURL returns the browser URL for installing octomerge on org. When the
// org's numeric ID is known it deep-links via suggested_target_id, which
// pre-selects that account and sends the user straight to its install/permissions
// page — no second account pick. Without an ID it falls back to the generic
// install flow, where GitHub shows the account picker.
func InstallURL(org Org) string {
	if org.ID != 0 {
		return fmt.Sprintf("%s/installations/new/permissions?suggested_target_id=%d", appPage, org.ID)
	}
	return appPage + "/installations/new"
}

// OpenApp opens url using gh's configured browser, honoring GH_BROWSER, then the
// gh config `browser` option, then BROWSER, then the OS default.
func OpenApp(url string) error {
	b := browser.New("", os.Stdout, os.Stderr)
	if err := b.Browse(url); err != nil {
		return fmt.Errorf("opening browser: %w", err)
	}
	return nil
}

// Run orchestrates the flow: gather the target org and confirmation (from flags
// and/or the TUI), then open the install page.
func Run(ctx context.Context, opts Options) error {
	_ = ctx

	target := Org{Login: opts.Org}
	confirmed := opts.AutoConfirm

	if target.Login == "" || !confirmed {
		orgs, err := ListUserOrgs()
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not list your organizations: %v\n", err)
		}
		res, err := runForm(orgs, opts.Org)
		if err != nil {
			return err
		}
		if !res.Confirmed {
			fmt.Println("Aborted. No changes made.")
			return nil
		}
		target = res.Org
	}

	if target.Login == "" {
		return errors.New("no organization was provided")
	}

	if target.ID == 0 {
		if id, err := LookupOrgID(target.Login); err != nil {
			fmt.Fprintf(os.Stderr, "warning: couldn't pre-select %q (%v); opening the account picker instead\n", target.Login, err)
		} else {
			target.ID = id
		}
	}

	url := InstallURL(target)
	fmt.Printf("Opening %s\n", url)
	if target.ID != 0 {
		fmt.Printf("This opens the install page for %q — review the permissions and click Install.\n", target.Login)
	} else {
		fmt.Printf("On the install page, choose %q and click Install.\n", target.Login)
	}
	return OpenApp(url)
}
