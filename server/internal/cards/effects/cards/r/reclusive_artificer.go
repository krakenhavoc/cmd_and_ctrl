package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reclusive Artificer — Creature — Human Artificer {2}{U}{R}, 2/3
// (Magic Origins):
//
//	"Haste
//	 When this creature enters, you may have it deal damage to target
//	 creature equal to the number of artifacts you control."
//
// An optional enters trigger with a target (Hissing Iguanar's
// Optional-around-Targeting shape). The artifacts are counted as the
// trigger resolves, the Artificer's own entry included if it is one;
// it deals the damage, so a creature that has left is read as it last
// was (CR 608.2h). With no artifacts nothing is dealt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "de2c7724-647c-4ca2-be4d-f0e348a59262",
		Name:            "Reclusive Artificer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(
					effects.WhenThisEnters("Reclusive Artificer — damage to target creature equal to the number of artifacts you control",
						func(g *game.Game, item *game.StackItem) error {
							ctx := effects.NewContext(g, item)
							n := effects.RfCreatureAArtifactsYouControl(g, item.Controller)
							if n <= 0 {
								return nil
							}
							ts := ctx.LegalTargets()
							if len(ts) == 0 {
								return nil
							}
							return effects.DealDamage{Source: ctx.Source(), Target: ts[0].ID, Amount: n}.Apply(ctx)
						}),
					"Reclusive Artificer — deal damage to the target creature equal to the number of artifacts you control?"),
				effects.TargetCreature("target creature")),
		},
	})
}
