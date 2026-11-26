package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ignition Team — Creature — Goblin Warrior {5}{R}{R}, 0/0 (Conspiracy):
//
//	"This creature enters with X +1/+1 counters on it, where X is the number
//	 of tapped lands on the battlefield.
//	 {2}{R}, Remove a +1/+1 counter from this creature: Target land becomes a
//	 4/4 red Elemental creature until end of turn. It's still a land."
//
// X counts every tapped land, anyone's, as the Team enters (CR 614.1c). The
// animation is the land creature of Hissing Quagmire's family (one effect,
// layers 4, 5 and 7b, pinned to the land, CR 611.2c).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9edaadc9-6bbd-4727-9884-1c6b13c778bb",
		Name:         "Ignition Team",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, _ *game.Card) int {
				return effects.S82OldB365TappedLands(g)
			}, "Ignition Team enters with X +1/+1 counters, X the number of tapped lands on the battlefield"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, Remove a +1/+1 counter from this creature: Target land becomes a 4/4 red Elemental creature until end of turn. It's still a land.",
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.RemoveCountersFromThis(game.CounterPlusOne, 1)),
			Targets: effects.TargetPermanent("target land", effects.Land()),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.Cmdr2B13Animate(ctx, id,
					"Ignition Team — a 4/4 red Elemental creature until end of turn", 4, 4, "Elemental", []string{"R"})
			},
		}},
	})
}
