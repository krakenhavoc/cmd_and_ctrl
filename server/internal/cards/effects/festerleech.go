package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Festerleech — Creature — Zombie Leech {B}, 1/1 (Murders at Karlov
// Manor, #4469):
//
//	"Whenever this creature deals combat damage to a player, you mill two
//	 cards.
//	 {1}{B}: This creature gets +2/+2 until end of turn. Activate only
//	 once each turn."
//
// The mill is its controller's, from the combat-damage trigger. The pump
// is Kraven's Cats' row at {1}{B}: CR 602.1b's once-a-turn limit counts
// activations, so a second activation is never offered.
//
// No simplification.
func init() {
	const label = "{1}{B}: This creature gets +2/+2 until end of turn. Activate only once each turn."
	Register(Spec{
		OracleID:     "d14fdc1d-62bc-4e9c-b79b-8c9c3232f0bc",
		Name:         "Festerleech",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Festerleech — you mill two cards", Do(MillCards{N: 2})),
		},
		Activated: []ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      ManaCost("{1}{B}"),
			Condition: OncePerTurnActivation(label),
			Effect:    thisGetsUntilEndOfTurn(2, 2, "Festerleech — +2/+2 until end of turn"),
		}},
	})
}
