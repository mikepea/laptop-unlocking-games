// Command aido is the launcher the laptop boots into.
//
//	aido            play
//	aido version    print build identity
//	aido update     check for a newer build, and with --apply, install it
//	aido profile    print the profile path and a summary
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mikepea/aido-linux-unlocker/internal/games"
	"github.com/mikepea/aido-linux-unlocker/internal/games/typing"
	"github.com/mikepea/aido-linux-unlocker/internal/launcher"
	"github.com/mikepea/aido-linux-unlocker/internal/points"
	"github.com/mikepea/aido-linux-unlocker/internal/profile"
	"github.com/mikepea/aido-linux-unlocker/internal/update"
	"github.com/mikepea/aido-linux-unlocker/internal/version"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "aido:", err)
		os.Exit(1)
	}
}

type config struct {
	profilePath string
	playerName  string
	manifestURL string
}

func (c *config) bind(fs *flag.FlagSet) {
	fs.StringVar(&c.profilePath, "profile", os.Getenv("AIDO_PROFILE"),
		"path to the profile JSON (default: $XDG_DATA_HOME/aido/profile.json)")
	fs.StringVar(&c.playerName, "name", envOr("AIDO_PLAYER", "player"),
		"player name, used on a first run")
	fs.StringVar(&c.manifestURL, "manifest-url", os.Getenv("AIDO_MANIFEST_URL"),
		"release manifest URL; empty disables update checks")
}

func (c *config) store() (*profile.Store, error) {
	path := c.profilePath
	if path == "" {
		var err error
		path, err = profile.DefaultPath()
		if err != nil {
			return nil, err
		}
	}
	return profile.NewStore(path), nil
}

func run(args []string) error {
	cmd := ""
	if len(args) > 0 && len(args[0]) > 0 && args[0][0] != '-' {
		cmd, args = args[0], args[1:]
	}

	switch cmd {
	case "", "play":
		return runPlay(args)
	case "version":
		return runVersion()
	case "update":
		return runUpdate(args)
	case "profile":
		return runProfile(args)
	case "help", "-h", "--help":
		usage()
		return nil
	default:
		usage()
		return fmt.Errorf("unknown command %q", cmd)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `aido — the laptop that unlocks

usage:
  aido [play]     start the launcher
  aido version    print build identity
  aido update     check for a newer build (--apply to install it)
  aido profile    show where progress is stored, and a summary

flags are shared by all commands; see aido play -h
`)
}

func runPlay(args []string) error {
	var cfg config
	fs := flag.NewFlagSet("play", flag.ContinueOnError)
	cfg.bind(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}

	store, err := cfg.store()
	if err != nil {
		return err
	}
	prof, err := store.Load(cfg.playerName)
	if err != nil {
		return err
	}

	registry := games.NewRegistry(typing.New())

	model := launcher.New(launcher.Options{
		Store:   store,
		Profile: prof,
		Games:   registry,
		Ledger:  points.NewLocal(prof),
		Checker: &update.Checker{ManifestURL: cfg.manifestURL, Current: version.Version},
	})

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("run launcher: %w", err)
	}
	// The launcher saves after every run; this is the belt to that braces,
	// covering a quit straight from the menu.
	if err := store.Save(prof); err != nil {
		return fmt.Errorf("save profile: %w", err)
	}
	return nil
}

func runVersion() error {
	fmt.Printf("aido %s (%s, built %s)\n", version.Version, version.Commit, version.BuildDate)
	return nil
}

func runUpdate(args []string) error {
	var cfg config
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	cfg.bind(fs)
	apply := fs.Bool("apply", false, "install the update if one is available")
	target := fs.String("target", "", "path to replace (default: this executable)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	checker := &update.Checker{ManifestURL: cfg.manifestURL, Current: version.Version}
	if !checker.Enabled() {
		return errors.New("no manifest URL configured; set -manifest-url or AIDO_MANIFEST_URL")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	available, err := checker.Check(ctx)
	if err != nil {
		return err
	}
	if available == nil {
		fmt.Printf("aido %s is up to date\n", version.Version)
		return nil
	}

	fmt.Printf("update available: %s → %s\n", version.Version, available.Version)
	if available.Notes != "" {
		fmt.Printf("  %s\n", available.Notes)
	}
	if !*apply {
		fmt.Println("re-run with -apply to install it")
		return nil
	}

	path := *target
	if path == "" {
		path, err = os.Executable()
		if err != nil {
			return fmt.Errorf("locate this executable: %w", err)
		}
	}

	applyCtx, applyCancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer applyCancel()
	if err := update.Apply(applyCtx, available, path, &http.Client{Timeout: 10 * time.Minute}); err != nil {
		return err
	}
	fmt.Printf("installed %s to %s\n", available.Version, path)
	return nil
}

func runProfile(args []string) error {
	var cfg config
	fs := flag.NewFlagSet("profile", flag.ContinueOnError)
	cfg.bind(fs)
	asJSON := fs.Bool("json", false, "print the whole profile as JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}

	store, err := cfg.store()
	if err != nil {
		return err
	}
	prof, err := store.Load(cfg.playerName)
	if err != nil {
		return err
	}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(prof)
	}

	fmt.Printf("profile:      %s\n", store.Path())
	fmt.Printf("player:       %s\n", prof.Name)
	fmt.Printf("points:       %d earned, %d available\n", prof.PointsEarned, prof.Balance())
	fmt.Printf("badges:       %d\n", len(prof.Achievements))
	fmt.Printf("unlocked:     %d stages\n", len(prof.Unlocked))
	for id, s := range prof.Games {
		fmt.Printf("  %-10s %d plays, best %.0f wpm, %d lessons cleared\n",
			id, s.Plays, s.BestWPM, s.LessonsCleared)
	}
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
