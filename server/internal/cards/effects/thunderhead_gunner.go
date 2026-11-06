package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thunderhead Gunner — Creature — Shark Pirate {4}{R}, 4/5 (Aetherdrift,
// #4464):
//
//	"Reach
//	 Discard a card: Draw a card. Activate only as a sorcery and only once
//	 each turn."
//
// Charging Strifeknight's rummage without the tap, held to sorcery speed
// and to one activation a turn by DOT-7 Repair Squad's pair of
// restrictions (OncePerTurnActivation keyed on the label). The discard is
// the cost, so it is paid before the draw and an empty hand can't
// activate it. The once-a-turn record belongs to this object, so a change
// of control keeps it (CR 602.5b).
//
// No simplification.
func init() {
	const label = "Discard a card: Draw a card. Activate only as a sorcery and only once each turn."
	Register(Spec{
		OracleID:        "d9b6656d-d71f-4359-bf6f-b76089741e29",
		Name:            "Thunderhead Gunner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []ActivatedAbility{{
			Label:        label,
			Cost:         DiscardACard(),
			SorcerySpeed: true,
			Condition:    OncePerTurnActivation(label),
			Purpose:      game.Purpose{Draws: 1, Discards: 1},
			Effect:       Do(DrawCards{N: 1}),
		}},
	})
}
