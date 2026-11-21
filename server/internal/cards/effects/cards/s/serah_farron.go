package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Serah Farron // Crystallized Serah — a transforming card (Final Fantasy,
// #4462).
//
// Front, Serah Farron — Legendary Creature — Human Citizen {1}{G}{W}, 2/2:
//
//	"The first legendary creature spell you cast each turn costs {2} less
//	 to cast.
//	 At the beginning of combat on your turn, if you control two or more
//	 other legendary creatures, you may transform Serah Farron."
//
// Back, Crystallized Serah — Legendary Artifact:
//
//	"The first legendary creature spell you cast each turn costs {2} less
//	 to cast.
//	 Legendary creatures you control get +2/+2."
//
// The discount is a generic-only reduction (CR 601.2f) on the first
// legendary creature spell of the caster's turn, read off the new
// LegendaryCreature cast tally; each face carries its own, so it works
// whichever face is up. The transform is an intervening-if trigger
// (CR 603.4): asked as combat begins and again as it resolves, with the
// printed "you may" asked on resolution. The anthem is a layer 7c +2/+2 on
// every legendary creature its controller controls (Crystallized Serah is
// not a creature).
func init() {
	const serahOracle = "a8a7a3b7-015c-4fbd-8b90-e0ddf71663e4"
	discount := func() game.CostModifier {
		return effects.CostsLess(2, "The first legendary creature spell you cast each turn costs {2} less to cast.",
			effects.YourSpell(), effects.FirstLegendaryCreatureSpellYouCastThisTurn())
	}
	effects.Register(effects.Spec{
		OracleID:      serahOracle,
		Name:          "Serah Farron",
		Completeness:  effects.CompletenessFull,
		CostModifiers: []game.CostModifier{discount()},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you control two or more other legendary creatures",
				Test: func(ctx *effects.Context) bool {
					return effects.ControlsAtLeastOtherLegendaryCreatures(ctx.Game, ctx.Controller(), ctx.Source(), 2)
				},
			}, effects.Optional(effects.AtBeginningOfYourCombat("Serah Farron — you may transform Serah Farron",
				func(g *game.Game, item *game.StackItem) error {
					return effects.Transform{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				}), "Serah Farron — transform her?")),
		},
	})
	effects.Register(effects.Spec{
		OracleID:      serahOracle + "#1",
		Name:          "Crystallized Serah",
		Completeness:  effects.CompletenessFull,
		CostModifiers: []game.CostModifier{discount()},
		Static: []game.StaticAbility{{
			Label:    "Legendary creatures you control get +2/+2.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.IsCreature() && target.IsLegendary()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power += 2
				c.Toughness += 2
			},
		}},
	})
}
