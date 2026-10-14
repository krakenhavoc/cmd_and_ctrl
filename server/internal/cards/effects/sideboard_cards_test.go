package effects

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sideboard_cards_test.go — ADR 0191: the cards that bring a card in
// from outside the game, over a deck's sideboard.

const (
	extrapolateOracle = "1f47c869-f4a9-43b0-aec0-8da4e3a3fec5"
	burningWishOracle = "13a884b5-8970-4d71-9082-a8b6c8db9a01"
	livingWishOracle  = "b9d186f5-1c92-4539-a5b5-e7fbd4b726fe"
	mastermindsOracle = "6aec5b1f-cea6-4ec3-9759-e4544da8c52b"
)

func outsideCard(name, typeLine string) game.Card {
	c := game.NewCard(name, uuid.Nil)
	c.TypeLine = typeLine
	c.IsSideboard = true
	return c
}

// newSideboardCatalogGame is newCatalogGame with `seats` seats, seat 0
// owning `side` outside the game.
func newSideboardCatalogGame(t *testing.T, seats int, side ...game.Card) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < seats; i++ {
		deck := make([]game.Card, 20)
		for j := range deck {
			deck[j] = game.NewCard("basic-filler", uuid.Nil)
		}
		if i == 0 {
			deck = append(deck, side...)
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(11, 22))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	for i := 0; g.Turn.Step != game.StepDraw; i++ {
		if i >= 8 {
			t.Fatalf("never reached the draw step")
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	return g
}

func sideNamed(p *game.Player) []string {
	var out []string
	for _, c := range p.SideboardCards() {
		out = append(out, c.Name)
	}
	return out
}

func outsideHandHas(p *game.Player, name string) bool {
	for _, c := range p.Hand.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}

func outsideExileHas(g *game.Game, name string) bool {
	for _, c := range g.Exile.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}

func outsideGraveyardHas(p *game.Player, name string) bool {
	for _, c := range p.Graveyard.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}

func sideCardID(p *game.Player, name string) uuid.UUID {
	for _, c := range p.SideboardCards() {
		if c.Name == name {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

func candidateNames(g *game.Game, c *game.PendingChoice) []string {
	var out []string
	g.WithWriteLock(func() {
		for _, id := range c.ChooseCards {
			if card, ok := g.LookupOutsideCardForEffect(id); ok {
				out = append(out, card.Name)
			}
		}
	})
	return out
}

func equalNames(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// --- Burning Wish -----------------------------------------------------

func TestBurningWishBringsASorceryAndExilesItself(t *testing.T) {
	g := newSideboardCatalogGame(t, 2,
		outsideCard("Lightning Bolt", "Instant"), outsideCard("Rift Bolt", "Sorcery"), outsideCard("Char", "Sorcery"))
	me, opp := g.Seats[0], g.Seats[1]
	handBefore := me.Hand.Size()

	castCatalogSpell(t, g, "Burning Wish", "Sorcery", burningWishOracle, nil)
	passPriorityAroundTable(t, g)

	c := latestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatalf("no prompt for the sorcery: %+v", g.PendingChoices)
	}
	if got := candidateNames(g, c); !equalNames(got, []string{"Rift Bolt", "Char"}) {
		t.Errorf("candidates %v, want only the sorceries", got)
	}
	if c.ChooseMin != 0 || c.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want a \"may\" of one", c.ChooseMin, c.ChooseMax)
	}
	// An instant is not a candidate, an opponent cannot answer.
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{sideCardID(me, "Lightning Bolt")}); err == nil {
		t.Error("the wish accepted an instant")
	}
	if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{sideCardID(me, "Rift Bolt")}); err == nil {
		t.Error("an opponent answered my wish")
	}
	riftID := sideCardID(me, "Rift Bolt")
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{riftID}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !outsideHandHas(me, "Rift Bolt") || me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d (had %d), has Rift Bolt %v", me.Hand.Size(), handBefore, outsideHandHas(me, "Rift Bolt"))
	}
	if got := sideNamed(me); !equalNames(got, []string{"Lightning Bolt", "Char"}) {
		t.Errorf("sideboard %v, want the other two to stay outside", got)
	}
	if !outsideExileHas(g, "Burning Wish") || outsideGraveyardHas(me, "Burning Wish") {
		t.Error("Burning Wish did not exile itself")
	}
	// Revealed: the opponent knows the card now in my hand.
	for _, card := range me.Hand.Cards {
		if card.InstanceID == riftID && !card.IsKnownTo(opp.ID) {
			t.Error("the revealed card is not known to the opponent")
		}
	}
	revealed := false
	for _, ev := range g.Events {
		if ev.Kind == game.EventRevealCards && ev.CardID == riftID && ev.OldZone == game.ZoneOutside {
			revealed = true
		}
	}
	if !revealed {
		t.Error("no reveal event for the chosen card")
	}
}

