package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fumiko the Lowblood — Legendary Creature — Human Samurai {2}{R}{R}, 3/2
// (Commander 2020, S82 OLD, #5513):
//
//	"Fumiko has bushido X, where X is the number of attacking creatures.
//	 (Whenever this creature blocks or becomes blocked, it gets +X/+X until
//	 end of turn.)
//	 Creatures your opponents control attack each combat if able."
//
// Bushido is the shared trigger (CR 702.45a) with X counted as it resolves
// (CR 702.45a: X is the value then), over every attacking creature whoever
// controls it. The second ability is Warmonger Hellkite's attack requirement
// scoped to creatures an opponent of Fumiko's controller controls: a
// requirement of the source (CR 508.1d), so a tapped or summoning-sick
// creature stays home and a creature that loses abilities still has to attack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cba58942-b98c-45aa-8651-7dc666e31c93",
		Name:         "Fumiko the Lowblood",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventBlock, game.EventBecomesBlocked}, effects.ThisBlocksOrBecomesBlocked,
				"Fumiko the Lowblood — bushido X, where X is the number of attacking creatures",
				func(g *game.Game, item *game.StackItem) error {
					x := effects.S82BroB20AttackingCreatures(g)
					return effects.ThisGetsUntilEndOfTurn(x, x, "Fumiko the Lowblood — bushido X")(g, item)
				}),
		},
		Static: []game.StaticAbility{
			effects.AttacksEachCombatWhere(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller != source.Controller
			}),
		},
	})
}
