package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Holy Justiciar — Creature — Human Cleric {3}{W}, 2/1:
//
//	"{2}{W}, {T}: Tap target creature. If that creature is a Zombie, exile it."
//
// The Zombie check reads the creature after it is tapped, through the layered
// view, so a creature that is a Zombie by an effect counts.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9edcf49b-2767-48bd-8fbe-968527cbddf4",
		Name:         "Holy Justiciar",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}, {T}: Tap target creature. If that creature is a Zombie, exile it.",
			Cost:    game.AbilityCost{Mana: "{2}{W}", Tap: true},
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				zombie := "Zombie"
				if src, ok := g.LookupCardForEffect(item.SourceCardID); ok {
					zombie = src.TextSubtype("Zombie")
				}
				for _, t := range ctx.LegalTargets() {
					if err := (effects.TapTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
					if c, ok := g.LookupCardForEffect(t.ID); ok && c.HasSubtype(zombie) {
						return effects.ExileTarget{Target: t.ID}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
