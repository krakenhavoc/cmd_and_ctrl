package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trimLibrary leaves exactly n cards in p's library.
func trimLibrary(t *testing.T, p *game.Player, n int) {
	t.Helper()
	for p.Library.Size() > n {
		if _, err := p.Library.PopTop(); err != nil {
			t.Fatalf("PopTop: %v", err)
		}
	}
}

func TestDredgeIsOfferedAndReplacesTheDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	imp := effectstest.GyCardWithOracle(p, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	libBefore, handBefore := p.Library.Size(), p.Hand.Size()

	if err := g.DrawCard(p.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	offer := effectstest.OptionalOffer(g, p)
	if offer == nil {
		t.Fatal("dredge was not offered")
	}
	if offer.Source != imp {
		t.Errorf("the offer names %v, want the Imp %v", offer.Source, imp)
	}
	if p.Hand.Size() != handBefore || p.Library.Size() != libBefore {
		t.Fatal("the draw went ahead before the question was answered")
	}

	effectstest.AnswerOptional(t, g, p, true)
	if !effectstest.ZoneHas(p.Hand, imp) || effectstest.ZoneHas(p.Graveyard, imp) {
		t.Error("the Imp did not return to the hand")
	}
	if got := p.Library.Size(); got != libBefore-5 {
		t.Errorf("library %d, want %d (milled five, drew nothing)", got, libBefore-5)
	}
	if got := p.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand %d, want %d (only the Imp)", got, handBefore+1)
	}
	if got := p.Graveyard.Size(); got != 5 {
		t.Errorf("graveyard %d, want the five milled cards", got)
	}
	if len(g.DrawnThisTurn[p.ID]) != 0 {
		t.Error("a dredged card is not a drawn card")
	}
}

func TestDecliningDredgeDrawsNormally(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	imp := effectstest.GyCardWithOracle(p, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	libBefore, handBefore := p.Library.Size(), p.Hand.Size()

	_ = g.DrawCard(p.ID)
	effectstest.AnswerOptional(t, g, p, false)

	if p.Hand.Size() != handBefore+1 || p.Library.Size() != libBefore-1 {
		t.Errorf("hand %d library %d, want an ordinary draw", p.Hand.Size(), p.Library.Size())
	}
	if !effectstest.ZoneHas(p.Graveyard, imp) {
		t.Error("a declined dredge must leave the card in the graveyard")
	}
	if effectstest.OptionalOffer(g, p) != nil {
		t.Error("the offer is asked once per draw")
	}
}

func TestDredgeNeedsNCardsInTheLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	imp := effectstest.GyCardWithOracle(p, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	trimLibrary(t, p, 4)
	handBefore := p.Hand.Size()

	_ = g.DrawCard(p.ID)
	if effectstest.OptionalOffer(g, p) != nil {
		t.Fatal("dredge 5 offered with four cards in the library (CR 702.52b)")
	}
	if p.Hand.Size() != handBefore+1 || !effectstest.ZoneHas(p.Graveyard, imp) {
		t.Error("the draw should have been an ordinary one")
	}

	// Exactly N is enough.
	g2 := effectstest.NewCatalogGame(t)
	q := g2.Seats[1]
	effectstest.GyCardWithOracle(q, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	trimLibrary(t, q, 5)
	_ = g2.DrawCard(q.ID)
	if effectstest.OptionalOffer(g2, q) == nil {
		t.Error("dredge 5 with exactly five cards must be offered")
	}
}

func TestDredgeDoesNotApplyToAnotherPlayersGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, other := g.Seats[1], g.Seats[2]
	effectstest.GyCardWithOracle(other, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	_ = g.DrawCard(me.ID)
	if effectstest.OptionalOffer(g, me) != nil || effectstest.OptionalOffer(g, other) != nil {
		t.Error("dredge is \"your graveyard\" only")
	}
}

func TestMultiCardDrawAsksOncePerCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	thug := effectstest.GyCardWithOracle(p, "Golgari Thug", "Creature — Human Warrior", effectstest.GolgariThugOracle)
	handBefore, libBefore := p.Hand.Size(), p.Library.Size()

	effectstest.DrawN(g, p, 3)
	// First card: dredge it.
	effectstest.AnswerOptional(t, g, p, true)
	if !effectstest.ZoneHas(p.Hand, thug) {
		t.Fatal("the Thug did not return")
	}
	// The draw is still owed two more cards, and the Thug is back in
	// hand, so nothing more is offered: the rest are ordinary draws.
	if effectstest.OptionalOffer(g, p) != nil {
		t.Fatal("nothing is left to dredge")
	}
	if got := p.Hand.Size(); got != handBefore+1+2 {
		t.Errorf("hand %d, want %d: the Thug and the two remaining draws", got, handBefore+3)
	}
	if got := p.Library.Size(); got != libBefore-4-2 {
		t.Errorf("library %d, want %d", got, libBefore-6)
	}
}

func TestMultiCardDrawPausesAndAsksAgainOnEachCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	thug := effectstest.GyCardWithOracle(p, "Golgari Thug", "Creature — Human Warrior", effectstest.GolgariThugOracle)
	handBefore := p.Hand.Size()

	effectstest.DrawN(g, p, 3)
	effectstest.AnswerOptional(t, g, p, false) // card one: ordinary
	if p.Hand.Size() != handBefore+1 {
		t.Fatalf("hand %d after the first answer, want %d", p.Hand.Size(), handBefore+1)
	}
	effectstest.AnswerOptional(t, g, p, false) // card two: asked again
	effectstest.AnswerOptional(t, g, p, false) // card three
	if effectstest.OptionalOffer(g, p) != nil {
		t.Fatal("a fourth question")
	}
	if p.Hand.Size() != handBefore+3 || !effectstest.ZoneHas(p.Graveyard, thug) {
		t.Errorf("hand %d, want three ordinary draws with the Thug untouched", p.Hand.Size())
	}
}

