package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grafted Wargear — Artifact — Equipment {3} (Fifth Dawn, #5604):
//
//	"Equipped creature gets +3/+2.
//	 Whenever this Equipment becomes unattached from a permanent,
//	 sacrifice that permanent.
//	 Equip {0} ({0}: Attach to target creature you control. Equip only
//	 as a sorcery.)"
//
// Hero's Blade's +3/+2 and Grafted Exoskeleton's "becomes unattached"
// sacrifice (the shared GraftedSacrificeOnUnattach row: moved to another
// creature, an illegal host, or the Equipment leaving), with equip {0}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c48cf069-eb20-4a61-b58c-b031ec22a5f0",
		Name:         "Grafted Wargear",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(3, 2)},
		Triggered: []game.TriggeredAbility{
			GraftedSacrificeOnUnattach("Grafted Wargear — sacrifice the permanent it became unattached from"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{0}"),
		},
	})
}
