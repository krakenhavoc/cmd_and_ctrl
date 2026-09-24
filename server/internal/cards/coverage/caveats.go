package coverage

import (
	"fmt"
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
		Name:    "trigger doubling",
		Phrases: []string{"trigger doubling", "trigger-doubling", "trigger an additional time", "triggers an additional time"},
		Implements: func(s effects.Spec) bool {
			return game.CatalogTriggerDoublers != nil && len(game.CatalogTriggerDoublers(s.OracleID)) > 0
		},
		Evidence:   "game.CatalogTriggerDoublers(oracleID) declares a doubler",
		Confidence: Exact,
		Adopt:      "Declare Spec.TriggerDoublers using the cause helpers in trigger_doubling.go",
	},
	{
		Name:       "flashback",
		Phrases:    []string{"flashback"},
		Implements: altCost("flashback"),
		Evidence:   `game.AlternativeCostByKey(oracleID, "flashback") resolves`,
		Confidence: Exact,
		Adopt:      `CastableZones: []game.ZoneKind{game.ZoneGraveyard} plus Flashback("{cost}") — see alternative_cost.go`,
	},
	{
		Name:       "escape",
		Phrases:    []string{"escape"},
		Implements: altCost("escape"),
		Evidence:   `game.AlternativeCostByKey(oracleID, "escape") resolves`,
		Confidence: Exact,
		Adopt:      `CastableZones: []game.ZoneKind{game.ZoneGraveyard} plus Escape("{cost}", n) — EscapeWithCounters when the creature escapes with counters`,
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
		// S24: "loses all abilities" was a declared machinery gap on
		// #76 until the layer-6 removal reached the Catalog* hooks.
		// The probe is exact because the removal is a DECLARATION on
		// the static (game.StaticAbility.RemovesAbilities) rather
		// than something an opaque Apply closure does — which is
		// itself half of why the field exists.
		Name: "loses all abilities",
		Phrases: []string{
			"loses all abilities", "lose all abilities",
			"loses its abilities", "loses all its abilities",
		},
		Implements: func(s effects.Spec) bool {
			for _, ab := range game.CatalogStaticAbilities(s.OracleID) {
				if ab.RemovesAbilities {
					return true
				}
			}
			return false
		},
		Evidence:   "the spec declares a layer-6 static with RemovesAbilities",
		Confidence: Exact,
		Adopt:      `Static: []game.StaticAbility{LoseAllAbilities()} — see effects/attachments.go`,
	},
	{
		// #746: "this spell costs {1} less to cast", affinity, strive.
		// Blasphemous Act shipped for months with "The cost reduction
		// is missing"; the day a card's own cost modifier is declared,
		// that sentence is false. Phrases stay about the SPELL'S OWN
		// cost — "discount" is left out on purpose, because Blossoming
		// Tortoise's caveat is about ability costs, which this slot
		// does not touch.
		Name:    "a spell's own cost modifier",
		Phrases: []string{"cost reduction", "affinity", "strive", "undaunted"},
		Implements: func(s effects.Spec) bool {
			// Through the engine's own reader, so a broken
			// CardDef wiring goes quiet here too (see altCost).
			return len(game.SelfCostModifiersFor(game.Card{OracleID: s.OracleID})) > 0
		},
		Evidence:   "game.SelfCostModifiersFor returns the spec's self cost modifiers",
		Confidence: Exact,
		Adopt:      `SelfCostModifiers: []game.CostModifier{CostsLessEach(…) / AffinityFor(…) / CostsMorePerTargetBeyondFirst(…)} — see effects/self_cost_modifier.go`,
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
		// grows an AsEnters hook or a Triggered ability, the caveat saying
		// its entry trigger "never happens" needs re-reading.
		//
		// When surveil ships as a primitive, replace this probe with
		// one that looks for the primitive and promote it to Exact.
		Name:    "surveil",
		Phrases: []string{"surveil"},
		Implements: func(s effects.Spec) bool {
			return s.AsEnters != nil || len(s.Triggered) > 0
		},
		Evidence:   "the spec declares an entry hook (AsEnters or Triggered), which it did not when the caveat was written",
		Confidence: Heuristic,
		Adopt:      "implement the surveil on entry once the keyword exists, then drop the caveat",
	},
	{
		// #761: a spell that reads the COLOURS of the mana that paid
		// for it. Converge and sunburst are the two, and both
		// declare Spec.WantsDistinctColors — not because the words
		// mean the same thing (converge counts at resolution,
		// sunburst at entry) but because the declaration is the
		// engine-visible fact both depend on: without it the cast
		// gate pays colourless-first and the card converges for less
		// than the board allowed.
		//
		// "Adamant" is deliberately NOT in this row's phrases and has
		// its own below: an adamant card does not want the colours
		// spread, so it never sets the flag, and a shared row would
		// report every adamant card as unimplemented forever.
		Name:    "converge and sunburst",
		Phrases: []string{"converge", "sunburst"},
		Implements: func(s effects.Spec) bool {
			return s.WantsDistinctColors
		},
		Evidence:   "the spec declares WantsDistinctColors, the colour-spreading payment #761 added",
		Confidence: Exact,
		Adopt:      `WantsDistinctColors: true plus ctx.ColorsSpentCount() (converge) or SunburstCounters(kind) in OnResolve — see effects/mana_spent.go`,
	},
	{
		// #761's other reader, and a Heuristic one because there is
		// nothing on the Spec to point at: adamant is an ability word
		// (CR 207.2c), so the whole mechanic lives inside OnResolve,
		// and "does this closure call ManaSpentOfColor?" is not a
		// question a probe can ask. What IS observable is that a card
		// whose caveat says its adamant does nothing has grown a
		// resolution body at all.
		//
		// The false positive is a card that gains an unrelated
		// OnResolve; the fix then is to tighten this probe, not to
		// delete the row. When adamant becomes a declared slot rather
		// than a closure, probe for the slot and promote to Exact.
		Name:    "adamant",
		Phrases: []string{"adamant"},
		Implements: func(s effects.Spec) bool {
			return s.OnResolve != nil
		},
		Evidence:   "the spec declares an OnResolve body, which it did not when the caveat was written",
		Confidence: Heuristic,
		Adopt:      `AdamantSpent(ctx, "R", 3) inside OnResolve — see effects/mana_spent.go`,
	},
	{
		// #625, then #789: removing counters as part of paying an
		// ability's cost, on either ability kind. The probe reads the
		// DECLARATION off the cost — one CounterRemovalCost with two
		// owners — so it is exact and covers a card that puts the
		// component on a mana ability (Vivid Creek, Ramos) as well as
		// on a CR 602 one (Heart of Kiran).
		//
		// Phrases stay about a counter being REMOVED TO PAY. Bare
		// "counter" is far too common in caveat prose (countering
		// spells, +1/+1 counters arriving), and a probe that fired on
		// those would name the wrong mechanic with total confidence.
		Name: "a counter-removal cost",
		Phrases: []string{
			"counter-removal cost", "remove counters", "removing counters",
			"remove a counter", "removing a counter",
		},
		Implements: func(s effects.Spec) bool {
			for _, ab := range s.Activated {
				if ab.Cost.RemoveCounters != nil {
					return true
				}
			}
			for _, ma := range s.ManaAbilities {
				if ma.Cost.RemoveCounters != nil {
					return true
				}
			}
			return false
		},
		Evidence:   "the spec declares RemoveCounters on an activated or mana ability's cost",
		Confidence: Exact,
		Adopt:      `RemoveCountersFromThis / RemoveCountersFrom / RemoveCountersXFromThis / RemoveCountersAmong — see effects/activated.go`,
	},
	{
		// #1273: exactly the #412 failure mode. Force of Negation's
		// caveat said the countered spell "goes to its owner's
		// graveyard instead of being exiled" for a sprint after #1230
		// closed the gap — `Game.CounterTargetToZoneForEffect` (rode
		// onto the card catalog as `CounterTarget.Dest`) has shipped
		// on Devious Cover-Up, Remand, Memory Lapse and Dissipate
		// since. This row is the durable half: the caveat text was
		// fixed by hand, but the curated table had no entry for the
		// mechanic it names, so nothing would have caught the NEXT
		// card that ships the same stale sentence.
		//
		// Heuristic, and narrower than "adamant"'s bare OnResolve
		// check: CounterTarget.Dest is set inside an opaque OnResolve
		// closure, so there is no declared Spec field an Exact probe
		// could point at the way AlternativeCosts or a Static ability
		// gives one. What IS observable is that the card targets a
		// spell on the stack at all (TargetSpell's "stack_spell"
		// mode) and has grown a resolution body — which at least
		// excludes every card whose caveat happens to share the
		// phrase for an unrelated reason (Whip of Erebos' "graveyard
		// instead of being exiled" is about a returned creature, and
		// declares a graveyard target, not a stack one).
		//
		// The false positive is an ordinary "Counter target spell."
		// card whose caveat is about something else and happens to
		// contain one of these phrases; the fix then is to tighten
		// the phrase list, not delete the row. When counter-to-zone
		// becomes a declared Spec field, probe for the field and
		// promote to Exact.
		Name: "counter to a zone",
		Phrases: []string{
			"graveyard instead of being exiled", "graveyard instead of exile",
			"countered this way",
		},
		Implements: func(s effects.Spec) bool {
			return s.Targets != nil && s.Targets.Mode == "stack_spell" && s.OnResolve != nil
		},
		Evidence:   `the spec targets a spell on the stack (Targets.Mode == "stack_spell") and declares an OnResolve body`,
		Confidence: Heuristic,
		Adopt:      `CounterTarget{StackID: …, Dest: game.ZoneRef{Kind: game.ZoneExile}} (or ZoneHand / ZoneLibrary) inside OnResolve — see effects/force_of_negation.go, effects/devious_cover_up.go`,
	},
	{
		// ADR 0089 (#1267): gift is one declaration, Spec.Gift, and
		// the keyword's cost and gift both grow from it — so the probe
		// is exact. Six catalog cards carried "the gift can't be
		// promised" when it landed; the ones it did not adopt are
		// pinned in caveats_test.go.
		Name:    "gift",
		Phrases: []string{"gift"},
		Implements: func(s effects.Spec) bool {
			return s.Gift != nil
		},
		Evidence:   "the spec declares Spec.Gift",
		Confidence: Exact,
		Adopt:      `Gift: GiftACard() / GiftAFood() / GiftATappedFish() / GiftATreasure(), plus .Instead(clause) when the promise swaps the target — see effects/gift.go`,
	},
	{
		// #1258: a keyword TRIGGER had no machine-readable name, so
		// nothing could ask whether a card has cascade. The
		// constructors now stamp game.TriggeredAbility.Keyword, and
		// this reads it back through the same catalog hook the
		// harvester uses.
		Name:       "cascade",
		Phrases:    []string{"cascade"},
		Implements: keywordTrigger(effects.KeywordCascade),
		Evidence:   `a trigger in game.CatalogTriggers(oracleID) is named "cascade"`,
		Confidence: Exact,
		Adopt:      `Triggered: []game.TriggeredAbility{Cascade()} — or GrantsCascade(label, when) for a permanent that gives it`,
	},
	{
		Name:       "storm",
		Phrases:    []string{"storm"},
		Implements: keywordTrigger(effects.KeywordStorm),
		Evidence:   `a trigger in game.CatalogTriggers(oracleID) is named "storm"`,
		Confidence: Exact,
		Adopt:      `Triggered: []game.TriggeredAbility{Storm()}`,
	},
	{
		// #706: prowess is a canonicalKeywords token, not a
		// constructor — the engine derives the trigger from the
		// ability list (game/prowess.go). A card that DECLARES it in
		// PrintedKeywords has it; the probe reads that declaration
		// through the hook printedCharacteristic reads. A card that
		// relies on the deck importer alone is invisible here, which
		// is the direction a curated table is allowed to miss in.
		Name:    "prowess",
		Phrases: []string{"prowess"},
		Implements: func(s effects.Spec) bool {
			if game.CatalogPrintedKeywords == nil {
				return false
			}
			for _, kw := range game.CatalogPrintedKeywords(s.OracleID) {
				if kw == game.KeywordProwess {
					return true
				}
			}
			return false
		},
		Evidence:   `game.CatalogPrintedKeywords(oracleID) contains "prowess"`,
		Confidence: Exact,
		Adopt:      `PrintedKeywords: []string{"prowess"} — the engine does the rest`,
	},
}

