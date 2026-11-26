package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nettle Drone — Creature — Eldrazi Drone {2}{R}, 3/1 (Battle for Zendikar,
// #5547):
//
//	"Devoid (This card has no color.)
//	 {T}: This creature deals 1 damage to each opponent.
//	 Whenever you cast a colorless spell, untap this creature."
//
// The tap ability needs the creature to have been under its controller's
// control since their turn began (CR 302.6). The cast trigger untaps it even
// in response to its own activation, so it can ping again.
//
// No simplification.
func init() {
	const name = "Nettle Drone"
	effects.Register(effects.Spec{
		OracleID:        "cc4e8b8e-f615-4842-84c5-d6154986b666",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to each opponent.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect:  effects.Do(effects.DamageEachOpponent{N: 1}),
		}},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Colorless(), name+" — untap this creature", effects.B31UntapSelf),
		},
	})
}
