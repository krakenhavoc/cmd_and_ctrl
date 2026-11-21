package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mistblade Shinobi — Creature — Human Ninja {2}{U}, 1/1 (Betrayers of
// Kamigawa, #5602):
//
//	"Ninjutsu {U} ({U}, Return an unblocked attacker you control to hand:
//	 Put this card onto the battlefield from your hand tapped and
//	 attacking.)
//	 Whenever this creature deals combat damage to a player, you may
//	 return target creature that player controls to its owner's hand."
//
// The keyword is Ninjutsu("{U}") (ninjutsu.go). The damage trigger is
// Trygon Predator's shape: "that player controls" is
// ControlledByThatPlayer, judged against the player this trigger's damage
// event named, and "you may" is the trigger's own yes/no as it resolves
// (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "16856e22-6f9e-4f8e-b4ee-622a7c2ee453",
		Name:         "Mistblade Shinobi",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Ninjutsu("{U}")},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WheneverThisDealsCombatDamageToAPlayer(
				"Mistblade Shinobi — return target creature that player controls to its owner's hand",
				effects.BounceUpToOneTarget),
				effects.ControlledByThatPlayer(effects.TargetCreature("target creature that player controls"))),
				"Mistblade Shinobi — return a creature that player controls to its owner's hand?"),
		},
	})
}
