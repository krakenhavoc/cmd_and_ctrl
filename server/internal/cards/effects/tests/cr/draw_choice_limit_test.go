package testscr

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// draw_choice_limit_test.go — #4016, CR 121.2b and CR 121.3: a player a
// draw limit stops drawing can't CHOOSE to draw, so a "you may draw" is
// not offered to them, and a "draw up to N" offers no more than they
// may still draw.

// An optional trigger (Geth's Grimoire's "you may draw a card"): asked
// while its controller may still draw, and not asked once Spirit of the
// Labyrinth's one draw is used; the trigger then resolves doing nothing.
func TestOptionalDrawIsNotOfferedToAPlayerWhoCantDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	// The game starts in the active player's draw step, after its draw,
	// so the Grimoire goes to the seat that has drawn nothing yet.
	active, me := g.Seats[0], g.Seats[1]
	effectstest.B21Push(g, me.ID, "Geth's Grimoire", "Artifact — Book", effectstest.B21GethsGrimoireOracle, 0, 0)
	dlPut(g, active.ID, "Spirit of the Labyrinth", dlSpirit, "Creature — Spirit")

	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(active.ID, 1) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.DrawN(g, me, 1)
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(active.ID, 1) })
	effectstest.SettleTriggers(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("a player with no draws left was asked \"draw a card?\" (CR 121.3)")
	}
	if !effectstest.StackFullyEmpty(g) {
		t.Fatal("the trigger did not finish resolving without its question")
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d, want no draw", hand, me.Hand.Size())
	}
}

// CR 121.2b: "you may draw two cards" is a choice to draw multiple
// cards, which a player under Narset can't take even before drawing any.
func TestOptionalDrawOfTwoIsNotOfferedUnderAOneCardLimit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp, me := g.Seats[0], g.Seats[1]
	effectstest.B21Push(g, me.ID, "Consecrated Sphinx", "Creature — Sphinx", effectstest.ConsecratedSphinxOracle, 4, 6)
	dlNarsetOn(g, opp.ID)
	hand := me.Hand.Size()

	effectstest.DrawN(g, opp, 1)
	effectstest.SettleTriggers(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("\"draw two cards?\" was offered to a player who may draw only one (CR 121.2b)")
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d, want no draw", hand, me.Hand.Size())
	}

	// Without the limit the same trigger asks.
	g2 := effectstest.NewCatalogGame(t)
	opp2, me2 := g2.Seats[0], g2.Seats[1]
	effectstest.B21Push(g2, me2.ID, "Consecrated Sphinx", "Creature — Sphinx", effectstest.ConsecratedSphinxOracle, 4, 6)
	effectstest.DrawN(g2, opp2, 1)
	if effectstest.AwaitTriggerPrompt(t, g2, me2.ID) == nil {
		t.Fatal("with no limit, Consecrated Sphinx asked nothing")
	}
}

// A hand-built prompt with a chooser (Edric's "its controller may
// draw") reads the chooser's limit: Maralen's "can't draw cards".
func TestChooserPromptedDrawIsNotOfferedUnderCantDrawCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dlPut(g, me.ID, "Maralen of the Mornsong", dlMaralen, "Legendary Creature — Elf Wizard")
	effectstest.PushPermanentForTest(g, me.ID, "Fecundity", effectstest.G2FecundityOracle, "Enchantment")
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	hand := opp.Hand.Size()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.SettleTriggers(t, g)
	if effectstest.LatestTriggerPrompt(g, opp.ID) != nil {
		t.Fatal("Fecundity asked a player who can't draw cards")
	}
	if opp.Hand.Size() != hand {
		t.Errorf("hand %d → %d, want no draw", hand, opp.Hand.Size())
	}
}