func TestBurningWishMayDecline(t *testing.T) {
	g := newSideboardCatalogGame(t, 2, outsideCard("Rift Bolt", "Sorcery"))
	me := g.Seats[0]
	castCatalogSpell(t, g, "Burning Wish", "Sorcery", burningWishOracle, nil)
	passPriorityAroundTable(t, g)
	c := latestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if outsideHandHas(me, "Rift Bolt") || len(me.SideboardCards()) != 1 {
		t.Error("declining took the card")
	}
	if !outsideExileHas(g, "Burning Wish") {
		t.Error("Burning Wish did not exile itself when declined")
	}
}

// No sorcery outside the game: nothing is asked, and the spell still
// exiles itself.
func TestBurningWishWithNothingToFetchAsksNothing(t *testing.T) {
	for name, side := range map[string][]game.Card{
		"no sideboard":     nil,
		"no sorcery in it": {outsideCard("Lightning Bolt", "Instant")},
	} {
		t.Run(name, func(t *testing.T) {
			g := newSideboardCatalogGame(t, 2, side...)
			castCatalogSpell(t, g, "Burning Wish", "Sorcery", burningWishOracle, nil)
			passPriorityAroundTable(t, g)
			if len(g.PendingChoices) != 0 {
				t.Errorf("a prompt went up: %+v", g.PendingChoices)
			}
			if !outsideExileHas(g, "Burning Wish") {
				t.Error("Burning Wish did not exile itself")
			}
		})
	}
}

// --- Living Wish ------------------------------------------------------

func TestLivingWishOffersCreaturesAndLands(t *testing.T) {
	g := newSideboardCatalogGame(t, 2,
		outsideCard("Lightning Bolt", "Instant"), outsideCard("Grizzly Bears", "Creature — Bear"),
		outsideCard("Forest", "Basic Land — Forest"), outsideCard("Dryad Arbor", "Land Creature — Forest Dryad"),
		outsideCard("Rift Bolt", "Sorcery"))
	me := g.Seats[0]
	castCatalogSpell(t, g, "Living Wish", "Sorcery", livingWishOracle, nil)
	passPriorityAroundTable(t, g)
	c := latestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	if got := candidateNames(g, c); !equalNames(got, []string{"Grizzly Bears", "Forest", "Dryad Arbor"}) {
		t.Errorf("candidates %v", got)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{sideCardID(me, "Forest")}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if !outsideHandHas(me, "Forest") || !outsideExileHas(g, "Living Wish") {
		t.Errorf("Forest in hand %v, Living Wish exiled %v", outsideHandHas(me, "Forest"), outsideExileHas(g, "Living Wish"))
	}
}

// --- Mastermind's Acquisition -----------------------------------------

func TestMastermindsAcquisitionOutsideModeIsMandatoryAndUnrevealed(t *testing.T) {
	g := newSideboardCatalogGame(t, 2, outsideCard("Rift Bolt", "Sorcery"), outsideCard("Lightning Bolt", "Instant"))
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpellWithModes(t, g, "Mastermind's Acquisition", "Sorcery", mastermindsOracle, []int{1})
	passPriorityAroundTable(t, g)
	c := latestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatalf("no prompt: %+v", g.PendingChoices)
	}
	if c.ChooseMin != 1 || c.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want a mandatory one", c.ChooseMin, c.ChooseMax)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err == nil {
		t.Error("the mandatory pick accepted nothing")
	}
	bolt := sideCardID(me, "Lightning Bolt")
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{bolt}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if !outsideHandHas(me, "Lightning Bolt") {
		t.Error("the card did not come into the hand")
	}
	// Not revealed, so the opponent has not learned it.
	for _, card := range me.Hand.Cards {
		if card.InstanceID == bolt && card.IsKnownTo(opp.ID) {
			t.Error("Mastermind's Acquisition revealed the card")
		}
	}
	if outsideExileHas(g, "Mastermind's Acquisition") || !outsideGraveyardHas(me, "Mastermind's Acquisition") {
		t.Error("Mastermind's Acquisition does not exile itself; it goes to the graveyard")
	}
}

