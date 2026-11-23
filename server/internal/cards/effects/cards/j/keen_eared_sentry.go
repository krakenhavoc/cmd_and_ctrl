package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keen-Eared Sentry — Creature — Human Soldier {1}{W}, 2/1:
//
//	"You have hexproof.
//	 Each opponent can't venture into the dungeon more than once each
//	 turn."
//
// Player hexproof is Leyline of Sanctity's PlayerKeywords. The venture
// limit is Spec.VentureLimits (game/venture_limit.go, #4100): a second
// venture by an opponent in one turn is not performed, the initiative's
// "venture into Undercity" included (CR 701.49d), and ventures made
// before the Sentry arrived count toward the one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "a5575201-0d73-43c4-9290-940bbaefc61b",
		Name:           "Keen-Eared Sentry",
		Completeness:   effects.CompletenessFull,
		PlayerKeywords: []string{effects.KeywordHexproof},
		VentureLimits: []game.VentureLimitStatic{{
			Label: "Each opponent can't venture into the dungeon more than once each turn.",
			Max:   1,
		}},
	})
}
