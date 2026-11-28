package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b839_helpers.go — the mana-type replacements of Ice Age's land-type
// enchantments (S82 OLD batch 839, #5636): Infernal Darkness, Ritual of Subdual
// and Naked Singularity.

// OldB839LandsProduce is "If a land is tapped for mana, it produces {X} instead
// of any other type": Contamination's CR 614 replacement on the mana production
// event (CR 106.12b) for every land tapped for mana, anyone's. It changes the
// type of each mana and not the amount; `colour` is "B" for Infernal Darkness
// and "C" for Ritual of Subdual's colorless mana.
func OldB839LandsProduce(label, colour string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:   []game.EventKind{game.EventManaAdded},
		AppliesTo: OldB839LandTappedForMana,
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			out := make([]string, len(ev.ManaColors))
			for i := range out {
				out[i] = colour
			}
			ev.ManaColors = out
			return nil
		},
		Controller: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) uuid.UUID { return ev.ManaPlayer },
		Label:      label,
	}
}

// OldB839LandTappedForMana is the shared condition of the land-tapped-for-mana replacements: the
// event is mana a land makes as it is tapped for mana (CR 106.12b).
func OldB839LandTappedForMana(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
	if ev.Kind != game.RepEventProduceMana || !ev.ManaFromTap || len(ev.ManaColors) == 0 {
		return false
	}
	land, ok := g.LookupCardForEffect(ev.ManaSource)
	return ok && land.IsLand()
}

// OldB839LandTypeSwap is Naked Singularity's "If tapped for mana, Plains
// produce {R}, Islands produce {G}, Swamps produce {W}, Mountains produce {U},
// and Forests produce {B} instead of any other type": the same replacement with
// an answer per basic land type, read through the layered view. Each mana is
// answered by the type its colour belongs to when the land has that type (a
// Taiga's {R} is a Mountain's, so it becomes {U}); a mana of a colour none of
// the land's types gives is answered by the land's first type in the order
// Plains, Island, Swamp, Mountain, Forest. The amount is unchanged; a land with
// none of the five types and every other land are untouched.
func OldB839LandTypeSwap(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventManaAdded},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
			if ev.Kind != game.RepEventProduceMana || !ev.ManaFromTap || len(ev.ManaColors) == 0 {
				return false
			}
			land, ok := g.LookupCardForEffect(ev.ManaSource)
			return ok && land.IsLand() && oldB839SwapIndex(land, "") >= 0
		},
		Replace: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) error {
			land, ok := g.LookupCardForEffect(ev.ManaSource)
			if !ok {
				return nil
			}
			out := make([]string, len(ev.ManaColors))
			for i, c := range ev.ManaColors {
				out[i] = c
				if k := oldB839SwapIndex(land, c); k >= 0 {
					out[i] = oldB839Swap[k].to
				}
			}
			ev.ManaColors = out
			return nil
		},
		Controller: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) uuid.UUID { return ev.ManaPlayer },
		Label:      label,
	}
}

// oldB839Swap is Naked Singularity's table: each basic land type, the colour its
// own mana has and the colour it produces under the card.
var oldB839Swap = []struct{ landType, natural, to string }{
	{"Plains", "W", "R"}, {"Island", "U", "G"}, {"Swamp", "B", "W"}, {"Mountain", "R", "U"}, {"Forest", "G", "B"},
}

// oldB839SwapIndex is the table row that answers a mana of `colour` from `land`:
// the row for the type that makes that colour when the land has it, else its
// first type's row. -1 when the land has none of the five types.
func oldB839SwapIndex(land game.Card, colour string) int {
	first := -1
	for i, row := range oldB839Swap {
		if !land.HasSubtype(row.landType) {
			continue
		}
		if first < 0 {
			first = i
		}
		if row.natural == colour {
			return i
		}
	}
	return first
}
