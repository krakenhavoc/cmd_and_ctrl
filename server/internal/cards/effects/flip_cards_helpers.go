package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// flip_cards_helpers.go — the pieces the follow-on flip cards of ADR
// 0151 (#3216) share. The verb is FlipThis (flip.go); this file is the
// shapes around it: "if <state>, flip this creature" as a clause that
// rechecks its state as it resolves, and the Kamigawa ki cycle.

// flipThisIf is "if <cond>, flip this creature" as an ability's whole
// effect (Kuon's end step, Student of Elements): the condition is asked
// as the ability resolves, which is also CR 603.4's second check for an
// intervening "if". A source that left and came back is a new object
// and flips nothing (#1432).
func flipThisIf(cond func(g *game.Game, controller uuid.UUID) bool) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if sourceIsNewObject(g, item) || !cond(g, item.Controller) {
			return nil
		}
		return FlipThis{}.Apply(NewContext(g, item))
	}
}

// flipThisWhen is the clause form: the same check, run from inside an
// effect that has already done the rest of its text (Jushi Apprentice's
// draw, Budoka Gardener's land). item is the resolving item.
func flipThisWhen(g *game.Game, item *game.StackItem, cond func(g *game.Game, controller uuid.UUID) bool) error {
	return flipThisIf(cond)(g, item)
}

// controlsAtLeast is "you control N or more <things>".
func controlsAtLeast(n int, match func(game.Card) bool) func(g *game.Game, controller uuid.UUID) bool {
	return func(g *game.Game, controller uuid.UUID) bool {
		return countControlled(g, controller, match) >= n
	}
}

// kiFlipTriggers is the two triggers the five ki flip cards print word
// for word (Budoka Pupil, Callow Jushi, Cunning Bandit, Faithful
// Squire, Hired Muscle):
//
//	"Whenever you cast a Spirit or Arcane spell, you may put a ki
//	 counter on this creature.
//	 At the beginning of the end step, if there are two or more ki
//	 counters on this creature, you may flip it."
//
// The counter is the Quillmane Baku shape (ADR 0109 §9). "The end step"
// is every turn's, so the flip trigger reads no actor; its intervening
// "if" is checked as the end step begins and again as the ability
// resolves (CR 603.4), and the "may" is asked as it resolves
// (MayChoice), not as it triggers.
func kiFlipTriggers(name string) []game.TriggeredAbility {
	return []game.TriggeredAbility{
		Optional(WheneverYouCast(Or(Subtype("Spirit"), Subtype("Arcane")),
			name+" — put a ki counter on this creature", putACounterOnThis("ki")),
			name+" — put a ki counter on it?"),
		On(game.EventBeginEndStep, func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return source.Counters["ki"] >= 2
		}, name+" — you may flip it", func(g *game.Game, item *game.StackItem) error {
			src, ok := g.LookupCardForEffect(item.SourceCardID)
			if !ok || !onBattlefield(g, item.SourceCardID) || src.Counters["ki"] < 2 || sourceIsNewObject(g, item) {
				return nil
			}
			return MayChoice{Question: name + " — flip it?", OnYes: FlipThis{}.Apply}.Apply(NewContext(g, item))
		}),
	}
}

// removeAKiCounter is the cost "Remove a ki counter from <this>".
func removeAKiCounter() game.AbilityCost { return RemoveCountersFromThis("ki", 1) }
