package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Primordial Hydra — Creature — Hydra {X}{G}{G}, 0/0 (EDHREC rank
// 2447):
//
//	"This creature enters with X +1/+1 counters on it.
//	 At the beginning of your upkeep, double the number of +1/+1
//	 counters on this creature.
//	 This creature has trample as long as it has ten or more +1/+1
//	 counters on it."
//
// The Hydra that doubles every turn: X, then 2X, then 4X, and trample
// once it passes ten. The upkeep trigger puts as many +1/+1 counters
// on the Hydra as it already has (through AddCounter, so Doubling
// Season and Hardened Scales apply to the doubling, as printed); the
// trample is a Layer 6 static gated on the live counter count, and a
// counter placement bumps the layer version so it appears the moment
// the tenth counter lands.
//
// Sandbox simplification, declared, for the X counters — Goldvein
// Hydra's: an entry replacement cannot see the X announced for the
// spell (the stack item is gone by the time the entry pipeline runs),
// so the counters go on as the spell RESOLVES, a beat before the
// card moves from the stack to the battlefield. They are on the card
// when it lands, so the 0/0 body never meets the state-based check
// without them, and a counter doubler applies. The one observable
// difference is that a "whenever you put counters on a permanent"
// payoff does not see them — weaker, never stronger.
//
// Cast for X=0 the Hydra enters as the printed 0/0 it is and the next
// state-based check puts it into its owner's graveyard (CR 704.5f),
// before any upkeep can double zero. CR 601.2b allows the
// announcement; it simply does not survive it. That was an engine gap
// until #691 — the toughness check read every printed 0/0 as the
// importer's stand-in — and what closed it is the printing behind the
// object (game.Card.ToughnessIsKnown).
func init() {
	Register(Spec{
		OracleID:     "1c36ed3a-c806-47e5-83f9-e44999c67fe5",
		Name:         "Primordial Hydra",
		XMatters:     true,
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The X +1/+1 counters are put on the Hydra as the spell resolves, a beat before it enters, so effects that watch you put counters on a permanent don't see them."},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: ctx.X()}.Apply(ctx)
		},
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && target.Counters["+1/+1"] >= 10
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				for _, k := range c.Abilities {
					if k == "trample" {
						return
					}
				}
				c.Abilities = append(c.Abilities, "trample")
			},
		}},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Primordial Hydra — double its +1/+1 counters", func(g *game.Game, item *game.StackItem) error {
				return b23DoublePlusOneCountersOn(NewContext(g, item), item.SourceCardID)
			}),
		},
	})
}
