package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "7060f2c8-fca0-4b7a-bddf-37682f434596",
		Name:         "Whirler Rogue",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Whirler Rogue — create two 1/1 colorless Thopter artifact creature tokens with flying",
				Do(CreateToken{Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: 2})),
		},
		Activated: []ActivatedAbility{{
			Label:   "Tap two untapped artifacts you control: Target creature can't be blocked this turn.",
			Purpose: ForTargets(CantBeBlockedUntilEOT(0, 0, 0)),
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  2,
				Filter: TargetPermanent(tapTwo, Artifact()),
				Label:  tapTwo,
			}},
			Targets: TargetCreature("target creature"),
			Effect:  targetCantBeBlockedThisTurn("Whirler Rogue"),
		}},
	})
}
