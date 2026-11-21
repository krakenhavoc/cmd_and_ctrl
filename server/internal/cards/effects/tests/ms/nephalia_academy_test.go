package testsms

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// revealWatcher records the reveal events the card is supposed to emit.
type revealWatcher struct{ reveals []game.Event }

func (w *revealWatcher) OnEvent(_ *game.Game, ev game.Event) {
	if ev.Kind == game.EventRevealCards {
		w.reveals = append(w.reveals, ev)
	}
}

// answerNephaliaPrompts answers the "may reveal" prompts in order: one
// per discarded card, in the order they were asked.
func answerNephaliaPrompts(t *testing.T, d effectstest.DboTable, answers ...bool) {
	t.Helper()
	for _, yes := range answers {
		effectstest.AnswerLengPrompt(t, d.G, d.Victim.ID, yes)
	}
}

// With the Academy out, an opponent's Mind Rot lets its controller reveal
// each discarded card and put it on top of the library instead. Saying yes
// to one and no to the other: the first is on top, revealed, and the
// second is in the graveyard.
func TestNephaliaAcademyPutsAnOpponentCausedDiscardOnTopOfTheLibrary(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	d.Permanent(d.Victim, "Nephalia Academy", "Land", effectstest.NephaliaAcademyOracle)
	a := d.HandCard(d.Victim, "Card A", "Sorcery", "{1}", "", 0, 0)
	b := d.HandCard(d.Victim, "Card B", "Sorcery", "{1}", "", 0, 0)
	w := d.Watch()
	rw := &revealWatcher{}
	d.G.RegisterListener(rw)

	effectstest.CastCatalogSpell(t, d.G, "Mind Rot", "Sorcery", effectstest.MindRotOracleID,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: d.Victim.ID}})
	effectstest.PassPriorityAroundTable(t, d.G)
	effectstest.AnswerDiscard(t, d.G, d.Victim.ID, a, b)
	// The prompts come one card at a time, in the order the cards were
	// discarded: yes to the first, no to the second.
	answerNephaliaPrompts(t, d, true)
	answerNephaliaPrompts(t, d, false)
	effectstest.PassPriorityAroundTable(t, d.G)

	var onTop, inYard int
	for _, id := range []uuid.UUID{a, b} {
		switch {
		case d.Victim.Library.Contains(id):
			onTop++
		case d.Victim.Graveyard.Contains(id):
			inYard++
		}
	}
	if onTop != 1 || inYard != 1 {
		t.Fatalf("library %d graveyard %d, want one card each", onTop, inYard)
	}
	top := d.Victim.Library.Cards[len(d.Victim.Library.Cards)-1]
	if top.InstanceID != a && top.InstanceID != b {
		t.Errorf("the top of the library is %q, want the discarded card", top.Name)
	}
	if len(rw.reveals) != 1 || rw.reveals[0].CardID != top.InstanceID {
		t.Errorf("reveals = %d, want one, of the card put on top", len(rw.reveals))
	}
	var sawLibrary bool
	for _, ev := range w.Events {
		if ev.CardID == top.InstanceID {
			sawLibrary = ev.NewZone == game.ZoneLibrary
		}
	}
	if !sawLibrary {
		t.Error("the discard event for the saved card does not name the library")
	}
}

// Declining leaves the discard in the graveyard.
func TestNephaliaAcademyMayBeDeclined(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	d.Permanent(d.Victim, "Nephalia Academy", "Land", effectstest.NephaliaAcademyOracle)
	a := d.HandCard(d.Victim, "Card A", "Sorcery", "{1}", "", 0, 0)
	b := d.HandCard(d.Victim, "Card B", "Sorcery", "{1}", "", 0, 0)

	effectstest.CastCatalogSpell(t, d.G, "Mind Rot", "Sorcery", effectstest.MindRotOracleID,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: d.Victim.ID}})
	effectstest.PassPriorityAroundTable(t, d.G)
	effectstest.AnswerDiscard(t, d.G, d.Victim.ID, a, b)
	answerNephaliaPrompts(t, d, false)
	answerNephaliaPrompts(t, d, false)
	effectstest.PassPriorityAroundTable(t, d.G)

	if !d.Victim.Graveyard.Contains(a) || !d.Victim.Graveyard.Contains(b) {
		t.Error("declining left a card out of the graveyard")
	}
}

// Your own discards are not an opponent's: no prompt, graveyard.
func TestNephaliaAcademyIgnoresYourOwnDiscards(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	d.Permanent(d.Caster, "Nephalia Academy", "Land", effectstest.NephaliaAcademyOracle)
	a := d.HandCard(d.Caster, "Card A", "Sorcery", "{1}", "", 0, 0)
	b := d.HandCard(d.Caster, "Card B", "Sorcery", "{1}", "", 0, 0)

	d.IDiscard(t, a, b)

	if len(d.G.PendingChoices) != 0 {
		t.Fatalf("a prompt is pending: %+v", d.G.PendingChoices)
	}
	if !d.Caster.Graveyard.Contains(a) || !d.Caster.Graveyard.Contains(b) {
		t.Error("the discarded cards are not both in the graveyard")
	}
}

// It is the Academy's controller's discards that it saves: an opponent's
// Academy does nothing for you.
func TestNephaliaAcademyOnlyHelpsItsController(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	d.Permanent(d.Caster, "Nephalia Academy", "Land", effectstest.NephaliaAcademyOracle)
	a := d.HandCard(d.Victim, "Card A", "Sorcery", "{1}", "", 0, 0)
	b := d.HandCard(d.Victim, "Card B", "Sorcery", "{1}", "", 0, 0)

	d.TheyDiscard(t, a, b)

	if len(d.G.PendingChoices) != 0 {
		t.Fatalf("a prompt is pending: %+v", d.G.PendingChoices)
	}
	if !d.Victim.Graveyard.Contains(a) || !d.Victim.Graveyard.Contains(b) {
		t.Error("the discarded cards are not both in the graveyard")
	}
}
