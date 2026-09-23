package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const azoriusChanceryOracle = "189fc8f4-17ac-4f1d-82c8-8401445bdaf4"

// chanceryPlay puts a Chancery into owner's hand and moves it to the
// battlefield, which is the path a played land takes.
func chanceryPlay(t *testing.T, g *game.Game, owner uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p := playerByIDForTest(g, owner)
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Azorius Chancery", TypeLine: "Land",
		OracleID: azoriusChanceryOracle, Owner: owner, Controller: owner,
	})
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneHand, Owner: owner},
		game.ZoneRef{Kind: game.ZoneBattlefield},
		id,
	); err != nil {
		t.Fatalf("play Chancery: %v", err)
	}
	return id
}

func playerByIDForTest(g *game.Game, id uuid.UUID) *game.Player {
	for _, p := range g.Seats {
		if p != nil && p.ID == id {
			return p
		}
	}
	return nil
}

// It arrives tapped, so it cannot be used for mana the turn it lands.
func TestAzoriusChanceryEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	p0 := g.Seats[0]
	aangPushLand(g, p0.ID, "Island", false) // a second land to bounce
	id := chanceryPlay(t, g, p0.ID)

	got, ok := aangCardOnBF(g, id)
	if !ok {
		t.Fatal("Chancery never reached the battlefield")
	}
	if !got.Tapped {
		t.Error("Chancery entered untapped; the enters-tapped clause is its drawback")
	}
}

// chanceryReturnPrompt is the resolution-time "return a land you
// control" prompt, once the trigger has resolved far enough to ask.
func chanceryReturnPrompt(t *testing.T, g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	c := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, chooser)
	if c == nil {
		t.Fatalf("no own_permanents prompt for the controller: %+v", g.PendingChoices)
	}
	return c
}

// chanceryInHand reports whether the card is in the player's hand.
func chanceryInHand(p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Hand.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

// The ETB returns the chosen land to its owner's hand, and the land is
// CHOSEN ON RESOLUTION: nothing is asked while the trigger waits on the
// stack, so opponents respond to the trigger without knowing which land
// is coming back.
func TestAzoriusChanceryReturnsChosenLand(t *testing.T) {
	g := newCatalogGame(t)
	p0 := g.Seats[0]
	island := aangPushLand(g, p0.ID, "Island", false)
	handBefore := len(p0.Hand.Cards)

	chancery := chanceryPlay(t, g, p0.ID)
	if pendingOfKind(g, game.PendingChoicePickTarget) != nil {
		t.Fatal("the bounce asked for a target; \"return a land you control\" names none")
	}
	if c := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, p0.ID); c != nil {
		t.Fatal("the land was chosen before the trigger resolved")
	}
	passPriorityAroundTable(t, g)

	prompt := chanceryReturnPrompt(t, g, p0.ID)
	if prompt.ChooseMin != 1 || prompt.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want exactly one land", prompt.ChooseMin, prompt.ChooseMax)
	}
	if len(prompt.ChooseCards) != 2 {
		t.Errorf("candidates = %v, want the Island and the Chancery", prompt.ChooseCards)
	}
	if err := g.ResolveOwnPermanents(prompt.ID, p0.ID, []uuid.UUID{island}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}

	if _, stillOut := aangCardOnBF(g, island); stillOut {
		t.Error("chosen land is still on the battlefield")
	}
	if !chanceryInHand(p0, island) {
		t.Error("chosen land did not return to its owner's hand")
	}
	if len(p0.Hand.Cards) != handBefore+1 {
		t.Errorf("hand %d, want %d", len(p0.Hand.Cards), handBefore+1)
	}
	// The Chancery itself stays put when something else was chosen.
	if _, ok := aangCardOnBF(g, chancery); !ok {
		t.Error("Chancery left the battlefield despite another land being chosen")
	}
}

// With no other land, the Chancery is still a candidate for its own
// trigger.
func TestAzoriusChanceryCanReturnItself(t *testing.T) {
	g := newCatalogGame(t)
	p0 := g.Seats[0]
	chancery := chanceryPlay(t, g, p0.ID)
	passPriorityAroundTable(t, g)

	prompt := chanceryReturnPrompt(t, g, p0.ID)
	if err := g.ResolveOwnPermanents(prompt.ID, p0.ID, []uuid.UUID{chancery}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if _, stillOut := aangCardOnBF(g, chancery); stillOut {
		t.Error("Chancery did not return itself")
	}
	if !chanceryInHand(p0, chancery) {
		t.Error("Chancery did not end up in its owner's hand")
	}
}

// Not a target, so a land of yours with hexproof is as
// good a choice as any — and an opponent's land is no choice at all.
func TestAzoriusChanceryReturnIgnoresHexproofAndOnlyOffersYourLands(t *testing.T) {
	g := newCatalogGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]
	field := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: field, Name: "Hexproof Land", TypeLine: "Land",
		Keywords: []string{"hexproof"}, Owner: p0.ID, Controller: p0.ID,
	})
	theirs := aangPushLand(g, p1.ID, "Plains", false)

	chanceryPlay(t, g, p0.ID)
	passPriorityAroundTable(t, g)

	prompt := chanceryReturnPrompt(t, g, p0.ID)
	for _, id := range prompt.ChooseCards {
		if id == theirs {
			t.Error("an opponent's land was offered for \"a land you control\"")
		}
	}
	if err := g.ResolveOwnPermanents(prompt.ID, p0.ID, []uuid.UUID{theirs}); err == nil {
		t.Error("the prompt accepted an opponent's land")
	}
	if err := g.ResolveOwnPermanents(prompt.ID, p0.ID, []uuid.UUID{field}); err != nil {
		t.Fatalf("ResolveOwnPermanents(hexproof land): %v", err)
	}
	if !chanceryInHand(p0, field) {
		t.Error("the hexproof land did not return; the bounce does not target")
	}
}

func TestAzoriusChanceryManaAbility(t *testing.T) {
	ab := game.ManaAbilitiesForCard(game.Card{OracleID: azoriusChanceryOracle})
	if len(ab) != 1 {
		t.Fatalf("%d mana abilities, want 1", len(ab))
	}
	if ab[0].Produced != "{W}{U}" {
		t.Errorf("produced %q, want {W}{U}", ab[0].Produced)
	}
}
