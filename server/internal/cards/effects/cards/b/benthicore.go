package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Benthicore — Creature — Elemental {6}{U}, 5/5 (Lorwyn, S82 OLD, #5593):
//
//	"When this creature enters, create two 1/1 blue Merfolk Wizard creature
//	 tokens.
//	 Tap two untapped Merfolk you control: Untap this creature. It gains shroud
//	 until end of turn. (It can't be the target of spells or abilities.)"
//
// Two Merfolk Wizard tokens on entry and Aphetto Grifter's tap-two cost over
// Merfolk. The cost has no {T} symbol, so a Merfolk that arrived this turn can
// pay it (CR 302.6); Benthicore is no Merfolk, so it can't tap itself for it.
// Both halves of the effect need the Benthicore that was activated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "68eb3566-c74d-42a6-89a6-b468677fc485",
		Name:         "Benthicore",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Benthicore — create two 1/1 blue Merfolk Wizard tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 blue Merfolk Wizard"), N: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap two untapped Merfolk you control: Untap this creature. It gains shroud until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.TapTwoUntapped("two untapped Merfolk you control", "Merfolk"),
			Effect:  effects.S82OldB554UntapThisAndItGainsShroud,
		}},
	})
}
