package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rampage of the Clans — Instant {3}{G} (Ravnica Allegiance):
//
//	"Destroy all artifacts and enchantments. For each permanent destroyed this
//	 way, its controller creates a 3/3 green Centaur creature token."
//
// Flood of Tears' mass-sweep shape: the tail is told the pre-move copies of the
// permanents actually destroyed (an indestructible one is not), and each
// controller gets one Centaur per permanent of theirs that died, all of them
// counted before any token is made.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5dc34e93-97e3-434b-9f12-f45c6d556f13",
		Name:         "Rampage of the Clans",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifactsAndEnchantments, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{
				Match:    effects.Or(effects.Artifact(), effects.Enchantment()),
				ThenTail: ctx.Tail(rampageOfTheClansCentaurs),
			}.Apply(ctx)
		},
	})
}

// rampageOfTheClansCentaurs is "for each permanent destroyed this way, its
// controller creates a 3/3 green Centaur creature token", keyed. An on-disk
// key: never renamed, never reused.
var rampageOfTheClansCentaurs = effects.MassTail("mass/rampage-of-the-clans-centaurs", func(ctx *effects.Context, swept []game.Card, _ int) error {
	var order []uuid.UUID
	count := map[uuid.UUID]int{}
	for _, c := range swept {
		if count[c.Controller] == 0 {
			order = append(order, c.Controller)
		}
		count[c.Controller]++
	}
	for _, who := range order {
		if err := (effects.CreateToken{Controller: who, Template: effects.TokenCard("3/3 green Centaur"), N: count[who]}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
