package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sudden Edict — Instant {1}{B}:
//
//	"Split second (As long as this spell is on the stack, players
//	 can't cast spells or activate abilities that aren't mana
//	 abilities.)
//	 Target player sacrifices a creature of their choice."
//
// The edict that can't be answered by flashing in a chump to sacrifice
// instead (#1519). The PLAYER is the target; the creature is their
// choice at resolution, through the same per-player sacrifice prompt
// Liliana of the Veil's −2 uses, so a hexproof creature is still a
// legal pick and a player with no creature sacrifices nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b95f704d-96b9-437e-8c49-aba874139e12",
		Name:            "Sudden Edict",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordSplitSecond},
		Targets:         effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetPlayer {
					continue
				}
				ctx.Game.PlayerSacrificesForEffect(ctx.Source(), t.ID,
					effects.SacrificeSpec("a creature", effects.Creature()),
					"Sudden Edict — sacrifice a creature")
			}
			return nil
		},
	})
}
