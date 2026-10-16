package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// play_from_outside_test.go — #3955: Spawnsire of Ulamog's {20} and Wish,
// over both pools (a deck's sideboard, the card database).

const (
	wishOracle           = "2ccbf0ee-c856-4f67-8579-811fe8571a2e"
	templeOfAbandonOracl = "3baa8e38-ef93-435d-b63e-f781d5bfcc68"
)

// fundOutside puts `n` colorless mana in the pool, for a paid play.
func fundOutside(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
}

func castFromOutside(g *game.Game, p *game.Player, id uuid.UUID, strict bool) error {
	return g.CastSpell(p.ID, id, game.CastSpellParams{FromZone: "outside", Strict: strict})
}

// installOutsideCollection is the test database for the cast tests.
func installOutsideCollection(t *testing.T) {
	t.Helper()
	prev := game.OutsideCollection
	game.OutsideCollection = testCollection{cards: []game.Card{
		dbCard("Grizzly Bears", "Creature — Bear", "", "G"),
		dbCard("Forest", "Basic Land — Forest", ""),
		dbCard("Temple of Abandon", "Land", templeOfAbandonOracl),
		dbCard("Steam Vents", "Land — Island Mountain", steamVentsOracle),
		dbCard("Ulamog's Crusher", "Creature — Eldrazi", "", ""),
		dbCard("Emrakul's Hatcher", "Creature — Eldrazi Drone", "", ""),
		dbCard("Lightning Bolt", "Instant", lightningBoltOracle, "R"),
	}}
	t.Cleanup(func() { game.OutsideCollection = prev })
}

func withManaCost(c game.Card, cost string) game.Card {
	c.ManaCost = cost
	return c
}

// --- Wish ---------------------------------------------------------------

func TestWishPlaysACardFromTheSideboardAtItsCost(t *testing.T) {
	bear := withManaCost(outsideCard("Grizzly Bears", "Creature — Bear"), "{2}")
	bear.Power, bear.Toughness = 2, 2
	g := newSideboardCatalogGame(t, 2, bear, outsideCard("Lightning Bolt", "Instant"))
	me := g.Seats[0]

	castCatalogSpell(t, g, "Wish", "Sorcery", wishOracle, nil)
	passPriorityAroundTable(t, g)
	c := latestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatalf("no prompt for the card: %+v", g.PendingChoices)
	}
	if c.ChooseDestination != game.ChoosePlayFromOutside || c.ChooseMin != 0 || c.ChooseMax != 1 {
		t.Errorf("destination %q bounds %d..%d, want a play of up to one", c.ChooseDestination, c.ChooseMin, c.ChooseMax)
	}
	if got := candidateNames(g, c); !equalNames(got, []string{"Grizzly Bears", "Lightning Bolt"}) {
		t.Errorf("candidates %v, want every card outside the game", got)
	}
	bearID := sideCardID(me, "Grizzly Bears")
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{bearID}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !outsideGraveyardHas(me, "Wish") {
		t.Error("Wish did not go to the graveyard")
	}

	// It is PLAYED, for its own cost: unpaid it is refused (strict), paid
	// it goes on the stack and the card leaves the container.
	if err := castFromOutside(g, me, bearID, true); err == nil {
		t.Fatal("Grizzly Bears was cast from outside the game for nothing")
	}
	fundOutside(me, 2)
	if err := castFromOutside(g, me, bearID, true); err != nil {
		t.Fatalf("paid play from outside: %v", err)
	}
	passPriorityAroundTable(t, g)
	if card, ok := battlefieldCard(g, bearID); !ok || card.Controller != me.ID {
		t.Error("Grizzly Bears did not resolve onto the battlefield")
	}
	if got := sideNamed(me); !equalNames(got, []string{"Lightning Bolt"}) {
		t.Errorf("sideboard %v, want the unchosen card to stay outside", got)
	}
}

