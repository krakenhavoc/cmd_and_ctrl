package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cloudthresher — Creature — Elemental {2}{G}{G}{G}{G}, 7/7 (Commander 2015,
// S82 OLD #5546):
//
//	"Flash
//	 Reach
//	 When this creature enters, it deals 2 damage to each creature with flying
//	 and each player.
//	 Evoke {2}{G}{G} (You may cast this spell for its evoke cost. If you do,
//	 it's sacrificed when it enters.)"
//
// Squallmonger's one damage instance over every flyer and every player, on an
// enters trigger; evoke is Mulldrifter's, the sacrifice a second trigger, so
// the damage is dealt whichever order they stack in. Cloudthresher is the
// source, read as it last existed if the evoke sacrifice got there first.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ab0e2aa0-1ddd-4a4c-b7cd-4063caaa8bcd",
		Name:            "Cloudthresher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "reach"},
		// ADR 0126 §6: only creatures with flying.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, Partial: true}},
		AlternativeCosts: []game.AlternativeCost{
			effects.Evoke("{2}{G}{G}"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Cloudthresher — 2 damage to each creature with flying and each player",
				effects.ThisDealsDamageToEachCreatureMatchingAndEachPlayer(effects.HasKeyword("flying"), 2)),
		},
	})
}
