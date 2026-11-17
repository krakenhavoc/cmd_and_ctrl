package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "b302bce5-d7ad-46d8-9dbf-8159e8709b4f",
		Name:         "Wing Shards",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPlayer("target player"),
		Triggered:    []game.TriggeredAbility{Storm()},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			player, ok := firstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			ctx.Game.PlayerSacrificesForEffect(ctx.Source(), player,
				sacrificeSpec("an attacking creature", And(Creature(), attackingCreature())),
				"Wing Shards — sacrifice an attacking creature")
			return nil
		},
	})
}
