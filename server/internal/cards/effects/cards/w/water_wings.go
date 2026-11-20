package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Water Wings — Instant {1}{U} (Wilds of Eldraine, #4471):
//
//	"Until end of turn, target creature you control has base power and
//	 toughness 4/4 and gains flying and hexproof. (It can't be the target
//	 of spells or abilities your opponents control.)"
//
// I Am Iron Man's shape without the type change: one until-end-of-turn
// record over the target with three mods at one timestamp, base power and
// toughness 4/4 at layer 7b (so counters and pumps still apply on top) and
// flying and hexproof at layer 6. A target that is gone by resolution
// leaves the spell with nothing to do (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "17f8e8cb-cd1b-4d94-9efa-ce268e8e7096",
		Name:         "Water Wings",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying", "hexproof")),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.UntilEndOfTurn(ctx, id, nil,
				"Water Wings — base power and toughness 4/4, flying and hexproof",
				game.SetBasePowerMod(4), game.SetBaseToughnessMod(4), game.AddKeywordsMod("flying", "hexproof"))
		},
	})
}