func TestWishPermissionIsForTheChosenCardOnly(t *testing.T) {
	g := newSideboardCatalogGame(t, 2,
		withManaCost(outsideCard("Grizzly Bears", "Creature — Bear"), "{1}{G}"),
		withManaCost(outsideCard("Runeclaw Bear", "Creature — Bear"), "{1}{G}"))
	me := g.Seats[0]
	castCatalogSpell(t, g, "Wish", "Sorcery", wishOracle, nil)
	passPriorityAroundTable(t, g)
	c := latestChooseCardsFor(g, me.ID)
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{sideCardID(me, "Grizzly Bears")}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	fundOutside(me, 4)
	if err := castFromOutside(g, me, sideCardID(me, "Runeclaw Bear"), true); err == nil {
		t.Error("Wish let a second card be played from outside the game")
	}
}

func TestWishPlaysALandFromTheDatabaseAndItCountsAsTheLandDrop(t *testing.T) {
	installOutsideCollection(t)
	g := newSideboardCatalogGame(t, 2) // no sideboard: the database is the pool
	me := g.Seats[0]
	castCatalogSpell(t, g, "Wish", "Sorcery", wishOracle, nil)
	passPriorityAroundTable(t, g)

	c := latestCollectionPickFor(g, me.ID)
	if c == nil {
		t.Fatalf("no collection pick: %+v", g.PendingChoices)
	}
	if c.ChooseDestination != game.ChoosePlayFromOutside || c.ChooseMax != 1 {
		t.Errorf("destination %q max %d", c.ChooseDestination, c.ChooseMax)
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Forest"}); err != nil {
		t.Fatalf("ResolveOutsideCollection: %v", err)
	}
	forest := sideCardID(me, "Forest")
	if forest == uuid.Nil {
		t.Fatal("the named card was not minted outside the game")
	}
	drops := me.LandDropsPerTurn
	var left int
	g.WithWriteLock(func() { left = g.LandDropsRemainingLocked(me.ID) })
	if err := castFromOutside(g, me, forest, true); err != nil {
		t.Fatalf("play the land: %v", err)
	}
	if card, ok := battlefieldCard(g, forest); !ok || card.Controller != me.ID {
		t.Fatal("the land is not on the battlefield")
	}
	var after int
	g.WithWriteLock(func() { after = g.LandDropsRemainingLocked(me.ID) })
	if after != left-1 {
		t.Errorf("land drops left %d (of %d), want %d: a land played from outside is the land drop", after, drops, left-1)
	}
}

// The entry replacement of a land played from outside the game still
// applies: Temple of Abandon enters TAPPED, as a replacement (no tap
// event), and a shockland pauses its entry on the payment prompt and
// resumes it.
func TestWishedLandKeepsItsOwnEntryReplacement(t *testing.T) {
	installOutsideCollection(t)
	g := newSideboardCatalogGame(t, 2)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Wish", "Sorcery", wishOracle, nil)
	passPriorityAroundTable(t, g)
	c := latestCollectionPickFor(g, me.ID)
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Temple of Abandon"}); err != nil {
		t.Fatal(err)
	}
	temple := sideCardID(me, "Temple of Abandon")
	if err := castFromOutside(g, me, temple, true); err != nil {
		t.Fatalf("play the Temple: %v", err)
	}
	card, ok := battlefieldCard(g, temple)
	if !ok || !card.Tapped {
		t.Fatalf("Temple of Abandon on the battlefield %v, tapped %v: want it to enter tapped", ok, card.Tapped)
	}
	if n := tapEventsFor(g, temple); n != 0 {
		t.Errorf("%d tap events; the land should have ENTERED tapped", n)
	}
}

