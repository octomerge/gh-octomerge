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

const (
	// appPage is the octomerge GitHub App's public base URL on github.com. The
	// install flow lives beneath it; see InstallURL.
	appPage = "https://github.com/apps/octomerge"

	// appSlug is the App's slug on github.com - the trailing segment of appPage.
	// It is how an existing installation is spotted among an org's installed apps.
	appSlug = "octomerge"

	// templateOwner/templateRepo identify the template repository the config repo
	// is generated from, via POST /repos/{owner}/{repo}/generate.
	templateOwner = "octomerge"
	templateRepo  = "octomerge"

	// configRepo is the per-org configuration repository octomerge reads. It is
	// created under the org the App is installed on, mirroring GitHub's .github
	// special-repo convention.
	configRepo = ".octomerge"
)

// Options configure a single run. They come from command flags; anything left
// unset is gathered interactively via the TUI form.
type Options struct {
	Org         string
	AutoConfirm bool
	Public      bool
}

// InstallURL returns the browser URL for installing octomerge on org. When the
// org's numeric ID is known it deep-links via suggested_target_id, which
// pre-selects that account and sends the user straight to its install/permissions
// page - no second account pick. Without an ID it falls back to the generic
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

// Run orchestrates the flow: choose the target org, then - unless octomerge is
// already installed on it - open the App install page. Either way it finishes by
// setting up the org's .octomerge configuration repository.
func Run(ctx context.Context, opts Options) error {
	_ = ctx

	target := Org{Login: opts.Org}

	if target.Login == "" {
		orgs, err := ListUserOrgs()
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not list your organizations: %v\n", err)
		}
		org, ok, err := selectOrg(orgs, opts.Org)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Aborted. No changes made.")
			return nil
		}
		target = org
	}

	if target.Login == "" {
		return errors.New("no organization was provided")
	}

	// Best-effort: if the App is already installed, skip the browser entirely and
	// go straight to the config repo. On a detection error, fall back to showing
	// the install step so users without the admin:org scope see no regression.
	installed, err := AppInstalled(target.Login)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: couldn't check whether octomerge is installed on %q (%v); continuing to the install step\n", target.Login, err)
	}

	if installed {
		fmt.Printf("octomerge is already installed on %q; skipping the install page.\n", target.Login)
	} else {
		if target.ID == 0 {
			if id, err := LookupOrgID(target.Login); err != nil {
				fmt.Fprintf(os.Stderr, "warning: couldn't pre-select %q (%v); opening the account picker instead\n", target.Login, err)
			} else {
				target.ID = id
			}
		}

		if !opts.AutoConfirm {
			ok, err := confirmInstall(target)
			if err != nil {
				return err
			}
			if !ok {
				fmt.Println("Aborted. No changes made.")
				return nil
			}
		}

		url := InstallURL(target)
		fmt.Printf("Opening %s\n", url)
		if target.ID != 0 {
			fmt.Printf("This opens the install page for %q - review the permissions and click Install.\n", target.Login)
		} else {
			fmt.Printf("On the install page, choose %q and click Install.\n", target.Login)
		}
		if err := OpenApp(url); err != nil {
			return err
		}
	}

	return setupConfigRepo(target.Login, opts)
}

// setupConfigRepo is the second half of the flow: after the App install page is
// open, generate the org's .octomerge configuration repository from the octomerge
// template. If the repo already exists it says so and stops. Otherwise it decides
// visibility - interactively (a confirm plus a Private/Public picker) or, with
// --yes, unattended and private unless --public - then creates it and prints the
// new repository's URL.
func setupConfigRepo(owner string, opts Options) error {
	exists, err := RepoExists(owner, configRepo)
	if err != nil {
		return err
	}
	if exists {
		fmt.Printf("The repository %s already exists\n", configRepo)
		return nil
	}

	private := !opts.Public
	if !opts.AutoConfirm {
		res, err := runConfigForm(owner)
		if err != nil {
			return err
		}
		if !res.Proceed {
			fmt.Println("Skipped. Run `gh octomerge` again to create the .octomerge repository later.")
			return nil
		}
		private = res.Private
	}

	fmt.Printf("Creating %s/%s from the octomerge template…\n", owner, configRepo)
	repo, err := CreateFromTemplate(owner, configRepo, private)
	if err != nil {
		return err
	}
	fmt.Printf("Created %s (%s): %s\n", repo.FullName, visibilityLabel(private), repo.HTMLURL)
	return nil
}
