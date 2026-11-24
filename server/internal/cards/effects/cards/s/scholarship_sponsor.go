package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scholarship Sponsor — Creature — Human Advisor {3}{W}, 3/3 (Commander 2021):
//
//	"When this creature enters, each player who controls fewer lands than the
//	 player who controls the most lands searches their library for a number
//	 of basic land cards less than or equal to the difference, puts those
//	 cards onto the battlefield tapped, then shuffles."
//
// Every land count is read once, as the ability resolves, before any
// search (CR 608.2h); the difference is the most lands any player controls
// less the searcher's own. A player level with the leader does not search
// and does not shuffle. Each search is that player's own, "up to" the
// difference, the lands entering tapped; a player who has left the game
// does not take part (CR 800.4a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "21e3895b-f6b6-48af-9161-71b422818004",
		Name:         "Scholarship Sponsor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Scholarship Sponsor — each player with fewer lands than the most searches for basic lands up to the difference, tapped",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					lands := map[uuid.UUID]int{}
					most := 0
					for _, p := range g.Seats {
						if p == nil || p.Eliminated {
							continue
						}
						n := effects.CountControlled(g, p.ID, func(c game.Card) bool { return c.IsLand() })
						lands[p.ID] = n
						most = max(most, n)
					}
					for _, p := range g.Seats {
						if p == nil || p.Eliminated || most-lands[p.ID] <= 0 {
							continue
						}
						if err := (effects.SearchLibrary{
							Player:        p.ID,
							Predicate:     effects.IsBasicLand,
							Dest:          game.ZoneBattlefield,
							Limit:         most - lands[p.ID],
							TappedOnEntry: true,
							Shuffle:       true,
							Reason:        "Scholarship Sponsor — search for basic lands up to the difference",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
