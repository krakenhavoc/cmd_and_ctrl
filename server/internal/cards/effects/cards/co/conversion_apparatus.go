package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Conversion Apparatus — Artifact {3}:
//
//	"{T}: Add {C}.
//	 {3}, {T}: You get {E}{E}{E} (three energy counters).
//	 {T}, Pay {E}{E}{E}: Add three mana in any combination of colors."
//
// ADR 0129 §5. The middle ability adds no mana, so it is an ordinary
// CR 602 activation on the stack. The last is a mana ability paying
// three energy (CR 605.1a, CR 107.14), planned by the auto-tapper in its
// energy tier: its three slots are any colour each, so one activation
// can pay a {W}{U}{B}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "931d3dcb-4bbb-4f1a-95e2-7e315faf4158",
		Name:         "Conversion Apparatus",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			effects.ColorlessTapRow(),
			{
				Cost:     effects.ManaAbilityCost{Tap: true, Energy: 3},
				Produced: effects.AnyCombinationOfColors(3),
				Label:    "{T}, Pay {E}{E}{E}: Add three mana in any combination of colors.",
			},
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: You get {E}{E}{E}.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerValue, Energy: 3},
			Effect:  effects.Do(effects.GetEnergy{N: 3}),
		}},
	})
}
