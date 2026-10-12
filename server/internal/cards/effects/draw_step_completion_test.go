package effects

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestTeferisAgelessInsightDoublesOnlyLaterDrawsInEachDrawStep(t *testing.T) {
	g := newCatalogGame(t)
	p := g.Seats[1]
	bfCardWithOracle(g, p, "Teferi's Ageless Insight", "Legendary Enchantment", "1c9e1f75-73f0-4846-b53f-458a0984b1bb")
	before := p.Hand.Size()
	advanceToDrawStepOfSeat(t, g, 1)
	if p.Hand.Size() != before+1 {
		t.Fatal("Insight doubled the first draw in its controller's draw step")
	}
	if got := drawFor(t, g, p, 1); got != 2 {
		t.Fatalf("later draw gained %d cards, want 2", got)
	}
}

func TestNotionThiefExemptsFirstActualDrawThenStealsLaterDraws(t *testing.T) {
	g := newCatalogGame(t)
	thief, drawer := g.Seats[0], g.Seats[1]
	bfCardWithOracle(g, thief, "Notion Thief", "Creature — Human Rogue", b10NotionThiefOracle)
	before := drawer.Hand.Size()
	advanceToDrawStepOfSeat(t, g, 1)
	if drawer.Hand.Size() != before+1 {
		t.Fatal("the first draw in the opponent's draw step was stolen")
	}
	thiefBefore := thief.Hand.Size()
	if got := drawFor(t, g, drawer, 2); got != 0 || thief.Hand.Size() != thiefBefore+2 {
		t.Fatalf("later draws: drawer gained %d, thief gained %d; want 0 and 2", got, thief.Hand.Size()-thiefBefore)
	}
}

func TestNotionThiefLeavesFirstDoubledDescendantWithDrawer(t *testing.T) {
	g := newCatalogGame(t)
	thief, drawer := g.Seats[0], g.Seats[1]
	bfCardWithOracle(g, thief, "Notion Thief", "Creature — Human Rogue", b10NotionThiefOracle)
	bfCardWithOracle(g, drawer, "Thought Reflection", "Enchantment", thoughtReflectionOracle)
	before, thiefBefore := drawer.Hand.Size(), thief.Hand.Size()
	advanceToDrawStepOfSeat(t, g, 1)
	if drawer.Hand.Size() != before+1 || thief.Hand.Size() != thiefBefore+1 {
		t.Fatalf("doubled first draw: drawer gained %d, thief gained %d; want one each", drawer.Hand.Size()-before, thief.Hand.Size()-thiefBefore)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatal("one applicable replacement per descendant should not prompt")
	}
}

func TestDredgedFirstDrawKeepsNotionThiefExemption(t *testing.T) {
	g := newCatalogGame(t)
	thief, drawer := g.Seats[0], g.Seats[1]
	bfCardWithOracle(g, thief, "Notion Thief", "Creature — Human Rogue", b10NotionThiefOracle)
	gyCardWithOracle(drawer, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)
	advanceToDrawStepOfSeat(t, g, 1)
	answerOptional(t, g, drawer, true)
	if g.DrawnThisStep[drawer.ID] != 0 {
		t.Fatal("dredging incorrectly consumed the first draw")
	}
	if got := drawFor(t, g, drawer, 1); got != 1 {
		t.Fatalf("first actual draw was stolen after dredging: gained %d", got)
	}
	if got := drawFor(t, g, drawer, 1); got != 0 {
		t.Fatalf("later actual draw was not stolen: gained %d", got)
	}
}

func TestAdditionalDrawStepResetsNotionThiefExemption(t *testing.T) {
	g := newCatalogGame(t)
	thief, drawer := g.Seats[0], g.Seats[1]
	bfCardWithOracle(g, thief, "Notion Thief", "Creature — Human Rogue", b10NotionThiefOracle)
	advanceToDrawStepOfSeat(t, g, 1)
	g.WithWriteLock(func() {
		if !g.AddStepAfterCurrentForEffect(uuid.Nil, game.StepDraw) {
			t.Fatal("could not add a second draw step")
		}
	})
	before := drawer.Hand.Size()
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	if g.Turn.Step != game.StepDraw || drawer.Hand.Size() != before+1 || g.DrawnThisStep[drawer.ID] != 1 {
		t.Fatal("the second draw step did not get its own exempt first draw")
	}
}

