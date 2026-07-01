package install

import "testing"

func TestInstallURL(t *testing.T) {
	const want = "https://github.com/apps/octomerge"
	if got := InstallURL(); got != want {
		t.Errorf("InstallURL() = %q, want %q", got, want)
	}
}

func TestResolveOrg(t *testing.T) {
	tests := []struct {
		name          string
		picked, typed string
		want          string
	}{
		{"picked real org", "acme", "", "acme"},
		{"manual sentinel uses typed", manualSentinel, "  my-org ", "my-org"},
		{"empty picked uses typed", "", "typed-org", "typed-org"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveOrg(tt.picked, tt.typed); got != tt.want {
				t.Errorf("resolveOrg(%q, %q) = %q, want %q", tt.picked, tt.typed, got, tt.want)
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
