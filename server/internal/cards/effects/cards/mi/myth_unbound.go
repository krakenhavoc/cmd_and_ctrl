package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Myth Unbound — Enchantment {2}{G} (Commander 2018, S82 OLD batch 193,
// #5526):
//
//	"Your commander costs {1} less to cast for each time it's been cast from
//	 the command zone this game.
//	 Whenever your commander is put into the command zone from anywhere, draw
//	 a card."
//
// The reduction is a generic cost reduction (CR 601.2f) on a spell that is
// your commander and that you cast, from any zone, by the number of times THAT
// commander has been cast from the command zone (Player.CommanderCasts, the
// CR 903.8 tally, which only a cast from the command zone raises and which
// follows the card, not the player's other commanders). The tally is read
// as the cast is priced, so the cast that is being announced is not yet
// counted. The draw is a trigger on any card of yours that is a commander
// arriving in the command zone (CR 903.9), from the battlefield, a graveyard,
// exile, a hand or a library; "you draw" is Myth Unbound's controller.
//
// No simplification.
func init() {
	reduction := effects.CostsLessEach(func(q game.CostQuery) int {
		p := q.Game.PlayerByIDForEffect(q.Card.Owner)
		if p == nil {
			return 0
		}
		return p.CommanderCasts[q.Card.InstanceID]
	}, "Your commander costs {1} less to cast for each time it's been cast from the command zone this game.",
		func(q game.CostQuery) bool {
			return q.Card.IsCommander && q.Card.Owner == q.Source.Controller && q.Controller == q.Source.Controller
		})
	effects.Register(effects.Spec{
		OracleID:      "83e047ac-aaa9-4bd8-9ea0-49761d39139c",
		Name:          "Myth Unbound",
		Completeness:  effects.CompletenessFull,
		CostModifiers: []game.CostModifier{reduction},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventZoneMove, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.NewZone != game.ZoneCommand {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCommander && c.Owner == source.Controller
			}, "Myth Unbound — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
