package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Throat Slitter — Creature — Rat Ninja {4}{B}, 2/2 (Betrayers of
// Kamigawa, #5602):
//
//	"Ninjutsu {2}{B} ({2}{B}, Return an unblocked attacker you control
//	 to hand: Put this card onto the battlefield from your hand tapped
//	 and attacking.)
//	 Whenever this creature deals combat damage to a player, destroy
//	 target nonblack creature that player controls."
//
// The keyword is Ninjutsu("{2}{B}") (ninjutsu.go). The damage trigger is
// Scion of Calamity's: the target is chosen among what the damaged
// player controls (ControlledByThatPlayer) as the trigger goes on the
// stack and re-checked as it resolves. It is not optional.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "56f782b7-d46c-4aa1-ae16-66d6401c72ae",
		Name:         "Throat Slitter",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Ninjutsu("{2}{B}")},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisDealsCombatDamageToAPlayer(
				"Throat Slitter — destroy target nonblack creature that player controls",
				destroyFirstLegalTarget),
				ControlledByThatPlayer(TargetCreature("target nonblack creature that player controls", NonBlack()))),
		},
	})
}
