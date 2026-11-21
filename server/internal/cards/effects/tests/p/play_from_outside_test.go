package testsp

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestSpawnsireCastsAnyNumberOfEldraziFromTheSideboardFree(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2,
		effectstest.WithManaCost(effectstest.OutsideCard("Ulamog's Crusher", "Creature — Eldrazi"), "{8}"),
		effectstest.WithManaCost(effectstest.OutsideCard("Emrakul's Hatcher", "Creature — Eldrazi Drone"), "{3}"),
		effectstest.WithManaCost(effectstest.OutsideCard("Annihilator Drone", "Creature — Eldrazi"), "{5}"),
		effectstest.WithManaCost(effectstest.OutsideCard("Grizzly Bears", "Creature — Bear"), "{1}{G}"),
		effectstest.OutsideCard("Eldrazi Temple", "Land"))
	me := g.Seats[0]
	src := effectstest.SpawnsireReady(t, g, me)

	if err := g.ActivateCatalogAbility(me.ID, src, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the {20} ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatalf("no prompt: %+v", g.PendingChoices)
	}
	if c.ChooseDestination != game.ChooseCastFreeFromOutside || c.ChooseMin != 0 {
		t.Errorf("destination %q min %d", c.ChooseDestination, c.ChooseMin)
	}
	if got := effectstest.CandidateNames(g, c); !effectstest.EqualNames(got, []string{"Ulamog's Crusher", "Emrakul's Hatcher", "Annihilator Drone"}) {
		t.Errorf("candidates %v, want the three Eldrazi creature cards and nothing else", got)
	}
	if c.ChooseMax != 3 {
		t.Errorf("max %d, want any number of the three candidates", c.ChooseMax)
	}
	crusher, hatcher := effectstest.SideCardID(me, "Ulamog's Crusher"), effectstest.SideCardID(me, "Emrakul's Hatcher")
	drone := effectstest.SideCardID(me, "Annihilator Drone")
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{crusher, hatcher}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}

	// Both are cast for free (strict: nothing in the pool), in either order.
	if err := effectstest.CastFromOutside(g, me, hatcher, true); err != nil {
		t.Fatalf("free cast of the first: %v", err)
	}
	if err := effectstest.CastFromOutside(g, me, crusher, true); err != nil {
		t.Fatalf("free cast of the second: %v", err)
	}
	// The one not chosen has no permission.
	if err := effectstest.CastFromOutside(g, me, drone, true); err == nil {
		t.Error("an Eldrazi that was not chosen was cast from outside the game")
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{crusher, hatcher} {
		if card, ok := effectstest.BattlefieldCard(g, id); !ok || card.Controller != me.ID {
			t.Errorf("%s did not resolve onto the battlefield", id)
		}
	}
	if got := effectstest.SideNamed(me); !effectstest.EqualNames(got, []string{"Annihilator Drone", "Grizzly Bears", "Eldrazi Temple"}) {
		t.Errorf("sideboard %v, want the uncast cards to stay outside the game", got)
	}
}

func TestSpawnsireUncastEldraziStayOutsideWhenTheWindowCloses(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2,
		effectstest.WithManaCost(effectstest.OutsideCard("Ulamog's Crusher", "Creature — Eldrazi"), "{8}"),
		effectstest.WithManaCost(effectstest.OutsideCard("Emrakul's Hatcher", "Creature — Eldrazi Drone"), "{3}"))
	me := g.Seats[0]
	src := effectstest.SpawnsireReady(t, g, me)
	if err := g.ActivateCatalogAbility(me.ID, src, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	crusher, hatcher := effectstest.SideCardID(me, "Ulamog's Crusher"), effectstest.SideCardID(me, "Emrakul's Hatcher")
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{crusher, hatcher}); err != nil {
		t.Fatal(err)
	}
	if err := effectstest.CastFromOutside(g, me, crusher, true); err != nil {
		t.Fatalf("cast one: %v", err)
	}
	// The holder passes priority: the other window closes, the card stays.
	effectstest.PassPriorityAroundTable(t, g)
	if err := effectstest.CastFromOutside(g, me, hatcher, true); err == nil {
		t.Error("the free cast outlived the holder's pass")
	}
	if got := effectstest.SideNamed(me); !effectstest.EqualNames(got, []string{"Emrakul's Hatcher"}) {
		t.Errorf("sideboard %v, want the uncast Eldrazi to stay outside", got)
	}
}

func TestSpawnsireCastsEldraziFromTheDatabaseWithoutASideboard(t *testing.T) {
	effectstest.InstallOutsideCollection(t)
	g := effectstest.NewSideboardCatalogGame(t, 2)
	me := g.Seats[0]
	src := effectstest.SpawnsireReady(t, g, me)
	if err := g.ActivateCatalogAbility(me.ID, src, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestCollectionPickFor(g, me.ID)
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
		if err := effectstest.CastFromOutside(g, me, id, true); err != nil {
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
	effects.Register(effects.Spec{
		OracleID:      "test-outside-castable-zone",
		Name:          "Sideboard Opener",
		CastableZones: []game.ZoneKind{game.ZoneOutside},
	})
}

func TestSpawnsireAndWishAreFull(t *testing.T) {
	for _, id := range []string{effectstest.SpawnsireOracle, effectstest.WishOracle} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Fatalf("%s is not in the catalog", id)
		}
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s is %v, want Full", spec.Name, spec.Completeness)
		}
	}
}
