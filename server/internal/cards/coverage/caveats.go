package coverage

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// caveats.go — the second drift guard: a Spec.Caveats entry that has
// stopped being true.
//
// # The failure being hunted
//
// Spec.Caveats (ADR 0042) is published verbatim on the public catalog
// page, so a caveat is a promise to a player about what the engine
// will not do. #412's author measured 24 of 87 declared
// simplifications already describing gaps that had since closed. The
// mechanism is structural, not sloppiness: a card file's caveat is
// written in one PR and the mechanic it names is implemented in
// another, weeks later, by someone who never opens that file. Four
// more turned up in a single session in September — faithless_looting
// ("Flashback isn't implemented", flashback shipped in #411),
// weftstalker_ardent ("Warp isn't implemented", #413),
// fierce_guardianship ("the free cast is not offered", #428) and the
// surveil lands ("the surveil 1 on entry never happens", #405).
//
// # Scope, stated honestly
//
// Free text cannot be verified in general. There is no check that
// reads "the sacrifice is an additional cost to cast, not a
// resolution-time action" and decides whether it is still true, and
// pretending otherwise would produce a guard that passes on
// everything it cannot see — worse than a small guard that says what
// it covers.
//
// So this is a CURATED TABLE, not an analysis. Each entry pairs a
// mechanic's name (as a caveat would spell it) with a probe that asks
// the registry a question it can actually answer. A caveat that names
// a mechanic the card demonstrably implements is a contradiction, and
// the build says so. A caveat about anything not in the table is
// invisible to this file, on purpose, and the table is the honest
// statement of coverage. Widening it is one struct literal.
//
// Two probe strengths, and the difference is stated on every failure:
//
//   - Exact: the spec LITERALLY DECLARES the thing — an
//     AlternativeCost under a known key, a zone in CastableZones.
//     There is no judgement in it and no false positive is possible.
//   - Heuristic: the spec grew the hook the caveat says it lacks. The
//     surveil lands are the case that needs it: nothing in the engine
//     is named "surveil" today, so the only observable signal is that
//     a land whose caveat says its entry trigger never happens has
//     acquired an entry trigger. A land that grows an ETB hook for
//     some OTHER reason would trip it — that is the known false
//     positive, and the fix when it happens is to tighten the probe
//     here (once surveil is a primitive, probe for the primitive),
//     not to delete the entry.

// Confidence records how much a probe's answer is worth, and is
// quoted in every failure so the reader knows whether they are
// looking at a fact or an inference.
type Confidence int

const (
	// Exact: the probe reads a declaration off the Spec. If it says
	// the card implements the mechanic, the card implements it.
	Exact Confidence = iota

	// Heuristic: the probe reads a proxy. A true answer is strong
	// evidence and not proof; the failure message says so and names
	// this file as the place to tighten it.
	Heuristic
)

func (c Confidence) String() string {
	if c == Exact {
		return "exact"
	}
	return "heuristic"
}

// Mechanic is one row of the curated table: a thing a caveat can
// claim the engine does not do, and a way to ask the registry whether
// this card now does it.
type Mechanic struct {
	// Name is the canonical spelling, used in failure text.
	Name string

	// Phrases are the ways a caveat might name the mechanic. Matched
	// case-insensitively on word boundaries, so "flashback" does not
	// match "flashbacks" only by accident of substring — it matches
	// because \b lets it, and "warp" does not match "warpath".
	Phrases []string

	// Implements reports whether THIS spec demonstrably implements
	// the mechanic.
	Implements func(effects.Spec) bool

	// Evidence says in one clause what Implements looked at. It is
	// printed in the failure so the reader can check the probe's
	// reasoning without opening this file.
	Evidence string

	// Confidence grades Implements. See the constants.
	Confidence Confidence

	// Adopt is the one-line instruction for a card that could now
	// carry the mechanic but does not. Printed by the adoptable-gap
	// report, where the whole point is that the change is small.
	Adopt string
}

// altCost builds the exact probe shared by every keyword that rides
// Spec.AlternativeCosts: ask the live engine, through the same
// exported lookup the cast path uses, whether this card offers a cost
// under that key.
//
// Routing through game.AlternativeCostByKey rather than ranging over
// spec.AlternativeCosts is deliberate. It goes through the
// CatalogAlternativeCosts hook that effects installs at init, so if
// that wiring ever breaks, these probes go quiet — and
// TestEveryExactMechanicHasAnImplementor notices the silence.
func altCost(key string) func(effects.Spec) bool {
	return func(s effects.Spec) bool {
		return game.AlternativeCostByKey(s.OracleID, key) != nil
	}
}

