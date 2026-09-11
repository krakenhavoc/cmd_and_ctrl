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
// ## Why "could produce" is computed from two sources
//
// A land's producible mana is read from its mana abilities
// (ManaAbilitiesForCard — catalog entry, intrinsic token ability, or
// the synthetic basic-land shape) UNIONED with Card.ProducedMana, the
// `produced_mana` array Scryfall stamps at deck import. Neither alone
// is enough: the catalog has entries for 300-odd cards and a
// decklist's other 60 lands only have the Scryfall field, while a
// token or a demo-seed card has the ability and no Scryfall data.
//
// ## The recursion guard
//
// An ability that itself has a ProducedFunc is SKIPPED when deriving.
// Two Exotic Orchards facing each other, or an Exotic Orchard and a
// Reflecting Pool, would otherwise recurse until the stack ran out.
// CR 106.6b answers the circular case with "no mana", and so does
// this: the other Orchard contributes nothing. It also means a
// Reflecting Pool does not see a Cabal Coffers' black — a
// simplification in the weaker-than-printed direction, declared here
// rather than on four card files.

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

// producibleFrom returns the distinct mana symbols `c` could produce,
// in WUBRGC order. Includes "C" — callers that mean *colours* filter
// it out (see colorsOnly).
func producibleFrom(c game.Card) []string {
	seen := map[string]bool{}
	for _, ab := range game.ManaAbilitiesForCard(c) {
		// The recursion guard — see the file comment.
		if ab.ProducedFunc != nil {
			continue
		}
		slots, err := game.ParseProducedMana(ab.Produced)
		if err != nil {
			continue
		}
		for _, slot := range slots {
			for _, opt := range slot.Options {
				seen[opt] = true
			}
		}
	}
	for _, m := range c.ProducedMana {
		u := strings.ToUpper(m)
		switch u {
		case "W", "U", "B", "R", "G", "C":
			seen[u] = true
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
// a colour.
func ProducedFromOpponentLands() func(*game.Game, uuid.UUID, uuid.UUID) string {
	return func(g *game.Game, controller, _ uuid.UUID) string {
		seen := map[string]bool{}
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == controller || !c.IsLand() {
				continue
			}
			for _, m := range producibleFrom(c) {
				seen[m] = true
			}
		}
		return pipeString(colorsOnly(orderedManaSymbols(seen)))
	}
}

// ProducedFromOwnLands is Reflecting Pool's "one mana of any TYPE
// that a land you control could produce" — type, so {C} counts.
//
// The Pool sees itself only through the recursion guard, which is to
// say not at all: its own ability has a ProducedFunc and is skipped.
// That matches the printed card, whose ruling is that a lone
// Reflecting Pool produces nothing.
func ProducedFromOwnLands() func(*game.Game, uuid.UUID, uuid.UUID) string {
	return func(g *game.Game, controller, _ uuid.UUID) string {
		seen := map[string]bool{}
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller != controller || !c.IsLand() {
				continue
			}
			for _, m := range producibleFrom(c) {
				seen[m] = true
			}
		}
		return pipeString(orderedManaSymbols(seen))
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
