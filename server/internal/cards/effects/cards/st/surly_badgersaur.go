package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Surly Badgersaur — Creature — Badger Dinosaur {3}{R}, 3/3 (EDHREC
// rank 3180):
//
//	"Whenever you discard a creature card, put a +1/+1 counter on
//	 this creature.
//	 Whenever you discard a land card, create a Treasure token. (It's
//	 an artifact with "{T}, Sacrifice this token: Add one mana of
//	 any color.")
//	 Whenever you discard a noncreature, nonland card, this creature
//	 fights up to one target creature you don't control."
//
// The discard deck's three-way payoff. Three triggers on the same
// event, split by the discarded card's printed type line, read from
// where the card landed; each discard fires exactly one of them,
// once per card, so a wheel that pitches five cards fires five
// times. The fight is a targeted trigger — "up to one", so the
// controller may decline it when the Badgersaur would lose — and
// with no creature an opponent controls the trigger is removed with
// no prompt (CR 603.3d), which is what "up to one" with nothing to
// fight means. The Treasure is the shared template.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0209dc74-ac49-4deb-907a-e9fa49d27a0f",
		Name:         "Surly Badgersaur",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B30YouDiscardedCardWhere(ev, source, g, func(c game.Card) bool { return c.IsCreature() })
			}, "Surly Badgersaur — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B30YouDiscardedCardWhere(ev, source, g, func(c game.Card) bool { return c.IsLand() })
			}, "Surly Badgersaur — create a Treasure", effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
			{
				Watches: []game.EventKind{game.EventDiscardCard},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.B30YouDiscardedCardWhere(ev, source, g, func(c game.Card) bool { return !c.IsCreature() && !c.IsLand() })
				},
				Targets: effects.TargetCreature("up to one target creature you don't control", effects.OpponentControls()).WithCount(0, 1),
				Key:     "Surly Badgersaur — fights up to one target creature you don't control",
				Effect:  effects.B30SourceFightsFirstLegalTarget,
			},
		},
	})
}
