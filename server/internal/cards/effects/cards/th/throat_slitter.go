package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "56f782b7-d46c-4aa1-ae16-66d6401c72ae",
		Name:         "Throat Slitter",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Ninjutsu("{2}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisDealsCombatDamageToAPlayer(
				"Throat Slitter — destroy target nonblack creature that player controls",
				effects.DestroyFirstLegalTarget),
				effects.ControlledByThatPlayer(effects.TargetCreature("target nonblack creature that player controls", effects.NonBlack()))),
		},
	})
}