func TestWishedShocklandPausesAndResumesItsEntry(t *testing.T) {
	installOutsideCollection(t)
	g := newSideboardCatalogGame(t, 2)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Wish", "Sorcery", wishOracle, nil)
	passPriorityAroundTable(t, g)
	c := latestCollectionPickFor(g, me.ID)
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Steam Vents"}); err != nil {
		t.Fatal(err)
	}
	vents := sideCardID(me, "Steam Vents")
	life := me.Life
	if err := castFromOutside(g, me, vents, true); err != nil {
		t.Fatalf("play the shockland: %v", err)
	}
	if entryPayLifeChoiceFor(g, me.ID) == nil {
		t.Fatal("the shockland asked nothing as it entered from outside the game")
	}
	if _, ok := battlefieldCard(g, vents); ok {
		t.Fatal("the land entered before the choice was made")
	}
	answerEntryPayLife(t, g, me.ID, true)
	card, ok := battlefieldCard(g, vents)
	if !ok || card.Tapped {
		t.Fatalf("on the battlefield %v, tapped %v: want it untapped after paying", ok, card.Tapped)
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
	if got := sideNamed(me); slices.Contains(got, "Steam Vents") {
		t.Errorf("the land is still outside the game: %v", got)
	}
}

// --- Spawnsire of Ulamog ------------------------------------------------

func spawnsireReady(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	toMain(t, g)
	return pushAnnCatalogCreature(g, me.ID, "Spawnsire of Ulamog", spawnsireOracle, 7, 11)
}

func TestSpawnsireCastsAnyNumberOfEldraziFromTheSideboardFree(t *testing.T) {
	g := newSideboardCatalogGame(t, 2,
		withManaCost(outsideCard("Ulamog's Crusher", "Creature — Eldrazi"), "{8}"),
		withManaCost(outsideCard("Emrakul's Hatcher", "Creature — Eldrazi Drone"), "{3}"),
		withManaCost(outsideCard("Annihilator Drone", "Creature — Eldrazi"), "{5}"),
		withManaCost(outsideCard("Grizzly Bears", "Creature — Bear"), "{1}{G}"),
		outsideCard("Eldrazi Temple", "Land"))
	me := g.Seats[0]
	src := spawnsireReady(t, g, me)

	if err := g.ActivateCatalogAbility(me.ID, src, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the {20} ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := latestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatalf("no prompt: %+v", g.PendingChoices)
	}
	if c.ChooseDestination != game.ChooseCastFreeFromOutside || c.ChooseMin != 0 {
		t.Errorf("destination %q min %d", c.ChooseDestination, c.ChooseMin)
	}
	if got := candidateNames(g, c); !equalNames(got, []string{"Ulamog's Crusher", "Emrakul's Hatcher", "Annihilator Drone"}) {
		t.Errorf("candidates %v, want the three Eldrazi creature cards and nothing else", got)
	}
	if c.ChooseMax != 3 {
		t.Errorf("max %d, want any number of the three candidates", c.ChooseMax)
	}
	crusher, hatcher := sideCardID(me, "Ulamog's Crusher"), sideCardID(me, "Emrakul's Hatcher")
	drone := sideCardID(me, "Annihilator Drone")
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{crusher, hatcher}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}

	// Both are cast for free (strict: nothing in the pool), in either order.
	if err := castFromOutside(g, me, hatcher, true); err != nil {
		t.Fatalf("free cast of the first: %v", err)
	}
	if err := castFromOutside(g, me, crusher, true); err != nil {
		t.Fatalf("free cast of the second: %v", err)
	}
	// The one not chosen has no permission.
	if err := castFromOutside(g, me, drone, true); err == nil {
		t.Error("an Eldrazi that was not chosen was cast from outside the game")
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{crusher, hatcher} {
		if card, ok := battlefieldCard(g, id); !ok || card.Controller != me.ID {
			t.Errorf("%s did not resolve onto the battlefield", id)
		}
	}
	if got := sideNamed(me); !equalNames(got, []string{"Annihilator Drone", "Grizzly Bears", "Eldrazi Temple"}) {
		t.Errorf("sideboard %v, want the uncast cards to stay outside the game", got)
	}
}

