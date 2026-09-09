package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// scry_test.go — CR 701.18 scry: Viscera Seer, Preordain.
//
// Three things here are worth more than "the cards moved":
//
//   - "Scry 2, THEN draw a card" really is ordered. A draw that runs
//     before the player answers takes one of the cards they were still
//     deciding about, and leaves the prompt unanswerable because that
//     card is no longer in the library to put back. That was a real bug
//     in the first draft of this code.
//   - Scry is LOOK AT, not reveal. Only the scrying player may see the
//     cards; anything else hands the table the top of a library.
//   - Nothing moves until the answer arrives, so an unanswered scry
//     leaves the library exactly as it was.

const (
	visceraSeerOracle = "f82a4e85-526d-4456-b700-7760043a31be"
	preordainOracle   = "ac641490-ca14-48d7-8cc4-b69ce984befa"
)

// seedLibrary puts named cards on TOP of the player's library so that
// names[0] is the next card drawn.
//
// Note it uses PushTop directly rather than pushLibraryCardForTest,
// which pushes to the BOTTOM — newCatalogGame already seeds a library
// of filler cards, so a bottom push leaves the filler on top and every
// scry assertion here reads the wrong card.
func seedLibrary(p *game.Player, names ...string) []uuid.UUID {
	ids := make([]uuid.UUID, len(names))
	for i := range names {
		ids[i] = uuid.New()
	}
	// Last push wins the top slot, so walk backwards.
	for i := len(names) - 1; i >= 0; i-- {
		p.Library.PushTop(game.Card{
			InstanceID: ids[i],
			Name:       names[i],
			TypeLine:   "Sorcery",
			Owner:      p.ID,
			Controller: p.ID,
		})
	}
	return ids
}

// libraryTopNames reads the top n library card names, top-first.
func libraryTopNames(p *game.Player, n int) []string {
	out := make([]string, 0, n)
	for i := 0; i < n && i < p.Library.Size(); i++ {
		out = append(out, p.Library.Cards[p.Library.Size()-1-i].Name)
	}
	return out
}

func scryChoiceFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceScry && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// --- Viscera Seer ------------------------------------------------

func TestVisceraSeerSacrificesForScryOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seer := pushCatalogPermanent(g, me.ID, "Viscera Seer", "Creature — Vampire Wizard",
		visceraSeerOracle, false)
	fodder := pushCatalogPermanent(g, me.ID, "Doomed Traveler", "Creature — Human Soldier", "", false)
	seedLibrary(me, "Top", "Second", "Third")

	if err := g.ActivateCatalogAbility(me.ID, seer, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if _, ok := battlefieldCard(g, fodder); ok {
		t.Error("the sacrifice cost was not paid")
	}
	passPriorityAroundTable(t, g)

	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt after the Seer's ability resolved")
	}
	if len(c.ScryCards) != 1 {
		t.Fatalf("looked at %d cards, want 1", len(c.ScryCards))
	}
	// Nothing has moved yet.
	if got := libraryTopNames(me, 1); got[0] != "Top" {
		t.Errorf("library top is %q before the answer; scry must not move anything early", got[0])
	}

	// Bottom it.
	if err := g.ResolveScry(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	if got := libraryTopNames(me, 2); got[0] != "Second" || got[1] != "Third" {
		t.Errorf("top two are %v, want [Second Third] after bottoming Top", got)
	}
	if bottom, _ := me.Library.Bottom(); bottom.Name != "Top" {
		t.Errorf("bottom card is %q, want Top", bottom.Name)
	}
}

func TestVisceraSeerKeepingOnTopLeavesTheLibraryAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seer := pushCatalogPermanent(g, me.ID, "Viscera Seer", "Creature — Vampire Wizard",
		visceraSeerOracle, false)
	fodder := pushCatalogPermanent(g, me.ID, "Doomed Traveler", "Creature — Human Soldier", "", false)
	seedLibrary(me, "Top", "Second")

	if err := g.ActivateCatalogAbility(me.ID, seer, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)

	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	if err := g.ResolveScry(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	if got := libraryTopNames(me, 2); got[0] != "Top" || got[1] != "Second" {
		t.Errorf("top two are %v, want [Top Second] unchanged", got)
	}
}

// TestVisceraSeerCanEatItself — "Sacrifice a creature", and the Seer is
// one. The last activation is always available.
func TestVisceraSeerCanEatItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seer := pushCatalogPermanent(g, me.ID, "Viscera Seer", "Creature — Vampire Wizard",
		visceraSeerOracle, false)
	seedLibrary(me, "Top")

	if err := g.ActivateCatalogAbility(me.ID, seer, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{seer},
	}); err != nil {
		t.Fatalf("the Seer could not eat itself: %v", err)
	}
	if _, ok := battlefieldCard(g, seer); ok {
		t.Error("the Seer survived being fed to its own ability")
	}
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) == nil {
		t.Error("the ability didn't scry; it was already on the stack when the Seer died")
	}
}

// --- Preordain: the "then" ordering ------------------------------

