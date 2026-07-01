package install

import (
	"errors"
	"fmt"
	"strings"

	"charm.land/huh/v2"
)

// manualSentinel is the Select value meaning "let me type an org name instead".
// The NUL byte keeps it from ever colliding with a real organization login.
const manualSentinel = "\x00manual"

// selectOrg renders the org picker: choose from the user's orgs (with a
// manual-entry fallback) or, when none are listable, type one directly. It blocks
// until submitted or aborted, and returns the chosen org plus whether the user
// proceeded — a Ctrl-C / Esc abort is reported as proceeded=false rather than an
// error so the caller can exit cleanly.
//
// huh v2's conditional visibility is per-group, so the manual-entry Input lives in
// its own group gated by WithHideFunc. Confirming whether to open the browser is a
// separate step (confirmInstall), so it can be skipped when the App is already
// installed.
func selectOrg(orgs []Org, preselect string) (Org, bool, error) {
	picked := preselect
	typed := preselect

	var groups []*huh.Group

	if len(orgs) > 0 {
		opts := make([]huh.Option[string], 0, len(orgs)+1)
		for _, o := range orgs {
			opts = append(opts, huh.NewOption(o.Login, o.Login))
		}
		opts = append(opts, huh.NewOption("Enter an organization manually…", manualSentinel))

		groups = append(groups,
			huh.NewGroup(
				huh.NewSelect[string]().
					Key("org").
					Title("Which organization should octomerge be installed on?").
					Description("Pick one of your organizations, or choose manual entry.").
					Options(opts...).
					Value(&picked),
			),
			// Shown only when the user chooses manual entry.
			huh.NewGroup(
				huh.NewInput().
					Key("org_manual").
					Title("Organization login").
					Placeholder("my-org").
					Value(&typed).
					Validate(requireNonEmpty),
			).WithHideFunc(func() bool { return picked != manualSentinel }),
		)
	} else {
		// No orgs to list (none visible, or the API call failed): ask directly.
		picked = manualSentinel
		groups = append(groups, huh.NewGroup(
			huh.NewInput().
				Key("org_manual").
				Title("Organization login").
				Description("The GitHub organization to install octomerge on.").
				Placeholder("my-org").
				Value(&typed).
				Validate(requireNonEmpty),
		))
	}

	if err := huh.NewForm(groups...).Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return Org{}, false, nil
		}
		return Org{}, false, err
	}
	return resolvePickedOrg(orgs, picked, typed), true, nil
}

// confirmInstall asks whether to open the browser to install the App on org. It
// is shown only when the App is not already installed. A Ctrl-C / Esc abort is
// reported as false rather than an error so the caller can exit cleanly.
func confirmInstall(org Org) (bool, error) {
	proceed := true
	group := huh.NewGroup(
		huh.NewConfirm().
			Key("confirm").
			Title(fmt.Sprintf("Open your browser to install the octomerge GitHub App on %q?", org.Login)).
			Affirmative("Yes, open browser").
			Negative("Cancel").
			Value(&proceed),
	)
	if err := huh.NewForm(group).Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return false, nil
		}
		return false, err
	}
	return proceed, nil
}

// resolvePickedOrg maps the form selection back to an Org. A pick from the list
// returns the matching Org (carrying its numeric ID); manual entry, no selection,
// or a login not in the list returns an Org with just the login and ID 0, which
// the caller resolves via the API.
func resolvePickedOrg(orgs []Org, picked, typed string) Org {
	if picked != "" && picked != manualSentinel {
		for _, o := range orgs {
			if o.Login == picked {
				return o
			}
		}
		return Org{Login: picked}
	}
	return Org{Login: strings.TrimSpace(typed)}
}

func requireNonEmpty(s string) error {
	if strings.TrimSpace(s) == "" {
		return errors.New("organization is required")
	}
	return nil
}

// configResult carries the second form's answers: whether to proceed with
// creating the .octomerge repository, and whether it should be private.
type configResult struct {
	Proceed bool
	Private bool
}

// runConfigForm renders the second TUI, shown after the App install page opens:
// confirm the user has installed the App, then pick the .octomerge repository's
// visibility (Private by default). As in runForm, the visibility group is its own
// group gated by WithHideFunc, and a Ctrl-C / Esc abort is reported as a
// not-proceeded result rather than an error so the caller exits cleanly.
func runConfigForm(owner string) (configResult, error) {
	proceed := true
	visibility := "private"

	groups := []*huh.Group{
		huh.NewGroup(
			huh.NewConfirm().
				Key("proceed").
				Title(fmt.Sprintf("Create the %s/%s configuration repository?", owner, configRepo)).
				Description("Do this once you've installed octomerge in the browser.").
				Affirmative("Yes, create it").
				Negative("Skip").
				Value(&proceed),
		),
		// Shown only when the user chooses to proceed.
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("visibility").
				Title("Repository visibility").
				Options(
					huh.NewOption("Private", "private"),
					huh.NewOption("Public", "public"),
				).
				Value(&visibility),
		).WithHideFunc(func() bool { return !proceed }),
	}

	if err := huh.NewForm(groups...).Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return configResult{Proceed: false}, nil
		}
		return configResult{}, err
	}
	return configResult{Proceed: proceed, Private: visibility == "private"}, nil
}
