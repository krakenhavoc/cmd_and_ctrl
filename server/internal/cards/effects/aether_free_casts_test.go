package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// aether_free_casts_test.go — Aetherflux Conduit and Aetherworks Marvel
// (#2110): the last two cards of the "cast a spell for free as an ability
// resolves" row, on the #3270 pick and grant.

const (
	aetherfluxConduitOracle = "775617cd-7264-4365-94c4-ffae4eb9ec8b"
	aetherworksMarvelOracle = "470ec98e-120c-446d-ade6-7d80aa836b55"
)

// --- Aetherflux Conduit --------------------------------------------------

func aetherfluxTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	conduit := pushCatalogPermanent(g, me.ID, "Aetherflux Conduit", "Artifact", aetherfluxConduitOracle, false)
	return g, me, conduit
}

// The trigger gives energy equal to the mana spent, once per spell.
func TestAetherfluxConduitCastingASpellGivesEnergyEqualToTheManaSpent(t *testing.T) {
	g, me, _ := aetherfluxTable(t)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{R}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	castFromHandForTest(t, g, me, "Three Drop", "Sorcery", "{2}{R}", "", game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	if got := energyOf(me); got != 3 {
		t.Errorf("energy after a three-mana spell = %d, want 3", got)
	}
}

// A spell cast for nothing pays nothing, and an opponent's spell is not
// "you cast".
func TestAetherfluxConduitIgnoresFreeSpellsAndOpponentsSpells(t *testing.T) {
	g, me, _ := aetherfluxTable(t)
	free := handCardFull(me, "Free Sorcery", "Sorcery", "{3}{R}", "", []string{"R"})
	g.WithWriteLock(func() { g.FreeCastFromHandForEffect(me.ID, uuid.Nil, free, "Test") })
	if err := g.CastSpell(me.ID, free, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := energyOf(me); got != 0 {
		t.Errorf("a free spell gave %d energy, want 0", got)
	}
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	if opp.ID == me.ID {
		t.Fatal("test wants another seat")
	}
	castCatalogSpell(t, g, "Opp Instant", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if got := energyOf(opp); got != 0 {
		t.Errorf("an opponent has %d energy, want 0", got)
	}
}

// Paying fifty {E} is a cost: forty-nine refuses, and the tap is kept.
func TestAetherfluxConduitRefusesWithFortyNineEnergy(t *testing.T) {
	g, me, conduit := aetherfluxTable(t)
	if c := specActivatedEnergy(t, aetherfluxConduitOracle, 0); c.Energy != 50 || !c.Tap {
		t.Fatalf("cost energy %d tap %v, want 50 and tap", c.Energy, c.Tap)
	}
	setEnergy(t, g, me, 49)
	err := g.ActivateCatalogAbility(me.ID, conduit, 0, game.ActivateAbilityParams{})
	if !errors.Is(err, game.ErrInsufficientEnergy) {
		t.Fatalf("err = %v, want ErrInsufficientEnergy", err)
	}
	if energyOf(me) != 49 {
		t.Errorf("a refused activation spent energy: %d", energyOf(me))
	}
}

// The ability draws seven, pays fifty, and opens a free cast over every
// nonland card in hand (the drawn ones included), any number of them, in
// any timing; the next pass closes all of them and the cards stay.
func TestAetherfluxConduitDrawsSevenAndCastsAnyNumberOfSpellsFree(t *testing.T) {
	g, me, conduit := aetherfluxTable(t)
	older := handCardFull(me, "Held Sorcery", "Sorcery", "{4}{B}", "", []string{"B"})
	heldLand := handCardFull(me, "Held Swamp", "Basic Land — Swamp", "", "", nil)
	// Top first: five sorceries and two lands are the seven drawn; the
	// eighth stays in the library.
	sorcery := game.Card{Name: "Drawn Sorcery", TypeLine: "Sorcery", ManaCost: "{3}{R}", Colors: []string{"R"}}
	land := game.Card{Name: "Drawn Mountain", TypeLine: "Basic Land — Mountain"}
	ids := sunbirdLibrary(me, sorcery, sorcery, sorcery, sorcery, sorcery, land, land,
		game.Card{Name: "Eighth", TypeLine: "Sorcery", ManaCost: "{9}"})
	drawn := ids[:5]
	setEnergy(t, g, me, 50)

	p7Activate(t, g, me, conduit, 0, game.ActivateAbilityParams{})
	if energyOf(me) != 0 {
		t.Errorf("energy after paying fifty = %d, want 0", energyOf(me))
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
	// Several free casts in one go: the first cast leaves the others open.
	for _, id := range append([]uuid.UUID{older}, drawn[:2]...) {
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
			t.Fatalf("free cast of %v: %v", id, err)
		}
	}
	if err := g.CastSpell(me.ID, heldLand, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}); err == nil {
		t.Error("a land was cast as a spell")
	}
	// The window closes on the holder's next pass.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if err := g.CastSpell(me.ID, drawn[3], game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}); err == nil {
		t.Error("the free cast outlived the holder's pass")
	}
	if !me.Hand.Contains(drawn[3]) {
		t.Error("an uncast card stays in hand")
	}
}

// --- Aetherworks Marvel --------------------------------------------------

func aetherworksTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	marvel := pushCatalogPermanent(g, me.ID, "Aetherworks Marvel", "Legendary Artifact", aetherworksMarvelOracle, false)
	return g, me, marvel
}

