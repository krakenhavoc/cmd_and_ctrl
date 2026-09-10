// Package appenv models which deployment this process is: production
// or the develop preview environment. It is the single gate for every
// dev-only feature.
//
// The rule the rest of the server relies on: a dev feature is OFF in
// production and cannot be turned on there. Per-feature knobs only
// subtract inside a dev deployment; they can never add one to prod.
// That makes a leaked CMDCTRL_DEV_* line in the production env file
// inert rather than a privilege escalation, which matters because
// these features (spawn arbitrary cards, mutate state, drive other
// players' seats) would be outright cheats on a live table.
package appenv

import (
	"fmt"
	"os"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/envflag"
)

// Env is the deployment identity, from CMDCTRL_ENV.
type Env string

const (
	// EnvProd is the default when CMDCTRL_ENV is unset or empty.
	// Defaulting to prod means a forgotten variable fails closed.
	EnvProd Env = "prod"
	// EnvDev is the develop-branch preview deployment.
	EnvDev Env = "dev"
)

// EnvVar is the environment variable Parse reads.
const EnvVar = "CMDCTRL_ENV"

// IsDev reports whether dev-only surfaces are permitted at all.
func (e Env) IsDev() bool { return e == EnvDev }

// String implements fmt.Stringer.
func (e Env) String() string { return string(e) }

// Parse maps a raw CMDCTRL_ENV value onto an Env. Empty is prod.
// Anything unrecognised is an error rather than a silent fallback:
// "CMDCTRL_ENV=development" quietly behaving as prod would be a
// confusing way to lose every dev feature on the dev box.
func Parse(raw string) (Env, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "":
		return EnvProd, nil
	case "prod", "production":
		return EnvProd, nil
	case "dev", "develop", "development":
		return EnvDev, nil
	default:
		return EnvProd, fmt.Errorf("unknown %s value %q (want \"prod\" or \"dev\")", EnvVar, raw)
	}
}

// Features is the set of dev-only capabilities this process exposes.
// The zero value — everything off — is what production always gets.
type Features struct {
	// CardSpawn allows putting an arbitrary Scryfall card into an
	// arbitrary zone, and mutating life / counters / phase.
	CardSpawn bool `json:"card_spawn"`
	// SeatSwap lets one connection act as any seat in the game, so a
	// four-player game can be driven from a single browser.
	SeatSwap bool `json:"seat_swap"`
	// FrameInspector is a client-side raw WebSocket frame log. The
	// server exposes no route for it; the flag exists so the drawer
	// can be switched off independently while testing prod-alike UI.
	FrameInspector bool `json:"frame_inspector"`
	// ReplayScrubber enables stepping a downloaded replay frame by
	// frame, plus the deterministic-shuffle seed.
	ReplayScrubber bool `json:"replay_scrubber"`
}

// featureVars maps each feature onto its per-feature override
// variable. Order is the declaration order of Features.
var featureVars = []struct {
	name string
	get  func(*Features) *bool
}{
	{"CMDCTRL_DEV_CARD_SPAWN", func(f *Features) *bool { return &f.CardSpawn }},
	{"CMDCTRL_DEV_SEAT_SWAP", func(f *Features) *bool { return &f.SeatSwap }},
	{"CMDCTRL_DEV_FRAME_INSPECTOR", func(f *Features) *bool { return &f.FrameInspector }},
	{"CMDCTRL_DEV_REPLAY_SCRUBBER", func(f *Features) *bool { return &f.ReplayScrubber }},
}

// LoadFeatures returns the dev features enabled for env.
//
// In prod the answer is always the zero value, regardless of what the
// per-feature variables say — see the package doc. In dev every
// feature defaults ON (the point of the dev box is to exercise them)
// and each CMDCTRL_DEV_* variable can turn one OFF, which is how you
// check that a surface is genuinely hidden before promoting to main.
func LoadFeatures(env Env) Features {
	if !env.IsDev() {
		return Features{}
	}
	f := Features{
		CardSpawn:      true,
		SeatSwap:       true,
		FrameInspector: true,
		ReplayScrubber: true,
	}
	for _, fv := range featureVars {
		raw, present := os.LookupEnv(fv.name)
		if !present {
			continue
		}
		*fv.get(&f) = envflag.Truthy(raw)
	}
	return f
}

// StrayProdOverrides returns the CMDCTRL_DEV_* feature variables that
// are set to a truthy value while env is prod. They have no effect;
// main logs them at WARN so a copy-pasted env file is caught on the
// next boot rather than the next incident.
func StrayProdOverrides(env Env) []string {
	if env.IsDev() {
		return nil
	}
	var stray []string
	for _, fv := range featureVars {
		if envflag.Truthy(os.Getenv(fv.name)) {
			stray = append(stray, fv.name)
		}
	}
	return stray
}
