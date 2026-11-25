package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// each_player_exiles_board_and_hand_test.go — #6265: Descent into
// Madness's "each player exiles X permanents they control and/or cards
// from their hand" (CR 701.13a). One count from two zones, split as each
// player chooses; asked in APNAP order (CR 101.4); exiled together once
// the last player has chosen.

const testDescentOracle = "test-6265-descent"

func registerDescent(t *testing.T) {
	t.Helper()
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: testDescentOracle, Name: "Descent Test",
		OnResolve: effects.ResolveSteps(effects.Instruction(func(ctx *effects.Context) error {
			return effects.EachPlayerExilesPermanentsAndOrHandCards(ctx, 3)
		})),
	})
}

func pushPermanent(g *game.Game, controller uuid.UUID, name string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Artifact", Owner: controller, Controller: controller,
	})
}

func pushHand(g *game.Game, p *game.Player, name string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
	})
	return id
}

func inExile(g *game.Game, _, id uuid.UUID) bool { return effectstest.InExile(g, id) }

// TestEachPlayerExilesXFromPermanentsAndHand is the acceptance: with X = 3,
// a player with 2 permanents and 3 hand cards exiles exactly 3, split as
// they choose; the controller chooses first, then the next player; and
// nothing is exiled until every player has chosen.
func TestEachPlayerExilesXFromPermanentsAndHand(t *testing.T) {
	registerDescent(t)
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	for _, p := range g.Seats {
		if p.ID != me.ID && p.ID != opp.ID {
			g.WithWriteLock(func() { p.Hand.Cards = nil })
		}
	}
	g.WithWriteLock(func() { me.Hand.Cards, opp.Hand.Cards = nil, nil })

	p1, p2 := pushPermanent(g, me.ID, "Mind Stone"), pushPermanent(g, me.ID, "Sol Ring")
	h1, h2, h3 := pushHand(g, me, "Shock"), pushHand(g, me, "Opt"), pushHand(g, me, "Duress")
	op := pushPermanent(g, opp.ID, "Ornithopter")
	oh := pushHand(g, opp, "Ponder")

	effectstest.CastCatalogSpell(t, g, "Descent Test", "Sorcery", testDescentOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	mine := effectstest.LatestChooseCards(g, me.ID)
	if mine == nil {
		t.Fatal("the controller was not asked")
	}
	if mine.ChooseMin != 3 || mine.ChooseMax != 3 || len(mine.ChooseCards) != 5 {
		t.Fatalf("controller's pick: %d..%d of %d, want exactly 3 of their 2 permanents and 3 hand cards",
			mine.ChooseMin, mine.ChooseMax, len(mine.ChooseCards))
	}
	if err := g.ResolveChooseCards(mine.ID, me.ID, []uuid.UUID{p1, h1}); err == nil {
		t.Fatal("two picks accepted, want exactly three")
	}
	if err := g.ResolveChooseCards(mine.ID, me.ID, []uuid.UUID{p1, h1, h2}); err != nil {
		t.Fatalf("controller picks: %v", err)
	}
	if inExile(g, me.ID, p1) {
		t.Fatal("a pick was exiled before every player had chosen")
	}
	theirs := effectstest.LatestChooseCards(g, opp.ID)
	if theirs == nil {
		t.Fatal("the next player was not asked")
	}
	if theirs.ChooseMin != 2 || theirs.ChooseMax != 2 {
		t.Errorf("a player with two objects must choose both, got %d..%d", theirs.ChooseMin, theirs.ChooseMax)
	}
	if err := g.ResolveChooseCards(theirs.ID, opp.ID, []uuid.UUID{op, oh}); err != nil {
		t.Fatalf("opponent picks: %v", err)
	}
	for _, c := range []struct {
		owner, id uuid.UUID
		exiled    bool
	}{{me.ID, p1, true}, {me.ID, h1, true}, {me.ID, h2, true}, {me.ID, p2, false}, {me.ID, h3, false}, {opp.ID, op, true}, {opp.ID, oh, true}} {
		if got := inExile(g, c.owner, c.id); got != c.exiled {
			t.Errorf("card %v exiled = %v, want %v", c.id, got, c.exiled)
		}
	}
}

// TestEachPlayerExilesAsksTheControllerFirst: the first prompt queued is
// the controller's (APNAP, CR 101.4).
func TestEachPlayerExilesAsksTheControllerFirst(t *testing.T) {
	registerDescent(t)
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	for _, p := range g.Seats {
		pushPermanent(g, p.ID, "Mind Stone")
	}
	effectstest.CastCatalogSpell(t, g, "Descent Test", "Sorcery", testDescentOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var first *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards {
			first = c
			break
		}
	}
	if first == nil || first.Chooser != me.ID {
		t.Fatalf("first choose_cards prompt = %+v, want the controller's", first)
	}
}
