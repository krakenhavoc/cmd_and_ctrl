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

// #5363, CR 601.3b: Timely Ward cast through Omniscience's "without
// paying its mana cost" at instant speed is open only because of its own
// "if it targets a commander", so that offer's `legal_targets` hold the
// commander and not the plain creature, as CastSpell holds the claim. In
// a main phase the offer takes either.
func TestFlashIfTargetsNarrowsAnAlternativeCostOfferAtInstantSpeed(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ward := flashIfViewHand(t, g, me, game.Card{Name: "Timely Ward", TypeLine: "Enchantment — Aura",
		OracleID: "8e589dca-d31e-4d2f-81b1-18e5f9198795", ManaCost: "{2}{W}"})
	commander, bear := uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Omniscience", TypeLine: "Enchantment",
			OracleID: "730e39e6-c61d-48b5-8827-bfd952bf1be7", Owner: me.ID, Controller: me.ID})
		g.Battlefield.PushTop(game.Card{InstanceID: commander, Name: "My Commander", TypeLine: "Legendary Creature — Human",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, IsCommander: true})
		g.Battlefield.PushTop(game.Card{InstanceID: bear, Name: "Grizzly Bears", TypeLine: "Creature — Bear",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	})
	offerTargets := func() []string {
		t.Helper()
		c := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), ward.String())
		for _, a := range c.AlternativeCosts {
			if a.Key != "granted-free" {
				continue
			}
			if a.LegalTargets == nil {
				return nil
			}
			return a.LegalTargets.Cards
		}
		t.Fatalf("no Omniscience offer on Timely Ward: %+v", c.AlternativeCosts)
		return nil
	}
	flashIfViewTo(t, g, game.StepPrecombatMain)
	if got := offerTargets(); !slices.Contains(got, commander.String()) || !slices.Contains(got, bear.String()) {
		t.Errorf("main phase: the free offer's targets %v, want the commander and the bear", got)
	}
	flashIfViewTo(t, g, game.StepBeginCombat)
	if got := offerTargets(); !slices.Contains(got, commander.String()) || slices.Contains(got, bear.String()) {
		t.Errorf("beginning of combat: the free offer's targets %v, want the commander only", got)
	}
}

// #5363, CR 601.3c: Asinine Antics's "pay {2} more" opens the window.
// In a main phase no claim needs it; at instant speed the printed claim
// and Omniscience's free claim are each open only with it, so both carry
// opens_flash_required and the free offer is not timing_closed.
func TestOpensFlashRequiredMarksTheClaimsThatNeedTheCost(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	antics := flashIfViewHand(t, g, me, game.Card{Name: "Asinine Antics", TypeLine: "Sorcery",
		OracleID: "0a0a051c-59ed-4286-b842-a34da7ef15d5", ManaCost: "{2}{U}{U}"})
	card := func() CardView {
		return handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), antics.String())
	}
	flashIfViewTo(t, g, game.StepPrecombatMain)
	if c := card(); c.OpensFlashRequired {
		t.Error("main phase, no alternative costs: opens_flash_required is set")
	}
	flashIfViewTo(t, g, game.StepBeginCombat)
	if c := card(); !c.OpensFlashRequired {
		t.Error("beginning of combat, no alternative costs: opens_flash_required is not set")
	}

	g2 := buildActiveGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	antics2 := flashIfViewHand(t, g2, me2, game.Card{Name: "Asinine Antics", TypeLine: "Sorcery",
		OracleID: "0a0a051c-59ed-4286-b842-a34da7ef15d5", ManaCost: "{2}{U}{U}"})
	g2.WithWriteLock(func() {
		g2.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Omniscience", TypeLine: "Enchantment",
			OracleID: "730e39e6-c61d-48b5-8827-bfd952bf1be7", Owner: me2.ID, Controller: me2.ID})
	})
	free := func() (CardView, *AlternativeCostView) {
		t.Helper()
		c := handCardIn(t, ViewOfGameFor(g2, me2.ID.String()), me2.ID.String(), antics2.String())
		for i := range c.AlternativeCosts {
			if c.AlternativeCosts[i].Key == "granted-free" {
				return c, &c.AlternativeCosts[i]
			}
		}
		t.Fatalf("no Omniscience offer: %+v", c.AlternativeCosts)
		return c, nil
	}
	flashIfViewTo(t, g2, game.StepPrecombatMain)
	if c, a := free(); c.OpensFlashRequired || a.OpensFlashRequired || a.TimingClosed || c.PrintedCostTimingClosed {
		t.Errorf("main phase: printed required=%v closed=%v, free required=%v closed=%v; want all false",
			c.OpensFlashRequired, c.PrintedCostTimingClosed, a.OpensFlashRequired, a.TimingClosed)
	}
	flashIfViewTo(t, g2, game.StepBeginCombat)
	if c, a := free(); !c.OpensFlashRequired || !a.OpensFlashRequired || a.TimingClosed || c.PrintedCostTimingClosed {
		t.Errorf("beginning of combat: printed required=%v closed=%v, free required=%v closed=%v; want required and not closed",
			c.OpensFlashRequired, c.PrintedCostTimingClosed, a.OpensFlashRequired, a.TimingClosed)
	}
}
