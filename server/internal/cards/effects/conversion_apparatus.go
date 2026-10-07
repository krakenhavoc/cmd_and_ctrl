package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "931d3dcb-4bbb-4f1a-95e2-7e315faf4158",
		Name:         "Conversion Apparatus",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			colorlessTapRow(),
			{
				Cost:     ManaAbilityCost{Tap: true, Energy: 3},
				Produced: AnyCombinationOfColors(3),
				Label:    "{T}, Pay {E}{E}{E}: Add three mana in any combination of colors.",
			},
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}: You get {E}{E}{E}.",
			Cost:    Plus(ManaCost("{3}"), TapCost()),
			Purpose: game.Purpose{Energy: 3},
			Effect:  Do(GetEnergy{N: 3}),
		}},
	})
}
