package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Harmonized Trio // Brainstorm — Creature — Merfolk Bard Wizard {U}, 1/1
// // Instant {U} (Secrets of Strixhaven, #4456; preparation card, CR 722):
//
//	"{T}, Tap two untapped creatures you control: This creature becomes
//	 prepared. (While it's prepared, you may cast a copy of its spell.
//	 Doing so unprepares it.)"
//
//	Brainstorm — "Draw three cards, then put two cards from your hand on
//	 top of your library in any order."
//
// It does not enter prepared. The cost is {T} plus two OTHER untapped
// creatures you control: the Trio is already being tapped by {T}, so it
// cannot also be one of the two. The {T} is the symbol (CR 302.6), so a
// Trio that arrived this turn cannot activate it; the two others are not
// the symbol, so creatures that arrived this turn may be tapped for it,
// and the cost does not target. Activating it while already prepared is
// legal and does nothing (CR 722.3a). The prepare spell is Brainstorm
// word for word, so it shares that card's body (brainstorm.go).
//
// No simplification.
const HarmonizedTrioOracleID = "f6036065-7196-44d4-8a27-2199f0e1a03c"

func init() {
	effects.Register(effects.Spec{
		OracleID:     HarmonizedTrioOracleID,
		Name:         "Harmonized Trio",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{T}, Tap two untapped creatures you control: This creature becomes prepared.",
			Cost: effects.Plus(effects.TapCost(), game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:         2,
				Filter:        effects.TargetPermanent("two untapped creatures you control", effects.Creature()),
				ExcludeSource: true,
				Label:         "two untapped creatures you control",
			}}),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BecomePrepared{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     HarmonizedTrioOracleID + "#1",
		Name:         "Brainstorm",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.BrainstormOnResolve,
	})
}
