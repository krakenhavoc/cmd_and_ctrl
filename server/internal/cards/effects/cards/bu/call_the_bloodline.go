package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Call the Bloodline — Enchantment {1}{B} (Shadows over Innistrad, #5543):
//
//	"{1}, Discard a card: Create a 1/1 black Vampire Knight creature token
//	 with lifelink. Activate only once each turn."
//
// The discard is a cost of the activated ability, paid with the mana (CR
// 602.2b); "only once each turn" is an activation instruction (CR 602.1b),
// so a second activation is never offered in the same turn, and the token
// is made even if the enchantment leaves in response.
//
// No simplification.
func init() {
	const label = "{1}, Discard a card: Create a 1/1 black Vampire Knight creature token with lifelink. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "8e90300b-c0da-4484-b065-e2fafa4e1b15",
		Name:         "Call the Bloodline",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Cost:      effects.Plus(effects.ManaCost("{1}"), effects.DiscardACard()),
			Condition: effects.OncePerTurnActivation(label),
			Purpose:   game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1, CreatureTokens: 1},
			Effect:    effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 black Vampire Knight with lifelink"), N: 1}),
		}},
	})
}
