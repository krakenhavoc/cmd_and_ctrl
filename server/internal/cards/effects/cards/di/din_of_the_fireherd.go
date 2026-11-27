package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Din of the Fireherd — Sorcery {5}{B/R}{B/R}{B/R} (Shadowmoor, S82 older sets):
//
//	"Create a 5/5 black and red Elemental creature token. Target opponent
//	 sacrifices a creature of their choice for each black creature you control,
//	 then sacrifices a land of their choice for each red creature you control."
//
// The token is made first, so it counts as both a black and a red creature. Both
// counts are read as the spell resolves (CR 608.2h): the opponent's own
// sacrifices cannot change which creatures the caster controls. The opponent
// chooses what to sacrifice (CR 701.21a); the creature prompts are queued before
// the land prompts, so the lands go after the creatures, as printed. A count of
// zero asks nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6480ab32-3a29-4449-8ff7-c8e43c320ed3",
		Name:         "Din of the Fireherd",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerRemove},
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := (effects.CreateToken{Template: effects.TokenCard("5/5 black and red Elemental"), N: 1}).Apply(ctx); err != nil {
				return err
			}
			victim := effects.TargetedPlayerOf(ctx)
			if victim == uuid.Nil {
				return nil
			}
			black := effects.CountYouControl(ctx, effects.And(effects.Creature(), effects.OfColor("B")))
			red := effects.CountYouControl(ctx, effects.And(effects.Creature(), effects.OfColor("R")))
			if black > 0 {
				ctx.Game.PlayerSacrificesNForEffect(item.SourceCardID, victim,
					effects.SacrificeSpec("a creature", effects.Creature()), "Din of the Fireherd — sacrifice a creature", black)
			}
			if red > 0 {
				ctx.Game.PlayerSacrificesNForEffect(item.SourceCardID, victim,
					effects.SacrificeSpec("a land", effects.Land()), "Din of the Fireherd — sacrifice a land", red)
			}
			return nil
		},
	})
}
