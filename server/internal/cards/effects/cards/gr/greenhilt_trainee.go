package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Greenhilt Trainee — Creature — Elf Warrior {3}{G}, 2/3:
//
//	"{T}: Target creature gets +4/+4 until end of turn. Activate only if this
//	 creature's power is 4 or greater."
//
// The power condition is read as the ability is activated (CR 602.5); it is
// not rechecked on resolution.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1d7d340d-fe6f-4314-97ba-e0bf36a4ef6c",
		Name:         "Greenhilt Trainee",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{T}: Target creature gets +4/+4 until end of turn. Activate only if this creature's power is 4 or greater.",
			Purpose:   effects.ForTargets(effects.BenefitUntilEOT(0, 4, 4)),
			Cost:      effects.TapCost(),
			Condition: effects.FutB1SourcePowerAtLeast(4),
			Targets:   effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return effects.BoostUntilEOT{Target: t.ID, Power: 4, Toughness: 4, Label: "Greenhilt Trainee — +4/+4"}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
