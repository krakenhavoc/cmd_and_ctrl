package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Toxrill, the Corrosive — Legendary Creature — Slug Horror {5}{B}{B},
// 7/7 (EDHREC rank 2370):
//
//	"At the beginning of each end step, put a slime counter on each
//	 creature you don't control.
//	 Creatures you don't control get -1/-1 for each slime counter on
//	 them.
//	 Whenever a creature you don't control with a slime counter on it
//	 dies, create a 1/1 black Slug creature token.
//	 {U}{B}, Sacrifice a Slug: Draw a card."
//
// Four abilities, all on machinery main has:
//
//   - EACH end step (any player's), one slime counter on every
//     creature the controller does not control, the set snapshotted
//     before the first counter lands. "Slime" is a free-form counter
//     name; nothing else in the engine reads it.
//   - The -1/-1 is a layer 7c static reading the live counter map,
//     so the creature shrinks the moment the counter lands (a
//     counter placement invalidates the layer cache) and the
//     toughness SBA sweeps it when the counters catch up.
//   - The Slug trigger reads the dead creature's slime count back
//     off the log (b13LastKnownCounters — the card's counters are
//     cleared on the way out), so a creature Toxrill's own -1/-1
//     killed makes a Slug, as printed.
//   - "Sacrifice a Slug" is any permanent with the Slug subtype —
//     the tokens, and Toxrill itself, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c71b2325-bde6-4364-a93b-8477ffeb25d8",
		Name:         "Toxrill, the Corrosive",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{b22SlimeShrink()},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventBeginEndStep},
				AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Kind == game.EventBeginEndStep
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Toxrill — a slime counter on each creature you don't control",
						func(g *game.Game, item *game.StackItem) error {
							return b22PutCounterOnEachCreatureYouDontControl(g, item, "slime")
						})
				},
			},
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b22SlimedCreatureYouDontControlDied(ev, source, g)
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Toxrill — create a 1/1 black Slug",
						func(g *game.Game, item *game.StackItem) error {
							return CreateToken{Controller: item.Controller, Template: b22BlackSlugToken(), N: 1}.Apply(NewContext(g, item))
						})
				},
			},
		},
		Activated: []ActivatedAbility{{
			Label: "{U}{B}, Sacrifice a Slug: Draw a card.",
			Cost:  Plus(ManaCost("{U}{B}"), game.AbilityCost{SacrificeOther: sacrificeSpec("a Slug", Subtype("Slug"))}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
