package effectstest

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// modes_not_chosen_cards_test.go — ADR 0097 PR 2 (#1749): the cards
// whose modal clause says "choose one that hasn't been chosen THIS
// TURN". Each test plays the card through the engine's own verbs and
// checks three things: a used bullet is not offered again, every
// bullet does what it prints, and a case the card does not apply to
// asks nothing.

// chooseModeNow answers the newest mode_pick owed by `chooser` after
// checking what it offers.
func ChooseModeNow(t *testing.T, g *game.Game, chooser uuid.UUID, wantOffered []int, mode int) {
	t.Helper()
	c := ModePickChoiceFor(g, chooser)
	if c == nil {
		t.Fatalf("no mode_pick for %s; want one offering %v", chooser, wantOffered)
	}
	if !slices.Equal(c.ModeOptionIndex, wantOffered) {
		t.Fatalf("mode_pick offers %v, want %v (used %v)", c.ModeOptionIndex, wantOffered, c.ModeUsedIndex)
	}
	AnswerMode(t, g, c, chooser, mode)
}

// noModePick fails if `chooser` owes a mode_pick.
func NoModePick(t *testing.T, g *game.Game, chooser uuid.UUID, why string) {
	t.Helper()
	if c := ModePickChoiceFor(g, chooser); c != nil {
		t.Fatalf("%s: unexpected mode_pick offering %v", why, c.ModeOptionIndex)
	}
}

// advanceUntilModePick walks the turn on until `chooser` owes a
// mode_pick.
func AdvanceUntilModePick(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	for i := 0; i < 40 && ModePickChoiceFor(g, chooser) == nil; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if ModePickChoiceFor(g, chooser) == nil {
		t.Fatal("no mode_pick arrived")
	}
}

func HandSize(p *game.Player) int { return p.Hand.Size() }

func PoolCount(p *game.Player, color string) int {
	n := 0
	for _, m := range p.ManaPool.Units() {
		if m.Color == color {
			n++
		}
	}
	return n
}

func EffectiveOf(t *testing.T, g *game.Game, id uuid.UUID) game.Characteristic {
	t.Helper()
	var ch game.Characteristic
	var ok bool
	g.ReadSnapshot(func() {
		var c game.Card
		c, ok = BattlefieldCard(g, id)
		if ok {
			ch = c.Effective()
		}
	})
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return ch
}

func ZoneOf(g *game.Game, id uuid.UUID) game.ZoneKind {
	var kind game.ZoneKind
	g.ReadSnapshot(func() {
		if z := g.FindCardZoneForEffect(id); z != nil {
			kind = z.Kind
		}
	})
	return kind
}
