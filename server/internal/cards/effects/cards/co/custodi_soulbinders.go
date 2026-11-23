package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Custodi Soulbinders — Creature — Human Cleric {3}{W}, 0/0:
//
//	"This creature enters with X +1/+1 counters on it, where X is the
//	 number of other creatures on the battlefield.
//	 {2}{W}, Remove a +1/+1 counter from this creature: Create a 1/1
//	 white Spirit creature token with flying."
//
// The entry counters are a CR 614.1c replacement counted as it enters
// (Boss's Chauffeur's shape), over every other creature on the
// battlefield whoever controls it. The ability is Bladed Ambassador's
// mana-plus-counter cost; paying the last counter leaves a 0/0 that dies
// to the state-based check, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c7a8daff-9913-4970-8765-52ba9d9d01db",
		Name:         "Custodi Soulbinders",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, src *game.Card) int {
				n := 0
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.InstanceID != src.InstanceID && c.IsCreature() {
						n++
					}
				}
				return n
			}, "Custodi Soulbinders: enters with X +1/+1 counters, X the number of other creatures on the battlefield"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}, Remove a +1/+1 counter from this creature: Create a 1/1 white Spirit creature token with flying.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{2}{W}"), effects.RemoveCountersFromThis(game.CounterPlusOne, 1)),
			Effect:  effects.Do(effects.CreateToken{Template: effects.B28WhiteSpiritFlyingToken(), N: 1}),
		}},
	})
}
