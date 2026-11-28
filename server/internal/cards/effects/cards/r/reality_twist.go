package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reality Twist — Enchantment {U}{U}{U} (Ice Age, S82 OLD batch 838, #5636):
//
//	"Cumulative upkeep {1}{U}{U}
//	 If tapped for mana, Plains produce {R}, Swamps produce {G}, Mountains
//	 produce {W}, and Forests produce {B} instead of any other type."
//
// Contamination's CR 614 replacement on the mana production event (CR 106.12b)
// with a per-land-type answer. It applies to a land tapped for mana, anyone's,
// that has one of the four basic land types, read through the layered view, and
// changes the TYPE of each mana and not the amount: a Plains' {W} is {R}, a
// Swamp's {B} is {G}, a Mountain's {R} is {W}, a Forest's {G} is {B}. A land with
// two of the types (a Taiga, Mountain Forest) is answered per mana, by the type
// its colour belongs to: its {R} becomes {W} and its {G} becomes {B}; a mana of
// a colour none of its types gives is answered by its first type in the order
// Plains, Swamp, Mountain, Forest. Islands and every other land are unchanged.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c6f9fdb-ee27-4cd8-9d4d-1ea8b2830f73",
		Name:         "Reality Twist",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.CumulativeUpkeep("Reality Twist — cumulative upkeep {1}{U}{U}", "{1}{U}{U}")},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventManaAdded},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
				if ev.Kind != game.RepEventProduceMana || !ev.ManaFromTap || len(ev.ManaColors) == 0 {
					return false
				}
				land, ok := g.LookupCardForEffect(ev.ManaSource)
				return ok && land.IsLand() && realityTwistType(land, "") != ""
			},
			Replace: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) error {
				land, ok := g.LookupCardForEffect(ev.ManaSource)
				if !ok {
					return nil
				}
				out := make([]string, len(ev.ManaColors))
				for i, c := range ev.ManaColors {
					out[i] = realityTwistColor(realityTwistType(land, c))
					if out[i] == "" {
						out[i] = c
					}
				}
				ev.ManaColors = out
				return nil
			},
			Controller: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) uuid.UUID { return ev.ManaPlayer },
			Label:      "Reality Twist: Plains produce {R}, Swamps {G}, Mountains {W} and Forests {B}",
		}},
	})
}

// realityTwistOrder is the four land types and the colour each one's own mana
// has, in the order a mana of no matching colour is answered.
var realityTwistOrder = []struct{ landType, colour string }{
	{"Plains", "W"}, {"Swamp", "B"}, {"Mountain", "R"}, {"Forest", "G"},
}

// realityTwistType is the basic land type of `land` that answers a mana of
// `colour`: the type that makes that colour when it has it, else its first of
// the four. "" when it has none (or, for an empty colour, none at all).
func realityTwistType(land game.Card, colour string) string {
	first := ""
	for _, t := range realityTwistOrder {
		if !land.HasSubtype(t.landType) {
			continue
		}
		if first == "" {
			first = t.landType
		}
		if t.colour == colour {
			return t.landType
		}
	}
	return first
}

// realityTwistColor is the colour a land type produces under Reality Twist.
func realityTwistColor(landType string) string {
	switch landType {
	case "Plains":
		return "R"
	case "Swamp":
		return "G"
	case "Mountain":
		return "W"
	case "Forest":
		return "B"
	}
	return ""
}
