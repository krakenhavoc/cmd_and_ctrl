package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
//     off the departure record (the card's counters are
//     cleared on the way out), so a creature Toxrill's own -1/-1
//     killed makes a Slug, as printed.
//   - "Sacrifice a Slug" is any permanent with the Slug subtype —
//     the tokens, and Toxrill itself, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c71b2325-bde6-4364-a93b-8477ffeb25d8",
		Name:         "Toxrill, the Corrosive",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.B22SlimeShrink()},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.On(game.EventBeginEndStep, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventBeginEndStep
			}, "Toxrill — a slime counter on each creature you don't control", func(g *game.Game, item *game.StackItem) error {
				return effects.B22PutCounterOnEachCreatureYouDontControl(g, item, "slime")
			}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, OpponentsOnly: true}}),
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B22SlimedCreatureYouDontControlDied(ev, source, g)
			}, "Toxrill — create a 1/1 black Slug", effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 black Slug"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}{B}, Sacrifice a Slug: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{U}{B}"), game.AbilityCost{SacrificeOther: effects.SacrificeSpec("a Slug", effects.Subtype("Slug"))}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
