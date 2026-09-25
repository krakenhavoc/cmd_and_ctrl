package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// saga.go — the card-facing half of S27's Saga support (CR 714).
// The lifecycle (entry lore counter, the precombat-main advance, the
// CR 704.5s sacrifice) is engine-side and keys off the Saga subtype;
// see server/internal/game/sagas.go. What a card file declares here
// is the chapter ABILITIES.
//
// A chapter ability is an ordinary CR 603 triggered ability with an
// unusual trigger condition, so it is written as an ordinary
// game.TriggeredAbility and goes in Spec.Triggered like any other.
// ChapterTrigger is the constructor that gets the condition right,
// and it is the only supported way to declare one — the `Chapter`
// field it stamps is what the engine reads to learn the card's final
// chapter number, and a hand-rolled TriggeredAbility watching
// EventSagaChapter would leave that at zero. A Saga whose final
// chapter the engine reads as zero is never sacrificed: it sits on
// the battlefield forever, having already done everything it does.
//
//	Triggered: []game.TriggeredAbility{
//	    ChapterTrigger(1, "History of Benalia — create a Knight", makeKnight),
//	    ChapterTrigger(2, "History of Benalia — create a Knight", makeKnight),
//	    ChapterTrigger(3, "History of Benalia — Knights get +2/+1", pumpKnights),
//	},
//
// Chapters that share text ("I, II —") are two entries with the same
// effect, which is how the card is printed and what makes a Saga
// that skips a chapter (a second lore counter from somewhere) fire
// exactly the abilities it crossed.

// ChapterTrigger declares one chapter ability of a Saga: chapter
// number `n`, the stack label the overlay shows, and the effect that
// runs when the ability RESOLVES.
//
// `effect` has the same contract as any triggered ability's: it runs
// at resolution against the live game, and must read everything it
// needs off `item` (Controller, SourceCardID, Targets) rather than
// capturing a card pointer. See AGENTS.md §7 "Adding a triggered
// ability".
//
// For a chapter that targets, set `.Targets` on the returned value —
// the engine picks the target as the chapter goes on the stack
// (CR 603.3d) and re-checks it at resolution (CR 608.2b):
//
//	t := ChapterTrigger(2, "…", effect)
//	t.Targets = TargetCreature("target creature an opponent controls", OpponentControls())
func ChapterTrigger(n int, label string, effect func(g *game.Game, item *game.StackItem) error) game.TriggeredAbility {
	return game.TriggeredAbility{
		Chapter: n,
		Watches: []game.EventKind{game.EventSagaChapter},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.CardID == source.InstanceID && ev.Amount == n
		},
		Key:    label,
		Effect: effect,
	}
}

// ChapterTriggerTargeting is ChapterTrigger for a chapter with a
// target clause — "III — Return target creature or planeswalker card
// from your graveyard to the battlefield".
//
// The engine computes the legal set as the chapter goes on the stack
// (CR 603.3d): an empty set drops the chapter with no prompt, which
// is right — The Eldest Reborn's third chapter with every graveyard
// empty does nothing and still lets the Saga be sacrificed. The
// chosen ref lands in item.Targets[0] and resolution re-checks it
// (CR 608.2b).
func ChapterTriggerTargeting(n int, label string, targets *game.TargetSpec, effect func(g *game.Game, item *game.StackItem) error) game.TriggeredAbility {
	t := ChapterTrigger(n, label, effect)
	t.Targets = targets
	return t
}

// ChapterExileAndReturnTransformed is "Exile this Saga, then return it
// to the battlefield transformed under your control" (ADR 0079's
// second verb, CR 712.14a) — the shared body of every transforming
// Saga's final chapter (Fable of the Mirror-Breaker and the four
// Avatar Legend Sagas). One function instead of one copy per card,
// since the printed sentence never varies.
func ChapterExileAndReturnTransformed(g *game.Game, item *game.StackItem) error {
	return ExileAndReturnTransformed{
		Target:     item.SourceCardID,
		Controller: item.Controller,
	}.Apply(NewContext(g, item))
}

// SagaChapterLabel builds the conventional stack label for a
// chapter: "History of Benalia — I: create a Knight". Card files may
// write their own; this keeps the common case consistent across the
// eight Sagas that shipped together.
func SagaChapterLabel(name string, chapter int, text string) string {
	return name + " — " + romanChapter(chapter) + ": " + text
}

// romanChapter renders a chapter number as the Roman numeral the
// card prints. Sagas top out at IV in every printed set; anything
// beyond falls back to nothing rather than inventing a numeral, and
// the label degrades to "Name — : text" which is ugly and visible
// rather than wrong and silent.
func romanChapter(n int) string {
	switch n {
	case 1:
		return "I"
	case 2:
		return "II"
	case 3:
		return "III"
	case 4:
		return "IV"
	case 5:
		return "V"
	default:
		return ""
	}
}
