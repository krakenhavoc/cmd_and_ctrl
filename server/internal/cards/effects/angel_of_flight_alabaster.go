package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Angel of Flight Alabaster — Creature — Angel {4}{W}, 4/4 (Reprint, #5908):
//
//	"Flying
//	 At the beginning of your upkeep, return target Spirit card from your
//	 graveyard to your hand."
//
// Palace Siege's upkeep shape: the target is a Spirit card in the
// controller's graveyard, chosen as the trigger goes on the stack. With no
// Spirit card there, the trigger is removed (CR 603.3d); one that left in
// response makes the ability do nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9b6a74fb-c48f-4d85-b4b3-96eb80b9efc7",
		Name:            "Angel of Flight Alabaster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtYourUpkeep("Angel of Flight Alabaster — return target Spirit card from your graveyard to your hand", palaceSiegeReturn),
				TargetCardInGraveyard("target Spirit card from your graveyard", YouOwn(), Subtype("Spirit"))),
		},
	})
}