func TestTheDrawIsHeldBehindItsOwnPrompt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	effectstest.GyCardWithOracle(p, "Golgari Thug", "Creature — Human Warrior", effectstest.GolgariThugOracle)
	handBefore := p.Hand.Size()

	effectstest.DrawN(g, p, 3)
	if p.Hand.Size() != handBefore {
		t.Fatalf("hand %d: the later draws raced ahead of the first question (CR 121.6b)", p.Hand.Size())
	}
}

func TestTwoDredgersEachOfferTheChoice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	thug := effectstest.GyCardWithOracle(p, "Golgari Thug", "Creature — Human Warrior", effectstest.GolgariThugOracle)
	shell := effectstest.GyCardWithOracle(p, "Shambling Shell", "Creature — Plant Zombie", effectstest.ShamblingShellOracle)
	libBefore := p.Library.Size()

	_ = g.DrawCard(p.ID)
	first := effectstest.OptionalOffer(g, p)
	if first == nil {
		t.Fatal("no offer")
	}
	// Decline the first, take the second: one or none, never both.
	effectstest.AnswerOptional(t, g, p, false)
	second := effectstest.OptionalOffer(g, p)
	if second == nil {
		t.Fatal("the second dredger was never offered")
	}
	if second.Source == first.Source {
		t.Fatal("the same card was offered twice")
	}
	effectstest.AnswerOptional(t, g, p, true)

	var took, left uuid.UUID
	if second.Source == thug {
		took, left = thug, shell
	} else {
		took, left = shell, thug
	}
	if !effectstest.ZoneHas(p.Hand, took) || !effectstest.ZoneHas(p.Graveyard, left) {
		t.Error("only the accepted card returns")
	}
	if effectstest.OptionalOffer(g, p) != nil {
		t.Error("a draw is replaced by at most one dredge")
	}
	n := 4
	if took == shell {
		n = 3
	}
	if got := p.Library.Size(); got != libBefore-n {
		t.Errorf("library %d, want %d", got, libBefore-n)
	}
}

