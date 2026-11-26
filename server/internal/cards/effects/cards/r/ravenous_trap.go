package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ravenous Trap — Instant — Trap {2}{B}{B} (Zendikar, S82 OLD b495):
//
//	"If an opponent had three or more cards put into their graveyard from anywhere
//	 this turn, you may pay {0} rather than pay this spell's mana cost.
//	 Exile target player's graveyard."
//
// Whiplash Trap's alternative cost over the turn's graveyard arrivals
// (Old495AnOpponentHadThreeCardsPutIntoTheirGraveyard): one opponent has to have
// three cards put into their own graveyard, from any zone. The spell is Nihil
// Spellbomb's exile of the target player's graveyard as it is at resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e0bea449-7601-48ee-a522-f1bbe6905f6d",
		Name:         "Ravenous Trap",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		AlternativeCosts: []game.AlternativeCost{{
			Key:       "trap",
			Label:     "If an opponent had three or more cards put into their graveyard from anywhere this turn, pay {0}",
			ManaCost:  "{0}",
			Condition: effects.Old495AnOpponentHadThreeCardsPutIntoTheirGraveyard,
		}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileTargetPlayersGraveyard(ctx.Game, item)
		},
	})
}
