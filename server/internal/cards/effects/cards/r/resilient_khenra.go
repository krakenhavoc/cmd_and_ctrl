package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Resilient Khenra — {1}{G} Creature — Jackal Wizard 2/2 (Hour of Devastation):
//
//	"When this creature enters, you may have target creature get +X/+X until
//	 end of turn, where X is this creature's power.
//	 Eternalize {4}{G}{G}"
//
// The target is chosen as the trigger goes on the stack; the "may" is asked on
// resolution (Primal Boost's shape). X is the power as it resolves, or as the
// creature last had it (CR 608.2h): the eternalized 4/4 gives +4/+4.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eb23f999-e7f9-4082-b322-5b00ba64b458",
		Name:         "Resilient Khenra",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(
					effects.WhenThisEnters("Resilient Khenra — you may have target creature get +X/+X until end of turn, where X is its power",
						func(g *game.Game, item *game.StackItem) error {
							ctx := effects.NewContext(g, item)
							id, ok := effects.B16FirstLegalTargetCard(ctx)
							if !ok {
								return nil
							}
							info, ok := ctx.SourcePermanent()
							if !ok || info.Power == 0 {
								return nil
							}
							return effects.BoostUntilEOT{Target: id, Power: info.Power, Toughness: info.Power, Label: "Resilient Khenra — +X/+X"}.Apply(ctx)
						}),
					"Resilient Khenra — have target creature get +X/+X until end of turn?"),
				effects.TargetCreature("target creature"),
			),
		},
		Activated: []effects.ActivatedAbility{effects.Eternalize("{4}{G}{G}")},
	})
}