func TestAetherworksMarvelGivesEnergyForEachPermanentYouControlPutIntoAGraveyard(t *testing.T) {
	g, me, _ := aetherworksTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	mine := pushCatalogPermanent(g, me.ID, "My Bear", "Creature — Bear", "", false)
	token := pushCatalogPermanent(g, me.ID, "Bear Token", "Token Creature — Bear", "", false)
	theirs := pushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)
	bounced := pushCatalogPermanent(g, me.ID, "Bounced Bear", "Creature — Bear", "", false)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
	passPriorityAroundTable(t, g)
	if got := energyOf(me); got != 1 {
		t.Fatalf("energy after one of mine died = %d, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(token) })
	passPriorityAroundTable(t, g)
	if got := energyOf(me); got != 2 {
		t.Errorf("energy after a token of mine died = %d, want 2", got)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	passPriorityAroundTable(t, g)
	if got := energyOf(me); got != 2 {
		t.Errorf("an opponent's permanent changed my energy: %d", got)
	}
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(bounced) })
	passPriorityAroundTable(t, g)
	if got := energyOf(me); got != 2 {
		t.Errorf("a bounce is not a graveyard: energy %d", got)
	}
}

// aetherworksLibrary stacks six cards on top, top first: two lands, three
// spells and one more spell; and a distinct deep card below them.
func aetherworksLibrary(me *game.Player) (spells []uuid.UUID, lands []uuid.UUID, deep uuid.UUID) {
	deep = sunbirdLibrary(me, game.Card{Name: "Deep Card", TypeLine: "Basic Land — Plains"})[0]
	ids := sunbirdLibrary(me,
		game.Card{Name: "Top Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Huge Wurm", TypeLine: "Creature — Wurm", ManaCost: "{7}{G}", Power: 8, Toughness: 8},
		game.Card{Name: "Top Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Big Sorcery", TypeLine: "Sorcery", ManaCost: "{6}{R}"},
		game.Card{Name: "Cheap Elk", TypeLine: "Creature — Elk", ManaCost: "{G}"},
	)
	return []uuid.UUID{ids[1], ids[2], ids[4], ids[5]}, []uuid.UUID{ids[0], ids[3]}, deep
}

func aetherworksLook(t *testing.T, g *game.Game, me *game.Player, marvel uuid.UUID) *game.PendingChoice {
	t.Helper()
	setEnergy(t, g, me, 7)
	if err := g.ActivateCatalogAbility(me.ID, marvel, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityUntilChoice(t, g)
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the look should offer a free cast")
	}
	return c
}

