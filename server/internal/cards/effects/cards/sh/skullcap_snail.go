package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skullcap Snail — Creature — Fungus Snail {1}{B}, 1/1 (The Lost Caverns of
// Ixalan, #4470):
//
//	"When this creature enters, target opponent exiles a card from their
//	 hand."
//
// The opponent is targeted as the trigger goes on the stack and picks their
// own card (Ruthless Negotiation's exile, a floor of one: it is not a
// "may"). An empty hand exiles nothing (CR 609.3), and an opponent who has
// left by resolution is an illegal target (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6420e8a0-3ef4-4f95-bb6a-12409eef4d48",
		Name:         "Skullcap Snail",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Skullcap Snail — target opponent exiles a card from their hand",
				func(g *game.Game, item *game.StackItem) error {
					return effects.TargetsExileACardFromTheirHand(effects.NewContext(g, item), "Skullcap Snail")
				}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
