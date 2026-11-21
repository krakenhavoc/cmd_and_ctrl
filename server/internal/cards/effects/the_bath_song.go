package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Bath Song — Enchantment — Saga {3}{U} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I, II — Draw two cards, then discard a card.
//	 III — Shuffle any number of target cards from your graveyard into your
//	 library. Add {U}{U}."
//
// I and II are the shared loot body, two registered chapter triggers (the
// printed "I, II" is one line for two chapters). III is Perpetual Timepiece's
// "any number of target cards" shuffle, which shuffles once even when nothing
// was named, then adds the mana. The mana is added as the chapter resolves, so
// it empties with the step like any other mana (CR 106.4).
//
// No simplification.
func init() {
	const name = "The Bath Song"
	loot := func(g *game.Game, item *game.StackItem) error { return b16DrawThenDiscard(g, item, 2) }
	Register(Spec{
		OracleID:     "bc475376-54eb-4846-9cfc-5fa2e7908905",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "draw two cards, then discard a card"), loot),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "draw two cards, then discard a card"), loot),
			ChapterTriggerTargeting(3, SagaChapterLabel(name, 3, "shuffle any number of target cards from your graveyard into your library; add {U}{U}"),
				TargetCardInGraveyard("any number of target cards from your graveyard", YouOwn()).WithCount(0, 0),
				func(g *game.Game, item *game.StackItem) error {
					if err := shuffleTargetGraveyardCardsIntoLibrary(g, item); err != nil {
						return err
					}
					return AddMana{Produced: "{U}{U}"}.Apply(NewContext(g, item))
				}),
		},
	})
}
