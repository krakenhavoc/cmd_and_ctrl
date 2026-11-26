package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bludgeon Brawl — Enchantment {2}{R} (New Phyrexia, S82 OLD batch 463,
// #5575):
//
//	"Each noncreature, non-Equipment artifact is an Equipment with equip
//	 {X} and "Equipped creature gets +X/+0," where X is that artifact's
//	 mana value."
//
// The #6266 seam (artifacts_become_equipment_mv.go), wired as Dan Lewis is
// but over every player's artifacts: a layer-4 Equipment subtype, and two
// granted abilities whose equip cost and +X/+0 read each artifact's own
// mana value (CR 702.6a, CR 202.3), continuously. Artifact creatures and
// artifacts that already are Equipment are left alone.
//
// No simplification.
func init() {
	const (
		pumpGrant  = "bludgeon-brawl/artifact-pump"
		equipGrant = "bludgeon-brawl/artifact-equip"
	)
	effects.Register(effects.Spec{
		OracleID:     "a9aa3bc7-db5a-4792-a53a-40acb6ca4b20",
		Name:         "Bludgeon Brawl",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{
			{Key: pumpGrant, Static: []game.StaticAbility{effects.PumpAttachedBySourceManaValue("Equipped creature gets +X/+0, where X is this artifact's mana value.")},
				Text: "Equipped creature gets +X/+0."},
			{Key: equipGrant, Activated: []effects.ActivatedAbility{effects.EquipForItsManaValue()}, Text: "Equip {X}"},
		},
		Static: []game.StaticAbility{
			effects.AreAlsoEquipment("Each noncreature, non-Equipment artifact is an Equipment.", effects.NoncreatureNonEquipmentArtifacts),
			effects.Labelled(effects.GrantAbilities(effects.NoncreatureNonEquipmentArtifactsGrantSet, pumpGrant, equipGrant),
				"Each noncreature, non-Equipment artifact has equip {X} and \"Equipped creature gets +X/+0.\""),
		},
	})
}
