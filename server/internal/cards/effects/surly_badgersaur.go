package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "0209dc74-ac49-4deb-907a-e9fa49d27a0f",
		Name:         "Surly Badgersaur",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b30YouDiscardedCardWhere(ev, source, g, func(c game.Card) bool { return c.IsCreature() })
			}, "Surly Badgersaur — put a +1/+1 counter on it", b30PutCounterOnSelf),
			On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b30YouDiscardedCardWhere(ev, source, g, func(c game.Card) bool { return c.IsLand() })
			}, "Surly Badgersaur — create a Treasure", Do(CreateToken{Template: TreasureToken(), N: 1})),
			{
				Watches: []game.EventKind{game.EventDiscardCard},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b30YouDiscardedCardWhere(ev, source, g, func(c game.Card) bool { return !c.IsCreature() && !c.IsLand() })
				},
				Targets: TargetCreature("up to one target creature you don't control", OpponentControls()).WithCount(0, 1),
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Surly Badgersaur — fights up to one target creature you don't control", b30SourceFightsFirstLegalTarget)
				},
			},
		},
	})
}
