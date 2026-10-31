package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	tap := WheneverYouCast(OfColor("U"), "Iceman and Firestar — tap up to one target creature", msc13TapTheLegalTargets)
	tap.Targets = UpToOneTargetCreature("up to one target creature")
	const lootLabel = "Iceman and Firestar — you may discard a card; if you do, draw a card"
	loot := WheneverYouCast(OfColor("R"), lootLabel, func(g *game.Game, item *game.StackItem) error {
		return b39MayDiscardThenDraw(1, false, lootLabel, drawOneIfDiscarded)(NewContext(g, item))
	})
	Register(Spec{
		OracleID:        "f01c779a-97be-4727-ab62-33579f52447a",
		Name:            "Iceman and Firestar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{tap, loot},
	})
}
