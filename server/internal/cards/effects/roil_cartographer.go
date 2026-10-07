package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Roil Cartographer — Creature — Merfolk Rogue {1}{U}, 1/3:
//
//	"Landfall — Whenever a land you control enters, you get {E} (an
//	 energy counter).
//	 {T}, Pay six {E}: Draw three cards."
//
// ADR 0129 §2: six energy is a PayEnergy component, checked before the
// {T} is paid.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3cd7ef87-4cb4-42ff-842f-e9a465dab165",
		Name:         "Roil Cartographer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall("Roil Cartographer — you get {E}", Do(GetEnergy{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Pay six {E}: Draw three cards.",
			Cost:    Plus(TapCost(), PayEnergy(6)),
			Purpose: game.Purpose{Draws: 3},
			Effect:  Do(DrawCards{N: 3}),
		}},
	})
}
