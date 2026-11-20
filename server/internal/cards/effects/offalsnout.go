package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Offalsnout — Creature — Elemental {2}{B}, 2/2 (Morningtide, #5592):
//
//	"Flash
//	 When this creature leaves the battlefield, exile target card from a
//	 graveyard.
//	 Evoke {B} (You may cast this spell for its evoke cost. If you do,
//	 it's sacrificed when it enters.)"
//
// Reveillark's leave trigger (any exit) with Lantern of the Lost's exile
// of one target card from any graveyard; with no card in any graveyard
// the trigger is removed (CR 603.3d). Evoke is the alternative cost
// (CR 702.74a), and flash lets it be evoked at instant speed in response
// to a reanimation spell.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "f0a20c62-035d-408a-9385-8f2bc2a6d332",
		Name:             "Offalsnout",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flash"},
		AlternativeCosts: []game.AlternativeCost{Evoke("{B}")},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisLeaves("Offalsnout — exile target card from a graveyard", b27ExileChosenTarget),
				TargetCardInGraveyard("target card from a graveyard")),
		},
	})
}
