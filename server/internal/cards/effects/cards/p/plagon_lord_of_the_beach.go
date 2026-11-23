package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plagon, Lord of the Beach — Legendary Creature — Starfish Wizard {2}{U},
// 0/3 (Foundations Jumpstart):
//
//	"When Plagon enters, draw a card for each creature you control with
//	 toughness greater than its power.
//	 {W/U}: Target creature you control assigns combat damage equal to its
//	 toughness rather than its power this turn."
//
// The draw counts as the enter trigger resolves, on current characteristics,
// Plagon himself among the creatures (a 0/3 qualifies). The ability is Walking
// Bulwark's until-end-of-turn damage-by-toughness effect (the unconditional
// form, CR 510.1c), pinned to the creature chosen (CR 611.2c); it is a layer
// effect on the creature, not a granted ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1c95b7cf-be4d-4961-9073-3e39487492c8",
		Name:         "Plagon, Lord of the Beach",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Plagon, Lord of the Beach — draw a card for each creature you control with toughness greater than its power",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					n := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.IsCreature() && c.Controller == item.Controller && c.CurrentToughness() > c.CurrentPower() {
							n++
						}
					}
					return effects.DrawCards{N: n}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W/U}: Target creature you control assigns combat damage equal to its toughness rather than its power this turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{W/U}"),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := effects.FirstLegalTarget(ctx)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.UntilEndOfTurn(ctx, t.ID, nil, "Plagon, Lord of the Beach — assigns combat damage by toughness",
					game.DamageByToughnessMod())
			},
		}},
	})
}
