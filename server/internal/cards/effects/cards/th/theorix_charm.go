package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Theorix Charm — Instant {U}{B} (Reality Fracture, tracker #2795):
//
//	"Choose one —
//	 • Counter target noncreature spell unless its controller pays {2}.
//	 • Target creature gets -2/-2 until end of turn.
//	 • Mill three cards, then draw a card."
//
// The tax question goes to the countered spell's controller. The draw
// runs after the mill has landed (it is the mill's continuation).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf527bbd-e898-4aa9-909d-daec2f4b62ad",
		Name:         "Theorix Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Counter target noncreature spell unless its controller pays {2}.",
				effects.TargetSpell("target noncreature spell", effects.Noncreature()), effects.RfSpellBCounterNoncreatureUnlessPays),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gets -2/-2 until end of turn.",
				effects.TargetCreature("target creature"), effects.BoostTheModesTarget(-2, -2, "Theorix Charm — -2/-2")), effects.ForTargets(effects.HarmUntilEOT(0, -2, -2))),
			effects.ModeDoing("Mill three cards, then draw a card.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.MillToZone{N: 3, ThenTail: ctx.ItemTail(theorixCharmMilled)}.Apply(ctx)
				}),
		),
	})
}

// theorixCharmMilled is the clause after the mill (#4253 slice 9).
var theorixCharmMilled = effects.CardTail("theorix-charm/after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	c := ctx

	return effects.DrawCards{Player: c.Controller(), N: 1}.Apply(c)
})
