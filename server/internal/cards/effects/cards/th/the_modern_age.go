package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Modern Age // Vector Glider — a transforming Saga (Kamigawa: Neon
// Dynasty, #5493, S82).
//
// Front face, Enchantment — Saga {1}{U}:
//
//	"I, II — Draw a card, then discard a card.
//	 III — Exile this Saga, then return it to the battlefield transformed
//	       under your control."
//
// Back face, Enchantment Creature — Spirit, 2/3: "Flying".
//
// The loot is the ordinary draw-then-discard prompt (lootOne): the discard
// is chosen after the draw. Chapter III is Fable of the Mirror-Breaker's.
//
// No simplification.
const TheModernAgeOracleID = "a8995df6-99a7-41a7-bd3e-f34643b27947"

func init() {
	const name = "The Modern Age"
	loot := func(g *game.Game, item *game.StackItem) error { return effects.LootOne(g, item, 1) }
	effects.Register(effects.Spec{
		OracleID:     TheModernAgeOracleID,
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "draw a card, then discard a card"), loot),
				game.Purpose{Draws: 1, Discards: 1}),
			effects.TriggerWithPurpose(effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "draw a card, then discard a card"), loot),
				game.Purpose{Draws: 1, Discards: 1}),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "exile it, then return it transformed"),
				effects.ChapterExileAndReturnTransformed),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        TheModernAgeOracleID + "#1",
		Name:            "Vector Glider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
