package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Keeper of Fables — Creature — Cat {3}{G}{G}, 4/5 (EDHREC rank
// 3210):
//
//	"Whenever one or more non-Human creatures you control deal
//	 combat damage to a player, draw a card."
//
// The non-Human beatdown deck's Bident. "ONE OR MORE … to A PLAYER"
// is one trigger per PLAYER connected with (CR 603.2c, and this
// card's own ruling of 2019-10-04: "if non-Human creatures you
// control deal combat damage to two or more players at the same time,
// Keeper of Fables's ability triggers for each of those players"),
// not one per creature and not one per damage step. The engine emits
// one damage event per creature, so the guard is
// OncePerBatchPerPlayer (see AGENTS.md §7): three non-Humans on one
// opponent draw one card, three on three opponents draw three.
// First-strike and regular damage are two damage steps and two
// batches (CR 510.4), so a draw for each (step, player) pair, as in
// paper. Human is read off the creature's effective subtypes, so a
// changeling is a Human and does not count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c8ca3116-e0f0-4e27-aa0f-99ed85927040",
		Name:         "Keeper of Fables",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(b30NonHuman(),
				"Keeper of Fables — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
