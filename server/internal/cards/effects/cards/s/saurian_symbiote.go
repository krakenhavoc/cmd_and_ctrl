package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Saurian Symbiote — Creature — Fungus Dinosaur {3}{G}, 2/3
// (Foundations Jumpstart, #5464):
//
//	"Reach (This creature can block creatures with flying.)
//	 When this creature enters, choose one —
//	 • Put a +1/+1 counter on this creature.
//	 • Create a 1/1 green Saproling creature token."
//
// Trusty Retriever's modal enters trigger: the mode is chosen as the
// trigger is put on the stack (CR 700.2b). The counter bullet does
// nothing to a Symbiote that has left the battlefield (CR 400.7); the
// Saproling is made either way.
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Saurian Symbiote — choose one", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = effects.ChooseOne(
		effects.ModeDoing("Put a +1/+1 counter on this creature.", nil, effects.PutAPlusOneCounterOnThisCreatureMode),
		effects.ModeDoing("Create a 1/1 green Saproling creature token.", nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: 1}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:        "14b180d2-5701-452d-8deb-cf0845c95df5",
		Name:            "Saurian Symbiote",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}
