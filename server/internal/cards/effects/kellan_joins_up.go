package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kellan Joins Up — Legendary Enchantment {G}{W}{U} (Outlaws of Thunder
// Junction, #4468):
//
//	"When Kellan Joins Up enters, you may exile a nonland card with mana
//	 value 3 or less from your hand. If you do, it becomes plotted. (You
//	 may cast it as a sorcery on a later turn without paying its mana
//	 cost.)
//	 Whenever a legendary creature you control enters, put a +1/+1 counter
//	 on each creature you control."
//
// The first ability is the shared "exile from hand, it becomes plotted"
// pick (b35PlotFromHand). The second reads the entering creature through
// the layered view (a legendary creature may be legendary only as it
// enters) and puts a counter on every creature you control, the entering
// one included, snapshotting the set first.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9b3ddd31-c058-4846-8784-25c7a9bbcb31",
		Name:         "Kellan Joins Up",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Kellan Joins Up — you may exile a nonland card with mana value 3 or less from your hand; it becomes plotted",
				b35PlotFromHand("Kellan Joins Up — you may exile a nonland card with mana value 3 or less from your hand; it becomes plotted")),
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && c.IsLegendary()
			}, "Kellan Joins Up — put a +1/+1 counter on each creature you control", b11PutCounterOnEachCreatureYouControl),
		},
	})
}
