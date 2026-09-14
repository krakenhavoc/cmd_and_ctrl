package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lavaspur Boots — Artifact — Equipment for {1} (EDHREC rank 1506):
//
//	"Equipped creature gets +1/+0 and has haste and ward {1}.
//	 Equip {1}"
//
// The modern Lightning Greaves: the same "cast a commander and swing
// with it" job, trading the Greaves' shroud for a ward the controller
// can play through. One mana to cast and one to equip.
//
// THE FIRST GRANTED WARD in the catalog, and it is not a keyword
// grant. CR 702.21a makes ward a TRIGGERED ability, so ward.go models
// it as one and canonicalKeywords deliberately has no "ward" token —
// a bare token has nowhere to put the cost. Granting it therefore
// means the EQUIPMENT carries the ward trigger, watching for its HOST
// becoming the target of an opponent's spell; WardAttached is that
// composition and it reuses the same wardPayOrCounter resolution
// Hulking Raptor's own ward uses.
//
// The practical difference from the Greaves is the one the card is
// designed around: ward taxes your opponents, shroud locks YOU out
// too. Boots on a commander still lets you target it with your own
// Aura.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "63bfb9ca-d9d5-4c17-be39-82eb115bc20c",
		Name:         "Lavaspur Boots",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(1, 0),
			GrantToAttached("haste"),
		},
		Triggered: []game.TriggeredAbility{
			WardAttached(WardMana("{1}"), "Lavaspur Boots — ward {1}"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{1}"),
		},
	})
}
