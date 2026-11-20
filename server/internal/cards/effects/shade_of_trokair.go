package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shade of Trokair — Creature — Shade {3}{W}, 1/2 (Planar Chaos, #5595):
//
//	"{W}: This creature gets +1/+1 until end of turn.
//	 Suspend 3—{W} (Rather than cast this card from your hand, you may pay {W}
//	 and exile it with three time counters on it. At the beginning of your
//	 upkeep, remove a time counter. When the last is removed, you may cast it
//	 without paying its mana cost. It has haste.)"
//
// The shade pump is thisGetsUntilEndOfTurn on its own source; suspend is
// the Suspend special action (CR 702.62a), which gives the creature cast
// from it haste.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "61834b70-1b3e-49e6-aaf4-ee7d72155e51",
		Name:         "Shade of Trokair",
		Completeness: CompletenessFull,
		SpecialActions: []game.SpecialAction{
			Suspend(3, "{W}"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{W}: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    ManaCost("{W}"),
			Effect:  thisGetsUntilEndOfTurn(1, 1, "Shade of Trokair — +1/+1 until end of turn"),
		}},
	})
}
