package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prevention_choice_now_test.go — #2040 (ADR 0108 amendment of
// 2026-10-16): a prevention effect's additional effect that asks a
// question is answered as the damage is prevented. The rest of the
// damage instruction, the next follow-up, a modal spell's next bullet and
// the state-based actions after combat damage all wait for the answer,
// and a table waiting on it is still a restore point.

const kharnOracle = "3246eff5-8739-469f-a401-e5de927dd6bb"

// kharnPush puts Khârn the Betrayer (5/1) onto the battlefield under
// `controller`, sick-free.

// testBearDamageTail records the damage marked on its carried creature when
// the second instruction runs (#4253 retired DealDamageToPlayerThenForEffect's
// closure).
var (
	TestBearDamageAfter = -1
	TestBearDamageTail  = game.RegisterTail("test/4253-bear-damage-after", func(g *game.Game, o game.TailOutcome) error {
		if len(o.Tail.Carry) > 0 {
			TestBearDamageAfter = FindBattlefieldCardForTest(g, o.Tail.Carry[0]).DamageMarked
		}
		return nil
	})
)

func KharnPush(g *game.Game, owner, controller uuid.UUID) uuid.UUID {
	return ApaPush(g, owner, controller, game.Card{Name: "Khârn the Betrayer",
		TypeLine: "Legendary Creature — Astartes Berserker", OracleID: kharnOracle, Power: 5, Toughness: 1})
}

func KharnController(t *testing.T, g *game.Game, id uuid.UUID) uuid.UUID {
	t.Helper()
	c := FindBattlefieldCardForTest(g, id)
	if c == nil {
		t.Fatal("Khârn is not on the battlefield")
	}
	return c.Controller
}

// kharnAnswer has `chooser` answer the open "choose an opponent" prompt
// with `pick`, and fails when the prompt is not theirs.
func KharnAnswer(t *testing.T, g *game.Game, chooser, pick uuid.UUID) {
	t.Helper()
	c := LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, chooser)
	if c == nil {
		t.Fatalf("no option pick open for %s", chooser)
	}
	for i, o := range c.PickOptions {
		if o.Player == pick {
			if err := g.ResolveOptionPick(c.ID, chooser, i); err != nil {
				t.Fatal(err)
			}
			return
		}
	}
	t.Fatalf("%s is not offered: %+v", pick, c.PickOptions)
}

// The control change happens before the rest of the damage instruction:
// testKharnSeen and testKharnDealt are what testKharnSeenTail was told:
// the controller of the permanent dealt damage (as the tail ran) and the
// damage dealt. #4253 moved these tests onto the keyed damage door, whose
// continuation is a registered tail.
var (
	TestKharnSeen     uuid.UUID
	TestKharnDealt    int
	TestKharnSeenTail = game.RegisterTail("test/4253-kharn-seen", func(g *game.Game, o game.TailOutcome) error {
		if len(o.IDs) > 0 {
			if c, ok := g.LookupCardForEffect(o.IDs[0]); ok {
				TestKharnSeen = c.Controller
			}
		}
		TestKharnDealt = o.Amount
		return nil
	})
	// testDealOneMoreTail deals the told permanent 1 more damage from the
	// tail's Source.
	TestDealOneMoreTail = game.RegisterTail("test/4253-deal-one-more", func(g *game.Game, o game.TailOutcome) error {
		if len(o.IDs) == 0 {
			return nil
		}
		return g.DealDamageToCreatureForEffect(o.Tail.Source, o.IDs[0], 1)
	})
)
