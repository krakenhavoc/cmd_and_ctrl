package testsa

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Aetherflux Conduit --------------------------------------------------

func aetherfluxTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	conduit := effectstest.PushCatalogPermanent(g, me.ID, "Aetherflux Conduit", "Artifact", effectstest.AetherfluxConduitOracle, false)
	return g, me, conduit
}

// The trigger gives energy equal to the mana spent, once per spell.
func TestAetherfluxConduitCastingASpellGivesEnergyEqualToTheManaSpent(t *testing.T) {
	g, me, _ := aetherfluxTable(t)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{R}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	effectstest.CastFromHandForTest(t, g, me, "Three Drop", "Sorcery", "{2}{R}", "", game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EnergyOf(me); got != 3 {
		t.Errorf("energy after a three-mana spell = %d, want 3", got)
	}
}

// A spell cast for nothing pays nothing, and an opponent's spell is not
// "you cast".
func TestAetherfluxConduitIgnoresFreeSpellsAndOpponentsSpells(t *testing.T) {
	g, me, _ := aetherfluxTable(t)
	free := effectstest.HandCardFull(me, "Free Sorcery", "Sorcery", "{3}{R}", "", []string{"R"})
	g.WithWriteLock(func() { g.FreeCastFromHandForEffect(me.ID, uuid.Nil, free, "Test") })
	if err := g.CastSpell(me.ID, free, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EnergyOf(me); got != 0 {
		t.Errorf("a free spell gave %d energy, want 0", got)
	}
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	if opp.ID == me.ID {
		t.Fatal("test wants another seat")
	}
	effectstest.CastCatalogSpell(t, g, "Opp Instant", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EnergyOf(opp); got != 0 {
		t.Errorf("an opponent has %d energy, want 0", got)
	}
}

// Paying fifty {E} is a cost: forty-nine refuses, and the tap is kept.
func TestAetherfluxConduitRefusesWithFortyNineEnergy(t *testing.T) {
	g, me, conduit := aetherfluxTable(t)
	if c := effectstest.SpecActivatedEnergy(t, effectstest.AetherfluxConduitOracle, 0); c.Energy != 50 || !c.Tap {
		t.Fatalf("cost energy %d tap %v, want 50 and tap", c.Energy, c.Tap)
	}
	effectstest.SetEnergy(t, g, me, 49)
	err := g.ActivateCatalogAbility(me.ID, conduit, 0, game.ActivateAbilityParams{})
	if !errors.Is(err, game.ErrInsufficientEnergy) {
		t.Fatalf("err = %v, want ErrInsufficientEnergy", err)
	}
	if effectstest.EnergyOf(me) != 49 {
		t.Errorf("a refused activation spent energy: %d", effectstest.EnergyOf(me))
	}
}

// The ability draws seven, pays fifty, and opens a free cast over every
// nonland card in hand (the drawn ones included), any number of them, in
// any timing. #4250: the casts are made inside the ability's resolution,
// each its own answer to the resolution_cast prompt, and the decline
// closes the rest with the cards left in hand.
func TestAetherfluxConduitDrawsSevenAndCastsAnyNumberOfSpellsFree(t *testing.T) {
	g, me, conduit := aetherfluxTable(t)
	older := effectstest.HandCardFull(me, "Held Sorcery", "Sorcery", "{4}{B}", "", []string{"B"})
	heldLand := effectstest.HandCardFull(me, "Held Swamp", "Basic Land — Swamp", "", "", nil)
	// Top first: five sorceries and two lands are the seven drawn; the
	// eighth stays in the library.
	sorcery := game.Card{Name: "Drawn Sorcery", TypeLine: "Sorcery", ManaCost: "{3}{R}", Colors: []string{"R"}}
	land := game.Card{Name: "Drawn Mountain", TypeLine: "Basic Land — Mountain"}
	ids := effectstest.SunbirdLibrary(me, sorcery, sorcery, sorcery, sorcery, sorcery, land, land,
		game.Card{Name: "Eighth", TypeLine: "Sorcery", ManaCost: "{9}"})
	drawn := ids[:5]
	effectstest.SetEnergy(t, g, me, 50)

	effectstest.P7Activate(t, g, me, conduit, 0, game.ActivateAbilityParams{})
	if effectstest.EnergyOf(me) != 0 {
		t.Errorf("energy after paying fifty = %d, want 0", effectstest.EnergyOf(me))
	}
	if me.Hand.Size() != 9 {
		t.Fatalf("hand = %d, want 9 (two held + seven drawn)", me.Hand.Size())
	}
	granted := 0
	for _, perm := range me.CastPermissions {
		if perm.LapseOnPass == game.LapseStaysInHand && perm.Zone == game.ZoneHand {
			granted++
		}
	}
	if granted != 6 {
		t.Errorf("grants over hand cards = %d, want 6 (one held sorcery and five drawn; no land)", granted)
	}
	offer := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, older, drawn[0])
	if len(offer.DiscardOptions) != 6 || effectstest.HasID(offer.DiscardOptions, heldLand) {
		t.Fatalf("the offer is %v, want the six nonland cards", offer.DiscardOptions)
	}
	// Several free casts in one go: each cast leaves the others open.
	for _, id := range append([]uuid.UUID{older}, drawn[:2]...) {
		if err := effectstest.CastThroughResolutionCast(t, g, me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("free cast of %v: %v", id, err)
		}
		if effectstest.ResolutionCastOfferFor(g, me.ID) == nil {
			t.Fatal("one cast closed the \"any number\" offer")
		}
	}
	if err := g.CastSpell(me.ID, heldLand, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}); err == nil {
		t.Error("a land was cast as a spell")
	}
	// The decline closes the rest; the resolution then finishes.
	effectstest.DeclineResolutionCast(t, g, me.ID)
	effectstest.CastInsideTheResolution(t, g, drawn[1])
	if err := g.CastSpell(me.ID, drawn[3], game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}); err == nil {
		t.Error("the free cast outlived the decline")
	}
	if !me.Hand.Contains(drawn[3]) {
		t.Error("an uncast card stays in hand")
	}
}