func TestMastermindsAcquisitionOutsideModeWithNoSideboardDoesNothing(t *testing.T) {
	g := newSideboardCatalogGame(t, 2)
	castCatalogSpellWithModes(t, g, "Mastermind's Acquisition", "Sorcery", mastermindsOracle, []int{1})
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Errorf("a prompt went up: %+v", g.PendingChoices)
	}
}

func TestMastermindsAcquisitionSearchModeStillSearches(t *testing.T) {
	g := newSideboardCatalogGame(t, 2, outsideCard("Rift Bolt", "Sorcery"))
	castCatalogSpellWithModes(t, g, "Mastermind's Acquisition", "Sorcery", mastermindsOracle, []int{0})
	passPriorityAroundTable(t, g)
	if latestChoiceOfKind(g, game.PendingChoiceSearchLibrary) == nil {
		t.Errorf("the search mode asked for no search: %+v", g.PendingChoices)
	}
	if len(g.Seats[0].SideboardCards()) != 1 {
		t.Error("the search mode touched the sideboard")
	}
}

// --- Extrapolate the Impossible ---------------------------------------

func TestExtrapolateTheImpossibleRevealsTwoAndAnOpponentChoosesOne(t *testing.T) {
	g := newSideboardCatalogGame(t, 4,
		outsideCard("Alpha", "Sorcery"), outsideCard("Beta", "Instant"), outsideCard("Gamma", "Creature — Bear"))
	me, chooserSeat, third := g.Seats[0], g.Seats[2], g.Seats[1]
	handBefore := me.Hand.Size()
	alpha, beta, gamma := sideCardID(me, "Alpha"), sideCardID(me, "Beta"), sideCardID(me, "Gamma")

	castCatalogSpell(t, g, "Extrapolate the Impossible", "Sorcery", extrapolateOracle, nil)
	passPriorityAroundTable(t, g)

	// Beat 1: the caster chooses exactly two, or none.
	c := latestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatalf("no prompt for the pair: %+v", g.PendingChoices)
	}
	if c.ChooseMin != 0 || c.ChooseMax != 2 || len(c.ChooseCards) != 3 {
		t.Errorf("bounds %d..%d over %d cards, want up to two of all three", c.ChooseMin, c.ChooseMax, len(c.ChooseCards))
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{alpha}); err == nil {
		t.Error("a single card was accepted: it must be exactly two")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{alpha, beta, gamma}); err == nil {
		t.Error("three cards were accepted")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{alpha, gamma}); err != nil {
		t.Fatalf("a valid pair: %v", err)
	}

	// Beat 2: the caster names the opponent, who sees both revealed
	// cards and nothing else.
	answerChosenPlayer(t, g, me.ID, chooserSeat.ID)
	rp := latestRevealPickFor(g, chooserSeat.ID)
	if rp == nil {
		t.Fatalf("no reveal pick for the chosen opponent: %+v", g.PendingChoices)
	}
	if rp.ChooseMin != 1 || rp.ChooseMax != 1 || len(rp.ChooseCards) != 2 {
		t.Errorf("reveal pick %d..%d over %d cards, want one of two", rp.ChooseMin, rp.ChooseMax, len(rp.ChooseCards))
	}
	for _, id := range []uuid.UUID{alpha, gamma} {
		if !cardKnownToAll(g, me, id) {
			t.Errorf("revealed card %v is not known to every seat", id)
		}
	}
	if cardKnownTo(me, beta, chooserSeat.ID) {
		t.Error("the card that was not revealed became known to the chooser")
	}
	// Only the chosen opponent may answer, and only with a revealed card.
	if err := g.ResolveRevealPick(rp.ID, third.ID, []uuid.UUID{alpha}); err == nil {
		t.Error("another opponent answered")
	}
	if err := g.ResolveRevealPick(rp.ID, chooserSeat.ID, []uuid.UUID{beta}); err == nil {
		t.Error("the opponent chose a card that was not revealed")
	}

	// Beat 3: they choose Gamma. It goes to the hand; Alpha stays.
	if err := g.ResolveRevealPick(rp.ID, chooserSeat.ID, []uuid.UUID{gamma}); err != nil {
		t.Fatalf("the opponent's choice: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !outsideHandHas(me, "Gamma") || me.Hand.Size() != handBefore+1 {
		t.Errorf("Gamma in hand %v, hand %d (had %d)", outsideHandHas(me, "Gamma"), me.Hand.Size(), handBefore)
	}
	if got := sideNamed(me); !equalNames(got, []string{"Alpha", "Beta"}) {
		t.Errorf("sideboard %v: the other revealed card and the unrevealed one stay outside", got)
	}
	if !outsideGraveyardHas(me, "Extrapolate the Impossible") {
		t.Error("Extrapolate the Impossible did not go to the graveyard")
	}
	// The card that stays is still known to the table: it was revealed.
	if !cardKnownToAll(g, me, alpha) {
		t.Error("the revealed card that stayed outside is no longer known")
	}
}

