package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Paired Tactician — Creature — Human Warrior {2}{W}, 3/2 (Zendikar Rising,
// S82 OLD #5508):
//
//	"Whenever this creature and at least one other Warrior attack, put a
//	 +1/+1 counter on this creature."
//
// The trigger is checked as attackers are declared (CR 508.1m): the
// Tactician itself attacks, and so does at least one other Warrior (the
// Warrior type read through the layers).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "99262e50-94c1-4796-a7bf-11d7b6164839",
		Name:         "Paired Tactician",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventAttack || ev.CardID != source.InstanceID {
					return false
				}
				for _, id := range effects.B13AttackingCreaturesYouControl(g, source.Controller) {
					if id == source.InstanceID {
						continue
					}
					if c, ok := g.LookupCardForEffect(id); ok && c.HasSubtype(source.TextSubtype("Warrior")) {
						return true
					}
				}
				return false
			}, "Paired Tactician — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
