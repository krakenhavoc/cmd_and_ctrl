package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dig Site Inventory — Sorcery {W} (Secrets of Strixhaven, #4456):
//
//	"Put a +1/+1 counter on target creature you control. It gains
//	 vigilance until end of turn.
//	 Flashback {W} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Snakeskin Veil's shape: the counter, then the keyword as a second step so
// a counter that waits on a question still gets it (CR 608.2c). The
// flashback is the keyword helper's alternative cost; a flashbacked spell
// is exiled however it leaves the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "96742ebc-10a6-49e6-812a-631e5a60847d",
		Name:             "Dig Site Inventory",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetCreature("target creature you control", effects.YouControl()),
		Purpose:          effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "vigilance")),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{W}")},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"vigilance"},
					Label:    "Dig Site Inventory — vigilance until end of turn",
				}.Apply(ctx)
			}),
		),
	})
}