func TestDredgeAboveACardOnTheBattlefieldIsNotOffered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	effectstest.BfCardWithOracle(g, p, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	handBefore := p.Hand.Size()
	_ = g.DrawCard(p.ID)
	if effectstest.OptionalOffer(g, p) != nil || p.Hand.Size() != handBefore+1 {
		t.Error("dredge works from the graveyard only")
	}
}

// Library of Leng and Laboratory-Maniac-style effects are about other
// events and are untouched.
func TestOrdinaryDrawIsUnaffected(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	handBefore, libBefore := p.Hand.Size(), p.Library.Size()
	_ = g.DrawCard(p.ID)
	if p.Hand.Size() != handBefore+1 || p.Library.Size() != libBefore-1 {
		t.Error("a plain draw changed")
	}
}

// --- Underrealm Lich ---------------------------------------------------

func TestUnderrealmLichLooksAtThreeAndKeepsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	effectstest.BfCardWithOracle(g, p, "Underrealm Lich", "Creature — Zombie Elf Shaman", effectstest.UnderrealmLichOracle)
	libBefore, handBefore, gyBefore := p.Library.Size(), p.Hand.Size(), p.Graveyard.Size()
	top := make([]uuid.UUID, 3)
	for i := range top {
		top[i] = p.Library.Cards[len(p.Library.Cards)-1-i].InstanceID
	}

	_ = g.DrawCard(p.ID)
	pick := effectstest.ChooseCardsChoiceFor(g, p.ID)
	if pick == nil {
		t.Fatal("the Lich did not ask which card to take")
	}
	if len(pick.ChooseCards) != 3 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("prompt over %d cards, %d..%d; want exactly one of three", len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	if p.Hand.Size() != handBefore {
		t.Fatal("nothing moves before the answer")
	}
	effectstest.AnswerChooseCards(t, g, p.ID, top[1])

	if !effectstest.ZoneHas(p.Hand, top[1]) || p.Hand.Size() != handBefore+1 {
		t.Error("the chosen card is not in the hand")
	}
	if !effectstest.ZoneHas(p.Graveyard, top[0]) || !effectstest.ZoneHas(p.Graveyard, top[2]) || p.Graveyard.Size() != gyBefore+2 {
		t.Error("the other two did not reach the graveyard")
	}
	if p.Library.Size() != libBefore-3 {
		t.Errorf("library %d, want %d", p.Library.Size(), libBefore-3)
	}
	if len(g.DrawnThisTurn[p.ID]) != 0 {
		t.Error("the Lich replaces the draw, so no card was drawn")
	}
}

func TestUnderrealmLichAsksOncePerDrawOfAMultiCardDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	effectstest.BfCardWithOracle(g, p, "Underrealm Lich", "Creature — Zombie Elf Shaman", effectstest.UnderrealmLichOracle)
	handBefore, libBefore := p.Hand.Size(), p.Library.Size()

	effectstest.DrawN(g, p, 3)
	for i := 0; i < 3; i++ {
		pick := effectstest.ChooseCardsChoiceFor(g, p.ID)
		if pick == nil {
			t.Fatalf("no prompt for draw %d", i+1)
		}
		if p.Hand.Size() != handBefore+i {
			t.Fatalf("before answer %d the hand is %d, want %d", i+1, p.Hand.Size(), handBefore+i)
		}
		effectstest.AnswerChooseCards(t, g, p.ID, pick.ChooseCards[0])
	}
	if effectstest.ChooseCardsChoiceFor(g, p.ID) != nil {
		t.Fatal("a fourth prompt")
	}
	if p.Hand.Size() != handBefore+3 || p.Library.Size() != libBefore-9 {
		t.Errorf("hand %d library %d, want +3 and -9", p.Hand.Size(), p.Library.Size())
	}
}

