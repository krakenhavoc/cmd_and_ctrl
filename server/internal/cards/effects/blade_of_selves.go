package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blade of Selves — Artifact — Equipment {2} (EDHREC rank 892):
//
//	"Equipped creature has myriad. (Whenever it attacks, for each
//	 opponent other than defending player, you may create a token copy
//	 that's tapped and attacking that player or a planeswalker they
//	 control. Exile the tokens at end of combat.)
//	 Equip {4}"
//
// Legion Loyalty's myriad grant (myriad.go, CR 702.116) narrowed to
// the equipped creature: the trigger is the creature's own, granted in
// layer 6 while the Blade is attached, so moving the Blade or removing
// the creature's abilities takes myriad away. Each opponent other than
// the defending player is its own "you may", and the token copies are
// exiled at end of combat.
//
// No simplification.
func init() {
	const grant = "blade-of-selves/myriad"
	Register(Spec{
		OracleID:     "9c329f3d-4d0a-4012-8aa6-92afdd120a52",
		Name:         "Blade of Selves",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       grant,
			Triggered: []game.TriggeredAbility{GrantedMyriad("Blade of Selves")},
			Text:      "Myriad",
		}},
		Static: []game.StaticAbility{
			Labelled(GrantAbilitiesToAttached(grant), "Equipped creature has myriad."),
		},
		Activated: []ActivatedAbility{EquipAbility("{4}")},
	})
}
