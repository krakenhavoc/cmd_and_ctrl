package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Whack — Sorcery {3}{B} (Streets of New Capenna, #5491):
//
//	"This spell costs {3} less to cast if it targets a white creature.
//	 Target creature gets -4/-4 until end of turn."
//
// The reduction reads the announced target at CR 601.2f (Price of Fame's
// shape, CostsLessIfItTargets), so the cast and the legal-move enumerator
// both price it with the target chosen. The colour is read through the
// target's layered colours. A target that is gone or no longer legal on
// resolution is left alone (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "19becb14-3774-4732-b792-d2d0e5a94b35",
		Name:         "Whack",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -4, -4)),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessIfItTargets(3, "This spell costs {3} less to cast if it targets a white creature.",
				effects.And(effects.Creature(), effects.OfColor("W"))),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.BoostUntilEOT{Target: id, Power: -4, Toughness: -4, Label: "Whack — -4/-4 until end of turn"}.Apply(ctx)
		},
	})
}
