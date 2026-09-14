package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arcanis the Omnipotent — Legendary Creature — Wizard {3}{U}{U}{U},
// 3/4 (EDHREC rank 2997):
//
//	"{T}: Draw three cards.
//	 {2}{U}{U}: Return Arcanis to its owner's hand."
//
// The Onslaught card-draw engine. Two CR 602 abilities: a tap for
// three cards (summoning sickness applies — the engine enforces it
// for a creature source) and a mana-only self-bounce that returns
// Arcanis to its owner's hand if it is still on the battlefield when
// the ability resolves (b28ReturnSelfToOwnersHand) — the ability can
// be activated in response to removal, as in paper, and does nothing
// if the removal got there first.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8a7183cc-161c-444d-a889-a17519c8061b",
		Name:         "Arcanis the Omnipotent",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label: "{T}: Draw three cards.",
				Cost:  TapCost(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DrawCards{Player: item.Controller, N: 3}.Apply(NewContext(g, item))
				},
			},
			{
				Label:  "{2}{U}{U}: Return Arcanis to its owner's hand.",
				Cost:   ManaCost("{2}{U}{U}"),
				Effect: b28ReturnSelfToOwnersHand,
			},
		},
	})
}
