package aiseat_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// super_hero_civil_war_bot_test.go — #4616, whole-engine: the real
// catalog The Super Hero Civil War, the real enumerator, the real view
// and the real heuristic.
//
// Chapter I is "gain control of up to two target creatures with total
// mana value 6 or less". The bot answers the prompt from the moves the
// enumerator offers, so it never names a pair over the total (the
// engine would refuse it, and a refused answer is the #544 wedge), and
// it does take something: the opponent's board has creatures that fit.
//
// Not behind AISEAT_GAME_TESTS: one cast and one prompt, milliseconds.

const botSuperHeroCivilWarOracle = "11d74dc2-8c11-4d4b-b727-8216d2eb6b4e"

func TestBotStealsWithinTheSuperHeroCivilWarTotal(t *testing.T) {
	g := newSettledTable(t, 4616)
	advanceToStep(t, g, game.StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	everyone := seatIDs(g)
	mv := map[uuid.UUID]int{}
	var dragon, saga uuid.UUID
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		me.Command.Cards = nil
		for _, c := range []struct {
			name, cost string
			pt, v      int
		}{
			{"Test Dragon", "{6}{G}", 7, 7},
			{"Test Titan", "{5}{G}", 6, 6},
			{"Test Knight", "{3}{G}", 4, 4},
			{"Test Squire", "{1}{G}", 2, 2},
		} {
			card := game.Card{InstanceID: uuid.New(), Name: c.name, TypeLine: "Creature — Beast",
				ManaCost: c.cost, Power: c.pt, Toughness: c.pt, Owner: opp.ID, Controller: opp.ID}
			card.AddKnowersAll(everyone)
			g.Battlefield.PushTop(card)
			mv[card.InstanceID] = c.v
			if c.v == 7 {
				dragon = card.InstanceID
			}
		}
		for _, l := range []struct{ name, sub string }{
			{"Mountain", "Mountain"}, {"Mountain", "Mountain"}, {"Mountain", "Mountain"},
			{"Mountain", "Mountain"}, {"Plains", "Plains"},
		} {
			land := game.Card{InstanceID: uuid.New(), Name: l.name, TypeLine: "Basic Land — " + l.sub,
				Owner: me.ID, Controller: me.ID}
			land.AddKnowersAll(everyone)
			g.Battlefield.PushTop(land)
		}
		s := game.Card{InstanceID: uuid.New(), Name: "The Super Hero Civil War", OracleID: botSuperHeroCivilWarOracle,
			TypeLine: "Enchantment — Saga", ManaCost: "{3}{R}{W}", Owner: me.ID, Controller: me.ID}
		s.AddKnowersAll(everyone)
		me.Hand.Cards = append(me.Hand.Cards, s)
		saga = s.InstanceID
	})
	if err := g.CastSpell(me.ID, saga, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("cast The Super Hero Civil War: %v", err)
	}
	for i := 0; i < 8 && !owesChoice(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass: %v", err)
		}
	}
	if !owesChoice(g, me.ID) {
		t.Fatal("chapter I asked the bot for no targets")
	}

	driveChoices(t, g, heuristic.New(), me.ID, 4)

	var picked []uuid.UUID
	g.ReadSnapshot(func() {
		for _, it := range g.StackMeta {
			if it.Kind == game.StackItemTriggered && it.SourceCardID == saga {
				for _, r := range it.Targets {
					if r.Kind == game.TargetCard {
						picked = append(picked, r.ID)
					}
				}
			}
		}
	})
	if len(picked) == 0 {
		t.Fatal("the bot stole nothing, with creatures that fit on the opponent's side")
	}
	sum := 0
	for _, id := range picked {
		if id == dragon {
			t.Error("the bot named the mana value 7 creature, which fits no total of 6")
		}
		sum += mv[id]
	}
	if sum > 6 {
		t.Errorf("the bot's picks total mana value %d, over 6", sum)
	}

	// The chapter is an ability, which lives in StackMeta, not in the
	// stack zone's cards.
	for i := 0; i < 8 && len(g.StackMeta) > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass: %v", err)
		}
	}
	for _, id := range picked {
		var controller uuid.UUID
		g.ReadSnapshot(func() {
			for _, c := range g.Battlefield.Cards {
				if c.InstanceID == id {
					controller = c.Controller
				}
			}
		})
		if controller != me.ID {
			t.Errorf("the creature the bot picked (%s) is not under its control after chapter I", id)
		}
	}
}
