package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "d14fdc1d-62bc-4e9c-b79b-8c9c3232f0bc",
		Name:         "Festerleech",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Festerleech — you mill two cards", effects.Do(effects.MillCards{N: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.ManaCost("{1}{B}"),
			Condition: effects.OncePerTurnActivation(label),
			Effect:    effects.ThisGetsUntilEndOfTurn(2, 2, "Festerleech — +2/+2 until end of turn"),
		}},
	})
}
