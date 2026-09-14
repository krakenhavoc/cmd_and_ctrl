package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// look_at_top_test.go — "look at the top N cards of your library,
// then put them back in any order": the scry family's third member,
// and the cards built on it.
//
// The assertion that carries this file is the Sensei's Divining Top
// loop, which is the sprint's stated exit criterion: activate, see
// three cards nobody else can see, reorder them, confirm — then tap
// to draw the one you put on top and watch the Top itself land back
// on the library.

const (
	senseisTopOracle    = "13575cf9-65c1-4861-b21e-eb2155e07766"
	ponderOracle        = "02090581-61aa-4348-ad57-451be8ee91c2"
	crystalBallOracle   = "bd85fe4d-1d62-416f-ac2d-e287911c84e3"
	otherworldlyOracle  = "be668c2d-71ea-4346-8980-1fbf5e4cbed3"
	thoughtScourOracle  = "83101ba8-a569-4827-8c53-9ca0dfcd59a7"
	reliquaryTowerOracl = "c23e5b80-08d2-4e24-9908-fe2aa4f30f6f"
)

func lookAtTopChoiceFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceLookAtTop && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// --- Sensei's Divining Top: the exit criterion -----------------------

func TestSenseisTopLooksAtThreeAndReorders(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	top := pushCatalogPermanent(g, me.ID, "Sensei's Divining Top", "Artifact", senseisTopOracle, false)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	seedLibrary(me, "First", "Second", "Third")

	if err := g.ActivateCatalogAbility(me.ID, top, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the look-at ability: %v", err)
	}
	passPriorityAroundTable(t, g)

	c := lookAtTopChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("activating the Top queued no look-at prompt")
	}
	if len(c.ScryCards) != 3 {
		t.Fatalf("looked at %d cards, want 3", len(c.ScryCards))
	}
	// Nothing has moved.
	if got := libraryTopNames(me, 3); got[0] != "First" || got[2] != "Third" {
		t.Errorf("library is %v before the answer; the prompt must not move anything early", got)
	}
	libBefore := me.Library.Size()

	// Put the third card on top, keeping the other two in order.
	reordered := []uuid.UUID{c.ScryCards[2], c.ScryCards[0], c.ScryCards[1]}
	if err := g.ResolveLookAtTop(c.ID, me.ID, reordered); err != nil {
		t.Fatalf("ResolveLookAtTop: %v", err)
	}
	if got := libraryTopNames(me, 3); got[0] != "Third" || got[1] != "First" || got[2] != "Second" {
		t.Errorf("library top three are %v, want [Third First Second]", got)
	}
	if me.Library.Size() != libBefore {
		t.Errorf("library is %d, want %d — nothing leaves the library on a reorder", me.Library.Size(), libBefore)
	}
	if me.Graveyard.Size() != 0 {
		t.Error("a reorder put a card in the graveyard")
	}
}

// TestSenseisTopDrawsThenTucksItself — the order the card prints is
// the order that matters. "Draw a card, THEN put this on top" draws
// off the library as it stands; tucking first would draw the Top
// itself back into hand, which is a different card.
func TestSenseisTopDrawsThenTucksItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	top := pushCatalogPermanent(g, me.ID, "Sensei's Divining Top", "Artifact", senseisTopOracle, false)
	seedLibrary(me, "The Draw")
	handBefore := me.Hand.Size()

	if err := g.ActivateCatalogAbility(me.ID, top, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the tap ability: %v", err)
	}
	passPriorityAroundTable(t, g)

	if me.Hand.Size() != handBefore+1 {
		t.Fatalf("hand is %d, want %d", me.Hand.Size(), handBefore+1)
	}
	drawn, _ := me.Hand.Top()
	if drawn.Name != "The Draw" {
		t.Errorf("drew %q, want The Draw — the tuck must happen after the draw", drawn.Name)
	}
	if _, ok := battlefieldCard(g, top); ok {
		t.Error("the Top is still on the battlefield; it should have tucked itself")
	}
	if libTop, err := me.Library.Top(); err != nil || libTop.InstanceID != top {
		t.Error("the Top is not on top of its owner's library")
	}
	// A library is a hidden zone: the tuck must drop the card's
	// knowers, or its controller reads their own top card forever.
	if libTop, err := me.Library.Top(); err == nil && libTop.IsKnownTo(me.ID) {
		t.Error("the tucked Top is still known to its controller; a library is hidden")
	}
}

