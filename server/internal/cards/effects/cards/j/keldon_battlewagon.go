package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Keldon Battlewagon — Artifact Creature — Juggernaut {5}, 0/3 (Prophecy):
//
//	"Trample
//	 This creature can't block.
//	 When this creature attacks, sacrifice it at end of combat.
//	 Tap an untapped creature you control: This creature gets +X/+0 until end
//	 of turn, where X is the power of the creature tapped this way."
//
// Runaway Carriage's attack trigger over the delayed end-of-combat sacrifice
// (CR 511.2), the "can't block" restriction of Risen Executioner, and
// Impelled Giant's cost and pump: the tap is not the {T} symbol (CR 302.6) so a
// creature that just arrived, the Battlewagon itself included, may pay it,
// and X is that creature's power as the ability resolves (or as it last
// existed, CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4a2df44b-5d36-4c4e-b982-836d9d7e0505",
		Name:            "Keldon Battlewagon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static:          []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclared(ev, source)
			}, "Keldon Battlewagon — sacrifice it at end of combat",
				effects.SacrificeThisAtEndOfCombat("Keldon Battlewagon — sacrifice it")),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap an untapped creature you control: This creature gets +X/+0 until end of turn, where X is the power of the creature tapped this way.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.TapAnUntapped("an untapped creature you control", effects.Creature()),
			Effect:  effects.S82OldB523PumpByTappedPower,
		}},
	})
}
