package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dictate of Erebos — Enchantment {3}{B}{B}:
//
//	"Flash"
//	"Whenever a creature you control dies, each opponent sacrifices a
//	 creature of their choice."
//
// Grave Pact with flash, which is the whole reason to play both: flash
// it in during someone's declare-blockers step and the trade they were
// happy with costs them the rest of their board.
//
// "Each opponent" rather than Grave Pact's "each other player" — the
// same set in a normal game; they diverge only under an effect that
// makes a player not your opponent (Rules-wise, Two-Headed Giant and
// friends), which this sandbox has no notion of. Same machinery either
// way: ExceptController skips you.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7c777a41-e40a-4b40-96bf-8ddd5c12924c",
		Name:            "Dictate of Erebos",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Answers: game.AnswerRemove}, // #4734: an edict when a creature dies, as its trigger declares
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WheneverACreatureYouControlDies("Dictate of Erebos — each opponent sacrifices a creature", effects.Do(effects.EachPlayerSacrifices{
				ExceptController: true,
				Match:            effects.Creature(),
				Label:            "a creature",
			})), game.Purpose{DeathPayoff: true, Answers: game.AnswerRemove}),
		},
	})
}
