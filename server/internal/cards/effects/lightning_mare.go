package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lightning Mare — Creature — Elemental Horse {R}{R}, 3/1
// (Core Set 2019, #5527):
//
//	"This spell can't be countered.
//	This creature can't be blocked by blue creatures.
//	{1}{R}: This creature gets +1/+0 until end of turn."
//
// Spec.CantBeCountered (Carnage Tyrant's gate), the block rule Dauthi
// Horror uses for blue creatures, and a firebreathing pump on itself
// (thisGetsUntilEndOfTurn).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7c7948c6-5144-4df0-8e33-117710ae00af",
		Name:            "Lightning Mare",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), OfColor("U"), "blue creatures"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{1}{R}"),
			Effect:  thisGetsUntilEndOfTurn(1, 0, "Lightning Mare — +1/+0"),
		}},
	})
}
