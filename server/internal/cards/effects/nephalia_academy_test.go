package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// nephalia_academy_test.go — ADR 0153 PR 2 (#3168).

const nephaliaAcademyOracle = "3b7e7a11-bf59-413d-8796-640d17c2c1c6"

// revealWatcher records the reveal events the card is supposed to emit.
type revealWatcher struct{ reveals []game.Event }

func (w *revealWatcher) OnEvent(_ *game.Game, ev game.Event) {
	if ev.Kind == game.EventRevealCards {
		w.reveals = append(w.reveals, ev)
	}
}

// answerNephaliaPrompts answers the "may reveal" prompts in order: one
// per discarded card, in the order they were asked.
func answerNephaliaPrompts(t *testing.T, d dboTable, answers ...bool) {
	t.Helper()
	for _, yes := range answers {
		answerLengPrompt(t, d.g, d.victim.ID, yes)
	}
}

// With the Academy out, an opponent's Mind Rot lets its controller reveal
// each discarded card and put it on top of the library instead. Saying yes
// to one and no to the other: the first is on top, revealed, and the
// second is in the graveyard.
func TestNephaliaAcademyPutsAnOpponentCausedDiscardOnTopOfTheLibrary(t *testing.T) {
	d := newDBOTable(t)
	d.permanent(d.victim, "Nephalia Academy", "Land", nephaliaAcademyOracle)
	a := d.handCard(d.victim, "Card A", "Sorcery", "{1}", "", 0, 0)
	b := d.handCard(d.victim, "Card B", "Sorcery", "{1}", "", 0, 0)
	w := d.watch()
	rw := &revealWatcher{}
	d.g.RegisterListener(rw)

	castCatalogSpell(t, d.g, "Mind Rot", "Sorcery", mindRotOracleID,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: d.victim.ID}})
	passPriorityAroundTable(t, d.g)
	answerDiscard(t, d.g, d.victim.ID, a, b)
	// The prompts come one card at a time, in the order the cards were
	// discarded: yes to the first, no to the second.
	answerNephaliaPrompts(t, d, true)
	answerNephaliaPrompts(t, d, false)
	passPriorityAroundTable(t, d.g)

	var onTop, inYard int
	for _, id := range []uuid.UUID{a, b} {
		switch {
		case d.victim.Library.Contains(id):
			onTop++
		case d.victim.Graveyard.Contains(id):
			inYard++
		}
	}
	if onTop != 1 || inYard != 1 {
		t.Fatalf("library %d graveyard %d, want one card each", onTop, inYard)
	}
	top := d.victim.Library.Cards[len(d.victim.Library.Cards)-1]
	if top.InstanceID != a && top.InstanceID != b {
		t.Errorf("the top of the library is %q, want the discarded card", top.Name)
	}
	if len(rw.reveals) != 1 || rw.reveals[0].CardID != top.InstanceID {
		t.Errorf("reveals = %d, want one, of the card put on top", len(rw.reveals))
	}
	var sawLibrary bool
	for _, ev := range w.events {
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
	d := newDBOTable(t)
	d.permanent(d.victim, "Nephalia Academy", "Land", nephaliaAcademyOracle)
	a := d.handCard(d.victim, "Card A", "Sorcery", "{1}", "", 0, 0)
	b := d.handCard(d.victim, "Card B", "Sorcery", "{1}", "", 0, 0)

	castCatalogSpell(t, d.g, "Mind Rot", "Sorcery", mindRotOracleID,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: d.victim.ID}})
	passPriorityAroundTable(t, d.g)
	answerDiscard(t, d.g, d.victim.ID, a, b)
	answerNephaliaPrompts(t, d, false)
	answerNephaliaPrompts(t, d, false)
	passPriorityAroundTable(t, d.g)

	if !d.victim.Graveyard.Contains(a) || !d.victim.Graveyard.Contains(b) {
		t.Error("declining left a card out of the graveyard")
	}
}

// Your own discards are not an opponent's: no prompt, graveyard.
func TestNephaliaAcademyIgnoresYourOwnDiscards(t *testing.T) {
	d := newDBOTable(t)
	d.permanent(d.caster, "Nephalia Academy", "Land", nephaliaAcademyOracle)
	a := d.handCard(d.caster, "Card A", "Sorcery", "{1}", "", 0, 0)
	b := d.handCard(d.caster, "Card B", "Sorcery", "{1}", "", 0, 0)

	d.iDiscard(t, a, b)

	if len(d.g.PendingChoices) != 0 {
		t.Fatalf("a prompt is pending: %+v", d.g.PendingChoices)
	}
	if !d.caster.Graveyard.Contains(a) || !d.caster.Graveyard.Contains(b) {
		t.Error("the discarded cards are not both in the graveyard")
	}
}

// It is the Academy's controller's discards that it saves: an opponent's
// Academy does nothing for you.
func TestNephaliaAcademyOnlyHelpsItsController(t *testing.T) {
	d := newDBOTable(t)
	d.permanent(d.caster, "Nephalia Academy", "Land", nephaliaAcademyOracle)
	a := d.handCard(d.victim, "Card A", "Sorcery", "{1}", "", 0, 0)
	b := d.handCard(d.victim, "Card B", "Sorcery", "{1}", "", 0, 0)

	d.theyDiscard(t, a, b)

	if len(d.g.PendingChoices) != 0 {
		t.Fatalf("a prompt is pending: %+v", d.g.PendingChoices)
	}
	if !d.victim.Graveyard.Contains(a) || !d.victim.Graveyard.Contains(b) {
		t.Error("the discarded cards are not both in the graveyard")
	}
}
