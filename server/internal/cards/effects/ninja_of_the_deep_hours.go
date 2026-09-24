package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ninja of the Deep Hours — Creature — Human Ninja {3}{U}, 2/2:
//
//	"Ninjutsu {1}{U} ({1}{U}, Return an unblocked attacker you control
//	 to hand: Put this card onto the battlefield from your hand tapped
//	 and attacking.)
//	 Whenever this creature deals combat damage to a player, you may
//	 draw a card."
//
// #1227's canonical ninjutsu proof, and the card the keyword is named
// for in every ruling: swing with a 1/1, the defender declines to
// block, and the 1/1 goes home so a 2/2 that draws a card lands in its
// place already attacking and already unblocked.
//
// The second clause is ordinary machinery — the self-only combat
// damage trigger Scroll Thief uses, with "you may" as a real yes/no
// prompt (CR 601.2 is not involved; the prompt exists so an empty
// library is declinable). Everything that is NOT ordinary is the
// keyword, which is why this Spec is one Ninjutsu("{1}{U}") entry:
// see ninjutsu.go.
//
// No simplification. The ninja arrives tapped and attacking the player
// the returned creature was attacking (CR 702.49a), it was never
// DECLARED as an attacker so no "whenever ~ attacks" trigger sees it
// (CR 506.3c), and it is unblocked, so it connects in the combat
// damage step and the draw trigger fires.
func init() {
	Register(Spec{
		OracleID:     "1f3c2b00-0000-4ae1-9650-9553accac52e",
		Name:         "Ninja of the Deep Hours",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Ninjutsu("{1}{U}")},
		Triggered: []game.TriggeredAbility{
			Optional(
				WheneverThisDealsCombatDamageToAPlayer("Ninja of the Deep Hours — draw a card", Do(DrawCards{N: 1})),
				"Ninja of the Deep Hours — draw a card?",
			),
		},
	})
}