func TestAetherworksMarvelGivesEnergyForEachPermanentYouControlPutIntoAGraveyard(t *testing.T) {
	g, me, _ := effectstest.AetherworksTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	mine := effectstest.PushCatalogPermanent(g, me.ID, "My Bear", "Creature — Bear", "", false)
	token := effectstest.PushCatalogPermanent(g, me.ID, "Bear Token", "Token Creature — Bear", "", false)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)
	bounced := effectstest.PushCatalogPermanent(g, me.ID, "Bounced Bear", "Creature — Bear", "", false)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EnergyOf(me); got != 1 {
		t.Fatalf("energy after one of mine died = %d, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(token) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EnergyOf(me); got != 2 {
		t.Errorf("energy after a token of mine died = %d, want 2", got)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EnergyOf(me); got != 2 {
		t.Errorf("an opponent's permanent changed my energy: %d", got)
	}
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(bounced) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EnergyOf(me); got != 2 {
		t.Errorf("a bounce is not a graveyard: energy %d", got)
	}
}

func TestAetherworksMarvelPaysSixLooksAtSixAndCastsOneFreeOfThem(t *testing.T) {
	g, me, marvel := effectstest.AetherworksTable(t)
	if c := effectstest.SpecActivatedEnergy(t, effectstest.AetherworksMarvelOracle, 0); c.Energy != 6 || !c.Tap {
		t.Fatalf("cost energy %d tap %v, want 6 and tap", c.Energy, c.Tap)
	}
	spells, lands, deep := effectstest.AetherworksLibrary(me)
	c := effectstest.AetherworksLook(t, g, me, marvel)
	if effectstest.EnergyOf(me) != 1 {
		t.Errorf("energy after paying six from seven = %d, want 1", effectstest.EnergyOf(me))
	}
	for _, id := range lands {
		if effectstest.HasID(c.DiscardOptions, id) {
			t.Errorf("a land was offered as a spell: %v", id)
		}
	}
	for _, id := range spells {
		if !effectstest.HasID(c.DiscardOptions, id) {
			t.Errorf("a looked-at spell %v was not offered (no mana value cap)", id)
		}
	}
	if effectstest.HasID(c.DiscardOptions, deep) {
		t.Error("the seventh card was not looked at")
	}
	// Looked at, not revealed: only the looker knows any of the six.
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	for _, card := range me.Library.Cards {
		if card.IsKnownTo(opp.ID) {
			t.Errorf("%s is known to an opponent after a private look", card.Name)
		}
	}

	// An 8-mana creature cast for nothing, from where it sits in the
	// library, while the ability is still resolving.
	wurm := spells[1]
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, wurm, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("free cast from the library: %v", err)
	}
	effectstest.CastInsideTheResolution(t, g, wurm)
	// One spell: the cast spent the offer, and the rest went to the bottom.
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		t.Error("the offer outlived its one cast")
	}
	for _, id := range append(append([]uuid.UUID{}, lands...), spells[0], spells[2], spells[3]) {
		if i := effectstest.LibraryIndex(me, id); i < 0 || i > 4 {
			t.Errorf("the rest go to the bottom (indices 0-4 under the deep card), found %d", i)
		}
	}
	if i := effectstest.LibraryIndex(me, deep); i != len(me.Library.Cards)-1 {
		t.Errorf("the deep card is the top card once the rest go under, found %d", i)
	}
	if err := g.CastSpell(me.ID, spells[0], game.CastSpellParams{FromZone: "library"}); err == nil {
		t.Error("a second looked-at card was castable")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wurm) {
		t.Error("the free creature should resolve")
	}
}