// mechanics is the table. Order is the order failures are reported
// in.
var mechanics = []Mechanic{
	{
		Name:       "flashback",
		Phrases:    []string{"flashback"},
		Implements: altCost("flashback"),
		Evidence:   `game.AlternativeCostByKey(oracleID, "flashback") resolves`,
		Confidence: Exact,
		Adopt:      `CastableZones: []game.ZoneKind{game.ZoneGraveyard} plus Flashback("{cost}") — see alternative_cost.go`,
	},
	{
		Name:       "warp",
		Phrases:    []string{"warp"},
		Implements: altCost("warp"),
		Evidence:   `game.AlternativeCostByKey(oracleID, "warp") resolves`,
		Confidence: Exact,
		Adopt:      `AlternativeCosts: []game.AlternativeCost{Warp("{cost}")} — the exile clause rides the constructor`,
	},
	{
		Name: "free cast",
		Phrases: []string{
			"free cast", "free spell", "cast it for free",
			"without paying its mana cost",
		},
		Implements: altCost("free"),
		Evidence:   `game.AlternativeCostByKey(oracleID, "free") resolves`,
		Confidence: Exact,
		Adopt:      `AlternativeCosts: []game.AlternativeCost{FreeIfYouControlCommander("…")} — the Commander Legends cycle is two lines per card`,
	},
	{
		Name:       "evoke",
		Phrases:    []string{"evoke"},
		Implements: altCost("evoke"),
		Evidence:   `game.AlternativeCostByKey(oracleID, "evoke") resolves`,
		Confidence: Exact,
		Adopt:      `Evoke("{cost}") or EvokePitch(…) — both bundle SacrificeOnEntry`,
	},
	{
		Name:       "overload",
		Phrases:    []string{"overload"},
		Implements: altCost("overload"),
		Evidence:   `game.AlternativeCostByKey(oracleID, "overload") resolves`,
		Confidence: Exact,
		Adopt:      `Overload("{cost}") — it also clears the target clause`,
	},
	{
		Name:       "cleave",
		Phrases:    []string{"cleave"},
		Implements: altCost("cleave"),
		Evidence:   `game.AlternativeCostByKey(oracleID, "cleave") resolves`,
		Confidence: Exact,
		Adopt:      `Cleave("{cost}", widerTargets)`,
	},
	{
		Name:       "pitch cost",
		Phrases:    []string{"pitch", "force of will"},
		Implements: altCost("pitch"),
		Evidence:   `game.AlternativeCostByKey(oracleID, "pitch") resolves`,
		Confidence: Exact,
		Adopt:      `Pitch(label, life, from, payLabel)`,
	},
	{
		Name:       "pay life instead",
		Phrases:    []string{"pay life instead", "pay 4 life rather", "life rather than"},
		Implements: altCost("pay_life"),
		Evidence:   `game.AlternativeCostByKey(oracleID, "pay_life") resolves`,
		Confidence: Exact,
		Adopt:      `PayLifeInstead(label, life, condition)`,
	},
	{
		// Phrases here are deliberately CAST-shaped. "from your
		// graveyard" on its own is far too common in caveat prose —
		// City of Traitors talks about a land returning from the
		// graveyard and Impulsive Pilferer about encore — and a probe
		// that fired on those would be reporting the wrong mechanic
		// with total confidence, which is the failure mode this file
		// is supposed to be immune to.
		Name: "casting from the graveyard",
		Phrases: []string{
			"cast from the graveyard", "cast from your graveyard",
			"cast it from the graveyard", "cast it from your graveyard",
		},
		Implements: func(s effects.Spec) bool {
			for _, z := range s.CastableZones {
				if z == game.ZoneGraveyard {
					return true
				}
			}
			return false
		},
		Evidence:   "the spec lists game.ZoneGraveyard in CastableZones",
		Confidence: Exact,
		Adopt:      `CastableZones: []game.ZoneKind{game.ZoneGraveyard}`,
	},
	{
		// The heuristic one, and the reason Confidence exists.
		//
		// Surveil is not in the engine at the time of writing: there
		// is no primitive, no PendingChoice kind and no exported
		// symbol to point a probe at, so "does the engine surveil?"
		// has no honest answer from here. What IS observable is the
		// shape of the three cards that promise it — Meticulous
		// Archive, Shadowy Backstreet, Undercity Sewers are lands
		// whose ONLY printed trigger is the surveil, and whose specs
		// today declare no entry hook at all. The day one of them
		// grows an OnETB or a Triggered ability, the caveat saying
		// its entry trigger "never happens" needs re-reading.
		//
		// When surveil ships as a primitive, replace this probe with
		// one that looks for the primitive and promote it to Exact.
		Name:    "surveil",
		Phrases: []string{"surveil"},
		Implements: func(s effects.Spec) bool {
			return s.OnETB != nil || len(s.Triggered) > 0
		},
		Evidence:   "the spec declares an entry hook (OnETB or Triggered), which it did not when the caveat was written",
		Confidence: Heuristic,
		Adopt:      "implement the surveil on entry once the keyword exists, then drop the caveat",
	},
}

