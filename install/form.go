package install

import (
	"errors"
	"strings"

	"charm.land/huh/v2"
)

// manualSentinel is the Select value meaning "let me type an org name instead".
// The NUL byte keeps it from ever colliding with a real organization login.
const manualSentinel = "\x00manual"

type formResult struct {
	Org       Org
	Confirmed bool
}

// runForm renders the TUI: pick an organization (from the user's orgs, with a
// manual-entry fallback) and confirm. It blocks until the form is submitted or
// aborted. A Ctrl-C / Esc abort is reported as an unconfirmed result rather than
// an error, so the caller can exit cleanly.
//
// huh v2's conditional visibility is per-group, so the manual-entry Input lives
// in its own group gated by WithHideFunc; the org Select and the Confirm are
// their own groups. A huh.Form is itself a tea.Model, so this could later be
// embedded in a larger Bubble Tea program instead of calling Run directly.
func runForm(orgs []Org, preselect string) (formResult, error) {
	picked := preselect
	typed := preselect
	confirmed := true

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

	groups = append(groups, huh.NewGroup(
		huh.NewConfirm().
			Key("confirm").
			Title("Open your browser to install the octomerge GitHub App?").
			Affirmative("Yes, open browser").
			Negative("Cancel").
			Value(&confirmed),
	))

	if err := huh.NewForm(groups...).Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return formResult{Confirmed: false}, nil
		}
		return formResult{}, err
	}
	return formResult{Org: resolvePickedOrg(orgs, picked, typed), Confirmed: confirmed}, nil
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
