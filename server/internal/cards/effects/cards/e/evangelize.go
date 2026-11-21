package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Evangelize — Sorcery {4}{W}:
//
//	"Buyback {2}{W}{W} (You may pay an additional {2}{W}{W} as you cast
//	 this spell. If you do, put this card into your hand as it
//	 resolves.)
//	 Gain control of target creature of an opponent's choice they
//	 control."
//
// The target is an opponent's choice (#2055, ADR 0019's 2026-10-16
// amendment): as the spell is cast its caster chooses an opponent, and
// that opponent chooses a creature they control. The rulings: if the
// target isn't a creature that opponent controls as Evangelize resolves,
// it doesn't resolve, and a new target chosen for it must be a creature
// that opponent controls. The control change has no duration (CR
// 611.2a). Buyback is the engine's (ADR 0073 §6).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "23a60803-6b0f-48ee-aa3b-12a3c5fc62cf",
		Name:         "Evangelize",
		Completeness: effects.CompletenessFull,
		Targets: effects.TheyControl(effects.OfAnOpponentsChoice(
			effects.TargetCreature("target creature of an opponent's choice they control"))),
		OptionalCosts: []game.AdditionalCost{
			effects.Buyback("{2}{W}{W}"),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return effects.GainControl{Target: t.ID, Duration: game.IndefiniteDuration(), Label: "Evangelize"}.Apply(ctx)
		},
	})
}
