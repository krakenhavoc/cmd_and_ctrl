package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "b9a7c722-6b24-4383-ac6e-2a30f7834269",
		Name:         "Rhonas's Monument",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Green creature spells you cast cost {1} less to cast.",
				YourSpell(), CreatureSpell(), ColoredSpell("G")),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventCast, YouCast(Creature()), "Rhonas's Monument — target creature you control gets +2/+2 and gains trample until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Rhonas's Monument — +2/+2"}).Apply(ctx); err != nil {
							return err
						}
						return GrantKeywordUntilEOT{
							Target:   t.ID,
							Keywords: []string{"trample"},
							Label:    "Rhonas's Monument — trample",
						}.Apply(ctx)
					}
					return nil
				}), TargetCreature("target creature you control", YouControl())),
		},
	})
}
