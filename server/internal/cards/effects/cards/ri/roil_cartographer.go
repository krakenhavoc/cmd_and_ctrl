package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "3cd7ef87-4cb4-42ff-842f-e9a465dab165",
		Name:         "Roil Cartographer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.LandfallPaying("Roil Cartographer — you get {E}", game.LandfallPayoff{Energy: 1}, effects.Do(effects.GetEnergy{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Pay six {E}: Draw three cards.",
			Cost:    effects.Plus(effects.TapCost(), effects.PayEnergy(6)),
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 3},
			Effect:  effects.Do(effects.DrawCards{N: 3}),
		}},
	})
}