// Mechanics returns the curated table. Exported so a card author can
// read the coverage without reading the tests.
func Mechanics() []Mechanic { return mechanics }

// matcher caches the compiled word-boundary patterns for one
// mechanic's phrases.
var matchers = func() map[string][]*regexp.Regexp {
	out := map[string][]*regexp.Regexp{}
	for _, m := range mechanics {
		for _, p := range m.Phrases {
			out[m.Name] = append(out[m.Name], regexp.MustCompile(`(?i)\b`+regexp.QuoteMeta(p)+`\b`))
		}
	}
	return out
}()

// names reports whether a caveat names this mechanic.
func (m Mechanic) names(caveat string) bool {
	for _, re := range matchers[m.Name] {
		if re.MatchString(caveat) {
			return true
		}
	}
	return false
}

// Finding is one (card, mechanic, caveat) triple the guards object
// to.
type Finding struct {
	Card       string
	OracleID   string
	Mechanic   Mechanic
	Caveat     string
	ExampleFor string // for adoptable gaps: a card that already has it
}

// Contradictions is the hard guard: caveats that name a mechanic the
// SAME CARD demonstrably implements. The caveat is not merely stale
// in its reasoning — it is false about the card it is printed on, and
// the catalog page is showing that falsehood to a player.
//
// There is no allowlist. A finding here is always a bug in the card
// file (or, for a Heuristic mechanic, in the probe) and the fix is to
// delete the caveat and declare CompletenessFull, which Register will
// then hold you to.
func Contradictions() []Finding {
	var out []Finding
	for _, s := range effects.All() {
		for _, cv := range s.Caveats {
			for _, m := range mechanics {
				if m.names(cv) && m.Implements(s) {
					out = append(out, Finding{Card: s.Name, OracleID: s.OracleID, Mechanic: m, Caveat: cv})
				}
			}
		}
	}
	sortFindings(out)
	return out
}

// EngineImplements reports whether ANY registered card demonstrably
// implements the mechanic — which is this package's definition of
// "the engine can do this now". It is a registry question, not a
// source-code question, and that is the right one: a mechanic nothing
// uses is a mechanic no card file can copy.
//
// Only Exact mechanics are eligible. Answering "the engine has
// surveil" off a heuristic proxy would be a guess dressed as a fact,
// and the adoptable-gap report exists to be believed.
func EngineImplements(m Mechanic) (string, bool) {
	if m.Confidence != Exact {
		return "", false
	}
	best := ""
	for _, s := range effects.All() {
		if m.Implements(s) && (best == "" || s.Name < best) {
			best = s.Name
		}
	}
	return best, best != ""
}

// AdoptableGaps is the softer guard: caveats that name a mechanic the
// ENGINE has, on a card that has not taken it up. The caveat is
// honest about the card — the clause really does not happen — but its
// reason has expired: the machinery exists, some other card is
// already using it, and this one is a small edit away.
//
// This is the fierce_guardianship shape. Its caveat said the free
// cast "is not offered", which was true of the card and false about
// the engine from the moment #428 landed the conditional
// AlternativeCost.
//
// Unlike Contradictions this list is PINNED rather than required to
// be empty, because adopting a mechanic is card work with its own
// tests and cannot be demanded of the PR that happens to land the
// mechanic. The test's job is to make the set visible and to fail on
// any change to it in either direction.
func AdoptableGaps() []Finding {
	var out []Finding
	for _, s := range effects.All() {
		for _, cv := range s.Caveats {
			for _, m := range mechanics {
				if !m.names(cv) || m.Implements(s) {
					continue
				}
				if example, ok := EngineImplements(m); ok {
					out = append(out, Finding{
						Card: s.Name, OracleID: s.OracleID, Mechanic: m,
						Caveat: cv, ExampleFor: example,
					})
				}
			}
		}
	}
	sortFindings(out)
	return out
}

// Key is the stable identity of a finding, used as the pin in the
// test's table.
func (f Finding) Key() string { return f.Card + " / " + f.Mechanic.Name }

func sortFindings(f []Finding) {
	sort.Slice(f, func(i, j int) bool { return f[i].Key() < f[j].Key() })
}

// Describe renders a finding as the block of a failure message that
// says what was seen. Shared by both guards so the two failures read
// the same way.
func (f Finding) Describe() string {
	var b strings.Builder
	fmt.Fprintf(&b, "  card:     %s (%s)\n", f.Card, f.OracleID)
	fmt.Fprintf(&b, "  caveat:   %q\n", f.Caveat)
	fmt.Fprintf(&b, "  mechanic: %s (probe: %s, %s)\n", f.Mechanic.Name, f.Mechanic.Evidence, f.Mechanic.Confidence)
	return b.String()
}