func TestUnderrealmLichWithAnEmptyLibraryLosesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	effectstest.BfCardWithOracle(g, p, "Underrealm Lich", "Creature — Zombie Elf Shaman", effectstest.UnderrealmLichOracle)
	trimLibrary(t, p, 0)
	_ = g.DrawCard(p.ID)
	if effectstest.ChooseCardsChoiceFor(g, p.ID) != nil {
		t.Error("nothing to look at, so nothing to ask")
	}
	if p.AttemptedEmptyDraw {
		t.Error("nothing was drawn, so an empty library is not a failed draw (CR 704.5b)")
	}
}

func TestUnderrealmLichLooksAtWhatThereIsWithFewerThanThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	effectstest.BfCardWithOracle(g, p, "Underrealm Lich", "Creature — Zombie Elf Shaman", effectstest.UnderrealmLichOracle)
	trimLibrary(t, p, 2)
	_ = g.DrawCard(p.ID)
	pick := effectstest.ChooseCardsChoiceFor(g, p.ID)
	if pick == nil || len(pick.ChooseCards) != 2 {
		t.Fatalf("want a prompt over the two cards there are, got %+v", pick)
	}
}

func TestLichAndDredgeComposeTheAffectedPlayerOrders(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	effectstest.BfCardWithOracle(g, p, "Underrealm Lich", "Creature — Zombie Elf Shaman", effectstest.UnderrealmLichOracle)
	imp := effectstest.GyCardWithOracle(p, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	_ = g.DrawCard(p.ID)
	// Two replacements apply to one draw: CR 616.1, the player orders.
	order := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceReplacementOrder, p.ID)
	if order == nil {
		t.Fatal("expected a CR 616 ordering prompt")
	}
	// Whichever is first, the dredge is still answerable, and a "yes"
	// replaces the draw so the Lich never gets to look.
	ids := order.ReplacementEffectIDs
	if len(ids) != 2 {
		t.Fatalf("%d effects in the order prompt, want 2", len(ids))
	}
	var dredgeFirst []game.ReplacementEffectID
	for _, id := range ids {
		if label, card := g.ReplacementOptionMetaForEffect(id); label == "Dredge 5" && card == imp {
			dredgeFirst = append([]game.ReplacementEffectID{id}, dredgeFirst...)
		} else {
			dredgeFirst = append(dredgeFirst, id)
		}
	}
	if label, _ := g.ReplacementOptionMetaForEffect(dredgeFirst[0]); label != "Dredge 5" {
		t.Fatalf("the dredge offer has no label in the order prompt: %q", label)
	}
	if err := g.ResolveReplacementOrder(order.ID, p.ID, dredgeFirst); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	effectstest.AnswerOptional(t, g, p, true)
	if !effectstest.ZoneHas(p.Hand, imp) {
		t.Error("the dredge did not apply")
	}
	if effectstest.ChooseCardsChoiceFor(g, p.ID) != nil {
		t.Error("the draw was already replaced, so the Lich has nothing to replace")
	}
}

// --- Forbidden Crypt ---------------------------------------------------

func TestForbiddenCryptReturnsACardInsteadOfDrawing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	effectstest.BfCardWithOracle(g, p, "Forbidden Crypt", "Enchantment", effectstest.ForbiddenCryptOracle)
	a := effectstest.GyCardWithOracle(p, "A", "Creature — Test", "")
	b := effectstest.GyCardWithOracle(p, "B", "Creature — Test", "")
	libBefore, handBefore := p.Library.Size(), p.Hand.Size()

	_ = g.DrawCard(p.ID)
	pick := effectstest.ChooseCardsChoiceFor(g, p.ID)
	if pick == nil || len(pick.ChooseCards) != 2 {
		t.Fatalf("want a pick over the two graveyard cards, got %+v", pick)
	}
	effectstest.AnswerChooseCards(t, g, p.ID, b)
	if !effectstest.ZoneHas(p.Hand, b) || !effectstest.ZoneHas(p.Graveyard, a) {
		t.Error("the chosen card should be in the hand and the other still in the graveyard")
	}
	if p.Library.Size() != libBefore || p.Hand.Size() != handBefore+1 {
		t.Error("nothing is drawn from the library")
	}
}

