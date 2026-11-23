package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	loot := func(g *game.Game, item *game.StackItem) error { return effects.B16DrawThenDiscard(g, item, 2) }
	effects.Register(effects.Spec{
		OracleID:     "bc475376-54eb-4846-9cfc-5fa2e7908905",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "draw two cards, then discard a card"), loot),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "draw two cards, then discard a card"), loot),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel(name, 3, "shuffle any number of target cards from your graveyard into your library; add {U}{U}"),
				effects.TargetCardInGraveyard("any number of target cards from your graveyard", effects.YouOwn()).WithCount(0, 0),
				func(g *game.Game, item *game.StackItem) error {
					if err := effects.R38ShuffleTargetGraveyardCardsIntoLibrary(g, item); err != nil {
						return err
					}
					return effects.AddMana{Produced: "{U}{U}"}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
