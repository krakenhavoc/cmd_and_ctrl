package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wing Shards — Instant {1}{W}{W} (Scourge):
//
//	"Target player sacrifices an attacking creature of their choice.
//	 Storm (When you cast this spell, copy it for each spell cast before
//	 it this turn. You may choose new targets for the copies.)"
//
// Blessed Alliance's third mode on a target player of any kind, plus the
// storm keyword (CR 702.40a). The player chooses as the spell resolves, and
// the sacrifice does not target, so hexproof on the attacker does not
// matter. Each copy is its own edict and may name a different player (CR
// 707.10c). A player with no attacking creature sacrifices nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b302bce5-d7ad-46d8-9dbf-8159e8709b4f",
		Name:         "Wing Shards",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPlayer("target player"),
		Triggered:    []game.TriggeredAbility{effects.Storm()},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			ctx.Game.PlayerSacrificesForEffect(ctx.Source(), player,
				effects.SacrificeSpec("an attacking creature", effects.And(effects.Creature(), effects.IsAttackingCreature())),
				"Wing Shards — sacrifice an attacking creature")
			return nil
		},
	})
}
