package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ugin's Construct — {4} Artifact Creature — Construct 4/5 (Fate Reforged):
//
//	"When this creature enters, sacrifice a permanent that's one or more
//	 colors."
//
// You choose what to sacrifice as the trigger resolves (CR 701.17a), among the
// permanents you control with at least one color, on effective colors; the
// Construct itself is colorless and is never a candidate. With no colored
// permanent nothing is sacrificed.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6b87a933-b0d2-4120-b4a0-65ab0f3dbe48",
		Name:         "Ugin's Construct",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ugin's Construct — sacrifice a permanent that's one or more colors",
				func(g *game.Game, item *game.StackItem) error {
					g.PlayerSacrificesForEffect(item.SourceCardID, item.Controller,
						effects.SacrificeSpec("a permanent that's one or more colors", effects.Not(effects.Colorless())),
						"Ugin's Construct — sacrifice a permanent that's one or more colors")
					return nil
				}),
		},
	})
}
