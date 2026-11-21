package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rhonas's Monument — Legendary Artifact {3}:
//
//	"Green creature spells you cast cost {1} less to cast.
//	 Whenever you cast a creature spell, target creature you control gets
//	 +2/+2 and gains trample until end of turn."
//
// Oketra's Monument's reduction (green) and a cast trigger on any creature
// spell you cast. The target is a creature on the battlefield, so the
// spell that caused the trigger (still on the stack) can never be it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b9a7c722-6b24-4383-ac6e-2a30f7834269",
		Name:         "Rhonas's Monument",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Green creature spells you cast cost {1} less to cast.",
				effects.YourSpell(), effects.CreatureSpell(), effects.ColoredSpell("G")),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventCast, effects.YouCast(effects.Creature()), "Rhonas's Monument — target creature you control gets +2/+2 and gains trample until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Rhonas's Monument — +2/+2"}).Apply(ctx); err != nil {
							return err
						}
						return effects.GrantKeywordUntilEOT{
							Target:   t.ID,
							Keywords: []string{"trample"},
							Label:    "Rhonas's Monument — trample",
						}.Apply(ctx)
					}
					return nil
				}), effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
