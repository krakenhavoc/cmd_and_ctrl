package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Burning Wish -----------------------------------------------------

func TestBurningWishBringsASorceryAndExilesItself(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2,
		effectstest.OutsideCard("Lightning Bolt", "Instant"), effectstest.OutsideCard("Rift Bolt", "Sorcery"), effectstest.OutsideCard("Char", "Sorcery"))
	me, opp := g.Seats[0], g.Seats[1]
	handBefore := me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Burning Wish", "Sorcery", effectstest.BurningWishOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatalf("no prompt for the sorcery: %+v", g.PendingChoices)
	}
	if got := effectstest.CandidateNames(g, c); !effectstest.EqualNames(got, []string{"Rift Bolt", "Char"}) {
		t.Errorf("candidates %v, want only the sorceries", got)
	}
	if c.ChooseMin != 0 || c.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want a \"may\" of one", c.ChooseMin, c.ChooseMax)
	}
	// An instant is not a candidate, an opponent cannot answer.
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{effectstest.SideCardID(me, "Lightning Bolt")}); err == nil {
		t.Error("the wish accepted an instant")
	}
	if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{effectstest.SideCardID(me, "Rift Bolt")}); err == nil {
		t.Error("an opponent answered my wish")
	}
	riftID := effectstest.SideCardID(me, "Rift Bolt")
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{riftID}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.OutsideHandHas(me, "Rift Bolt") || me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d (had %d), has Rift Bolt %v", me.Hand.Size(), handBefore, effectstest.OutsideHandHas(me, "Rift Bolt"))
	}
	if got := effectstest.SideNamed(me); !effectstest.EqualNames(got, []string{"Lightning Bolt", "Char"}) {
		t.Errorf("sideboard %v, want the other two to stay outside", got)
	}
	if !effectstest.OutsideExileHas(g, "Burning Wish") || effectstest.OutsideGraveyardHas(me, "Burning Wish") {
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
	g := effectstest.NewSideboardCatalogGame(t, 2, effectstest.OutsideCard("Rift Bolt", "Sorcery"))
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Burning Wish", "Sorcery", effectstest.BurningWishOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OutsideHandHas(me, "Rift Bolt") || len(me.SideboardCards()) != 1 {
		t.Error("declining took the card")
	}
	if !effectstest.OutsideExileHas(g, "Burning Wish") {
		t.Error("Burning Wish did not exile itself when declined")
	}
}

// No sorcery outside the game: nothing is asked, and the spell still
// exiles itself.
func TestBurningWishWithNothingToFetchAsksNothing(t *testing.T) {
	for name, side := range map[string][]game.Card{
		"no sideboard":     nil,
		"no sorcery in it": {effectstest.OutsideCard("Lightning Bolt", "Instant")},
	} {
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewSideboardCatalogGame(t, 2, side...)
			effectstest.CastCatalogSpell(t, g, "Burning Wish", "Sorcery", effectstest.BurningWishOracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			if len(g.PendingChoices) != 0 {
				t.Errorf("a prompt went up: %+v", g.PendingChoices)
			}
			if !effectstest.OutsideExileHas(g, "Burning Wish") {
				t.Error("Burning Wish did not exile itself")
			}
		})
	}
}

// --- Living Wish ------------------------------------------------------

func TestLivingWishOffersCreaturesAndLands(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2,
		effectstest.OutsideCard("Lightning Bolt", "Instant"), effectstest.OutsideCard("Grizzly Bears", "Creature — Bear"),
		effectstest.OutsideCard("Forest", "Basic Land — Forest"), effectstest.OutsideCard("Dryad Arbor", "Land Creature — Forest Dryad"),
		effectstest.OutsideCard("Rift Bolt", "Sorcery"))
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Living Wish", "Sorcery", effectstest.LivingWishOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	if got := effectstest.CandidateNames(g, c); !effectstest.EqualNames(got, []string{"Grizzly Bears", "Forest", "Dryad Arbor"}) {
		t.Errorf("candidates %v", got)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{effectstest.SideCardID(me, "Forest")}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OutsideHandHas(me, "Forest") || !effectstest.OutsideExileHas(g, "Living Wish") {
		t.Errorf("Forest in hand %v, Living Wish exiled %v", effectstest.OutsideHandHas(me, "Forest"), effectstest.OutsideExileHas(g, "Living Wish"))
	}
}

// --- Mastermind's Acquisition -----------------------------------------

