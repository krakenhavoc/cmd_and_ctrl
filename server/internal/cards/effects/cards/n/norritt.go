package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Norritt — Creature — Imp {3}{B}, 1/1 (#4062):
//
//	"{T}: Untap target blue creature.
//	 {T}: Choose target non-Wall creature the active player has
//	 controlled continuously since the beginning of the turn. That
//	 creature attacks this turn if able. Destroy it at the beginning of
//	 the next end step if it didn't attack this turn. Activate only
//	 before attackers are declared."
//
// Nettling Imp's ability on any player's turn (its own, the second ability
// is gated by BeforeAttackersAreDeclared alone) and an untap of a blue
// creature, which is not gated. See Nettling Imp for the forced attack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4318c704-888f-4d85-b5f7-6d9abb14e60c",
		Name:         "Norritt",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Untap target blue creature.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target blue creature", effects.OfColor("U")),
				Effect:  effects.UntapFirstLegalTarget,
			},
			{
				Label:     "{T}: Choose target non-Wall creature the active player has controlled continuously since the beginning of the turn. That creature attacks this turn if able. Destroy it at the beginning of the next end step if it didn't attack this turn. Activate only before attackers are declared.",
				Cost:      effects.TapCost(),
				Targets:   effects.ForcedAttackTarget(),
				Condition: effects.BeforeAttackersAreDeclared(),
				Effect:    effects.ForceTheChosenCreature("Norritt"),
			},
		},
	})
}
