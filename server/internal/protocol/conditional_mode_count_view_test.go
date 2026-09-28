package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// conditional_mode_count_view_test.go — #1590, ADR 0065's 2026-09-27
// amendment: `modes.max` is the bound THIS caster is held to right
// now, so the picker offers "both" exactly when the announce gate
// would accept it. The printed Max stays on the public copy.

// viewControlsACommander is registered once, at package init, as a
// card file would register its condition.
var viewControlsACommander = game.ModeCondition("test/view-controls-a-commander", func(g *game.Game, chooser uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCommander && c.Controller == chooser {
			return true
		}
	}
	return false
})

func stubConditionalModeSpec(t *testing.T, oracle string) {
	t.Helper()
	prev := game.CatalogModeSpec
	spec := (&game.ModeSpec{
		Prompt: "Choose one", Min: 1, Max: 1,
		Options: []game.ModeOption{{Label: "Add mana."}, {Label: "Exile three."}},
	}).OrUpToIf(2, viewControlsACommander)
	game.CatalogModeSpec = func(id string) *game.ModeSpec {
		if id != oracle {
			return nil
		}
		return spec
	}
	t.Cleanup(func() { game.CatalogModeSpec = prev })
}

func handModes(t *testing.T, g *game.Game, viewer *game.Player, seat int, id uuid.UUID) *ModeSpecView {
	t.Helper()
	v := ViewOfGameFor(g, viewer.ID.String())
	for _, c := range v.Seats[seat].Hand.Cards {
		if c.InstanceID == id.String() {
			return c.Modes
		}
	}
	t.Fatalf("card %s missing from seat %d's hand", id, seat)
	return nil
}

func TestModesMaxIsTheCastersEffectiveBound(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const oracle = "test-view-conditional-mode-count"
	stubConditionalModeSpec(t, oracle)

	spell := game.NewCard("Will", me.ID)
	spell.TypeLine = "Sorcery"
	spell.OracleID = oracle
	spell.KnownBy = map[uuid.UUID]bool{me.ID: true}
	me.Hand.PushTop(spell)

	if m := handModes(t, g, me, 0, spell.InstanceID); m == nil || m.Max != 1 {
		t.Fatalf("no commander: max = %+v, want 1", m)
	}

	// An opponent's commander on their side does not raise it.
	theirs := game.NewCard("Their Commander", opp.ID)
	theirs.TypeLine = "Legendary Creature — Elf"
	theirs.IsCommander = true
	g.Battlefield.PushTop(theirs)
	if m := handModes(t, g, me, 0, spell.InstanceID); m.Max != 1 {
		t.Errorf("an opponent's commander: max = %d, want 1", m.Max)
	}

	// Yours does, and the minimum does not move ("you MAY choose both").
	mine := game.NewCard("My Commander", me.ID)
	mine.TypeLine = "Legendary Creature — Human"
	mine.IsCommander = true
	g.Battlefield.PushTop(mine)
	m := handModes(t, g, me, 0, spell.InstanceID)
	if m.Max != 2 || m.Min != 1 {
		t.Errorf("your commander: bounds = %d..%d, want 1..2", m.Min, m.Max)
	}
}

// The raise is the asking seat's answer; the public copy of the same
// card shows the printed bound.
func TestPublicModeSpecShowsThePrintedMax(t *testing.T) {
	g := buildActiveGame(t)
	const oracle = "test-view-conditional-mode-count-public"
	stubConditionalModeSpec(t, oracle)
	ms := game.ModeSpecFor(oracle)
	mine := game.NewCard("My Commander", g.Seats[0].ID)
	mine.TypeLine = "Legendary Creature — Human"
	mine.IsCommander = true
	g.Battlefield.PushTop(mine)

	v := viewOfModeSpec(g, g.Seats[0].ID, game.SourceObject(g.Seats[0].ID, nil), ms)
	if v.Max != 2 {
		t.Fatalf("the caster's own stamp: max = %d, want 2", v.Max)
	}
	if pub := publicModeSpec(v); pub.Max != 1 {
		t.Errorf("the public copy: max = %d, want the printed 1", pub.Max)
	}
	if v.Max != 2 {
		t.Error("publicModeSpec must copy, not overwrite the caster's own stamp")
	}
}
