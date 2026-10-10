package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arcane Lighthouse — Land (EDHREC rank 1043):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Until end of turn, creatures your opponents control lose
//	 hexproof and shroud and can't have hexproof or shroud."
//
// The proof card for #1651's second half (ADR 0038's amendment of
// 2026-09-28, B2), in its turn-scoped form. One cantHaveKeywords data
// record over the opponents' creatures:
//
//   - "Lose" and "can't have" are one record. The engine strips the
//     tokens after the whole layer-6 bucket, so a Heroic Intervention
//     or a Lightning Greaves re-equip afterwards does not give them
//     back this turn (CR 101.2 — "can't" beats "can").
//   - Unlike Detection Tower, this changes characteristics, so CR
//     611.2c locks the set to the creatures the opponents control as
//     the ability resolves. A creature cast afterwards keeps its
//     hexproof.
//   - Ward is not hexproof and is left alone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "30ac68e6-160a-41f9-9f0f-0e0eef383150",
		Name:         "Arcane Lighthouse",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Until end of turn, creatures your opponents control lose hexproof and shroud and can't have hexproof or shroud.",
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return LoseAndCantHaveUntilEOT{
					Match:    And(Creature(), OpponentControls()),
					Keywords: []string{"hexproof", "shroud"},
					Label:    "Arcane Lighthouse — no hexproof or shroud",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
