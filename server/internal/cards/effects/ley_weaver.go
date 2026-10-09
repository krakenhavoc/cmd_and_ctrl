package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ley Weaver — Creature — Human Druid {3}{G}, 2/2:
//
//	"Partner with Lore Weaver (When this creature enters, target player
//	 may put Lore Weaver into their hand from their library, then
//	 shuffle.)
//	 {T}: Untap two target lands."
//
// Not legendary, so it can't be a commander: of CR 702.124j's two
// abilities only the entry search does anything (PartnerWith, #2142).
// "Two target lands" is exactly two, any player's, and a land that has
// left by resolution is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "24f1f0e9-8c9b-4f32-95ec-7af883bbeef4",
		Name:         "Ley Weaver",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			PartnerWith("Ley Weaver", "Lore Weaver"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Untap two target lands.",
			Cost:    TapCost(),
			Targets: TargetPermanent("two target lands", Land()).WithCount(2, 2),
			Effect:  UntapEachLegalTarget,
		}},
	})
}
