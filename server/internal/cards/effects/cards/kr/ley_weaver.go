package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "24f1f0e9-8c9b-4f32-95ec-7af883bbeef4",
		Name:         "Ley Weaver",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.PartnerWith("Ley Weaver", "Lore Weaver"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Untap two target lands.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("two target lands", effects.Land()).WithCount(2, 2),
			Effect:  effects.UntapEachLegalTarget,
		}},
	})
}