func TestAetherworksMarvelPaysSixLooksAtSixAndCastsOneFreeOfThem(t *testing.T) {
	g, me, marvel := aetherworksTable(t)
	if c := specActivatedEnergy(t, aetherworksMarvelOracle, 0); c.Energy != 6 || !c.Tap {
		t.Fatalf("cost energy %d tap %v, want 6 and tap", c.Energy, c.Tap)
	}
	spells, lands, deep := aetherworksLibrary(me)
	c := aetherworksLook(t, g, me, marvel)
	if energyOf(me) != 1 {
		t.Errorf("energy after paying six from seven = %d, want 1", energyOf(me))
	}
	if c.ChooseDestination != game.ChooseCastFreeFromLibrary {
		t.Errorf("destination = %q, want the library free cast", c.ChooseDestination)
	}
	for _, id := range lands {
		if hasID(c.ChooseCards, id) {
			t.Errorf("a land was offered as a spell: %v", id)
		}
	}
	for _, id := range spells {
		if !hasID(c.ChooseCards, id) {
			t.Errorf("a looked-at spell %v was not offered (no mana value cap)", id)
		}
	}
	if hasID(c.ChooseCards, deep) {
		t.Error("the seventh card was not looked at")
	}
	// Looked at, not revealed: only the looker knows any of the six.
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	for _, card := range me.Library.Cards {
		if card.IsKnownTo(opp.ID) {
			t.Errorf("%s is known to an opponent after a private look", card.Name)
		}
	}

	wurm := spells[1]
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{wurm}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	if libraryIndex(me, wurm) != len(me.Library.Cards)-1 {
		t.Errorf("the pick waits on top of the library, at %d", libraryIndex(me, wurm))
	}
	for _, id := range append(append([]uuid.UUID{}, lands...), spells[0], spells[2], spells[3]) {
		if i := libraryIndex(me, id); i < 0 || i > 4 {
			t.Errorf("the rest go to the bottom (indices 0-4 above the deep card), found %d", i)
		}
	}
	if i := libraryIndex(me, deep); i != len(me.Library.Cards)-2 {
		t.Errorf("the deep card sits directly under the pick, found %d", i)
	}
	if err := g.CastSpell(me.ID, spells[0], game.CastSpellParams{FromZone: "library"}); err == nil {
		t.Error("the grant opens the pick and no other card")
	}
	// An 8-mana creature cast for nothing, in the main phase and at flash timing.
	if err := g.CastSpell(me.ID, wurm, game.CastSpellParams{FromZone: "library", Strict: true}); err != nil {
		t.Fatalf("free cast from the library: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wurm) {
		t.Error("the free creature should resolve")
	}
}

// An uncast pick, and a declined look, both end at the bottom.
func TestAetherworksMarvelUncastPickAndDeclinedLookGoToTheBottom(t *testing.T) {
	g, me, marvel := aetherworksTable(t)
	spells, _, _ := aetherworksLibrary(me)
	c := aetherworksLook(t, g, me, marvel)
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{spells[0]}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if i := libraryIndex(me, spells[0]); i < 0 || i > 5 {
		t.Errorf("an uncast pick goes to the bottom with the rest, found %d", i)
	}
	if err := g.CastSpell(me.ID, spells[0], game.CastSpellParams{FromZone: "library"}); err == nil {
		t.Error("the window closed on the pass")
	}

	g2, me2, marvel2 := aetherworksTable(t)
	spells2, _, deep2 := aetherworksLibrary(me2)
	c2 := aetherworksLook(t, g2, me2, marvel2)
	if err := g2.ResolveChooseCards(c2.ID, me2.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	for _, id := range spells2 {
		if i := libraryIndex(me2, id); i < 0 || i > 5 {
			t.Errorf("a declined look bottoms all six, found %d", i)
		}
	}
	if i := libraryIndex(me2, deep2); i != len(me2.Library.Cards)-1 {
		t.Errorf("the deep card is the top card after the six go under, found %d", i)
	}
}

// Six lands: nothing to cast, no prompt, and all six still go under.
func TestAetherworksMarvelLookAtAllLandsBottomsThemWithoutAPrompt(t *testing.T) {
	g, me, marvel := aetherworksTable(t)
	deep := sunbirdLibrary(me, game.Card{Name: "Deep Spell", TypeLine: "Sorcery", ManaCost: "{R}"})[0]
	var lands []uuid.UUID
	for i := 0; i < 6; i++ {
		lands = append(lands, sunbirdLibrary(me, game.Card{Name: "Land", TypeLine: "Basic Land — Forest"})...)
	}
	setEnergy(t, g, me, 6)
	if err := g.ActivateCatalogAbility(me.ID, marvel, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("no spell among the six, so no prompt")
	}
	for _, id := range lands {
		if i := libraryIndex(me, id); i < 0 || i > 5 {
			t.Errorf("a looked-at land goes to the bottom, found %d", i)
		}
	}
	if i := libraryIndex(me, deep); i != len(me.Library.Cards)-1 {
		t.Errorf("the deep card is on top, found %d", i)
	}
}

// Energy is a cost: five is one short.
func TestAetherworksMarvelRefusesWithFiveEnergy(t *testing.T) {
	g, me, marvel := aetherworksTable(t)
	setEnergy(t, g, me, 5)
	err := g.ActivateCatalogAbility(me.ID, marvel, 0, game.ActivateAbilityParams{})
	if !errors.Is(err, game.ErrInsufficientEnergy) {
		t.Fatalf("err = %v, want ErrInsufficientEnergy", err)
	}
}
