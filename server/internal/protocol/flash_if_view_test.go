package protocol

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Timely Ward, Asinine Antics
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// flash_if_view_test.go — #5207, the view halves of "as though it had
// flash if …".

func flashIfViewHand(t *testing.T, g *game.Game, me *game.Player, c game.Card) uuid.UUID {
	t.Helper()
	c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
	c.KnownBy = map[uuid.UUID]bool{me.ID: true}
	g.WithWriteLock(func() { me.Hand.PushTop(c) })
	return c.InstanceID
}

func flashIfViewTo(t *testing.T, g *game.Game, step game.Step) {
	t.Helper()
	for g.Turn.Step != step {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

// Timely Ward's "if it targets a commander" (CR 601.3b): at instant speed
// the picker offers the commander and not the plain creature; in a main
// phase it offers both.
func TestFlashIfTargetsNarrowsThePickerOnlyAtInstantSpeed(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ward := flashIfViewHand(t, g, me, game.Card{Name: "Timely Ward", TypeLine: "Enchantment — Aura",
		OracleID: "8e589dca-d31e-4d2f-81b1-18e5f9198795", ManaCost: "{2}{W}"})
	commander, bear := uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: commander, Name: "My Commander", TypeLine: "Legendary Creature — Human",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, IsCommander: true})
		g.Battlefield.PushTop(game.Card{InstanceID: bear, Name: "Grizzly Bears", TypeLine: "Creature — Bear",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	})
	targets := func() []string {
		c := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), ward.String())
		if c.LegalTargets == nil {
			return nil
		}
		return c.LegalTargets.Cards
	}
	flashIfViewTo(t, g, game.StepPrecombatMain)
	if got := targets(); !slices.Contains(got, commander.String()) || !slices.Contains(got, bear.String()) {
		t.Errorf("main phase: legal targets %v, want the commander and the bear", got)
	}
	flashIfViewTo(t, g, game.StepBeginCombat)
	if got := targets(); !slices.Contains(got, commander.String()) || slices.Contains(got, bear.String()) {
		t.Errorf("beginning of combat: legal targets %v, want the commander only", got)
	}
}

// "If you pay {2} more" is an optional cost stamped opens_flash (CR
// 601.3c), and the card offers no alternative cost for it.
func TestFlashForMoreShipsAsAnOptionalCostThatOpensFlash(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	antics := flashIfViewHand(t, g, me, game.Card{Name: "Asinine Antics", TypeLine: "Sorcery",
		OracleID: "0a0a051c-59ed-4286-b842-a34da7ef15d5", ManaCost: "{2}{U}{U}"})
	c := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), antics.String())
	if len(c.AlternativeCosts) != 0 {
		t.Errorf("alternative costs %+v, want none", c.AlternativeCosts)
	}
	if len(c.OptionalCosts) != 1 || !c.OptionalCosts[0].OpensFlash || c.OptionalCosts[0].ManaCost != "{2}" {
		t.Fatalf("optional costs %+v, want one {2} with opens_flash", c.OptionalCosts)
	}
}