func TestMastermindsAcquisitionOutsideModeIsMandatoryAndUnrevealed(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2, effectstest.OutsideCard("Rift Bolt", "Sorcery"), effectstest.OutsideCard("Lightning Bolt", "Instant"))
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpellWithModes(t, g, "Mastermind's Acquisition", "Sorcery", effectstest.MastermindsOracle, []int{1})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatalf("no prompt: %+v", g.PendingChoices)
	}
	if c.ChooseMin != 1 || c.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want a mandatory one", c.ChooseMin, c.ChooseMax)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err == nil {
		t.Error("the mandatory pick accepted nothing")
	}
	bolt := effectstest.SideCardID(me, "Lightning Bolt")
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{bolt}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OutsideHandHas(me, "Lightning Bolt") {
		t.Error("the card did not come into the hand")
	}
	// Not revealed, so the opponent has not learned it.
	for _, card := range me.Hand.Cards {
		if card.InstanceID == bolt && card.IsKnownTo(opp.ID) {
			t.Error("Mastermind's Acquisition revealed the card")
		}
	}
	if effectstest.OutsideExileHas(g, "Mastermind's Acquisition") || !effectstest.OutsideGraveyardHas(me, "Mastermind's Acquisition") {
		t.Error("Mastermind's Acquisition does not exile itself; it goes to the graveyard")
	}
}

func TestMastermindsAcquisitionOutsideModeWithNoSideboardDoesNothing(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2)
	effectstest.CastCatalogSpellWithModes(t, g, "Mastermind's Acquisition", "Sorcery", effectstest.MastermindsOracle, []int{1})
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Errorf("a prompt went up: %+v", g.PendingChoices)
	}
}

func TestMastermindsAcquisitionSearchModeStillSearches(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2, effectstest.OutsideCard("Rift Bolt", "Sorcery"))
	effectstest.CastCatalogSpellWithModes(t, g, "Mastermind's Acquisition", "Sorcery", effectstest.MastermindsOracle, []int{0})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceSearchLibrary) == nil {
		t.Errorf("the search mode asked for no search: %+v", g.PendingChoices)
	}
	if len(g.Seats[0].SideboardCards()) != 1 {
		t.Error("the search mode touched the sideboard")
	}
}

// --- Extrapolate the Impossible ---------------------------------------

func TestExtrapolateTheImpossibleRevealsTwoAndAnOpponentChoosesOne(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 4,
		effectstest.OutsideCard("Alpha", "Sorcery"), effectstest.OutsideCard("Beta", "Instant"), effectstest.OutsideCard("Gamma", "Creature — Bear"))
	me, chooserSeat, third := g.Seats[0], g.Seats[2], g.Seats[1]
	handBefore := me.Hand.Size()
	alpha, beta, gamma := effectstest.SideCardID(me, "Alpha"), effectstest.SideCardID(me, "Beta"), effectstest.SideCardID(me, "Gamma")

	effectstest.CastCatalogSpell(t, g, "Extrapolate the Impossible", "Sorcery", effectstest.ExtrapolateOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	// Beat 1: the caster chooses exactly two, or none.
	c := effectstest.LatestChooseCardsFor(g, me.ID)
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
	effectstest.AnswerChosenPlayer(t, g, me.ID, chooserSeat.ID)
	rp := effectstest.LatestRevealPickFor(g, chooserSeat.ID)
	if rp == nil {
		t.Fatalf("no reveal pick for the chosen opponent: %+v", g.PendingChoices)
	}
	if rp.ChooseMin != 1 || rp.ChooseMax != 1 || len(rp.ChooseCards) != 2 {
		t.Errorf("reveal pick %d..%d over %d cards, want one of two", rp.ChooseMin, rp.ChooseMax, len(rp.ChooseCards))
	}
	for _, id := range []uuid.UUID{alpha, gamma} {
		if !effectstest.CardKnownToAll(g, me, id) {
			t.Errorf("revealed card %v is not known to every seat", id)
		}
	}
	if effectstest.CardKnownTo(me, beta, chooserSeat.ID) {
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
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OutsideHandHas(me, "Gamma") || me.Hand.Size() != handBefore+1 {
		t.Errorf("Gamma in hand %v, hand %d (had %d)", effectstest.OutsideHandHas(me, "Gamma"), me.Hand.Size(), handBefore)
	}
	if got := effectstest.SideNamed(me); !effectstest.EqualNames(got, []string{"Alpha", "Beta"}) {
		t.Errorf("sideboard %v: the other revealed card and the unrevealed one stay outside", got)
	}
	if !effectstest.OutsideGraveyardHas(me, "Extrapolate the Impossible") {
		t.Error("Extrapolate the Impossible did not go to the graveyard")
	}
	// The card that stays is still known to the table: it was revealed.
	if !effectstest.CardKnownToAll(g, me, alpha) {
		t.Error("the revealed card that stayed outside is no longer known")
	}
}

// At two seats there is one opponent, so the engine does not ask whom.
func TestExtrapolateTheImpossibleAtTwoSeatsAsksTheOnlyOpponent(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2, effectstest.OutsideCard("Alpha", "Sorcery"), effectstest.OutsideCard("Beta", "Instant"))
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Extrapolate the Impossible", "Sorcery", effectstest.ExtrapolateOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{effectstest.SideCardID(me, "Alpha"), effectstest.SideCardID(me, "Beta")}); err != nil {
		t.Fatal(err)
	}
	if effectstest.PendingOfKind(g, game.PendingChoiceOptionPick) != nil {
		effectstest.AnswerChosenPlayer(t, g, me.ID, opp.ID)
	}
	rp := effectstest.LatestRevealPickFor(g, opp.ID)
	if rp == nil {
		t.Fatalf("the only opponent was not asked: %+v", g.PendingChoices)
	}
	if err := g.ResolveRevealPick(rp.ID, opp.ID, []uuid.UUID{effectstest.SideCardID(me, "Alpha")}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OutsideHandHas(me, "Alpha") || len(me.SideboardCards()) != 1 {
		t.Errorf("hand has Alpha %v, sideboard %v", effectstest.OutsideHandHas(me, "Alpha"), effectstest.SideNamed(me))
	}
}

