package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bellowing Mauler — Creature — Ogre Warrior {4}{B}, 4/6 (Streets of New
// Capenna Commander, S82):
//
//	"At the beginning of your end step, each player loses 4 life unless they
//	 sacrifice a nontoken creature of their choice."
//
// A punisher asked of every player, you included, in turn order from the
// controller (s82_snc_b13_helpers.go, after punisherRepeat): each player
// chooses between losing 4 life and sacrificing a nontoken creature, one
// question at a time, each queued by the answer to the one before it. A
// player with no nontoken creature isn't asked and loses the life (the
// choice is impossible); a token can't pay. Nothing targets.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa0a4765-5799-4464-88b6-6d1b15889eb5",
		Name:         "Bellowing Mauler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Bellowing Mauler — each player loses 4 life unless they sacrifice a nontoken creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.BellowingMaulerStep(ctx, effects.BellowingMaulerPlayers(ctx))
				}),
		},
	})
}
