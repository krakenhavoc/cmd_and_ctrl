package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prosperous Thief — Creature — Human Ninja {2}{U}, 3/2:
//
//	"Ninjutsu {1}{U} ({1}{U}, Return an unblocked attacker you control
//	 to hand: Put this card onto the battlefield from your hand tapped
//	 and attacking.)
//	 Whenever one or more Ninja or Rogue creatures you control deal
//	 combat damage to a player, create a Treasure token."
//
// #1227's fourth ninjutsu proof, and the one whose trigger is a BATCH
// rather than a creature: "one or more … deal" is one trigger per
// player connected with in a damage step (CR 603.2c), however many
// Ninja and Rogues got there. Two Ninja hitting one player make one
// Treasure; two Ninja hitting two different players make two.
//
// Everything is shared machinery — Ninjutsu("{1}{U}") for the keyword
// (ninjutsu.go), the once-per-batch-per-player trigger Professional
// Face-Breaker and Thopter Spy Network already use, and the Treasure
// the whole catalog mints. Subtypes are read post-layer off the damage
// source, so a lorded or changeling body counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6d03971a-8365-449a-8fbe-07b5b9bb42dc",
		Name:         "Prosperous Thief",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Ninjutsu("{1}{U}")},
		Triggered: []game.TriggeredAbility{
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(
				Or(HasSubtype("Ninja"), HasSubtype("Rogue")),
				"Prosperous Thief — create a Treasure token",
				Do(CreateToken{Template: TreasureToken(), N: 1}),
			),
		},
	})
}
