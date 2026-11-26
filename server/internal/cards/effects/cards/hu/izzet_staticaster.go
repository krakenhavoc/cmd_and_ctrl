package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Izzet Staticaster — Creature — Human Wizard {1}{U}{R}, 0/3 (Return to
// Ravnica, S82 OLD):
//
//	"Flash
//	 Haste
//	 {T}: This creature deals 1 damage to target creature and each other
//	 creature with the same name as that creature."
//
// Only the target is a real target (CR 115.1). The set is the target plus
// every other creature sharing its name as the ability resolves (Bile Blight's
// read); each takes 1 damage from the Staticaster at once. With the target
// illegal the ability does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d63adbed-07d3-44e4-83f8-1cbdfc186df4",
		Name:            "Izzet Staticaster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to target creature and each other creature with the same name as that creature.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item).AsGroupMember()
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					ids := effects.S82RprB17SameNameAsTarget(g, t.ID)
					return g.DamageInstanceForEffect(func() error {
						for _, id := range ids {
							c, ok := g.LookupCardForEffect(id)
							if !ok || !c.IsCreature() {
								continue
							}
							if err := (effects.DealDamage{Source: item.SourceCardID, Target: id, Amount: 1}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					})
				}
				return nil
			},
		}},
	})
}
