package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Antiquities War — Enchantment — Saga {3}{U} (Dominaria):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I, II — Look at the top five cards of your library. You may reveal an
//	 artifact card from among them and put it into your hand. Put the rest on
//	 the bottom of your library in a random order.
//	 III — Artifacts you control become artifact creatures with base power and
//	 toughness 5/5 until end of turn."
//
// I and II are Belisarius Cawl's look-and-take over five cards. III is Tezzeret
// the Seeker's -5: the artifacts you control as the chapter resolves are locked
// (CR 611.2c) and become creatures with base 5/5 in layer 7b, so counters and
// pumps still apply; one that arrives later this turn is not animated.
//
// No simplification.
func init() {
	const name = "The Antiquities War"
	look := effects.LookAtTopThenMayTakeToHand(5, effects.Artifact(), 1, name+" — you may reveal an artifact card and put it into your hand")
	effects.Register(effects.Spec{
		OracleID:     "ab5a9749-5c94-43e0-b9df-bd2c3601936f",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "look at the top five cards of your library; you may reveal an artifact card from among them and put it into your hand"), look),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "look at the top five cards of your library; you may reveal an artifact card from among them and put it into your hand"), look),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "artifacts you control become artifact creatures with base power and toughness 5/5 until end of turn"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.ScopedEffectFor{
						Match: effects.And(effects.Artifact(), effects.YouControl()),
						Mods: []game.Mod{
							game.AddTypesMod("Artifact", "Creature"),
							game.SetBasePowerMod(5),
							game.SetBaseToughnessMod(5),
						},
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    name + " — artifact creatures with base power and toughness 5/5",
					}.Apply(ctx)
				}),
		},
	})
}
