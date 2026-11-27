package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shinen of Fear's Chill — Creature — Spirit {4}{B}, 3/2 (Saviors of
// Kamigawa, S82 OLD b584, #5601):
//
//	"This creature can't block.
//	 Channel — {1}{B}, Discard this card: Target creature can't block this
//	 turn."
//
// Aesthir Glider's static restriction, and Abandon the Post's until-end-of-turn
// one on the channel target; the channel row works only from the hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "72e61b73-980c-4ee4-b529-a367897620ec",
		Name:         "Shinen of Fear's Chill",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
		Activated: []effects.ActivatedAbility{{
			Label:   "Channel — {1}{B}, Discard this card: Target creature can't block this turn.",
			Purpose: effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)),
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TargetCreatureCantBlockThisTurn("Shinen of Fear's Chill — target creature can't block this turn"),
		}},
	})
}
