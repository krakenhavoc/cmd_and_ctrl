package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mirror-Shield Hoplite — {R}{W} Creature — Human Soldier 2/2 (#5651).
//
//	"Vigilance
//	 Whenever a creature you control becomes the target of a backup
//	 ability, copy that ability. You may choose new targets for the
//	 copy. This ability triggers only once each turn."
//
// "A backup ability" is read off the event: Event.TargeterKeyword is the
// keyword the targeting trigger's catalog row carries, and effects.Backup
// stamps game.KeywordBackup on its row (CR 702.165a, "Backup is a
// triggered ability"). Whose backup ability it is does not matter — the
// clause says "a creature you control becomes the target", so an
// opponent's backup creature pointing its counters at one of yours
// triggers this too — and the Hoplite itself is a creature you control.
//
// The copy is the Hoplite's controller's (CR 707.10) and has the
// original's source (CR 707.10b), so "if that's another creature" is
// asked of the creature with backup, and the abilities it grants are
// that creature's, not the Hoplite's. New targets are optional
// (CR 707.10c); the copy resolves first, being put on the stack above
// the ability it copies. The body is copyTheTriggeringAbility, shared
// with every "copy that ability" trigger; a backup ability that has left
// the stack by then (countered in response) is not copied.
//
// A copy that takes a creature you control as its target is that
// creature becoming the target of a backup ability again, which is why
// the card prints "only once each turn": the tally counts the trigger
// from the moment it is put on the stack, so one that was countered
// still used the turn.
//
// No simplification.
func init() {
	const label = "Mirror-Shield Hoplite — copy that ability. You may choose new targets for the copy"
	Register(Spec{
		OracleID:        "23f01235-b5a5-4b2e-a90b-b49fc0c7b18d",
		Name:            "Mirror-Shield Hoplite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			TriggersOnlyOnceEachTurn(On(game.EventBecomesTarget,
				creatureYouControlBecameTargetOfABackupAbility, label, copyTheTriggeringAbility)),
		},
	})
}

// creatureYouControlBecameTargetOfABackupAbility is "a creature you
// control becomes the target of a backup ability": the event is a card
// target, the targeter is a backup ability (Event.TargeterKeyword,
// #5651 — a copy of one included, CR 707.10), and the target is a
// creature on the battlefield controlled by the source's controller.
func creatureYouControlBecameTargetOfABackupAbility(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBecomesTarget || ev.CardID == uuid.Nil || ev.TargeterKeyword != game.KeywordBackup {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.Controller == source.Controller
}
