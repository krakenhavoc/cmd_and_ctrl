package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Protective Parents — Creature — Human Peasant {2}{W}, 3/2:
//
//	"When this creature dies, create a Young Hero Role token attached to
//	 up to one target creature you control. (If you control another Role
//	 on it, put that one into the graveyard. Enchanted creature has
//	 "Whenever this creature attacks, if its toughness is 3 or less, put
//	 a +1/+1 counter on it.")"
//
// The Parents are in the graveyard when the trigger goes on the stack,
// so they are not a legal pick; "up to one" lets the controller choose
// nothing.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d9717522-358b-41ac-9183-8a6b3981f921",
		Name:         "Protective Parents",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisDies("Protective Parents — create a Young Hero Role token attached to up to one target creature you control",
					effects.CreateRoleOnFirstTarget(effects.RoleYoungHero)),
				effects.TargetCreature("up to one target creature you control", effects.YouControl()).WithCount(0, 1)),
		},
	})
}
