package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// emblem_gates_test.go — ADR 0109 §5 (#1899): an emblem is read by every
// rule gate whose slot it carries — the cast gate, the land-play gate,
// the game-end gates and the untap caps — with the emblem as the source,
// because CR 114.4 says an emblem's abilities function in the command
// zone. The real cards (Narset Transcendent, Dovin Baan) are tested in
// internal/cards/effects; these stub the catalog so each walk is pinned
// on its own.

const gatesWalker = "test-gates-walker"

// stubGateEmblem wires a catalog holding one card whose emblem carries
// `def`'s rule gates, and drops that card onto the battlefield under
// `owner` so CreateEmblemForEffect can derive the emblem's key. Returns
// the source card's ID.
func stubGateEmblem(t *testing.T, g *Game, owner uuid.UUID, def CardDef) uuid.UUID {
	t.Helper()
	def.Emblem = &EmblemDef{Label: "Gates Walker emblem", Text: "A rule gate."}
	defs := map[string]*CardDef{
		gatesWalker:            {},
		EmblemKey(gatesWalker): &def,
	}
	prev := CatalogLookup
	CatalogLookup = func(key string) *CardDef {
		if d, ok := defs[key]; ok {
			return d
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { CatalogLookup = prev })
	return pushEmblemSourceCard(g, owner, gatesWalker)
}

// opponentsOf is "your opponents" read off the gate's source.
func opponentsOf(source Card, player uuid.UUID) bool { return source.Controller != player }

// The cast gate refuses an opponent's matching spell from an emblem, and
// keeps refusing once the card that made the emblem has left: the emblem,
// not the card, is the source.
func TestCastGateReadsAnEmblem(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := stubGateEmblem(t, g, me.ID, CardDef{CastRestrictions: []CastRestriction{{
		Label: "Your opponents can't cast noncreature spells.",
		Forbids: func(q CastQuery) bool {
			return opponentsOf(q.Source, q.Controller) && !q.Card.IsCreature()
		},
	}}})
	if err := castGate(g, opp, instantCard(opp.ID), ZoneHand); err != nil {
		t.Fatalf("setup: the cast was refused before any emblem: %v", err)
	}
	giveEmblem(t, g, me.ID, src)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(src) })

	err := castGate(g, opp, instantCard(opp.ID), ZoneHand)
	var cant *CantCastError
	if !errors.As(err, &cant) || cant.Reason != "Your opponents can't cast noncreature spells." {
		t.Fatalf("an opponent's instant under the emblem: %v, want the emblem's clause", err)
	}
	if cant.Source != me.Emblems.Cards[0].InstanceID {
		t.Errorf("the refusal's source is %s, want the emblem", cant.Source)
	}
	if err := castGate(g, me, instantCard(me.ID), ZoneHand); err != nil {
		t.Errorf("the emblem's owner was refused their own instant: %v", err)
	}
	bear := NewCard("Test Bear", opp.ID)
	bear.TypeLine = "Creature — Bear"
	if err := castGate(g, opp, bear, ZoneHand); err != nil {
		t.Errorf("an opponent's creature spell was refused: %v", err)
	}
}

// CR 800.4a: an emblem leaves the game with its owner, and its "can't"
// with it.
func TestAnEmblemsCastBanEndsWhenItsOwnerLeaves(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := stubGateEmblem(t, g, me.ID, CardDef{CastRestrictions: []CastRestriction{{
		Label:   "Your opponents can't cast spells.",
		Forbids: func(q CastQuery) bool { return opponentsOf(q.Source, q.Controller) },
	}}})
	giveEmblem(t, g, me.ID, src)
	if castGate(g, opp, instantCard(opp.ID), ZoneHand) == nil {
		t.Fatal("setup: the emblem refuses nothing")
	}
	if err := g.Concede(me.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if err := castGate(g, opp, instantCard(opp.ID), ZoneHand); err != nil {
		t.Errorf("the emblem's ban outlived its owner: %v", err)
	}
}

// The land-play gate reads an emblem's restriction the same way.
func TestLandPlayGateReadsAnEmblem(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := stubGateEmblem(t, g, me.ID, CardDef{LandPlayRestrictions: []LandPlayRestriction{{
		Label:   "Your opponents can't play lands.",
		Forbids: func(q LandPlayQuery) bool { return opponentsOf(q.Source, q.Player) },
	}}})
	giveEmblem(t, g, me.ID, src)
	var oppErr, myErr error
	g.ReadSnapshot(func() {
		oppErr = g.LandPlayGateLocked(opp.ID, Card{Name: "Forest", TypeLine: "Land"}, ZoneHand)
		myErr = g.LandPlayGateLocked(me.ID, Card{Name: "Forest", TypeLine: "Land"}, ZoneHand)
	})
	var cant *CantPlayLandError
	if !errors.As(oppErr, &cant) || cant.Reason != "Your opponents can't play lands. — Gates Walker emblem" {
		t.Errorf("an opponent's land play under the emblem: %v, want the clause and the emblem", oppErr)
	}
	if myErr != nil {
		t.Errorf("the emblem's owner was refused a land play: %v", myErr)
	}
}

// The game-end gates read an emblem's gate, its While against the
// emblem, and name the emblem as the source.
func TestGameEndGatesReadAnEmblem(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	holds := true
	src := stubGateEmblem(t, g, me.ID, CardDef{GameEndGates: []GameEndGate{
		{Scope: GateYou, CantLose: true, While: func(*Game, Card) bool { return holds }},
		{Scope: GateOpponents, CantWin: true, While: func(*Game, Card) bool { return holds }},
	}})
	giveEmblem(t, g, me.ID, src)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(src) })

	var canLose, oppCanWin bool
	read := func() {
		g.ReadSnapshot(func() {
			canLose = g.canLoseLocked(me, LossLife)
			oppCanWin = g.canWinLocked(opp)
		})
	}
	read()
	if canLose || oppCanWin {
		t.Fatalf("under the emblem: can lose = %v, opponent can win = %v; want both false", canLose, oppCanWin)
	}
	var gates []GameEndGateSource
	g.ReadSnapshot(func() { gates = g.GameEndGatesForEffect(me) })
	if len(gates) != 1 || gates[0].Source != me.Emblems.Cards[0].InstanceID || gates[0].ThisTurn {
		t.Errorf("the seat's gates = %+v, want one from the emblem, not a granted one", gates)
	}
	holds = false
	read()
	if !canLose || !oppCanWin {
		t.Errorf("with the condition false: can lose = %v, opponent can win = %v; want both true", canLose, oppCanWin)
	}
}

// The untap caps read an emblem's cap, with the emblem as the source its
// Applies and Counts see.
func TestUntapCapsReadAnEmblem(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := stubGateEmblem(t, g, me.ID, CardDef{UntapCaps: []UntapCap{{
		Label: "Your opponents can't untap more than two permanents during their untap steps.",
		Max:   2,
		Applies: func(_ *Game, source *Card, active uuid.UUID) bool {
			return source.Controller != active
		},
	}}})
	giveEmblem(t, g, me.ID, src)
	var forOpp, forMe []boundUntapCap
	g.WithWriteLock(func() {
		forOpp = g.activeUntapCapsLocked(opp.ID)
		forMe = g.activeUntapCapsLocked(me.ID)
	})
	if len(forOpp) != 1 || forOpp[0].source.InstanceID != me.Emblems.Cards[0].InstanceID {
		t.Errorf("an opponent's untap step: %d caps, want the emblem's", len(forOpp))
	}
	if len(forMe) != 0 {
		t.Errorf("the emblem's owner's untap step: %d caps, want none", len(forMe))
	}
}