// CR 121.3a: Combustible Gearhulk's opponent chooses whether ITS
// CONTROLLER draws three. Under the opponent's Narset that controller
// can't draw three, so the opponent isn't asked and the "doesn't"
// branch happens: mill three, damage equal to their mana value.
func TestAnOpponentCantChooseADrawTheDrawerCantMake(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	// The Gearhulk's controller has drawn nothing this turn: one draw is
	// left, and three are offered.
	opp, me := g.Seats[0], g.Seats[1]
	for _, cost := range []string{"{3}", "{2}", "{1}"} {
		me.Library.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Cost " + cost, TypeLine: "Artifact",
			ManaCost: cost, Owner: me.ID, Controller: me.ID,
		})
	}
	dlNarsetOn(g, opp.ID)
	life, graves, hand := opp.Life, me.Graveyard.Size(), me.Hand.Size()

	gearhulk := effectstest.B12Push(g, me.ID, "Combustible Gearhulk", "Artifact Creature — Construct", effectstest.CombustibleGearhulkOracle, 6, 6)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: gearhulk}) })
	effectstest.B04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTarget(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, opp.ID) != nil {
		t.Fatal("the opponent was asked to let a player who can draw only one card draw three (CR 121.2b, 121.3a)")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d, want no draw", hand, me.Hand.Size())
	}
	if me.Graveyard.Size() != graves+3 {
		t.Errorf("they mill three: graveyard %d → %d", graves, me.Graveyard.Size())
	}
	if opp.Life != life-6 {
		t.Errorf("damage is the milled cards' total mana value (6): life %d → %d", life, opp.Life)
	}
}

// Each player's "may draw" in a chain (Kwain) skips the one who can't
// and still asks the rest.
func TestAChainOfMayDrawsSkipsOnlyThePlayerWhoCantDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dlNarsetOn(g, me.ID)
	effectstest.DrawN(g, opp, 1)
	ctx := effects.NewContext(g, &game.StackItem{Controller: me.ID})
	g.WithWriteLock(func() {
		if err := effects.B20KwainAsk(ctx, []uuid.UUID{opp.ID, me.ID}, nil); err != nil {
			t.Fatalf("b20KwainAsk: %v", err)
		}
	})
	if effectstest.LatestConfirmFor(g, opp.ID) != nil {
		t.Error("the opponent with no draws left was asked")
	}
	if effectstest.LatestConfirmFor(g, me.ID) == nil {
		t.Error("Narset's controller, who may draw, was not asked")
	}
}

// The readers: a choice to draw n, the ceiling of "draw up to n", and
// the badge's status.
func TestDrawChoiceReaders(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dlNarsetOn(g, me.ID)
	g.WithWriteLock(func() {
		if !g.CanChooseToDrawForEffect(opp.ID, 1) || g.CanChooseToDrawForEffect(opp.ID, 2) {
			t.Error("under Narset with no draw yet: one card may be chosen, two may not")
		}
		if got := g.DrawChoiceCeilingForEffect(opp.ID, 3); got != 1 {
			t.Errorf("\"draw up to three\" ceiling = %d, want 1", got)
		}
		if !g.CanChooseToDrawForEffect(me.ID, 7) || g.DrawChoiceCeilingForEffect(me.ID, 3) != 3 {
			t.Error("Narset's controller is not limited")
		}
		st, ok := g.DrawLimitStatusLocked(opp)
		if !ok || st.Max != 1 || st.Left != 1 || !slices.Equal(st.Sources, []string{"Narset, Parter of Veils"}) {
			t.Errorf("status = %+v (%v), want max 1, left 1, from Narset", st, ok)
		}
		if _, ok := g.DrawLimitStatusLocked(me); ok {
			t.Error("Narset's controller has a draw limit")
		}
	})
	effectstest.DrawN(g, opp, 1)
	g.WithWriteLock(func() {
		if g.CanChooseToDrawForEffect(opp.ID, 1) || g.DrawChoiceCeilingForEffect(opp.ID, 2) != 0 {
			t.Error("after the one draw, no choice to draw is left")
		}
	})
}