func TestForbiddenCryptLosesTheGameWithNothingToReturn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	effectstest.BfCardWithOracle(g, p, "Forbidden Crypt", "Enchantment", effectstest.ForbiddenCryptOracle)
	_ = g.DrawCard(p.ID)
	if !p.Eliminated {
		t.Error("\"if you can't, you lose the game\"")
	}
}

func TestForbiddenCryptExilesInsteadOfYourGraveyardOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, other := g.Seats[1], g.Seats[2]
	effectstest.BfCardWithOracle(g, me, "Forbidden Crypt", "Enchantment", effectstest.ForbiddenCryptOracle)
	g.WithWriteLock(func() {
		_, _ = g.MillToZoneForEffect(me.ID, 2, game.ZoneGraveyard)
		_, _ = g.MillToZoneForEffect(other.ID, 2, game.ZoneGraveyard)
	})
	if me.Graveyard.Size() != 0 {
		t.Errorf("your milled cards should be exiled, graveyard has %d", me.Graveyard.Size())
	}
	if other.Graveyard.Size() != 2 {
		t.Errorf("an opponent's graveyard is untouched, has %d", other.Graveyard.Size())
	}
}

func TestForbiddenCryptAsksOncePerCardOfAMultiCardDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	effectstest.BfCardWithOracle(g, p, "Forbidden Crypt", "Enchantment", effectstest.ForbiddenCryptOracle)
	ids := []uuid.UUID{
		effectstest.GyCardWithOracle(p, "A", "Creature — Test", ""),
		effectstest.GyCardWithOracle(p, "B", "Creature — Test", ""),
		effectstest.GyCardWithOracle(p, "C", "Creature — Test", ""),
	}
	effectstest.DrawN(g, p, 3)
	for i, id := range ids {
		if effectstest.ChooseCardsChoiceFor(g, p.ID) == nil {
			t.Fatalf("no prompt for draw %d", i+1)
		}
		effectstest.AnswerChooseCards(t, g, p.ID, id)
	}
	if p.Hand.Size() < 3 || p.Graveyard.Size() != 0 {
		t.Errorf("hand %d graveyard %d", p.Hand.Size(), p.Graveyard.Size())
	}
	if p.Eliminated {
		t.Error("each draw found a card to return")
	}
}

// --- The Necrobloom ----------------------------------------------------

func TestNecrobloomGivesLandsInTheGraveyardDredgeTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	effectstest.BfCardWithOracle(g, p, "The Necrobloom", "Legendary Creature — Plant", effectstest.NecrobloomOracle)
	forest := effectstest.GyCardWithOracle(p, "Forest", "Basic Land — Forest", "")
	bear := effectstest.GyCardWithOracle(p, "Bear", "Creature — Bear", "")
	libBefore := p.Library.Size()

	_ = g.DrawCard(p.ID)
	effectstest.AnswerOptional(t, g, p, true)
	pick := effectstest.ChooseCardsChoiceFor(g, p.ID)
	if pick == nil || len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != forest {
		t.Fatalf("only the land card is a candidate, got %+v", pick)
	}
	effectstest.AnswerChooseCards(t, g, p.ID, forest)
	if !effectstest.ZoneHas(p.Hand, forest) || !effectstest.ZoneHas(p.Graveyard, bear) {
		t.Error("the land returns and the creature stays")
	}
	if p.Library.Size() != libBefore-2 {
		t.Errorf("library %d, want %d (milled two)", p.Library.Size(), libBefore-2)
	}
}

