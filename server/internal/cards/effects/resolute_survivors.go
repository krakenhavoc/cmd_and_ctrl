package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Resolute Survivors — Creature — Human Warrior {1}{R}{W}, 3/3:
//
//	"You may exert this creature as it attacks. (It won't untap during
//	 your next untap step.)
//	 Whenever you exert a creature, this creature deals 1 damage to
//	 each opponent and you gain 1 life."
//
// ADR 0130 §11: exert as it attacks (CR 701.43d) with no linked
// trigger, and a "whenever you exert a creature" payoff that sees this
// creature and any other creature its controller exerts (the Amonkhet
// ruling). Damage, not life loss; the life gain happens whether or not
// any damage was dealt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "3d699db7-cc52-4ee3-947b-0ba291bf037a",
		Name:          "Resolute Survivors",
		Completeness:  CompletenessFull,
		ExertOnAttack: ExertAsItAttacks(),
		Triggered: []game.TriggeredAbility{
			WheneverYouExert("Resolute Survivors — 1 damage to each opponent, gain 1 life",
				damageEachOpponentThenGainLife(1)),
		},
	})
}
