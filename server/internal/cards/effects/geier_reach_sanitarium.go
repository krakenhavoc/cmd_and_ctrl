package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Geier Reach Sanitarium — Legendary Land (EDHREC rank 441):
//
//	"{T}: Add {C}.
//	 {2}, {T}: Each player draws a card, then discards a card."
//
// A symmetrical loot on a land: the discard deck's engine, and the
// reanimator deck's way to bin a fatty at instant speed. "Each
// player" is the table, controller first: every player draws and
// then chooses their own discard through the same discard prompt
// Faithless Looting uses, so nobody picks for anyone else.
//
// Sandbox simplification, cosmetic: in paper every player draws
// simultaneously and then every player discards; here each player
// draws-then-discards in seat order. No card in the catalog can
// observe the difference (a draw payoff sees the same draws either
// way), so it is recorded rather than modelled.
func init() {
	Register(Spec{
		OracleID:     "7b9fafe7-d26a-4ed5-b4c4-ce13763770b5",
		Name:         "Geier Reach Sanitarium",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}: Each player draws a card, then discards a card.",
			Cost:  Plus(ManaCost("{2}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, p := range tablePlayers(ctx) {
					if err := (DrawCards{Player: p, N: 1}).Apply(ctx); err != nil {
						return err
					}
					g.DiscardChoiceForEffect(p, 1)
				}
				return nil
			},
		}},
	})
}
