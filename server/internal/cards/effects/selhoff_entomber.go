package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Selhoff Entomber — Creature — Zombie {1}{U}, 1/3 (Innistrad: Crimson Vow):
//
//	"{T}, Discard a creature card: Draw a card."
//
// The discard is a cost (CR 602.2b) named at activation, as Fauna Shaman's
// is; the draw is the whole effect.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3b97794a-3aaa-45a6-9cad-f1b768615b6d",
		Name:         "Selhoff Entomber",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Discard a creature card: Draw a card.",
			Purpose: game.Purpose{Draws: 1, Discards: 1},
			Cost: Plus(TapCost(), DiscardCardsMatching(1, "a creature card", func(c game.Card) bool {
				return c.IsCreature()
			})),
			Effect: Do(DrawCards{N: 1}),
		}},
	})
}
