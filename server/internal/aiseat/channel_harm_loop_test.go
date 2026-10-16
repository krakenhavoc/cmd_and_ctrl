package aiseat_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// channel_harm_loop_test.go — #2040: two seats, each with a Channel Harm
// aimed at the other's creature. A prevented hit makes one Channel Harm
// deal damage the other prevents, and so on, until a player declines
// (the ruling). A policy that takes every confirm would never decline,
// and the table would stall. The second Channel Harm's question carries
// decline_ends_loop, and the heuristic declines it.

const oracleChannelHarmBot = "4b3d1cfb-9f1d-4e4e-a15a-2e248ff5bdfc"

func chBattlefieldCreature(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	c := game.NewCard(name, uuid.Nil)
	c.TypeLine, c.Power, c.Toughness = "Creature — Bear", 2, 6
	c.Owner, c.Controller = owner, owner
	g.WithWriteLock(func() {
		c.AddKnowersAll(seatIDs(g))
		g.Battlefield.PushTop(c)
	})
	return c.InstanceID
}

// chCast has `who` cast Channel Harm aimed at `target`, from a pool
// holding exactly its cost.
func chCast(t *testing.T, g *game.Game, who *game.Player, target uuid.UUID) {
	t.Helper()
	c := game.NewCard("Channel Harm", uuid.Nil)
	c.TypeLine, c.ManaCost, c.OracleID = "Instant", "{5}{W}", oracleChannelHarmBot
	id := tbHandCard(g, who, c)
	if err := g.AddManaForEffect(who.ID, uuid.Nil, "{W}{W}{W}{W}{W}{W}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(who.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err != nil {
		t.Fatalf("cast Channel Harm: %v", err)
	}
}

func chStackEmpty(g *game.Game) bool {
	empty := false
	g.ReadSnapshot(func() { empty = g.Stack.Size() == 0 })
	return empty
}

func TestHeuristicEndsAChannelHarmLoop(t *testing.T) {
	g := newSettledTable(t, 11)
	advanceToStep(t, g, game.StepPrecombatMain)
	a := g.Seats[g.Turn.ActiveSeat]
	b := g.Seats[1-g.Turn.ActiveSeat]
	g.WithWriteLock(func() { b.Hand.Cards = nil })
	aBear := chBattlefieldCreature(g, a.ID, "A's Bear")
	bBear := chBattlefieldCreature(g, b.ID, "B's Bear")

	chCast(t, g, a, bBear)
	if err := g.PassPriority(); err != nil { // to B, with A's Channel Harm on the stack
		t.Fatal(err)
	}
	chCast(t, g, b, aBear)
	for i := 0; i < 8 && !chStackEmpty(g); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass %d: %v", i, err)
		}
	}
	if !chStackEmpty(g) {
		t.Fatal("the two Channel Harms did not resolve")
	}

	// A's bear hits B: B's Channel Harm prevents it and asks B, whose
	// yes is prevented by A's Channel Harm, which asks A, and so on.
	life := b.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(aBear, b.ID, 2); err != nil {
			t.Fatal(err)
		}
	})
	g.SettleResolution()

	pol := heuristic.New()
	var taken []string
	for step := 0; ; step++ {
		var seat uuid.UUID
		for _, p := range []*game.Player{a, b} {
			if owesChoice(g, p.ID) {
				seat = p.ID
			}
		}
		if seat == uuid.Nil {
			break
		}
		if step >= 12 {
			t.Fatalf("the Channel Harm loop did not end in 12 answers: %v", taken)
		}
		moves := legal.EnumerateFor(g, seat)
		if len(moves) == 0 {
			t.Fatalf("step %d: a seat owes a choice and was offered nothing: %s", step, openChoices(g, seat))
		}
		d, err := pol.Decide(context.Background(), aiseat.Input{
			View: protocol.ViewOfGameFor(g, seat.String()), Seat: seat, Moves: moves,
		})
		if err != nil || d.Index < 0 || d.Index >= len(moves) {
			t.Fatalf("step %d: decision %+v, err %v", step, d, err)
		}
		mv := moves[d.Index]
		if err := actions.Dispatch(g, actions.Action{
			Type: actions.Type(mv.Type), Player: mv.Player, Caller: seat, Params: mv.Params,
		}); err != nil {
			t.Fatalf("step %d: the engine refused %q: %v", step, mv.Label, err)
		}
		taken = append(taken, mv.Label)
	}
	if len(taken) != 2 {
		t.Errorf("answers %v: want B's yes, then A's no", taken)
	}
	if b.Life != life {
		t.Errorf("B's life %d, want %d: the hit was prevented", b.Life, life)
	}
	var aDamage, bDamage int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			switch c.InstanceID {
			case aBear:
				aDamage = c.DamageMarked
			case bBear:
				bDamage = c.DamageMarked
			}
		}
	})
	if aDamage != 0 || bDamage != 0 {
		t.Errorf("damage: A's bear %d, B's bear %d; want 0 and 0 (each Channel Harm prevented the other's)", aDamage, bDamage)
	}
}
