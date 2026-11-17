package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Maestros Initiate — Creature — Human Citizen {2}{B}, 3/1 (Streets of New
// Capenna):
//
//	"{4}{U/R}, Exile this card from your graveyard: Draw two cards, then
//	 discard a card."
//
// Colossal Rattlewurm's graveyard ability: activated from the graveyard
// (Zones, CR 113.6), the exile of itself is part of the cost, and the loot is
// the ordinary draw-then-discard (the drawn cards are legal discards). The
// hybrid {U/R} is payable with either colour.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "85905607-dceb-4c37-81a3-ab441ce9490c",
		Name:         "Maestros Initiate",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{4}{U/R}, Exile this card from your graveyard: Draw two cards, then discard a card.",
			Cost:    Plus(ManaCost("{4}{U/R}"), ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Purpose: game.Purpose{Draws: 2, Discards: 1},
			Effect:  func(g *game.Game, item *game.StackItem) error { return b16DrawThenDiscard(g, item, 2) },
		}},
	})
}
