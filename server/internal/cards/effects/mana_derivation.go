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
//   - DERIVED colours. Exotic Orchard ("one mana of any color that a
//     land an opponent controls could produce") is the highest-ranked
//     card the catalog does not have — rank 9 of the whole format.
//     Reflecting Pool, Fellwar Stone and Mox Amber are the same
//     question asked of a different set of permanents.
//   - SCALED amounts. Cabal Coffers ("{B} for each Swamp you
//     control"), Gaea's Cradle.
//
// Everything here is READ-ONLY and runs under g.mu — held for write by
// ActivateManaAbility, for read by the auto-tapper. Use
// BattlefieldCardsForEffect and plain field reads; a public locking
// mutator deadlocks.
//
// ## Where "could produce" lives
//
// CR 106.7 is game.ProducibleManaLocked, in
// game/producible_mana.go — it asks each of the permanent's mana
// abilities what it would add NOW (a chosen colour, a board count, a
// commander-identity narrowing) through the same seam the activation
// reads, and falls back to Scryfall's `produced_mana` only for an
// imported card the catalog has never heard of. That file carries the
// rule, the fallback and the recursion guard; everything below is the
// card-side wrapper that turns its answer into a pipe string.
//
// Only the guard matters here: an ability marked
// DerivesFromOtherSources is skipped, which is exactly the three
// abilities in this file's own family (Exotic Orchard, Reflecting
// Pool, Fellwar Stone). Two of them facing each other would otherwise
// recurse until the stack ran out; CR 106.6b answers the circular
// case with "no mana" and so does the guard.

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

// producibleAcross unions CR 106.7's answer for every permanent on
// the battlefield that `match` accepts, in canonical WUBRGC order.
// The three derivations below differ only in their match and in
// whether they keep {C}.
func producibleAcross(g *game.Game, match func(game.Card) bool) []string {
	seen := map[string]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if !match(c) {
			continue
		}
		for _, m := range g.ProducibleManaLocked(c) {
			seen[m] = true
		}
	}
	return orderedManaSymbols(seen)
}

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

// ProducedFromOpponentLands is Exotic Orchard's and Fellwar Stone's
// "one mana of any color that a land an opponent controls could
// produce".
//
// CR 106.7 / the Exotic Orchard rulings: "could produce" asks what
// the land's abilities would add if activated, NOT whether it could
// legally be activated right now. A tapped opposing Island still
// offers {U}; a Temple of the False God its controller cannot
// activate still offers {C}, which then drops out because {C} is not
// a colour. What it does respect is the land's CURRENT choice: an
// opposing Thriving Isle that chose red offers {U} and {R}, and one
// with no colour chosen yet offers {U} alone (#782).
//
// Declare it with DerivesFromOtherSources: true — this is the
// recursion guard's whole membership list.
func ProducedFromOpponentLands() func(*game.Game, uuid.UUID, uuid.UUID) string {
	return func(g *game.Game, controller, _ uuid.UUID) string {
		return pipeString(colorsOnly(producibleAcross(g, func(c game.Card) bool {
			return c.Controller != controller && c.IsLand()
		})))
	}
}

// ProducedFromOwnLands is Reflecting Pool's "one mana of any TYPE
// that a land you control could produce" — type, so {C} counts.
//
// The Pool sees itself only through the recursion guard, which is to
// say not at all: its own ability is marked DerivesFromOtherSources
// and is skipped. That matches the printed card, whose ruling is that
// a lone Reflecting Pool produces nothing.
func ProducedFromOwnLands() func(*game.Game, uuid.UUID, uuid.UUID) string {
	return func(g *game.Game, controller, _ uuid.UUID) string {
		return pipeString(producibleAcross(g, func(c game.Card) bool {
			return c.Controller == controller && c.IsLand()
		}))
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
