package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mana_derivation.go — the constructors for `ManaAbility.ProducedFunc`,
// the slot that computes a mana ability's output at activation time
// instead of declaring it (#352 sub-gaps 3 and 4).
//
// Two card families, one mechanism:
//
//   - A colour or set read directly off the board: Mox Amber ("one
//     mana of any color among legendary creatures and planeswalkers
//     you control").
//   - SCALED amounts. Cabal Coffers ("{B} for each Swamp you
//     control"), Gaea's Cradle.
//
// Everything here is READ-ONLY and runs under g.mu — held for write by
// ActivateManaAbility, for read by the auto-tapper. Use
// BattlefieldCardsForEffect and plain field reads; a public locking
// mutator deadlocks.
//
// ## Where "could produce ANOTHER permanent's mana" lives
//
// Exotic Orchard ("one mana of any color that a land an opponent
// controls could produce" — the highest-ranked card the catalog did
// not have, rank 9 of the whole format), Reflecting Pool and Fellwar
// Stone are NOT built from a ProducedFunc closure here. CR 106.7 is
// game.ProducibleManaLocked, in game/producible_mana.go, and asking it
// about ANOTHER permanent is itself recursive whenever that permanent
// is ALSO one of these three — answering it correctly needs a visited
// set threaded the whole way down (#1323), which a
// `func(*Game, controller, source uuid.UUID) string` closure has no
// fourth argument to carry across the package boundary. So these three
// declare DerivedMatch instead — DerivedFromOpponentLands() /
// DerivedFromOwnLands() below, a plain predicate — and
// game.ProducibleManaLocked walks the board and recurses itself, where
// the visited set is an ordinary parameter. See that file's "recursion
// guard" section for the full account.

// Spend-restriction tags, re-exported from the game package so a card
// file reads `ManaRestrictSupertype("Legendary")` rather than
// `game.ManaRestrictSupertype(...)` — the same courtesy ZeroUUID does
// for uuid.Nil. The vocabulary, the AND semantics and the
// unknown-tag-denies rule all live in game/mana_restriction.go, which
// is also where they are enforced at spend time.
const (
	// ManaRestrictCast — spend only on a spell's cost.
	ManaRestrictCast = game.ManaRestrictCast
	// ManaRestrictActivate — spend only on an activated ability's
	// cost.
	ManaRestrictActivate = game.ManaRestrictActivate
	// ManaRestrictColorless — spend only on a colorless object.
	ManaRestrictColorless = game.ManaRestrictColorless
)

// ManaRestrictType restricts a token to objects with the named card
// type — ManaRestrictType("Creature") for Ancient Ziggurat.
func ManaRestrictType(t string) string { return game.ManaRestrictType(t) }

// ManaRestrictSubtype restricts a token to objects with the named
// subtype — ManaRestrictSubtype("Eldrazi") for Eldrazi Temple.
func ManaRestrictSubtype(t string) string { return game.ManaRestrictSubtype(t) }

// ManaRestrictSupertype restricts a token to objects with the named
// supertype — ManaRestrictSupertype("Legendary") for Delighted
// Halfling.
func ManaRestrictSupertype(t string) string { return game.ManaRestrictSupertype(t) }

// orderedManaSymbols flattens a symbol set into canonical WUBRGC
// order, so a derived pipe string is stable across activations and a
// test can assert on it.
func orderedManaSymbols(seen map[string]bool) []string {
	var out []string
	for _, m := range []string{"W", "U", "B", "R", "G", "C"} {
		if seen[m] {
			out = append(out, m)
		}
	}
	return out
}

// colorsOnly drops "C". "One mana of any COLOR" (Exotic Orchard,
// Fellwar Stone, Mox Amber) cannot make colorless — CR 105.1, {C} is
// not a colour. "One mana of any TYPE" (Reflecting Pool) can, which
// is the whole difference between those two cards' wordings and the
// reason this is a separate step rather than baked into
// producibleFrom.
func colorsOnly(symbols []string) []string {
	out := make([]string, 0, len(symbols))
	for _, s := range symbols {
		if s == "C" {
			continue
		}
		out = append(out, s)
	}
	return out
}

// pipeString renders a symbol set as a single produced-mana slot —
// ["W","U"] becomes "{W|U}". Empty input returns "", which
// ActivateManaAbility treats as "this ability produced no mana": the
// printed outcome for Exotic Orchard facing no opposing lands, and
// the reason the return type is a string rather than a slice.
func pipeString(symbols []string) string {
	if len(symbols) == 0 {
		return ""
	}
	return "{" + strings.Join(symbols, "|") + "}"
}

