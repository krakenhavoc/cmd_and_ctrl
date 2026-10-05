package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// escalate_view_test.go — CR 702.120a, #2126: `modes.escalate` is the
// printed per-mode cost, and `modes.max` is clamped to the mode count
// THIS caster could pay the non-mana half for, so the picker never
// offers a count the server refuses. The per-viewer half stays off the
// public copy.

func stubEscalateModeSpec(t *testing.T, oracle string, escalate *game.EscalateCost) {
	t.Helper()
	prev := game.CatalogModeSpec
	spec := &game.ModeSpec{
		Prompt: "Choose one or more", Min: 1, Max: 3,
		Options:  []game.ModeOption{{Label: "A."}, {Label: "B."}, {Label: "C."}},
		Escalate: escalate,
	}
	game.CatalogModeSpec = func(id string) *game.ModeSpec {
		if id != oracle {
			return nil
		}
		return spec
	}
	t.Cleanup(func() { game.CatalogModeSpec = prev })
}

func escalateHandCard(me *game.Player, oracle string) game.Card {
	spell := game.NewCard("Escalate Test", me.ID)
	spell.TypeLine = "Sorcery"
	spell.OracleID = oracle
	spell.KnownBy = map[uuid.UUID]bool{me.ID: true}
	me.Hand.PushTop(spell)
	return spell
}

func TestEscalateViewClampsMaxToTheDiscardsThePlayerCanPay(t *testing.T) {
	const oracle = "test-view-escalate-discard"
	g := buildActiveGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	stubEscalateModeSpec(t, oracle, &game.EscalateCost{DiscardCards: 1, Label: "Escalate—Discard a card"})
	spell := escalateHandCard(me, oracle)

	for _, tc := range []struct{ fodder, wantMax int }{{0, 1}, {1, 2}, {2, 3}, {6, 3}} {
		me.Hand.Cards = me.Hand.Cards[:1]
		for range tc.fodder {
			c := game.NewCard("Fodder", me.ID)
			c.KnownBy = map[uuid.UUID]bool{me.ID: true}
			me.Hand.PushTop(c)
		}
		m := handModes(t, g, me, 0, spell.InstanceID)
		if m == nil || m.Escalate == nil {
			t.Fatalf("%d fodder: no escalate on the view: %+v", tc.fodder, m)
		}
		if m.Max != tc.wantMax || m.Escalate.MaxExtra != tc.wantMax-1 {
			t.Errorf("%d fodder: max %d / max_extra %d, want %d / %d", tc.fodder, m.Max, m.Escalate.MaxExtra, tc.wantMax, tc.wantMax-1)
		}
		if m.Escalate.DiscardCards != 1 || m.Escalate.Label == "" {
			t.Errorf("escalate = %+v, want the printed discard clause", m.Escalate)
		}
	}
}

func TestEscalateViewClampsMaxToTheCreaturesThePlayerCanTap(t *testing.T) {
	const oracle = "test-view-escalate-tap"
	g := buildActiveGame(t)
	me := g.Seats[0]
	stubEscalateModeSpec(t, oracle, &game.EscalateCost{TapCreatures: 1, Label: "Escalate—Tap an untapped creature you control"})
	spell := escalateHandCard(me, oracle)

	if m := handModes(t, g, me, 0, spell.InstanceID); m.Max != 1 || len(m.Escalate.TapOptions) != 0 {
		t.Fatalf("no creatures: max %d options %v, want 1 and none", m.Max, m.Escalate.TapOptions)
	}
	for i := 1; i <= 3; i++ {
		c := game.NewCard("Bear", me.ID)
		c.TypeLine = "Creature — Bear"
		g.Battlefield.PushTop(c)
		want := min(1+i, 3)
		m := handModes(t, g, me, 0, spell.InstanceID)
		if m.Max != want || len(m.Escalate.TapOptions) != i {
			t.Errorf("%d creatures: max %d options %d, want %d and %d", i, m.Max, len(m.Escalate.TapOptions), want, i)
		}
	}
	// A tapped creature is not an option.
	for i := range g.Battlefield.Cards {
		g.Battlefield.Cards[i].Tapped = true
	}
	if m := handModes(t, g, me, 0, spell.InstanceID); m.Max != 1 {
		t.Errorf("all tapped: max %d, want 1", m.Max)
	}
}

func TestEscalateManaIsNotAskedByTheViewAndThePublicCopyDropsThePerViewerHalf(t *testing.T) {
	const oracle = "test-view-escalate-mana"
	g := buildActiveGame(t)
	me := g.Seats[0]
	stubEscalateModeSpec(t, oracle, &game.EscalateCost{ManaCost: "{G}", Label: "Escalate {G}"})
	spell := escalateHandCard(me, oracle)

	// Mana is paid after announcing (CR 601.2g): the full count is
	// offered with no mana source in play.
	m := handModes(t, g, me, 0, spell.InstanceID)
	if m.Max != 3 || m.Escalate.ManaCost != "{G}" {
		t.Errorf("mana escalate: max %d escalate %+v, want 3 and {G}", m.Max, m.Escalate)
	}

	pub := publicModeSpec(m)
	if pub.Escalate == nil || pub.Escalate.ManaCost != "{G}" {
		t.Errorf("the printed clause should survive on the public copy: %+v", pub.Escalate)
	}
	if pub.Escalate.MaxExtra != 0 || len(pub.Escalate.TapOptions) != 0 {
		t.Errorf("the per-viewer half leaked onto the public copy: %+v", pub.Escalate)
	}
	if m.Escalate.MaxExtra != 2 {
		t.Errorf("publicModeSpec mutated the owner's copy: %+v", m.Escalate)
	}
}