// At two seats there is one opponent, so the engine does not ask whom.
func TestExtrapolateTheImpossibleAtTwoSeatsAsksTheOnlyOpponent(t *testing.T) {
	g := newSideboardCatalogGame(t, 2, outsideCard("Alpha", "Sorcery"), outsideCard("Beta", "Instant"))
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpell(t, g, "Extrapolate the Impossible", "Sorcery", extrapolateOracle, nil)
	passPriorityAroundTable(t, g)
	c := latestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{sideCardID(me, "Alpha"), sideCardID(me, "Beta")}); err != nil {
		t.Fatal(err)
	}
	if pendingOfKind(g, game.PendingChoiceOptionPick) != nil {
		answerChosenPlayer(t, g, me.ID, opp.ID)
	}
	rp := latestRevealPickFor(g, opp.ID)
	if rp == nil {
		t.Fatalf("the only opponent was not asked: %+v", g.PendingChoices)
	}
	if err := g.ResolveRevealPick(rp.ID, opp.ID, []uuid.UUID{sideCardID(me, "Alpha")}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if !outsideHandHas(me, "Alpha") || len(me.SideboardCards()) != 1 {
		t.Errorf("hand has Alpha %v, sideboard %v", outsideHandHas(me, "Alpha"), sideNamed(me))
	}
}

func TestExtrapolateTheImpossibleMayRevealNothing(t *testing.T) {
	g := newSideboardCatalogGame(t, 2, outsideCard("Alpha", "Sorcery"), outsideCard("Beta", "Instant"))
	me := g.Seats[0]
	castCatalogSpell(t, g, "Extrapolate the Impossible", "Sorcery", extrapolateOracle, nil)
	passPriorityAroundTable(t, g)
	c := latestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 || len(me.SideboardCards()) != 2 {
		t.Errorf("declining left %d choices and %d sideboard cards", len(g.PendingChoices), len(me.SideboardCards()))
	}
}

