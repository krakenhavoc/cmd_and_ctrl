package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const secondDoctorOracle = "8a6eda31-2791-4def-b9e7-3adde155a4f0"

// answerSecondDoctor settles the end-step trigger, answering each
// seat's "draw a card?" from `yes` (default no) and recording the order
// the questions arrived in.
func answerSecondDoctor(t *testing.T, g *game.Game, yes map[uuid.UUID]bool) []uuid.UUID {
	t.Helper()
	var asked []uuid.UUID
	for i := 0; i < 64; i++ {
		var open *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceConfirm {
				open = c
				break
			}
		}
		if open != nil {
			asked = append(asked, open.Chooser)
			if err := g.ResolveConfirm(open.ID, open.Chooser, yes[open.Chooser]); err != nil {
				t.Fatalf("ResolveConfirm: %v", err)
			}
			continue
		}
		if stackFullyEmpty(g) {
			return asked
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("The Second Doctor's trigger never settled")
	return nil
}

func secondDoctorGame(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	mhTickingClock(t)
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Second Doctor", TypeLine: "Legendary Creature — Time Lord Doctor",
		OracleID: secondDoctorOracle, Power: 2, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	return g, me
}

func TestSecondDoctorEveryPlayerHasNoMaximumHandSize(t *testing.T) {
	g, _ := secondDoctorGame(t)
	for _, p := range g.Seats {
		if got := mhMax(g, p); got != game.NoMaxHandSize {
			t.Errorf("%s: maximum = %d, want no maximum", p.Name, got)
		}
	}
}

// Everyone is asked, the controller first and the rest in turn order.
// Those who said yes draw; each opponent among them can't attack the
// controller or their permanents during their next turn, and nobody
// else is touched.
func TestSecondDoctorAsksEveryoneAndRestrictsOpponentsWhoDrew(t *testing.T) {
	g, me := secondDoctorGame(t)
	seatIdx := g.Turn.ActiveSeat
	seats := func(k int) *game.Player { return g.Seats[(seatIdx+k)%len(g.Seats)] }
	a, b, c := seats(1), seats(2), seats(3)
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLoyalty: 3},
	})

	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Hand.Size()
	}
	advanceToEndStepOf(t, g, seatIdx)
	asked := answerSecondDoctor(t, g, map[uuid.UUID]bool{me.ID: true, a.ID: true, b.ID: false, c.ID: true})

	want := []uuid.UUID{me.ID, a.ID, b.ID, c.ID}
	if len(asked) != 4 {
		t.Fatalf("asked %v, want %v", asked, want)
	}
	for i := range want {
		if asked[i] != want[i] {
			t.Fatalf("asked %v, want %v", asked, want)
		}
	}
	for _, p := range []*game.Player{me, a, c} {
		if got := p.Hand.Size(); got != before[p.ID]+1 {
			t.Errorf("%s hand = %d, want %d", p.Name, got, before[p.ID]+1)
		}
	}
	if got := b.Hand.Size(); got != before[b.ID] {
		t.Errorf("the player who declined drew: hand = %d, want %d", got, before[b.ID])
	}

	restricted := func(p *game.Player) bool {
		for _, s := range p.Statics {
			if s.CantAttack.Protected == me.ID {
				return true
			}
		}
		return false
	}
	if restricted(me) || restricted(b) {
		t.Error("the controller or a player who declined was restricted")
	}
	if !restricted(a) || !restricted(c) {
		t.Error("an opponent who drew was not restricted")
	}

	// During a's next turn they can't attack me or my planeswalker, but
	// they can attack b; b (who declined) can attack me on their turn.
	attacker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Raider", TypeLine: "Creature — Test",
		Power: 2, Toughness: 2, Owner: a.ID, Controller: a.ID,
	})
	for g.Turn.ActiveSeat != (seatIdx+1)%len(g.Seats) || g.Turn.Step != game.StepDeclareAttackers {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	for _, target := range []uuid.UUID{me.ID, walker} {
		err := g.Clone().DeclareAttacker(attacker, target)
		var pe *game.PlayerCantAttackError
		if !errors.As(err, &pe) {
			t.Fatalf("attack at %s: err = %v, want *PlayerCantAttackError", target, err)
		}
	}
	if err := g.Clone().DeclareAttacker(attacker, b.ID); err != nil {
		t.Errorf("attack at a player the Doctor doesn't protect: %v", err)
	}
}

// Nobody drawing means nobody restricted.
func TestSecondDoctorWithNoTakersRestrictsNobody(t *testing.T) {
	g, _ := secondDoctorGame(t)
	advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	answerSecondDoctor(t, g, nil)
	for _, p := range g.Seats {
		if len(p.Statics) != 0 {
			t.Errorf("%s has statics %+v", p.Name, p.Statics)
		}
	}
}
