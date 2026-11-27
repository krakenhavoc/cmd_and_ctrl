package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Molder Slug — Creature — Slug Beast {3}{G}{G}, 4/6 (Mirrodin, #5606):
//
//	"At the beginning of each player's upkeep, that player sacrifices an
//	 artifact of their choice."
//
// Ruthless Winnower's upkeep edict on artifacts: it fires in every player's
// upkeep and that player chooses what to sacrifice among the artifacts they
// control (a player with exactly one is not asked, one with none sacrifices
// nothing).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d7c8b636-b7ff-446d-aa51-516e2cfe2730",
		Name:         "Molder Slug",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Molder Slug — that player sacrifices an artifact of their choice",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					g.PlayerSacrificesForEffect(item.SourceCardID, ctx.Trigger().Event.Actor,
						effects.SacrificeSpec("an artifact", effects.Artifact()), "Molder Slug — sacrifice an artifact")
					return nil
				}),
		},
	})
}
