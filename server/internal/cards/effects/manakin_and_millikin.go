package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Manakin and Millikin — Artifact Creature — Construct {3}, 1/2:
//
//	"{T}, Mill a card: Add {C}{C}."
//
// Millikin's ability for twice the mana, and an ordinary activated
// ability for the same reason: its cost mills a card from a library, so
// CR 605.1a says it is not a mana ability (#3971). It uses the stack and
// the mana is added on resolution. The mill is a real cost component
// (AbilityCost.Mill, #3845, ADR 0020's 2026-10-15 amendment), validated
// with the tap and paid last (CR 601.2h), so an empty library can't pay
// it and the ability is not offered (CR 701.17b). The {T} is on a
// creature, so summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ceda813a-7c15-47a9-a8a3-4724f1dfbf35",
		Name:         "Manakin and Millikin",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Mill a card: Add {C}{C}",
			Cost:    Plus(TapCost(), MillCardsCost(1)),
			Purpose: game.Purpose{Answers: game.AnswerValue}, // ADR 0142: mana answers nothing
			Effect:  Do(AddMana{Produced: "{C}{C}"}),
		}},
	})
}