func TestExtrapolateTheImpossibleMayRevealNothing(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2, effectstest.OutsideCard("Alpha", "Sorcery"), effectstest.OutsideCard("Beta", "Instant"))
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Extrapolate the Impossible", "Sorcery", effectstest.ExtrapolateOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 || len(me.SideboardCards()) != 2 {
		t.Errorf("declining left %d choices and %d sideboard cards", len(g.PendingChoices), len(me.SideboardCards()))
	}
}

// Two cards with the same name are not "different names": a sideboard of
// copies has nothing to reveal and is asked nothing.
func TestExtrapolateTheImpossibleNeedsTwoDifferentNames(t *testing.T) {
	for name, side := range map[string][]game.Card{
		"no sideboard": nil,
		"one card":     {effectstest.OutsideCard("Alpha", "Sorcery")},
		"two copies":   {effectstest.OutsideCard("Alpha", "Sorcery"), effectstest.OutsideCard("Alpha", "Sorcery")},
	} {
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewSideboardCatalogGame(t, 2, side...)
			effectstest.CastCatalogSpell(t, g, "Extrapolate the Impossible", "Sorcery", effectstest.ExtrapolateOracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			if len(g.PendingChoices) != 0 {
				t.Errorf("a prompt went up: %+v", g.PendingChoices)
			}
			if !effectstest.OutsideGraveyardHas(g.Seats[0], "Extrapolate the Impossible") {
				t.Error("the spell did not finish resolving")
			}
		})
	}
}

// Three cards, two of them copies: the pair has to be different names.
func TestExtrapolateTheImpossibleRefusesACopyPair(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2,
		effectstest.OutsideCard("Alpha", "Sorcery"), effectstest.OutsideCard("Alpha", "Sorcery"), effectstest.OutsideCard("Beta", "Instant"))
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Extrapolate the Impossible", "Sorcery", effectstest.ExtrapolateOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
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
	g := effectstest.NewSideboardCatalogGame(t, 2, effectstest.OutsideCard("Rift Bolt", "Sorcery"))
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Burning Wish", "Sorcery", effectstest.BurningWishOracle, nil)
	g.WithWriteLock(func() {
		g.PlayerControls = append(g.PlayerControls, game.PlayerControl{
			Controller: opp.ID, Controlled: me.ID, Until: game.ControlUntilTurn,
			Stage: game.ControlTurn, TurnSeq: g.Turn.Seq,
		})
	})
	effectstest.PassPriorityAroundTable(t, g)
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
	for _, id := range []string{effectstest.ExtrapolateOracle, effectstest.BurningWishOracle, effectstest.LivingWishOracle, effectstest.MastermindsOracle} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Fatalf("%s is not in the catalog", id)
		}
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s is %v, want Full", spec.Name, spec.Completeness)
		}
	}
}
