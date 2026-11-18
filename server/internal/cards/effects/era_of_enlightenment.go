package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Era of Enlightenment // Hand of Enlightenment — a transforming Saga
// (Kamigawa: Neon Dynasty, #5493, S82).
//
// Front face, Enchantment — Saga {1}{W}:
//
//	"I — Scry 2.
//	 II — You gain 2 life.
//	 III — Exile this Saga, then return it to the battlefield transformed
//	       under your control."
//
// Back face, Enchantment Creature — Human Monk, 2/2: "First strike".
//
// Chapter III is Fable of the Mirror-Breaker's exile-and-return, so the
// Hand is a new object with no lore counters (CR 400.7, 712.14a).
//
// No simplification.
const eraOfEnlightenmentOracleID = "a6d646e8-0d84-489b-a46a-3bd5b1141864"

func init() {
	const name = "Era of Enlightenment"
	Register(Spec{
		OracleID:     eraOfEnlightenmentOracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 2}, // chapter I's rider is the card's own (#4978)
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "scry 2"), Do(Scry{N: 2})),
			TriggerWithPurpose(ChapterTrigger(2, SagaChapterLabel(name, 2, "you gain 2 life"),
				Do(GainLife{Amount: 2})), game.Purpose{LifeGain: 2}),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "exile it, then return it transformed"),
				ChapterExileAndReturnTransformed),
		},
	})
	Register(Spec{
		OracleID:        eraOfEnlightenmentOracleID + "#1",
		Name:            "Hand of Enlightenment",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
	})
}