// DerivedFromOpponentLands is Exotic Orchard's and Fellwar Stone's
// DerivedMatch: "a land an opponent controls" — pair it with
// DerivedColorsOnly: true, since "any COLOR" cannot make colorless
// (CR 105.1).
//
// CR 106.7 / the Exotic Orchard rulings: "could produce" asks what
// the land's abilities would add if activated, NOT whether it could
// legally be activated right now. A tapped opposing Island still
// offers {U}; a Temple of the False God its controller cannot
// activate still offers {C}, which then drops out via
// DerivedColorsOnly. What it does respect is the land's CURRENT
// choice: an opposing Thriving Isle that chose red offers {U} and
// {R}, and one with no colour chosen yet offers {U} alone (#782). The
// recursion itself, and the visited set that keeps it from running
// forever, live in game.ProducibleManaLocked (#1323) — this is only
// the "which candidates" half.
//
// Declare it with DerivesFromOtherSources: true, so CR 106.7's reader
// knows to route through the visited set rather than a plain
// ProducedFunc call.
func DerivedFromOpponentLands() func(game.Card, uuid.UUID) bool {
	return func(c game.Card, controller uuid.UUID) bool {
		return c.Controller != controller && c.IsLand()
	}
}

// DerivedFromOwnLands is Reflecting Pool's DerivedMatch: "a land you
// control" — DerivedColorsOnly stays false, since "any TYPE" includes
// colorless.
//
// A lone Reflecting Pool sees only itself among "lands you control",
// and game.derivedManaLocked excludes the source from its own match —
// the printed ruling is that a lone Reflecting Pool produces nothing,
// and that holds independently of the visited-set cycle guard.
func DerivedFromOwnLands() func(game.Card, uuid.UUID) bool {
	return func(c game.Card, controller uuid.UUID) bool {
		return c.Controller == controller && c.IsLand()
	}
}

// ProducedFromLegendaryPermanents is Mox Amber's "one mana of any
// color among legendary creatures and planeswalkers you control".
//
// This one reads the permanents' COLOURS, not what they could
// produce — a legendary green creature offers {G} whether or not it
// taps for anything.
func ProducedFromLegendaryPermanents() func(*game.Game, uuid.UUID, uuid.UUID) string {
	return func(g *game.Game, controller, _ uuid.UUID) string {
		seen := map[string]bool{}
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller != controller {
				continue
			}
			ch := c.Effective()
			if !hasFold(ch.Supertypes, "Legendary") {
				continue
			}
			if !hasFold(ch.Types, "Creature") && !hasFold(ch.Types, "Planeswalker") {
				continue
			}
			for _, col := range c.EffectiveColors() {
				seen[strings.ToUpper(col)] = true
			}
		}
		return pipeString(colorsOnly(orderedManaSymbols(seen)))
	}
}

// ProducedPerPermanent is the scaled shape: "Add {SYMBOL} for each
// permanent you control matching `match`". Cabal Coffers is
// ProducedPerPermanent("B", ControlsLandSubtype("Swamp")); Gaea's
// Cradle is ProducedPerPermanent("G", IsCreaturePermanent).
//
// A count of zero returns "", which adds no mana and still taps the
// land — the printed outcome for Gaea's Cradle with an empty board.
func ProducedPerPermanent(symbol string, match func(game.Card) bool) func(*game.Game, uuid.UUID, uuid.UUID) string {
	slot := "{" + strings.ToUpper(symbol) + "}"
	return func(g *game.Game, controller, _ uuid.UUID) string {
		n := 0
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller != controller {
				continue
			}
			if match(c) {
				n++
			}
		}
		return strings.Repeat(slot, n)
	}
}

// MatchLandSubtype matches a land with the given subtype, read off
// the EFFECTIVE type line — so Urborg, Tomb of Yawgmoth turning every
// land into a Swamp really does grow Cabal Coffers, which is half of
// why anyone plays either card.
func MatchLandSubtype(subtype string) func(game.Card) bool {
	return func(c game.Card) bool {
		if !c.IsLand() {
			return false
		}
		return hasFold(c.Effective().Subtypes, subtype)
	}
}

// MatchCreature matches any creature permanent — Gaea's Cradle.
func MatchCreature(c game.Card) bool { return c.IsCreature() }

// MatchEnchantment matches any enchantment permanent — Serra's
// Sanctum.
func MatchEnchantment(c game.Card) bool { return c.IsEnchantment() }

// ControlsAtLeast builds a `ManaAbility.Condition`: "Activate only if
// you control N or more permanents matching `match`". Temple of the
// False God is ControlsAtLeast(5, matchAnyLand); Mox Opal is
// ControlsAtLeast(3, matchArtifact).
func ControlsAtLeast(n int, match func(game.Card) bool) func(*game.Game, uuid.UUID, uuid.UUID) bool {
	return func(g *game.Game, controller, _ uuid.UUID) bool {
		count := 0
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller != controller {
				continue
			}
			if !match(c) {
				continue
			}
			count++
			if count >= n {
				return true
			}
		}
		return false
	}
}

// MatchLand matches any land permanent — Temple of the False God,
// Shrine of the Forsaken Gods.
func MatchLand(c game.Card) bool { return c.IsLand() }

// MatchArtifact matches any artifact permanent — Mox Opal's
// metalcraft.
func MatchArtifact(c game.Card) bool { return c.IsArtifact() }

// hasFold is a case-insensitive membership test over a type-line
// slice.
func hasFold(xs []string, want string) bool {
	for _, x := range xs {
		if strings.EqualFold(x, want) {
			return true
		}
	}
	return false
}
