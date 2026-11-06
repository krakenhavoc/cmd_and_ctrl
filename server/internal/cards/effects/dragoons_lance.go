package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragoon's Lance — Artifact — Equipment {1}{W} (Final Fantasy, #4462):
//
//	"Job select (When this Equipment enters, create a 1/1 colorless Hero
//	 creature token, then attach this to it.)
//	 Equipped creature gets +1/+0 and is a Knight in addition to its
//	 other types.
//	 During your turn, equipped creature has flying.
//	 Gae Bolg — Equip {4}"
//
// Paladin's Arms' job-select shape (CR 702.182a): the shared JobSelect
// trigger, a layer 7c pump and a layer 4 creature-type addition. The
// flying is Hook Swords' turn-gated grant (GrantToAttachedWhile with
// duringSourceControllersTurn), re-read on every recompute: "your" is
// the Equipment's controller (CR 301.5d), so the creature flies on that
// player's turn only. "Gae Bolg" is a flavor word with no rules meaning;
// the ability is an ordinary equip (CR 702.6a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a66c35fc-d5db-4cfd-9b85-d70b001d83d3",
		Name:         "Dragoon's Lance",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(1, 0),
			AttachedIsAlso("Knight"),
			GrantToAttachedWhile(duringSourceControllersTurn, "flying"),
		},
		Triggered: []game.TriggeredAbility{JobSelect("Dragoon's Lance")},
		Activated: []ActivatedAbility{EquipAbility("{4}")},
	})
}
