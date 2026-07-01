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

// appPage is the octomerge GitHub App's public landing page. It carries the
// "Install" button and lets the user pick which account/org to install on.
const appPage = "https://github.com/apps/octomerge"

// Options configure a single run. They come from command flags; anything left
// unset is gathered interactively via the TUI form.
type Options struct {
	Org         string
	AutoConfirm bool
}

// InstallURL returns the page to open in the browser. It is a function (rather
// than a bare constant) so callers read intent clearly and so switching to a
// deep-link flow later (e.g. appPage+"/installations/new") is a one-line change.
func InstallURL() string { return appPage }

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

	org := opts.Org
	confirmed := opts.AutoConfirm

	if org == "" || !confirmed {
		orgs, err := ListUserOrgs()
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not list your organizations: %v\n", err)
		}
		res, err := runForm(orgs, org)
		if err != nil {
			return err
		}
		if !res.Confirmed {
			fmt.Println("Aborted. No changes made.")
			return nil
		}
		org = res.Org
	}

	if org == "" {
		return errors.New("no organization was provided")
	}

	url := InstallURL()
	fmt.Printf("Opening %s\n", url)
	fmt.Printf("On the install page, choose the %q organization and click Install.\n", org)
	return OpenApp(url)
}
