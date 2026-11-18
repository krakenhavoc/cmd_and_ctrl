package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
const theModernAgeOracleID = "a8995df6-99a7-41a7-bd3e-f34643b27947"

func init() {
	const name = "The Modern Age"
	loot := func(g *game.Game, item *game.StackItem) error { return lootOne(g, item, 1) }
	Register(Spec{
		OracleID:     theModernAgeOracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(ChapterTrigger(1, SagaChapterLabel(name, 1, "draw a card, then discard a card"), loot),
				game.Purpose{Draws: 1, Discards: 1}),
			TriggerWithPurpose(ChapterTrigger(2, SagaChapterLabel(name, 2, "draw a card, then discard a card"), loot),
				game.Purpose{Draws: 1, Discards: 1}),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "exile it, then return it transformed"),
				ChapterExileAndReturnTransformed),
		},
	})
	Register(Spec{
		OracleID:        theModernAgeOracleID + "#1",
		Name:            "Vector Glider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
