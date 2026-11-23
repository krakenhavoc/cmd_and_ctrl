package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Termagant Swarm — Creature — Tyranid {X}{G}, 0/0:
//
//	"Ravenous (This creature enters with X +1/+1 counters on it. If X is 5 or
//	 more, draw a card when it enters.)
//	 Death Frenzy — When this creature dies, create a number of 1/1 green
//	 Tyranid creature tokens equal to this creature's power."
//
// The power is the last-known power it had on the battlefield (CR 608.2h).
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "e7d8b8ea-07d7-4b6c-aa49-6a2ad185a304",
		Name:                       "Termagant Swarm",
		Completeness:               effects.CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: effects.RavenousEntry(),
		Triggered: []game.TriggeredAbility{
			effects.RavenousDraw("Termagant Swarm"),
			effects.WhenThisDies("Termagant Swarm — create a number of 1/1 green Tyranid creature tokens equal to its power",
				func(g *game.Game, item *game.StackItem) error {
					info, ok := g.LastKnownPermanentForEffect(item.SourceCardID)
					if !ok || info.Power <= 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.TokenCard("1/1 green Tyranid"), N: info.Power}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