// keywordTrigger builds the exact probe for a keyword that ships as a
// triggered-ability constructor (#1258): does any of this card's
// catalog triggers carry that keyword's name? Routed through the
// game.CatalogTriggers hook, not spec.Triggered, for the reason altCost
// gives — if the wiring breaks, the probe goes quiet and
// TestEveryExactMechanicHasAnImplementor notices.
func keywordTrigger(name string) func(effects.Spec) bool {
	return func(s effects.Spec) bool {
		if game.CatalogTriggers == nil {
			return false
		}
		for _, t := range game.CatalogTriggers(s.OracleID) {
			if t.Keyword == name {
				return true
			}
		}
		return false
	}
}

// Mechanics returns the curated table. Exported so a card author can
// read the coverage without reading the tests.
func Mechanics() []Mechanic { return mechanics }

// matchers caches the compiled word-boundary patterns for each
// mechanic's phrases (see PhraseMatcher).
var matchers = func() map[string]func(string) bool {
	out := map[string]func(string) bool{}
	for _, m := range mechanics {
		out[m.Name] = PhraseMatcher(m.Phrases)
	}
	return out
}()

// names reports whether a caveat names this mechanic.
func (m Mechanic) names(caveat string) bool {
	match, ok := matchers[m.Name]
	return ok && match(caveat)
}

// Names is names, exported: whether a caveat names this mechanic,
// under the same word-boundary rule the guards apply.
func (m Mechanic) Names(caveat string) bool { return m.names(caveat) }

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
