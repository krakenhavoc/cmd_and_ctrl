package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mikokoro, Center of the Sea — Legendary Land (EDHREC rank 1172):
//
//	"{T}: Add {C}.
//	 {2}, {T}: Each player draws a card."
//
// The group-hug land. A two-component activated ability with no
// target; every seated player draws, the controller first and then
// the table in seat order, so a draw payoff on either side of the
// table sees its own draw. Any player may activate it on any turn,
// which is what makes it a group-hug card rather than a card-draw
// engine.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a4580a1d-141e-449b-9018-e0258130634b",
		Name:         "Mikokoro, Center of the Sea",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}: Each player draws a card.",
			Cost:  Plus(ManaCost("{2}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, id := range tablePlayers(ctx) {
					if err := (DrawCards{Player: id, N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
