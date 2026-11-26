package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Renowned Weaver — Creature — Human Shaman {G}, 1/1 (Journey into Nyx, S82
// OLD, #5558):
//
//	"{1}{G}, Sacrifice this creature: Create a 1/3 green Spider enchantment
//	 creature token with reach. (It can block creatures with flying.)"
//
// The sacrifice is a cost, so the Weaver is gone before the token is made; the
// token is a creature and an enchantment, so it is also an enchantment-enters
// trigger for constellation.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "562684ca-6e3c-47df-9e94-8ff772c25a96",
		Name:         "Renowned Weaver",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}, Sacrifice this creature: Create a 1/3 green Spider enchantment creature token with reach.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1, CreatureTokens: 1},
			Cost:    effects.Plus(effects.ManaCost("{1}{G}"), effects.SacrificeThis()),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/3 green Spider enchantment with reach"), N: 1}),
		}},
	})
}
