package install

import "testing"

func TestInstallURL(t *testing.T) {
	tests := []struct {
		name string
		org  Org
		want string
	}{
		{
			"deep link when id known",
			Org{Login: "acme", ID: 12345},
			"https://github.com/apps/octomerge/installations/new/permissions?suggested_target_id=12345",
		},
		{
			"fallback when id unknown",
			Org{Login: "acme"},
			"https://github.com/apps/octomerge/installations/new",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InstallURL(tt.org); got != tt.want {
				t.Errorf("InstallURL(%+v) = %q, want %q", tt.org, got, tt.want)
			}
		})
	}
}

func TestResolvePickedOrg(t *testing.T) {
	orgs := []Org{{Login: "acme", ID: 1}, {Login: "beta", ID: 2}}
	tests := []struct {
		name          string
		picked, typed string
		want          Org
	}{
		{"pick from list keeps id", "acme", "", Org{Login: "acme", ID: 1}},
		{"manual sentinel trims typed", manualSentinel, "  my-org ", Org{Login: "my-org"}},
		{"empty picked uses typed", "", "typed-org", Org{Login: "typed-org"}},
		{"picked not in list", "ghost", "", Org{Login: "ghost"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolvePickedOrg(orgs, tt.picked, tt.typed); got != tt.want {
				t.Errorf("resolvePickedOrg(_, %q, %q) = %+v, want %+v", tt.picked, tt.typed, got, tt.want)
			}
		})
	}
}

func TestRequireNonEmpty(t *testing.T) {
	if err := requireNonEmpty("  "); err == nil {
		t.Error("requireNonEmpty(blank) = nil, want error")
	}
	if err := requireNonEmpty("acme"); err != nil {
		t.Errorf("requireNonEmpty(acme) = %v, want nil", err)
	}
}
