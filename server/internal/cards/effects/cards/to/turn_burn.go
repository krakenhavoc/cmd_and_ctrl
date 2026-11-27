package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Turn // Burn — split card, oracle 41841bbf-1c51-494b-b299-c997cce88e44
// (Dragon's Maze, S82 OLD batch 402, #5564):
//
//	Turn — Instant {2}{U}: "Until end of turn, target creature loses all
//	       abilities and becomes a red Weird with base power and toughness 0/1."
//	Burn — Instant {1}{R}: "Burn deals 2 damage to any target."
//	       Fuse (You may cast one or both halves of this card from your hand.)
//
// Both halves register as ADR 0034 faces (Turn under the bare oracle ID, Burn
// under "#1"); the fused cast is the engine's (CR 702.102): Turn's clause first,
// Burn's second, resolving Turn then Burn.
//
// Turn is Snakeform's one continuous effect until end of turn: layer 4 the
// creature type becomes Weird, layer 5 the colour is set to red, layer 6 all
// abilities are lost and layer 7b the base power and toughness are 0/1
// (counters and pumps still apply). Burn is Lightning Bolt for 2.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "41841bbf-1c51-494b-b299-c997cce88e44",
		Name:         "Turn",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.UntilEndOfTurn(ctx, id, nil,
					"Turn — loses all abilities and becomes a red Weird with base power and toughness 0/1 until end of turn",
					game.SetCreatureTypesMod(ctx.TextSubtypes("Weird")...),
					game.SetColorsMod(ctx.TextColors("R")...),
					game.LoseAllAbilitiesMod(),
					game.SetBasePowerMod(0), game.SetBaseToughnessMod(1))
			}),
		),
	})
	effects.Register(effects.Spec{
		OracleID:     "41841bbf-1c51-494b-b299-c997cce88e44#1",
		Name:         "Burn",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 2)), Answers: game.AnswerRemove},
		OnResolve:    effects.DamageToFirstTargetStep(2),
	})
}