func TestNecrobloomOffersNothingWithoutALandCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	effectstest.BfCardWithOracle(g, p, "The Necrobloom", "Legendary Creature — Plant", effectstest.NecrobloomOracle)
	effectstest.GyCardWithOracle(p, "Bear", "Creature — Bear", "")
	_ = g.DrawCard(p.ID)
	if effectstest.OptionalOffer(g, p) != nil {
		t.Error("no land card in the graveyard, no dredge")
	}
}

// --- Snapshots mid-pause ----------------------------------------------

func TestDredgeSnapshotTakenMidPauseReplaysTheAnswer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	imp := effectstest.GyCardWithOracle(p, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	effectstest.DrawN(g, p, 2)
	if effectstest.OptionalOffer(g, p) == nil {
		t.Fatal("no offer")
	}
	snap := g.Clone()
	sp := snap.Seats[1]

	// Answer on the live game: dredge, then the second card is an
	// ordinary draw (the Imp is in hand).
	effectstest.AnswerOptional(t, g, p, true)
	if !effectstest.ZoneHas(p.Hand, imp) {
		t.Fatal("live game did not dredge")
	}

	// The snapshot is still paused, still owes both cards, and replays
	// the same answer to the same result.
	if effectstest.OptionalOffer(snap, sp) == nil {
		t.Fatal("the snapshot lost its pending question")
	}
	offer := effectstest.OptionalOffer(snap, sp)
	if err := snap.ResolveOptionalReplacement(offer.ID, sp.ID, true); err != nil {
		t.Fatalf("answer on snapshot: %v", err)
	}
	if !effectstest.ZoneHas(sp.Hand, imp) || sp.Hand.Size() != p.Hand.Size() || sp.Library.Size() != p.Library.Size() {
		t.Errorf("snapshot diverged: hand %d/%d library %d/%d", sp.Hand.Size(), p.Hand.Size(), sp.Library.Size(), p.Library.Size())
	}
}

func TestLichSnapshotMidPauseKeepsTheRestOfTheDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	effectstest.BfCardWithOracle(g, p, "Underrealm Lich", "Creature — Zombie Elf Shaman", effectstest.UnderrealmLichOracle)
	handBefore := p.Hand.Size()
	effectstest.DrawN(g, p, 2)
	snap := g.Clone()
	sp := snap.Seats[1]

	pick := effectstest.ChooseCardsChoiceFor(snap, sp.ID)
	if pick == nil {
		t.Fatal("snapshot lost the prompt")
	}
	if err := snap.ResolveChooseCards(pick.ID, sp.ID, []uuid.UUID{pick.ChooseCards[0]}); err != nil {
		t.Fatalf("answer on snapshot: %v", err)
	}
	// The second draw's prompt is queued on the snapshot, and the live
	// game is untouched.
	if effectstest.ChooseCardsChoiceFor(snap, sp.ID) == nil {
		t.Error("the snapshot dropped the second draw")
	}
	if p.Hand.Size() != handBefore || effectstest.ChooseCardsChoiceFor(g, p.ID) == nil {
		t.Error("answering the snapshot reached the live game")
	}
}

func TestLifeFromTheLoamDredgesThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	loam := effectstest.GyCardWithOracle(p, "Life from the Loam", "Sorcery", effectstest.G3LifeFromLoamOracle)
	libBefore := p.Library.Size()
	_ = g.DrawCard(p.ID)
	effectstest.AnswerOptional(t, g, p, true)
	if !effectstest.ZoneHas(p.Hand, loam) || p.Library.Size() != libBefore-3 {
		t.Errorf("Loam in hand: %v, library %d want %d", effectstest.ZoneHas(p.Hand, loam), p.Library.Size(), libBefore-3)
	}
}

