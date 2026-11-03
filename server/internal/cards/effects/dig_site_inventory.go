package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "96742ebc-10a6-49e6-812a-631e5a60847d",
		Name:             "Dig Site Inventory",
		Completeness:     CompletenessFull,
		Targets:          TargetCreature("target creature you control", YouControl()),
		Purpose:          ForTargets(BenefitUntilEOT(0, 0, 0, "vigilance")),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{W}")},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"vigilance"},
					Label:    "Dig Site Inventory — vigilance until end of turn",
				}.Apply(ctx)
			}),
		),
	})
}
