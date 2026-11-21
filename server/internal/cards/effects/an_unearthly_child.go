package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// An Unearthly Child — Enchantment — Saga {1}{U}{U} (Doctor Who):
//
//	"I, II, III — Reveal cards from the top of your library until you
//	 reveal a Doctor card, a card with doctor's companion, or a Vehicle
//	 card. Put that card into your hand and the rest on the bottom of your
//	 library in a random order."
//
// Each chapter is Evolutionary Leap's reveal run (socB30RevealUntil...):
// the whole run is revealed, the first matching card goes to the hand and
// the rest under the library in a random order. "A card with doctor's
// companion" reads gapDoctorsCompanionCards (the keyword is kept nowhere on
// a card).
//
// No simplification.
func init() {
	const name = "An Unearthly Child"
	run := socB30RevealUntilToHandRestBottom(name+" — a Doctor card, a card with doctor's companion, or a Vehicle card",
		func(c game.Card) bool {
			return gapIsDoctorCard(c) || gapHasDoctorsCompanion(c) || c.HasSubtypeWord("Vehicle")
		})
	Register(Spec{
		OracleID:     "9527d2dc-d2e2-4259-aefe-2342ea07bc63",
		Name:         name,
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "reveal until a Doctor, doctor's companion or Vehicle card"), run),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "reveal until a Doctor, doctor's companion or Vehicle card"), run),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "reveal until a Doctor, doctor's companion or Vehicle card"), run),
		},
	})
}