// A declined look ends with all six at the bottom, and opens nothing.
func TestAetherworksMarvelUncastPickAndDeclinedLookGoToTheBottom(t *testing.T) {
	g2, me2, marvel2 := effectstest.AetherworksTable(t)
	spells2, _, deep2 := effectstest.AetherworksLibrary(me2)
	effectstest.AetherworksLook(t, g2, me2, marvel2)
	effectstest.DeclineResolutionCast(t, g2, me2.ID)
	if g2.ResolutionPaused() {
		t.Error("the decline did not let the resolution finish")
	}
	if err := g2.CastSpell(me2.ID, spells2[0], game.CastSpellParams{FromZone: "library"}); err == nil {
		t.Error("the decline left a looked-at card castable")
	}
	for _, id := range spells2 {
		if i := effectstest.LibraryIndex(me2, id); i < 0 || i > 5 {
			t.Errorf("a declined look bottoms all six, found %d", i)
		}
	}
	if i := effectstest.LibraryIndex(me2, deep2); i != len(me2.Library.Cards)-1 {
		t.Errorf("the deep card is the top card after the six go under, found %d", i)
	}
}

// Six lands: nothing to cast, no prompt, and all six still go under.
func TestAetherworksMarvelLookAtAllLandsBottomsThemWithoutAPrompt(t *testing.T) {
	g, me, marvel := effectstest.AetherworksTable(t)
	deep := effectstest.SunbirdLibrary(me, game.Card{Name: "Deep Spell", TypeLine: "Sorcery", ManaCost: "{R}"})[0]
	var lands []uuid.UUID
	for i := 0; i < 6; i++ {
		lands = append(lands, effectstest.SunbirdLibrary(me, game.Card{Name: "Land", TypeLine: "Basic Land — Forest"})...)
	}
	effectstest.SetEnergy(t, g, me, 6)
	if err := g.ActivateCatalogAbility(me.ID, marvel, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("no spell among the six, so no prompt")
	}
	for _, id := range lands {
		if i := effectstest.LibraryIndex(me, id); i < 0 || i > 5 {
			t.Errorf("a looked-at land goes to the bottom, found %d", i)
		}
	}
	if i := effectstest.LibraryIndex(me, deep); i != len(me.Library.Cards)-1 {
		t.Errorf("the deep card is on top, found %d", i)
	}
}

// Energy is a cost: five is one short.
func TestAetherworksMarvelRefusesWithFiveEnergy(t *testing.T) {
	g, me, marvel := effectstest.AetherworksTable(t)
	effectstest.SetEnergy(t, g, me, 5)
	err := g.ActivateCatalogAbility(me.ID, marvel, 0, game.ActivateAbilityParams{})
	if !errors.Is(err, game.ErrInsufficientEnergy) {
		t.Fatalf("err = %v, want ErrInsufficientEnergy", err)
	}
}
