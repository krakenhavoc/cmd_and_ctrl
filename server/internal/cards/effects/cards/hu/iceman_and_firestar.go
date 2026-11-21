package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Iceman and Firestar — {2}{U/R}{U/R} Legendary Creature — Mutant Hero
// 3/3 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Whenever you cast a blue spell, tap up to one target creature.
//	 Whenever you cast a red spell, you may discard a card. If you do, draw
//	 a card."
//
// Two cast triggers, one per colour; a blue-and-red spell triggers both.
// The tap trigger's "up to one" lets it go on the stack with no target. The
// loot is the shared discard-then-draw prompt: the draw only follows a
// discard that happened.
//
// No simplification.
func init() {
	tap := effects.WheneverYouCast(effects.OfColor("U"), "Iceman and Firestar — tap up to one target creature", effects.Msc13TapTheLegalTargets)
	tap.Targets = effects.UpToOneTargetCreature("up to one target creature")
	const lootLabel = "Iceman and Firestar — you may discard a card; if you do, draw a card"
	loot := effects.WheneverYouCast(effects.OfColor("R"), lootLabel, func(g *game.Game, item *game.StackItem) error {
		return effects.B39MayDiscardThenDraw(1, false, lootLabel, effects.DrawOneIfDiscarded)(effects.NewContext(g, item))
	})
	effects.Register(effects.Spec{
		OracleID:        "f01c779a-97be-4727-ab62-33579f52447a",
		Name:            "Iceman and Firestar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{tap, loot},
	})
}