// TestSenseisTopIsLookAtNotReveal — same privacy rule as scry.
func TestSenseisTopIsLookAtNotReveal(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	top := pushCatalogPermanent(g, me.ID, "Sensei's Divining Top", "Artifact", senseisTopOracle, false)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	seedLibrary(me, "Secret One", "Secret Two", "Secret Three")

	if err := g.ActivateCatalogAbility(me.ID, top, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	find := func(v protocol.GameView) *protocol.PendingChoiceView {
		for i := range v.PendingChoices {
			if v.PendingChoices[i].Kind == "look_at_top" && v.PendingChoices[i].Chooser == me.ID.String() {
				return &v.PendingChoices[i]
			}
		}
		return nil
	}

	mine := find(protocol.ViewOfGameFor(g, me.ID.String()))
	if mine == nil {
		t.Fatal("the activating player can't see their own prompt")
	}
	if len(mine.Options) != 3 {
		t.Fatalf("chooser sees %d options, want 3", len(mine.Options))
	}
	if mine.Options[0].Name != "Secret One" {
		t.Errorf("chooser sees %q on top, want Secret One", mine.Options[0].Name)
	}

	theirs := find(protocol.ViewOfGameFor(g, opp.ID.String()))
	if theirs != nil {
		// The COUNT is public — "look at the top three" is printed —
		// but the identities are not.
		for _, o := range theirs.Options {
			if o.Name != "" || o.KnownByYou {
				t.Error("an opponent can read the top of my library")
			}
		}
	}
}

func TestLookAtTopRejectsBadAnswers(t *testing.T) {
	setup := func(t *testing.T) (*game.Game, *game.Player, *game.PendingChoice) {
		t.Helper()
		g := newCatalogGame(t)
		me := g.Seats[0]
		top := pushCatalogPermanent(g, me.ID, "Sensei's Divining Top", "Artifact", senseisTopOracle, false)
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
		seedLibrary(me, "A", "B", "C")
		if err := g.ActivateCatalogAbility(me.ID, top, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		passPriorityAroundTable(t, g)
		c := lookAtTopChoiceFor(g, me.ID)
		if c == nil {
			t.Fatal("no look-at prompt")
		}
		return g, me, c
	}

	t.Run("a card left out", func(t *testing.T) {
		g, me, c := setup(t)
		if err := g.ResolveLookAtTop(c.ID, me.ID, c.ScryCards[:2]); err == nil {
			t.Error("a partial permutation was accepted; every looked-at card must be placed")
		}
		if got := libraryTopNames(me, 3); got[0] != "A" {
			t.Error("a rejected answer reordered the library")
		}
	})

	t.Run("a card listed twice", func(t *testing.T) {
		g, me, c := setup(t)
		dup := []uuid.UUID{c.ScryCards[0], c.ScryCards[0], c.ScryCards[1]}
		if err := g.ResolveLookAtTop(c.ID, me.ID, dup); err == nil {
			t.Error("a duplicate was accepted")
		}
	})

	t.Run("someone else answering", func(t *testing.T) {
		g, me, c := setup(t)
		if err := g.ResolveLookAtTop(c.ID, g.Seats[1].ID, c.ScryCards); err == nil {
			t.Error("another player answered my prompt")
		}
		_ = me
	})

	t.Run("the scry resolver refuses it", func(t *testing.T) {
		g, me, c := setup(t)
		if err := g.ResolveScry(c.ID, me.ID, nil, c.ScryCards); err == nil {
			t.Error("ResolveScry answered a look-at-top prompt")
		}
	})
}

// --- Ponder ---------------------------------------------------------

// TestPonderDrawsAfterTheReorder — "then" is load-bearing here for
// the same reason it is on Preordain: the card left on top is the
// card drawn.
func TestPonderDrawsAfterTheReorder(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Worst", "Middling", "Best")
	handBefore := me.Hand.Size()

	castCatalogSpell(t, g, "Ponder", "Sorcery", ponderOracle, nil)
	passPriorityAroundTable(t, g)

	c := lookAtTopChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("Ponder queued no look-at prompt")
	}
	if len(c.ScryCards) != 3 {
		t.Fatalf("looked at %d cards, want 3", len(c.ScryCards))
	}
	if me.Hand.Size() != handBefore {
		t.Fatal("Ponder drew before the player had arranged the cards")
	}

	// Put "Best" on top. The draw does NOT happen yet: "you may
	// shuffle" comes between the reorder and the draw, so the reorder
	// answer queues a second prompt and the draw rides on its
	// branches.
	if err := g.ResolveLookAtTop(c.ID, me.ID,
		[]uuid.UUID{c.ScryCards[2], c.ScryCards[0], c.ScryCards[1]}); err != nil {
		t.Fatalf("ResolveLookAtTop: %v", err)
	}
	if me.Hand.Size() != handBefore {
		t.Fatal("Ponder drew before the shuffle question was answered")
	}
	shuffle := confirmChoiceFor(g, me.ID)
	if shuffle == nil {
		t.Fatal("the reorder answer queued no shuffle prompt")
	}
	// Decline: keep the order, then draw.
	if err := g.ResolveConfirm(shuffle.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if me.Hand.Size() != handBefore+1 {
		t.Fatalf("hand is %d, want %d — the draw never ran", me.Hand.Size(), handBefore+1)
	}
	drawn, _ := me.Hand.Top()
	if drawn.Name != "Best" {
		t.Errorf("drew %q, want Best — the card left on top is the card drawn", drawn.Name)
	}
}

// TestPonderShuffleIsASecondPromptOffTheFirst is the composition this
// card exists in the suite to pin: "you may shuffle" is not asked until
// the reorder has been answered, and answering it yes throws the
// arrangement away before the draw.
func TestPonderShuffleIsASecondPromptOffTheFirst(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	// Enough cards that a shuffle can actually move the top one; the
	// assertion below is on the DRAW, not on a particular permutation.
	seedLibrary(me, "L1", "L2", "L3", "L4", "L5", "L6", "L7", "L8", "Keep")
	handBefore := me.Hand.Size()
	libBefore := me.Library.Size()

	castCatalogSpell(t, g, "Ponder", "Sorcery", ponderOracle, nil)
	passPriorityAroundTable(t, g)

	c := lookAtTopChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("Ponder queued no look-at prompt")
	}
	// Exactly one prompt is open: the shuffle question does not exist
	// yet, because the player has not answered the first one.
	if n := len(g.PendingChoices); n != 1 {
		t.Fatalf("%d prompts open at once, want 1 — the chain must not front-load", n)
	}
	if err := g.ResolveLookAtTop(c.ID, me.ID, c.ScryCards); err != nil {
		t.Fatalf("ResolveLookAtTop: %v", err)
	}

	shuffle := confirmChoiceFor(g, me.ID)
	if shuffle == nil {
		t.Fatal("no shuffle prompt after the reorder")
	}
	if shuffle.AcceptLabel == "" || shuffle.DeclineLabel == "" {
		t.Errorf("the shuffle prompt has no branch labels: %+v", shuffle)
	}
	if err := g.ResolveConfirm(shuffle.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand is %d, want %d — the draw must happen on BOTH branches",
			me.Hand.Size(), handBefore+1)
	}
	if me.Library.Size() != libBefore-1 {
		t.Errorf("library is %d, want %d", me.Library.Size(), libBefore-1)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d prompts left open after the chain finished", len(g.PendingChoices))
	}
	// A shuffle un-knows the library (CR 701.20): the three cards the
	// player just looked at are no longer theirs to read.
	for _, card := range me.Library.Cards {
		if len(card.KnownBy) != 0 {
			t.Fatalf("%q is still known after the shuffle", card.Name)
		}
	}
}

func confirmChoiceFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceConfirm && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// --- Crystal Ball ----------------------------------------------------

func TestCrystalBallScriesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ball := pushCatalogPermanent(g, me.ID, "Crystal Ball", "Artifact", crystalBallOracle, false)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	seedLibrary(me, "Bad", "Alright", "Good")

	if err := g.ActivateCatalogAbility(me.ID, ball, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("Crystal Ball queued no scry")
	}
	if len(c.ScryCards) != 2 {
		t.Fatalf("looked at %d cards, want 2", len(c.ScryCards))
	}
	if card, ok := battlefieldCard(g, ball); !ok || !card.Tapped {
		t.Error("the Ball should be tapped — the ability costs {T}")
	}
	if err := g.ResolveScry(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	if got := libraryTopNames(me, 1); got[0] != "Good" {
		t.Errorf("library top is %q after bottoming both, want Good", got[0])
	}
}

// --- Otherworldly Gaze -----------------------------------------------

func TestOtherworldlyGazeSurveilsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Bin One", "Bin Two", "Keep")

	castCatalogSpell(t, g, "Otherworldly Gaze", "Instant", otherworldlyOracle, nil)
	passPriorityAroundTable(t, g)

	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("Otherworldly Gaze queued no surveil")
	}
	if len(c.ScryCards) != 3 {
		t.Fatalf("looked at %d cards, want 3", len(c.ScryCards))
	}
	// Bin the first two, keep the third.
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards[:2], c.ScryCards[2:]); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	// The Gaze itself is already in the graveyard: a spell routes
	// there on resolution, and the surveil prompt is answered
	// afterwards. Assert on the binned cards, not the count.
	got := graveyardNames(me)
	binned := map[string]bool{}
	for _, n := range got {
		binned[n] = true
	}
	if !binned["Bin One"] || !binned["Bin Two"] {
		t.Fatalf("graveyard is %v, want both binned cards in it", got)
	}
	if binned["Keep"] {
		t.Errorf("graveyard is %v; the kept card was binned too", got)
	}
	if libTop := libraryTopNames(me, 1); libTop[0] != "Keep" {
		t.Errorf("library top is %q, want Keep", libTop[0])
	}
}

// --- Thought Scour ---------------------------------------------------

func TestThoughtScourMillsTargetAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oppLibBefore := opp.Library.Size()
	handBefore := me.Hand.Size()

	castCatalogSpell(t, g, "Thought Scour", "Instant", thoughtScourOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

	if opp.Graveyard.Size() != 2 {
		t.Errorf("opponent graveyard has %d cards, want 2", opp.Graveyard.Size())
	}
	if opp.Library.Size() != oppLibBefore-2 {
		t.Errorf("opponent library is %d, want %d", opp.Library.Size(), oppLibBefore-2)
	}
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("caster hand is %d, want %d — the cantrip half is not targeted",
			me.Hand.Size(), handBefore+1)
	}
}

// --- Reliquary Tower -------------------------------------------------

func TestReliquaryTowerRemovesTheHandLimit(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Reliquary Tower", "Land", reliquaryTowerOracl, false)

	if got := g.EffectiveMaxHandSizeLocked(me); got != game.NoMaxHandSize {
		t.Errorf("max hand size is %d with a Tower out, want the no-limit sentinel %d",
			got, game.NoMaxHandSize)
	}
}
