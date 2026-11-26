package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Join Shields — Instant {3}{G}{W} (Guilds of Ravnica, #5525):
//
//	"Untap all creatures you control. They gain hexproof and indestructible
//	 until end of turn."
//
// Merciful Mandate's untap, then Aerie Mystics' grant to the creatures the
// caster controls as the spell resolves: the set is fixed then (CR 611.2c), so
// a creature that arrives later in the turn does not get it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1102524d-7287-496f-a18a-5fdae6d4c271",
		Name:         "Join Shields",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.UntapAllCreaturesYouControl{}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl()),
				Keywords: []string{"hexproof", "indestructible"},
				Label:    "Join Shields — hexproof and indestructible",
			}.Apply(ctx)
		},
	})
}
