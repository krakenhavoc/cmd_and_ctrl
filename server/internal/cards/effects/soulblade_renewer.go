package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soulblade Renewer — Creature — Elf Warrior {4}{G}, 2/2 (Battlebond,
// #5529):
//
//	"Partner with Soulblade Corrupter (When this creature enters, target
//	 player may put Soulblade Corrupter into their hand from their
//	 library, then shuffle.)
//	 When this creature enters, support 2. (Put a +1/+1 counter on each of
//	 up to two other target creatures.)"
//
// Not legendary, so of CR 702.124j's two abilities only the entry search
// does anything (PartnerWith). Support is the shared keyword trigger
// (WhenThisEntersSupport, CR 701.41), up to two OTHER target creatures,
// each still legal getting one counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7ccc2740-e659-452d-95e8-5dee958322fa",
		Name:         "Soulblade Renewer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			PartnerWith("Soulblade Renewer", "Soulblade Corrupter"),
			WhenThisEntersSupport("Soulblade Renewer", 2, true),
		},
	})
}