func TestDoubledDrawOffersDredgeIndependentlyOnBothDescendants(t *testing.T) {
	g := newCatalogGame(t)
	p := g.Seats[1]
	bfCardWithOracle(g, p, "Thought Reflection", "Enchantment", thoughtReflectionOracle)
	imp := gyCardWithOracle(p, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)
	thug := gyCardWithOracle(p, "Golgari Thug", "Creature — Human Warrior", golgariThugOracle)
	before, libBefore := p.Hand.Size(), p.Library.Size()
	drawN(g, p, 1)
	orderFirst(t, g, p, "Thought Reflection")
	answerOptional(t, g, p, true)
	answerOptional(t, g, p, true)
	if !p.Hand.Contains(imp) || !p.Hand.Contains(thug) || p.Hand.Size() != before+2 || p.Library.Size() != libBefore-9 {
		t.Fatal("each expanded draw should independently dredge its selected card")
	}
	if g.DrawnThisStep[p.ID] != 0 {
		t.Fatal("dredged descendants are not actual draws")
	}
}

func TestExpandedDrawPauseCloneAndThenWaitForEverySibling(t *testing.T) {
	g := newCatalogGame(t)
	p := g.Seats[1]
	bfCardWithOracle(g, p, "Thought Reflection", "Enchantment", thoughtReflectionOracle)
	bfCardWithOracle(g, p, "Underrealm Lich", "Creature — Zombie Elf Shaman", underrealmLichOracle)
	then := game.RegisterDrawThen("test-expanded-draw-finish", func(g *game.Game, d game.DrawThen) error {
		return g.ChangePlayerLifeForEffect(uuid.Nil, d.Player, 1)
	})
	lifeBefore, handBefore := p.Life, p.Hand.Size()
	g.WithWriteLock(func() {
		if err := g.DrawNThenForEffect(p.ID, 2, game.DrawThen{Ref: then, Player: p.ID}); err != nil {
			t.Fatal(err)
		}
	})
	orderFirst(t, g, p, "Thought Reflection")
	snap := g.Clone()
	for _, table := range []*game.Game{g, snap} {
		seat := table.Seats[1]
		for i := 0; i < 4; i++ {
			// The second original draw opens another ordering window.
			if latestChoiceOfKindFor(table, game.PendingChoiceReplacementOrder, seat.ID) != nil {
				orderFirst(t, table, seat, "Thought Reflection")
			}
			if seat.Life != lifeBefore {
				t.Fatalf("then ran before descendant %d finished", i+1)
			}
			pick := chooseCardsChoiceFor(table, seat.ID)
			if pick == nil {
				t.Fatalf("missing descendant %d's Lich prompt", i+1)
			}
			answerChooseCards(t, table, seat.ID, pick.ChooseCards[0])
		}
		if seat.Life != lifeBefore+1 || seat.Hand.Size() != handBefore+4 || len(table.PendingChoices) != 0 {
			t.Fatal("expanded draw did not finish its then exactly once after four picks")
		}
	}
}

func TestOrcishBowmastersExemptsOnlyFirstCompletedDraw(t *testing.T) {
	g := newCatalogGame(t)
	drawer := g.Seats[1]
	bfCardWithOracle(g, g.Seats[0], "Orcish Bowmasters", "Creature — Orc Archer", orcishBowmastersOracle)
	advanceToDrawStepOfSeat(t, g, 1)
	if g.Stack.Size() != 0 || len(g.PendingChoices) != 0 {
		t.Fatal("Bowmasters fired on the first actual draw")
	}
	drawFor(t, g, drawer, 1)
	if g.Stack.Size() == 0 && len(g.PendingChoices) == 0 {
		t.Fatal("Bowmasters did not fire on the next draw in the same step")
	}
}
