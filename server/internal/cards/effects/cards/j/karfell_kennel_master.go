package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Karfell Kennel-Master — Creature — Zombie Berserker {4}{B}, 4/4
// (Kaldheim, #5505):
//
//	"When this creature enters, up to two target creatures each get
//	 +1/+0 and gain indestructible until end of turn."
//
// Invasion of Kylem's enters trigger at +1/+0 and indestructible: each
// target still legal as it resolves gets both, and none chosen is a
// legal answer.
//
// No simplification.
func init() {
	const name = "Karfell Kennel-Master"
	enters := effects.WhenThisEnters(name+" — up to two target creatures each get +1/+0 and gain indestructible until end of turn",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 1, Label: name + " — +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"indestructible"}, Label: name + " — indestructible"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		})
	enters.Targets = effects.TargetCreature("up to two target creatures").WithCount(0, 2)
	effects.Register(effects.Spec{
		OracleID:     "2cde1ca7-d126-4d3a-82a6-0acfd7c85cdb",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}
