package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
		return CostsLess(2, "The first legendary creature spell you cast each turn costs {2} less to cast.",
			YourSpell(), firstLegendaryCreatureSpellYouCastThisTurn())
	}
	Register(Spec{
		OracleID:      serahOracle,
		Name:          "Serah Farron",
		Completeness:  CompletenessFull,
		CostModifiers: []game.CostModifier{discount()},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "you control two or more other legendary creatures",
				Test: func(ctx *Context) bool {
					return controlsAtLeastOtherLegendaryCreatures(ctx.Game, ctx.Controller(), ctx.Source(), 2)
				},
			}, Optional(AtBeginningOfYourCombat("Serah Farron — you may transform Serah Farron",
				func(g *game.Game, item *game.StackItem) error {
					return Transform{Target: item.SourceCardID}.Apply(NewContext(g, item))
				}), "Serah Farron — transform her?")),
		},
	})
	Register(Spec{
		OracleID:      serahOracle + "#1",
		Name:          "Crystallized Serah",
		Completeness:  CompletenessFull,
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
