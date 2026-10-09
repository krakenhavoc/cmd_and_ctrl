package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Splashy Spellcaster — Creature — Elemental Wizard {3}{U}, 2/4:
//
//	"Whenever you cast an instant or sorcery spell, create a Sorcerer
//	 Role token attached to up to one other target creature you control.
//	 (If you control another Role on it, put that one into the
//	 graveyard. Enchanted creature gets +1/+1 and has "Whenever this
//	 creature attacks, scry 1.")"
//
// The trigger goes on the stack above the spell that caused it, with
// "up to one" meaning it may choose no creature and do nothing.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "21486218-70ec-4c60-9844-8f7c80912333",
		Name:         "Splashy Spellcaster",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventCast, YouCast(Or(Instant(), Sorcery())),
					"Splashy Spellcaster — create a Sorcerer Role token attached to up to one other target creature you control",
					createRoleOnFirstTarget(RoleSorcerer)),
				Another(TargetCreature("up to one other target creature you control", YouControl()).WithCount(0, 1))),
		},
	})
}
