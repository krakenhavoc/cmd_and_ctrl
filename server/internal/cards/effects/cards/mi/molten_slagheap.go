package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Molten Slagheap — Land (Starter Commander Decks):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Put a storage counter on this land.
//	 {1}, Remove X storage counters from this land: Add X mana in any
//	 combination of {B} and/or {R}."
//
// Crucible of the Spirit Dragon's storage shape: the counter is a plain
// activated ability, and the payout is a mana ability whose cost is {1} plus
// the removal of X counters (no {T}, so the land can be tapped for its {C}
// as well). X may be 0. The mana is X independent picks of {B} or {R}, X the
// counters actually removed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "33587cb2-0fd3-4e4c-bc5e-e7299cc9dab5",
		Name:         "Molten Slagheap",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost: effects.ManaAbilityCost{
					Mana:           "{1}",
					RemoveCounters: effects.RemoveCountersXFromThis(game.CounterStorage, 0).RemoveCounters,
				},
				ProducedForPaid: moltenSlagheapMana,
				Label:           "{1}, Remove X storage counters: Add X mana in any combination of {B} and/or {R}",
			},
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Put a storage counter on this land.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Effect:  effects.PutStorageCounterOnThis,
		}},
	})
}

// moltenSlagheapMana is X mana, each {B} or {R}, X the counters removed to pay.
func moltenSlagheapMana(_ *game.Game, _, _ uuid.UUID, paid game.PaidCost) game.ManaOutput {
	return game.ManaOutput{Amount: game.ManaAmount{Options: []string{"B", "R"}, N: paid.CountersRemoved}}
}
