package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
const EraOfEnlightenmentOracleID = "a6d646e8-0d84-489b-a46a-3bd5b1141864"

func init() {
	const name = "Era of Enlightenment"
	effects.Register(effects.Spec{
		OracleID:     EraOfEnlightenmentOracleID,
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 2}, // chapter I's rider is the card's own (#4978)
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "scry 2"), effects.Do(effects.Scry{N: 2})),
			effects.TriggerWithPurpose(effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "you gain 2 life"),
				effects.Do(effects.GainLife{Amount: 2})), game.Purpose{LifeGain: 2}),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "exile it, then return it transformed"),
				effects.ChapterExileAndReturnTransformed),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        EraOfEnlightenmentOracleID + "#1",
		Name:            "Hand of Enlightenment",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
	})
}
