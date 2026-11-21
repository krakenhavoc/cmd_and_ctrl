package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frantic Confrontation — Instant {X}{R} (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"Target creature you control gets +X/+0 and gains first strike and
//	 trample until end of turn."
//
// X is announced with the cast (CR 107.3, 601.2b) and read back with
// ctx.X(); the pump (layer 7c) and the two keywords (layer 6) are both
// until-end-of-turn effects on the one target, applied as it resolves. A
// target that has become illegal gets neither (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a866cc05-69b0-45f2-8829-b6ef08193ed1",
		Name:         "Frantic Confrontation",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      effects.ForTargets(effects.PowerXUntilEOT(0, "first strike", "trample")),
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: t.ID, Power: ctx.X(), Label: "Frantic Confrontation — +X/+0"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"first strike", "trample"},
				Label: "Frantic Confrontation — first strike and trample"}.Apply(ctx)
		},
	})
}
