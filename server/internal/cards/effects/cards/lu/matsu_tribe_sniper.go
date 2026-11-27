package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Matsu-Tribe Sniper — Creature — Snake Warrior Archer {1}{G}, 1/1
// (Betrayers of Kamigawa, #5602):
//
//	"{T}: This creature deals 1 damage to target creature with flying.
//	 Whenever this creature deals damage to a creature, tap that creature
//	 and it doesn't untap during its controller's next untap step."
//
// The activated ability is Centaur Archer's. The trigger fires once per
// damage event the Sniper is the source of whose recipient is a creature on
// the battlefield (ThisDealtDamageToACreature), the Sniper's own ping or
// combat damage alike, and taps and freezes the creature that was dealt the
// damage (TapAndFreeze), read off the event the trigger carries.
//
// No simplification.
func init() {
	const label = "Matsu-Tribe Sniper — tap that creature and it doesn't untap during its controller's next untap step"
	effects.Register(effects.Spec{
		OracleID:     "c20728ac-c01c-470f-a43b-de6104836645",
		Name:         "Matsu-Tribe Sniper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to target creature with flying.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
			Effect:  effects.DealDamageToTheTarget(1),
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisDealtDamageToACreature, label,
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return effects.TapAndFreeze{Targets: []uuid.UUID{item.Trigger.Event.Target}, Label: label}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
