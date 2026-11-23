package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:         "f0a20c62-035d-408a-9385-8f2bc2a6d332",
		Name:             "Offalsnout",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flash"},
		AlternativeCosts: []game.AlternativeCost{effects.Evoke("{B}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisLeaves("Offalsnout — exile target card from a graveyard", effects.B27ExileChosenTarget),
				effects.TargetCardInGraveyard("target card from a graveyard")),
		},
	})
}