// TestPreordainDrawsTheCardLeftOnTop is the ordering assertion. Leave a
// known card on top, and that is the card drawn. If the draw ran before
// the scry was answered it would take whatever was on top at cast time
// instead — and would leave the prompt unanswerable.
func TestPreordainDrawsTheCardLeftOnTop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedLibrary(me, "Junk", "Gas", "Third")
	handBefore := me.Hand.Size()

	castCatalogSpell(t, g, "Preordain", "Sorcery", preordainOracle, nil)
	passPriorityAroundTable(t, g)

	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	if len(c.ScryCards) != 2 {
		t.Fatalf("looked at %d cards, want 2", len(c.ScryCards))
	}
	// The draw must NOT have happened yet: handBefore was sampled
	// before Preordain was seeded, so its arrival and its departure to
	// the graveyard cancel, and a delta of 0 means nothing was drawn.
	if got := me.Hand.Size() - handBefore; got != 0 {
		t.Fatalf("hand delta %d before the scry was answered, want 0; the draw jumped the queue", got)
	}

	// Bottom "Junk", keep "Gas" on top.
	junk, gas := ids[0], ids[1]
	if err := g.ResolveScry(c.ID, me.ID, []uuid.UUID{junk}, []uuid.UUID{gas}); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Fatalf("drew %d cards, want 1", got)
	}
	drew := false
	for _, c := range me.Hand.Cards {
		if c.InstanceID == gas {
			drew = true
		}
	}
	if !drew {
		t.Error("Preordain drew something other than the card left on top")
	}
	if me.Library.Contains(gas) {
		t.Error("the kept card is still in the library")
	}
	if bottom, _ := me.Library.Bottom(); bottom.InstanceID != junk {
		t.Errorf("bottom card is %q, want the bottomed Junk", bottom.Name)
	}
}

// TestPreordainBottomingBothStillDraws — both cards to the bottom, then
// the draw comes off what was underneath them.
func TestPreordainBottomingBothStillDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "One", "Two", "Wanted")
	handBefore := me.Hand.Size()

	castCatalogSpell(t, g, "Preordain", "Sorcery", preordainOracle, nil)
	passPriorityAroundTable(t, g)

	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	if err := g.ResolveScry(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Fatalf("drew %d cards, want 1", got)
	}
	found := false
	for _, card := range me.Hand.Cards {
		if card.Name == "Wanted" {
			found = true
		}
	}
	if !found {
		t.Error("drew something other than the card under the two bottomed ones")
	}
}

// --- validation ---------------------------------------------------

func TestResolveScryRejectsBadAnswers(t *testing.T) {
	setup := func(t *testing.T) (*game.Game, *game.Player, *game.PendingChoice) {
		t.Helper()
		g := newCatalogGame(t)
		me := g.Seats[0]
		seer := pushCatalogPermanent(g, me.ID, "Viscera Seer", "Creature — Vampire Wizard",
			visceraSeerOracle, false)
		fodder := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
		seedLibrary(me, "Top", "Second")
		if err := g.ActivateCatalogAbility(me.ID, seer, 0, game.ActivateAbilityParams{
			SacrificeIDs: []uuid.UUID{fodder},
		}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		passPriorityAroundTable(t, g)
		c := scryChoiceFor(g, me.ID)
		if c == nil {
			t.Fatal("no scry prompt")
		}
		return g, me, c
	}

	t.Run("omitting a looked-at card", func(t *testing.T) {
		g, me, c := setup(t)
		if err := g.ResolveScry(c.ID, me.ID, nil, nil); err == nil {
			t.Error("an answer that moves neither card was accepted")
		}
	})

	t.Run("same card in both lists", func(t *testing.T) {
		g, me, c := setup(t)
		id := c.ScryCards[0]
		if err := g.ResolveScry(c.ID, me.ID, []uuid.UUID{id}, []uuid.UUID{id}); err == nil {
			t.Error("a card put on the bottom AND the top was accepted")
		}
	})

	t.Run("a card that wasn't looked at", func(t *testing.T) {
		g, me, c := setup(t)
		// The second library card is real but was not part of a scry 1.
		var other uuid.UUID
		for _, card := range me.Library.Cards {
			if card.InstanceID != c.ScryCards[0] {
				other = card.InstanceID
				break
			}
		}
		if err := g.ResolveScry(c.ID, me.ID, []uuid.UUID{other}, c.ScryCards); err == nil {
			t.Error("a card the player never looked at was accepted")
		}
	})

	t.Run("someone else answering", func(t *testing.T) {
		g, me, c := setup(t)
		other := g.Seats[1]
		if err := g.ResolveScry(c.ID, other.ID, c.ScryCards, nil); err == nil {
			t.Error("another player answered my scry")
		}
		if got := libraryTopNames(me, 1); got[0] != "Top" {
			t.Error("the library moved on a rejected answer")
		}
	})
}

// --- privacy ------------------------------------------------------

// TestScryIsLookAtNotReveal — the whole table must not see the top of a
// library. The chooser sees faces; everyone else sees backs.
func TestScryIsLookAtNotReveal(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seer := pushCatalogPermanent(g, me.ID, "Viscera Seer", "Creature — Vampire Wizard",
		visceraSeerOracle, false)
	fodder := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	seedLibrary(me, "Secret Top")

	if err := g.ActivateCatalogAbility(me.ID, seer, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	find := func(v protocol.GameView, chooser uuid.UUID) *protocol.PendingChoiceView {
		for i := range v.PendingChoices {
			if v.PendingChoices[i].Kind == "scry" && v.PendingChoices[i].Chooser == chooser.String() {
				return &v.PendingChoices[i]
			}
		}
		return nil
	}

	mine := find(protocol.ViewOfGameFor(g, me.ID.String()), me.ID)
	if mine == nil {
		t.Fatal("the scrying player can't see their own prompt")
	}
	if len(mine.Options) != 1 {
		t.Fatalf("chooser sees %d options, want 1", len(mine.Options))
	}
	if mine.Options[0].Name != "Secret Top" {
		t.Errorf("chooser sees %q, want the real card name", mine.Options[0].Name)
	}

	theirs := find(protocol.ViewOfGameFor(g, opp.ID.String()), me.ID)
	if theirs != nil {
		for _, o := range theirs.Options {
			if o.Name == "Secret Top" {
				t.Error("an opponent can read the top of my library; scry is look-at, not reveal")
			}
			if o.KnownByYou {
				t.Error("an opponent is marked a knower of a scried card")
			}
		}
	}
}