func TestGolgariGraveTrollEntersWithACounterPerCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[0]
	effectstest.GyCardWithOracle(p, "A", "Creature — Test", "")
	effectstest.GyCardWithOracle(p, "B", "Creature — Test", "")
	effectstest.GyCardWithOracle(p, "Spell", "Sorcery", "")
	id := effectstest.CastCatalogSpell(t, g, "Golgari Grave-Troll", "Creature — Troll Skeleton", effectstest.GraveTrollOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var got int
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			got = c.Counters["+1/+1"]
		}
	}
	if got != 2 {
		t.Errorf("%d +1/+1 counters, want 2 (one per creature card, not the sorcery)", got)
	}
}

// --- Nightmare Void ----------------------------------------------------

const nightmareVoidOracle = "9a53ef0d-8e29-4128-9cc4-fbdd3e4fdb84"

func TestNightmareVoidTargetsAnyPlayerAndOffersTheWholeHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	ids := effectstest.RevealHand(victim, effectstest.PoolAHand()...)
	effectstest.CastCatalogSpell(t, g, "Nightmare Void", "Sorcery", nightmareVoidOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenPick(t, g)
	if pick.Chooser != caster.ID || pick.FromPlayer != victim.ID || len(pick.DiscardOptions) != len(ids) {
		t.Fatalf("chooser %v from %v with %d options, want the caster over the victim's whole hand (%d)",
			pick.Chooser, pick.FromPlayer, len(pick.DiscardOptions), len(ids))
	}
}

func TestNightmareVoidDredgesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	void := effectstest.GyCardWithOracle(p, "Nightmare Void", "Sorcery", nightmareVoidOracle)
	libBefore := p.Library.Size()
	_ = g.DrawCard(p.ID)
	effectstest.AnswerOptional(t, g, p, true)
	if !effectstest.ZoneHas(p.Hand, void) || p.Library.Size() != libBefore-2 {
		t.Errorf("Void in hand %v, library %d want %d", effectstest.ZoneHas(p.Hand, void), p.Library.Size(), libBefore-2)
	}
}

// A Lich replaces EVERY draw of a doubled event: two draws, two looks.
func TestLichReplacesBothDrawsOfADoubledDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	effectstest.BfCardWithOracle(g, p, "Underrealm Lich", "Creature — Zombie Elf Shaman", effectstest.UnderrealmLichOracle)
	effectstest.BfCardWithOracle(g, p, "Thought Reflection", "Enchantment", effectstest.ThoughtReflectionOracle)
	handBefore := p.Hand.Size()

	_ = g.DrawCard(p.ID)
	effectstest.OrderFirst(t, g, p, "Thought Reflection")
	for i := 0; i < 2; i++ {
		pick := effectstest.ChooseCardsChoiceFor(g, p.ID)
		if pick == nil {
			t.Fatalf("no look for draw %d of the doubled draw", i+1)
		}
		effectstest.AnswerChooseCards(t, g, p.ID, pick.ChooseCards[0])
	}
	if effectstest.ChooseCardsChoiceFor(g, p.ID) != nil || p.Hand.Size() != handBefore+2 {
		t.Errorf("hand %d, want %d", p.Hand.Size(), handBefore+2)
	}
}

// Dredge replaces ONE draw of a doubled event; the other is ordinary.
func TestDredgeReplacesOneDrawOfADoubledDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	effectstest.BfCardWithOracle(g, p, "Thought Reflection", "Enchantment", effectstest.ThoughtReflectionOracle)
	imp := effectstest.GyCardWithOracle(p, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	handBefore, libBefore := p.Hand.Size(), p.Library.Size()

	_ = g.DrawCard(p.ID)
	effectstest.OrderFirst(t, g, p, "Thought Reflection")
	effectstest.AnswerOptional(t, g, p, true)
	if !effectstest.ZoneHas(p.Hand, imp) || p.Hand.Size() != handBefore+2 {
		t.Errorf("hand %d, want the Imp plus one ordinary draw (%d)", p.Hand.Size(), handBefore+2)
	}
	if p.Library.Size() != libBefore-5-1 {
		t.Errorf("library %d, want %d", p.Library.Size(), libBefore-6)
	}
}
