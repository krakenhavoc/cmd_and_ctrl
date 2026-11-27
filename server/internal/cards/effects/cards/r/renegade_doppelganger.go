package cardsr

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Renegade Doppelganger — Creature — Shapeshifter {1}{U}, 0/1 (Rise of the
// Eldrazi, S82 OLD batch 484, #5581):
//
//	"Whenever another creature you control enters, you may have this creature
//	 become a copy of that creature until end of turn. (If it does, it loses
//	 this ability for the rest of the turn.)"
//
// The copy lasts until end of turn (CR 707.2, 611.2). "Loses this ability for
// the rest of the turn" is a note on the Doppelganger naming this turn: once
// it really became a copy this turn the trigger no longer applies, even if
// the creature it copied has this very ability. A "no" spends nothing, and a
// new turn starts fresh.
//
// No simplification.
func init() {
	const label = "Renegade Doppelganger — become a copy of that creature until end of turn"
	t := effects.WheneverAnotherCreatureEntersUnderYourControl(label, func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil || item.Trigger.Event.CardID == uuid.Nil || !effects.SourceIsStillThisPermanent(g, item) {
			return nil
		}
		ctx := effects.NewContext(g, item)
		if game.NotedIn(g.NotesForEffect(item.SourceCardID), renegadeDoppelgangerNote(g)) {
			return nil
		}
		copied, err := effects.BecomeCopy{
			Targets: []uuid.UUID{item.SourceCardID},
			Of:      item.Trigger.Event.CardID,
			Label:   label,
		}.ApplyReports(ctx)
		if err != nil || !copied {
			return err
		}
		g.AddNoteForEffect(item.SourceCardID, renegadeDoppelgangerNote(g))
		return nil
	})
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return !game.NotedIn(source.Notes, renegadeDoppelgangerNote(g)) && (when == nil || when(ev, source, lki, g))
	}
	effects.Register(effects.Spec{
		OracleID:     "08528c5e-9a01-42a4-a00b-c1ec5eaae64d",
		Name:         "Renegade Doppelganger",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Optional(t, "Renegade Doppelganger — become a copy of that creature until end of turn?")},
	})
}

// renegadeDoppelgangerNote is the word recording that the Doppelganger became
// a copy during the current turn (game.Turn.Seq is the turn's identity).
func renegadeDoppelgangerNote(g *game.Game) string {
	return "renegade-doppelganger/copied-turn-" + strconv.Itoa(g.Turn.Seq)
}
