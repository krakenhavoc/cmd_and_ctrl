package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Broodlord — Creature — Tyranid {X}{3}{G}, 3/3:
//
//	"Ravenous (This creature enters with X +1/+1 counters on it. If X is 5 or
//	 more, draw a card when it enters.)
//	 Brood Telepathy — When this creature enters, distribute X +1/+1 counters
//	 among any number of other target creatures you control."
//
// Ravenous is Mawloc's two halves. The distribution is announced with the
// targets (CR 601.2d, 603.3d): X is the X announced for the spell, read off the
// Broodlord as the trigger is put on the stack, and each chosen creature gets at
// least one counter. Zero targets is legal and puts none.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "e3a73311-25f7-4bd7-af1d-ab7b83ad01f2",
		Name:                       "Broodlord",
		Completeness:               effects.CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(game.CounterPlusOne)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.Self(ev, source, lki, g) && source.CastX() >= 5
			}, "Broodlord — ravenous: X was 5 or more, draw a card", effects.Do(effects.DrawCards{N: 1})),
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: effects.Self,
				Targets: effects.Another(effects.TargetCreature("any number of other target creatures you control", effects.YouControl())).
					WithCount(0, 0).Dividing(effects.DivideBy(effects.S82GapB191DivideSourceCastX)),
				Key: "Broodlord — distribute X +1/+1 counters among any number of other target creatures you control",
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.PutDividedCounters(effects.NewContext(g, item), game.CounterPlusOne)
				},
			},
		},
	})
}