// Two cards with the same name are not "different names": a sideboard of
// copies has nothing to reveal and is asked nothing.
func TestExtrapolateTheImpossibleNeedsTwoDifferentNames(t *testing.T) {
	for name, side := range map[string][]game.Card{
		"no sideboard": nil,
		"one card":     {outsideCard("Alpha", "Sorcery")},
		"two copies":   {outsideCard("Alpha", "Sorcery"), outsideCard("Alpha", "Sorcery")},
	} {
		t.Run(name, func(t *testing.T) {
			g := newSideboardCatalogGame(t, 2, side...)
			castCatalogSpell(t, g, "Extrapolate the Impossible", "Sorcery", extrapolateOracle, nil)
			passPriorityAroundTable(t, g)
			if len(g.PendingChoices) != 0 {
				t.Errorf("a prompt went up: %+v", g.PendingChoices)
			}
			if !outsideGraveyardHas(g.Seats[0], "Extrapolate the Impossible") {
				t.Error("the spell did not finish resolving")
			}
		})
	}
}

// Three cards, two of them copies: the pair has to be different names.
func TestExtrapolateTheImpossibleRefusesACopyPair(t *testing.T) {
	g := newSideboardCatalogGame(t, 2,
		outsideCard("Alpha", "Sorcery"), outsideCard("Alpha", "Sorcery"), outsideCard("Beta", "Instant"))
	me := g.Seats[0]
	castCatalogSpell(t, g, "Extrapolate the Impossible", "Sorcery", extrapolateOracle, nil)
	passPriorityAroundTable(t, g)
	c := latestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	cards := me.SideboardCards()
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{cards[0].InstanceID, cards[1].InstanceID}); err == nil {
		t.Error("two cards with the same name were accepted")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{cards[0].InstanceID, cards[2].InstanceID}); err != nil {
		t.Errorf("a pair with different names: %v", err)
	}
}

// ADR 0167 §5 (CR 723.4): while another player controls the caster's
// turn, nothing is offered from the caster's sideboard.
func TestAWishFindsNothingWhileAnotherPlayerControlsTheTurn(t *testing.T) {
	g := newSideboardCatalogGame(t, 2, outsideCard("Rift Bolt", "Sorcery"))
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpell(t, g, "Burning Wish", "Sorcery", burningWishOracle, nil)
	g.WithWriteLock(func() {
		g.PlayerControls = append(g.PlayerControls, game.PlayerControl{
			Controller: opp.ID, Controlled: me.ID, Until: game.ControlUntilTurn,
			Stage: game.ControlTurn, TurnSeq: g.Turn.Seq,
		})
	})
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Errorf("the controller was offered %d prompts over a sideboard they may not see", len(g.PendingChoices))
	}
	if len(me.SideboardCards()) != 1 {
		t.Error("a card left the sideboard")
	}
}

// Tests that these are not in the sandbox: a card the catalog does not
// know stays inert.
func TestEveryOutsideGameCardIsFull(t *testing.T) {
	for _, id := range []string{extrapolateOracle, burningWishOracle, livingWishOracle, mastermindsOracle} {
		spec, ok := Lookup(id)
		if !ok {
			t.Fatalf("%s is not in the catalog", id)
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s is %v, want Full", spec.Name, spec.Completeness)
		}
	}
}

// cardKnownTo reports whether `viewer` is a knower of the sideboard card.
func cardKnownTo(p *game.Player, id, viewer uuid.UUID) bool {
	for _, c := range p.SideboardCards() {
		if c.InstanceID == id {
			return c.IsKnownTo(viewer)
		}
	}
	return false
}

func cardKnownToAll(g *game.Game, p *game.Player, id uuid.UUID) bool {
	for _, s := range g.Seats {
		if !cardKnownTo(p, id, s.ID) {
			return false
		}
	}
	return true
}
