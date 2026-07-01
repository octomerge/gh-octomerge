---
name: charm-tui
description: >
  Build terminal UIs in Go with the 2026 Charm stack - huh v2 forms and Bubble Tea v2.
  Use when adding or editing TUI forms or prompts: documents the charm.land vanity import
  paths, the huh form pattern, reading values, conditional (per-group) fields, theming,
  and embedding a huh.Form as a tea.Model.
---

# Charm TUIs: huh v2 + Bubble Tea v2

Guidance for the 2026 Charm ecosystem, as used by this repo's `install/form.go`.

## When to activate

- Adding or editing any terminal form, prompt, or Bubble Tea UI in Go
- Migrating Charm code from v1 to v2
- Debugging huh field visibility, values, or theming

## Vanity import paths (v2)

Charm moved to `charm.land/*` vanity paths on Bubble Tea v2. Use these, not the old
`github.com/charmbracelet/*` v1 paths:

    huh          charm.land/huh/v2         (v2.0.3)
    bubbletea    charm.land/bubbletea/v2   (v2.0.2)
    bubbles      charm.land/bubbles/v2     (v2.0.0)
    lipgloss     charm.land/lipgloss/v2    (v2.0.1)

Exception: fang is still `github.com/charmbracelet/fang`.

## Minimal huh form

A `huh.Form` bundles one or more `*huh.Group`s of fields. Bind each field's value with
`.Value(&ptr)` and read the variables after `Run()`:

    var name string
    var ok bool
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewInput().Title("Name").Value(&name),
            huh.NewConfirm().Title("Continue?").Value(&ok),
        ),
    )
    if err := form.Run(); err != nil {   // huh.ErrUserAborted on Ctrl-C / Esc
        return err
    }

Values can also be read with `form.GetString("key")` / `form.GetBool("key")` when you set
`.Key(...)` on the field.

## Select with a typed value

`NewSelect` is generic; build options with `NewOption[T](label, value)`:

    huh.NewSelect[string]().
        Title("Pick one").
        Options(huh.NewOption("A", "a"), huh.NewOption("B", "b")).
        Value(&picked)

## Conditional fields are PER-GROUP

huh v2 has no per-field hide. Put the conditional field in its OWN group and gate the
group with `WithHideFunc`:

    huh.NewGroup(
        huh.NewInput().Title("Manual entry").Value(&typed),
    ).WithHideFunc(func() bool { return picked != manualSentinel })

Groups render as sequential pages; a hidden group is skipped during navigation.

## Theming

The default theme is already "Charm". Choose another with
`form.WithTheme(huh.ThemeCharm(true))`, `huh.ThemeDracula(true)`, etc. - theme
constructors take an `isDark bool`.

## huh.Form is a tea.Model

For chrome around the form, embed it in your own Bubble Tea model. `Form.Update` returns
the `tea.Model` interface, so type-assert back:

    func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
        f, cmd := m.form.Update(msg)
        if hf, ok := f.(*huh.Form); ok { m.form = hf }
        if m.form.State == huh.StateCompleted { return m, tea.Quit }
        return m, cmd
    }

Otherwise `form.Run()` spins up its own program and is simplest.

## Gotchas

- Handle `huh.ErrUserAborted` (Ctrl-C / Esc) as a clean exit, not a hard error.
- A Select needs ≥1 option; guard the empty case (fall back to an Input).
- Forms need a TTY; provide a non-interactive fallback (flags) for scripts/CI.
