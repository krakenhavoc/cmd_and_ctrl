package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Limited Resources — Enchantment {W} (Exodus, S82 OLD batch 744, #5626):
//
//	"When this enchantment enters, each player chooses five lands they control
//	 and sacrifices the rest.
//	 Players can't play lands as long as ten or more lands are on the
//	 battlefield."
//
// The enters trigger is Global Ruin's ordered pick without a rule: each player
// with more than five lands, in turn order, picks the five to keep, then
// everything else of theirs goes at once (CR 701.21a); a player with five or
// fewer is not asked. The ban is a land-play restriction (CR 305.1, CR 101.2)
// that holds on every player, the Enchantment's controller included, while the
// battlefield holds ten or more lands, counted as each play is announced.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5b287dcc-f853-4df1-8ddf-5b8b2adc2873",
		Name:         "Limited Resources",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepSacrifice, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Limited Resources — each player chooses five lands they control and sacrifices the rest", effects.S82OldB744LimitedResources),
		},
		LandPlayRestrictions: []game.LandPlayRestriction{{
			Label: "Players can't play lands as long as ten or more lands are on the battlefield.",
			Forbids: func(q game.LandPlayQuery) bool {
				return effects.S82OldB744TenOrMoreLands(q.Game)
			},
		}},
	})
}