func TestSpawnsireUncastEldraziStayOutsideWhenTheWindowCloses(t *testing.T) {
	g := newSideboardCatalogGame(t, 2,
		withManaCost(outsideCard("Ulamog's Crusher", "Creature — Eldrazi"), "{8}"),
		withManaCost(outsideCard("Emrakul's Hatcher", "Creature — Eldrazi Drone"), "{3}"))
	me := g.Seats[0]
	src := spawnsireReady(t, g, me)
	if err := g.ActivateCatalogAbility(me.ID, src, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	c := latestChooseCardsFor(g, me.ID)
	crusher, hatcher := sideCardID(me, "Ulamog's Crusher"), sideCardID(me, "Emrakul's Hatcher")
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{crusher, hatcher}); err != nil {
		t.Fatal(err)
	}
	if err := castFromOutside(g, me, crusher, true); err != nil {
		t.Fatalf("cast one: %v", err)
	}
	// The holder passes priority: the other window closes, the card stays.
	passPriorityAroundTable(t, g)
	if err := castFromOutside(g, me, hatcher, true); err == nil {
		t.Error("the free cast outlived the holder's pass")
	}
	if got := sideNamed(me); !equalNames(got, []string{"Emrakul's Hatcher"}) {
		t.Errorf("sideboard %v, want the uncast Eldrazi to stay outside", got)
	}
}

func TestSpawnsireCastsEldraziFromTheDatabaseWithoutASideboard(t *testing.T) {
	installOutsideCollection(t)
	g := newSideboardCatalogGame(t, 2)
	me := g.Seats[0]
	src := spawnsireReady(t, g, me)
	if err := g.ActivateCatalogAbility(me.ID, src, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	c := latestCollectionPickFor(g, me.ID)
	if c == nil {
		t.Fatalf("no collection pick: %+v", g.PendingChoices)
	}
	if want := []string{"Ulamog's Crusher", "Emrakul's Hatcher"}; !slices.Equal(c.CollectionSuggestions, want) {
		t.Errorf("suggestions %v, want only the Eldrazi %v", c.CollectionSuggestions, want)
	}
	if c.ChooseDestination != game.ChooseCastFreeFromOutside || c.ChooseMin != 0 || c.ChooseMax != game.OutsideAnyNumber {
		t.Errorf("destination %q bounds %d..%d", c.ChooseDestination, c.ChooseMin, c.ChooseMax)
	}
	// A non-Eldrazi is refused by name.
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Grizzly Bears"}); err == nil {
		t.Fatal("the pick accepted a card that is not an Eldrazi")
	}
	// "Any number": two copies of one card are two cards you own.
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Ulamog's Crusher", "Ulamog's Crusher", "Emrakul's Hatcher"}); err != nil {
		t.Fatalf("ResolveOutsideCollection: %v", err)
	}
	var ids []uuid.UUID
	for _, card := range me.SideboardCards() {
		ids = append(ids, card.InstanceID)
	}
	if len(ids) != 3 {
		t.Fatalf("minted %d cards, want 3", len(ids))
	}
	for _, id := range ids {
		if err := castFromOutside(g, me, id, true); err != nil {
			t.Fatalf("free cast of a minted Eldrazi: %v", err)
		}
	}
	n := 0
	for _, id := range ids {
		if g.Stack.Contains(id) {
			n++
		}
	}
	if n != 3 {
		t.Errorf("%d of the 3 minted Eldrazi are on the stack", n)
	}
	if me.Sideboard.Size() != 0 {
		t.Errorf("%d cards are still outside the game", me.Sideboard.Size())
	}
}

// A card file cannot open the container for every copy of itself: a
// CastableZones that names it is refused at boot, as exile is.
func TestRegisterRefusesACardLevelOutsideCastZone(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("a card declaring ZoneOutside in CastableZones was registered")
		}
	}()
	Register(Spec{
		OracleID:      "test-outside-castable-zone",
		Name:          "Sideboard Opener",
		CastableZones: []game.ZoneKind{game.ZoneOutside},
	})
}

func TestSpawnsireAndWishAreFull(t *testing.T) {
	for _, id := range []string{spawnsireOracle, wishOracle} {
		spec, ok := Lookup(id)
		if !ok {
			t.Fatalf("%s is not in the catalog", id)
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s is %v, want Full", spec.Name, spec.Completeness)
		}
	}
}
