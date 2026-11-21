package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Whirler Rogue — Creature — Human Rogue Artificer {2}{U}{U}, 2/2:
//
//	"When this creature enters, create two 1/1 colorless Thopter artifact
//	 creature tokens with flying.
//	 Tap two untapped artifacts you control: Target creature can't be
//	 blocked this turn."
//
// The entry trigger is Broadcast Rambler's Thopter, twice. The cost is
// Adaptive Gemguard's TapOthersCost over artifacts only: it is not the {T}
// symbol (CR 302.6), so a Thopter that just arrived may pay it, the Rogue
// itself pays only if it is an artifact (it is not), and nothing in the
// cost targets. The grant is Elvenking's Harper's.
//
// No simplification.
func init() {
	const tapTwo = "two untapped artifacts you control"
	effects.Register(effects.Spec{
		OracleID:     "7060f2c8-fca0-4b7a-bddf-37682f434596",
		Name:         "Whirler Rogue",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Whirler Rogue — create two 1/1 colorless Thopter artifact creature tokens with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap two untapped artifacts you control: Target creature can't be blocked this turn.",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  2,
				Filter: effects.TargetPermanent(tapTwo, effects.Artifact()),
				Label:  tapTwo,
			}},
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TargetCantBeBlockedThisTurn("Whirler Rogue"),
		}},
	})
}
