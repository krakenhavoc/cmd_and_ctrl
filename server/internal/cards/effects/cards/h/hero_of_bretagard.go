package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hero of Bretagard — Creature — Human Warrior {2}{W}, 1/1 (Kaldheim
// Commander, #5506):
//
//	"Whenever one or more cards are put into exile from your hand or a spell
//	 or ability you control exiles one or more permanents from the battlefield,
//	 put that many +1/+1 counters on this creature.
//	 As long as this creature has five or more counters on it, it has flying
//	 and is an Angel in addition to its other types.
//	 As long as this creature has ten or more counters on it, it has
//	 indestructible and is a God in addition to its other types."
//
// The trigger is Ranar the Ever-Watchful's: once for each batch of exiles
// (CR 603.2c), whoever or whatever exiled the cards from your hand and a
// spell or ability YOU control for permanents (a cost is not that ability).
// "That many" is read as the trigger resolves, off the batch: the distinct
// cards and permanents it exiled. The counters go on this creature if it is
// still the permanent that triggered. The two thresholds count counters of
// every kind and switch the keyword (layer 6) and the type (layer 4) on and
// off with the count.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "83f1408e-8270-49b4-9d3f-eacfd3645543",
		Name:         "Hero of Bretagard",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			OncePerBatch: true,
			Watches:      []game.EventKind{game.EventZoneMove, game.EventDiscardCard},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.S82OldB075ExiledForController(ev, source.Controller, g)
			},
			Key: "Hero of Bretagard — put that many +1/+1 counters on this creature",
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil || effects.SourceIsNewObject(g, item) || !effects.OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				n := effects.S82OldB075ExiledInBatch(g, item.Trigger.Event, item.Controller)
				if n <= 0 {
					return nil
				}
				return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: n}.Apply(effects.NewContext(g, item))
			},
		}},
		Static: []game.StaticAbility{
			heroOfBretagardType("Angel", 5, "As long as this creature has five or more counters on it, it is an Angel in addition to its other types."),
			heroOfBretagardKeyword("flying", 5, "As long as this creature has five or more counters on it, it has flying."),
			heroOfBretagardType("God", 10, "As long as this creature has ten or more counters on it, it is a God in addition to its other types."),
			heroOfBretagardKeyword("indestructible", 10, "As long as this creature has ten or more counters on it, it has indestructible."),
		},
	})
}

// heroOfBretagardType adds a creature type (layer 4) while the permanent has
// at least n counters.
func heroOfBretagardType(subtype string, n int, label string) game.StaticAbility {
	return game.StaticAbility{
		Label:     label,
		Layer:     game.Layer4Type,
		AppliesTo: effects.S82OldB075CountersAtLeast(n),
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			if !effects.HasFold(c.Subtypes, subtype) {
				c.Subtypes = append(c.Subtypes, subtype)
			}
		},
	}
}

// heroOfBretagardKeyword grants a keyword (layer 6) while the permanent has
// at least n counters.
func heroOfBretagardKeyword(keyword string, n int, label string) game.StaticAbility {
	s := effects.KeywordGrant(effects.S82OldB075CountersAtLeast(n), keyword)
	s.Label = label
	return s
}
