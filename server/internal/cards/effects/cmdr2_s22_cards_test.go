package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s22_cards_test.go — the "draw-matters, hand size and flash
// grants" Commander batch (#5368). Each test plays the real catalog entry.

const (
	s22TriskOracle     = "ef238ba8-acf0-4e62-9abd-5291886a32e8"
	s22TishanaOracle   = "479dd04b-0048-446a-9202-a4dc9472e5c3"
	s22KefnetOracle    = "caf626a8-7abc-496b-9d3a-23194ce70e01"
	s22HoofprintOracle = "17739f55-486b-46ea-8785-70cc2bd2d872"
	s22JolraelOracle   = "f8640899-2458-40f2-98a3-35c402ee9a7a"
	s22KnowledgeOracle = "8806ee05-10d5-4492-b672-2f6bcab94be0"
	s22AlandraOracle   = "97bb1589-2941-4a75-a5f6-916871e51e9e"
	s22KianneOracle    = "7fc9e800-da66-4dda-a56d-79daec9e943e"
	s22PhantomOracle   = "f07db0f7-7d25-4395-9af3-6b0ca9197b61"
	s22SphinxOracle    = "e6284fb3-7cf8-4730-b156-10085b70b0e8"
	s22ProgSphxOracle  = "92c470c9-1373-4f3a-9805-5e9613874793"
)

func TestS22EveryBuiltCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s22TriskOracle:    "Triskaidekaphile",
		s22SphinxOracle:   "Sphinx of Enlightenment",
		s22ProgSphxOracle: "Prognostic Sphinx",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull {
			t.Errorf("%s: name %q, completeness %s", name, spec.Name, spec.Completeness)
		}
	}
}

// s22FillHand sets a seat's hand to exactly n filler cards.
func s22FillHand(p *game.Player, n int) {
	for p.Hand.Size() > 0 {
		p.Hand.Remove(p.Hand.Cards[0].InstanceID)
	}
	for i := 0; i < n; i++ {
		p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
	}
}

// --- Triskaidekaphile ---

func TestS22TriskaidekaphileWinsWithExactlyThirteenAtUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	pushPermanentForTest(g, me.ID, "Triskaidekaphile", s22TriskOracle, "Creature — Human Wizard")
	s22FillHand(me, 13)
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if g.State != game.StateEnded || g.Outcome == nil || g.Outcome.Winner != me.ID {
		t.Fatalf("state %s outcome %+v, want a win for the controller", g.State, g.Outcome)
	}
}

func TestS22TriskaidekaphileDoesNotWinAtTwelveOrFourteen(t *testing.T) {
	for _, n := range []int{12, 14} {
		g := newCatalogGame(t)
		me := g.Seats[1]
		pushPermanentForTest(g, me.ID, "Triskaidekaphile", s22TriskOracle, "Creature — Human Wizard")
		s22FillHand(me, n)
		advanceToUpkeepOf(t, g, 1)
		passPriorityAroundTable(t, g)
		if g.State == game.StateEnded {
			t.Errorf("hand of %d won the game", n)
		}
	}
}

func TestS22TriskaidekaphileInterveningIfRechecksOnResolution(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	pushPermanentForTest(g, me.ID, "Triskaidekaphile", s22TriskOracle, "Creature — Human Wizard")
	s22FillHand(me, 13)
	advanceToUpkeepOf(t, g, 1)
	// Triggered as the upkeep began; a card leaves the hand before it resolves.
	s22FillHand(me, 12)
	passPriorityAroundTable(t, g)
	if g.State == game.StateEnded {
		t.Fatal("won with twelve cards in hand as the trigger resolved")
	}
}

func TestS22TriskaidekaphileHasNoMaximumHandSizeAndDrawsForFourMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	phile := pushCatalogPermanent(g, me.ID, "Triskaidekaphile", "Creature — Human Wizard", s22TriskOracle, false)
	if got := g.EffectiveMaxHandSizeLocked(me); got != game.NoMaxHandSize {
		t.Errorf("max hand size %d, want no limit", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, phile, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without mana")
	}
	stdB23Mana(t, g, me, "{U}{U}{U}{U}")
	before := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, phile, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), before+1)
	}
}

// --- Sphinx of Enlightenment ---

func TestS22SphinxOfEnlightenmentDrawsForTargetOpponentAndYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	if opp == me {
		opp = g.Seats[2]
	}
	meBefore, oppBefore := me.Hand.Size(), opp.Hand.Size()
	castCreatureCard(t, g, "Sphinx of Enlightenment", "Creature — Sphinx", s22SphinxOracle, 5, 5)
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Hand.Size() != oppBefore+1 {
		t.Errorf("opponent hand %d, want %d", opp.Hand.Size(), oppBefore+1)
	}
	if me.Hand.Size() != meBefore+3 {
		t.Errorf("my hand %d, want %d (the Sphinx left it, three drawn)", me.Hand.Size(), meBefore+3)
	}
}

// --- Prognostic Sphinx ---

func TestS22PrognosticSphinxDiscardGivesHexproofAndTapsIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	sphinx := pushCatalogPermanent(g, me.ID, "Prognostic Sphinx", "Creature — Sphinx", s22ProgSphxOracle, false)
	if err := g.ActivateCatalogAbility(me.ID, sphinx, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without discarding a card")
	}
	card := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: card, Name: "Filler", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, sphinx, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{card}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Hand.Contains(card) {
		t.Fatal("the discard was not paid")
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, sphinx, "hexproof") {
		t.Error("no hexproof")
	}
	if !dskRT2Tapped(g, sphinx) {
		t.Error("not tapped")
	}
}

func TestS22PrognosticSphinxScriesThreeWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	seedLibrary(me, "One", "Two", "Three", "Four")
	sphinx := pushCatalogPermanent(g, me.ID, "Prognostic Sphinx", "Creature — Sphinx", s22ProgSphxOracle, false)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(sphinx, them.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no scry prompt: %+v", g.PendingChoices)
	}
	if len(c.ScryCards) != 3 {
		t.Fatalf("scry 3 offered %d cards", len(c.ScryCards))
	}
}
