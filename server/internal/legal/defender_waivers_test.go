package legal_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// defender_waivers_test.go — ADR 0106's 2026-10-16 amendment (#1936),
// the #544 half. Weathered Sentinels (the real catalog card) is offered
// only the players who attacked its controller during their last turn,
// every offered attack is accepted, the withheld ones are refused, and a
// goaded Sentinels owes an attack only where it may attack (CR 508.1d).

const legalSentinelsOracle = "0903b941-82f1-4d3e-9717-29ab7d0c8cef"

func readySentinels(g *game.Game, me *game.Player) uuid.UUID {
	id := enteredCard(g, me, game.Card{
		Name: "Weathered Sentinels", TypeLine: "Artifact Creature — Wall",
		OracleID: legalSentinelsOracle, Power: 2, Toughness: 5,
	})
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.InstanceID == id {
			c.SummonedThisTurn = false
		}
	}
	return id
}

func attackTargetsOffered(t *testing.T, moves []legal.Move, attacker uuid.UUID) map[string]bool {
	t.Helper()
	out := map[string]bool{}
	for _, m := range moves {
		if m.Kind == legal.KindAttack && m.Source == attacker {
			out[decodeAttack(t, m).Target] = true
		}
	}
	return out
}

func TestWeatheredSentinelsMovesAgreeWithTheEngine(t *testing.T) {
	g := newTable(t)
	s := g.Turn.ActiveSeat
	me, hitMe, other, third := g.Seats[s], g.Seats[(s+1)%4], g.Seats[(s+2)%4], g.Seats[(s+3)%4]
	clearHand(me)
	sentinels := readySentinels(g, me)
	walker := battlefieldCard(g, hitMe, game.Card{
		Name: "Their Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 3},
	})
	hitMe.LastTurnAttacks = []game.AttackRecord{{Attacker: uuid.New(), Defender: me.ID}}
	other.LastTurnAttacks = []game.AttackRecord{{Attacker: uuid.New(), Defender: third.ID}}
	advanceTo(t, g, game.StepDeclareAttackers)

	moves := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)
	got := attackTargetsOffered(t, moves, sentinels)
	if len(got) != 1 || !got[hitMe.ID.String()] {
		t.Fatalf("offered %v: want only the player who attacked you", got)
	}
	for _, refused := range []uuid.UUID{other.ID, third.ID, walker} {
		if err := g.Clone().DeclareAttacker(sentinels, refused); !errors.Is(err, game.ErrDefender) {
			t.Errorf("engine accepted the withheld attack at %s: %v", refused, err)
		}
	}
}

func TestWeatheredSentinelsWithNobodyToAttackIsNotOffered(t *testing.T) {
	g := newTable(t)
	s := g.Turn.ActiveSeat
	me := g.Seats[s]
	clearHand(me)
	sentinels := readySentinels(g, me)
	advanceTo(t, g, game.StepDeclareAttackers)

	moves := legal.EnumerateFor(g, me.ID)
	if n := attacksBy(moves, sentinels); n != 0 {
		t.Errorf("offered %d attacks though nobody attacked you: %v", n, labels(moves))
	}
}

// A goaded Sentinels attacks if able and, if able, a player other than
// its goader (CR 701.15b). Able means "as though it didn't have
// defender" reaches the target, so with two players who attacked you it
// owes the attack at the one who did not goad it, and with nobody it
// owes nothing.
func TestGoadedWeatheredSentinelsOwesOnlyWhereItMayAttack(t *testing.T) {
	t.Run("two players attacked you", func(t *testing.T) {
		g := newTable(t)
		s := g.Turn.ActiveSeat
		me, goader, other := g.Seats[s], g.Seats[(s+1)%4], g.Seats[(s+2)%4]
		clearHand(me)
		sentinels := readySentinels(g, me)
		goader.LastTurnAttacks = []game.AttackRecord{{Attacker: uuid.New(), Defender: me.ID}}
		other.LastTurnAttacks = []game.AttackRecord{{Attacker: uuid.New(), Defender: me.ID}}
		g.WithWriteLock(func() { g.GoadForEffect(sentinels, goader.ID) })
		advanceTo(t, g, game.StepDeclareAttackers)

		moves := legal.EnumerateFor(g, me.ID)
		dispatchAll(t, g, me.ID, moves)
		if countKind(moves, legal.KindPass) != 0 {
			t.Fatalf("pass offered while the goaded Sentinels can attack: %v", labels(moves))
		}
		got := attackTargetsOffered(t, moves, sentinels)
		if len(got) != 1 || !got[other.ID.String()] {
			t.Fatalf("offered %v: want only the player who attacked you and didn't goad it", got)
		}
	})
	t.Run("nobody attacked you", func(t *testing.T) {
		g := newTable(t)
		s := g.Turn.ActiveSeat
		me, goader := g.Seats[s], g.Seats[(s+1)%4]
		clearHand(me)
		sentinels := readySentinels(g, me)
		g.WithWriteLock(func() { g.GoadForEffect(sentinels, goader.ID) })
		advanceTo(t, g, game.StepDeclareAttackers)

		moves := legal.EnumerateFor(g, me.ID)
		if countKind(moves, legal.KindPass) != 1 {
			t.Errorf("pass withheld although the goaded Sentinels can't attack anyone: %v", labels(moves))
		}
	})
}
