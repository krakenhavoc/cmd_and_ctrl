package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Balloon Peddler — Creature — Human Spellshaper {2}{U}, 2/2 (Mercadian
// Masques, S82 OLD b681, #5619):
//
//	"{U}, {T}, Discard a card: Target creature gains flying until end of
//	 turn."
//
// Boros Guildmage's grant with a Spellshaper cost (mana, tap and a discard).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "10f182a7-ffa3-4128-8b2e-e05b97f92e00",
		Name:         "Balloon Peddler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}, Discard a card: Target creature gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying"))},
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost(), effects.DiscardACard()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("flying", "Balloon Peddler — flying until end of turn"),
		}},
	})
}
