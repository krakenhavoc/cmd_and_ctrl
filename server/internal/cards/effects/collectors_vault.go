package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Collector's Vault — Artifact {2} (EDHREC rank 1796):
//
//	"{2}, {T}: Draw a card, then discard a card. Create a Treasure
//	 token."
//
// A loot that pays a Treasure back. One activated ability: the loot
// is lootOne (draw first, then the controller's discard choice) and
// the Treasure is minted in the same resolution, so the token is on
// the battlefield while the discard prompt is open — nothing the
// prompt offers is affected, as the discard is from hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f460ff80-8175-43f7-9810-540cf3817085",
		Name:         "Collector's Vault",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}: Draw a card, then discard a card. Create a Treasure token.",
			Cost:  Plus(ManaCost("{2}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := lootOne(g, item, 1); err != nil {
					return err
				}
				return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
