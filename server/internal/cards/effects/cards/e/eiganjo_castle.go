package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eiganjo Castle — Legendary Land (Champions of Kamigawa, #5603):
//
//	"{T}: Add {W}.
//	 {W}, {T}: Prevent the next 2 damage that would be dealt to target
//	 legendary creature this turn."
//
// Shizo, Death's Storehouse's shape with Battlefield Medic's shield: the
// prevention is a 2-point shield on the creature that lasts the turn (CR
// 615.7). A target that stops being legendary in response gets nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "895a0e00-20a9-44f8-9215-66edcdf016b7",
		Name:         "Eiganjo Castle",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: Prevent the next 2 damage that would be dealt to target legendary creature this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target legendary creature", effects.Legendary()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.PreventNextDamage{Target: id, Amount: 2, Label: "Eiganjo Castle: prevent the next 2 damage"}.Apply(ctx)
			},
		}},
	})
}
