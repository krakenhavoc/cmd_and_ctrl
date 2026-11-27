package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dragonshift — Instant {1}{U}{R} (Dragon's Maze, S82 OLD batch 403, #5564):
//
//	"Until end of turn, target creature you control becomes a blue and red
//	 Dragon with base power and toughness 4/4, loses all abilities, and gains
//	 flying.
//	 Overload {3}{U}{U}{R}{R} (You may cast this spell for its overload cost. If
//	 you do, change "target" in its text to "each.")"
//
// Scale Up's one continuous effect until end of turn: layer 4 the creature type
// becomes Dragon, layer 5 the colours are set to blue and red, layer 6 all
// abilities are lost and flying is gained (Hellcat's pair, in that order, with one
// timestamp), and layer 7b the base power and toughness are 4/4 (counters and
// pumps still apply). Overload is the engine's (CR 702.96): it clears the target,
// and the effect then covers each creature you control, the set fixed as the spell
// resolves (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b02d24fd-62b4-4f78-96af-57f88c98c5c8",
		Name:         "Dragonshift",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		Purpose:      game.Purpose{Answers: game.AnswerAnimate},
		AlternativeCosts: []game.AlternativeCost{
			effects.CostWithPurpose(effects.Overload("{3}{U}{U}{R}{R}"), game.Purpose{Answers: game.AnswerAnimate}),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			mods := []game.Mod{
				game.SetCreatureTypesMod(ctx.TextSubtype("Dragon")),
				game.SetColorsMod(ctx.TextColors("U", "R")...),
				game.LoseAllAbilitiesMod(),
				game.AddKeywordsMod("flying"),
				game.SetBasePowerMod(4),
				game.SetBaseToughnessMod(4),
			}
			const label = "Dragonshift — becomes a blue and red Dragon with base power and toughness 4/4, loses all abilities, and gains flying"
			if ctx.PaidAltCost("overload") {
				return effects.UntilEndOfTurn(ctx, uuid.Nil, effects.And(effects.Creature(), effects.YouControl()), label, mods...)
			}
			id := effects.FirstLegalBattlefieldTarget(ctx)
			if id == uuid.Nil {
				return nil
			}
			return effects.UntilEndOfTurn(ctx, id, nil, label, mods...)
		},
	})
}
