package testsu

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// wish_test.go — #3955. Wish does not name its card when it resolves: it
// stores a standing, one-use, until-end-of-turn right, and the card is
// chosen when it is played. A sideboard casts the card it wants; the
// database pool names one through the outside_play action.

// resolveWish casts Wish and lets it resolve.
func resolveWish(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Wish", "Sorcery", effectstest.WishOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
}

func rightIsLive(g *game.Game, p *game.Player) bool {
	var live bool
	g.WithWriteLock(func() { live = g.OutsidePlayRightLocked(p.ID) != nil })
	return live
}

// Resolving Wish asks for nothing: the card is chosen later, from the
// sideboard tab, at its own timing and cost. The first card played spends
// the right, and the others stay outside the game.
func TestWishNamesNoCardAndAnySideboardCardIsPlayedLater(t *testing.T) {
	bear := effectstest.WithManaCost(effectstest.OutsideCard("Grizzly Bears", "Creature — Bear"), "{2}")
	bear.Power, bear.Toughness = 2, 2
	g := effectstest.NewSideboardCatalogGame(t, 2, bear, effectstest.WithManaCost(effectstest.OutsideCard("Runeclaw Bear", "Creature — Bear"), "{2}"))
	me := g.Seats[0]

	resolveWish(t, g)
	if c := effectstest.LatestChooseCardsFor(g, me.ID); c != nil {
		t.Fatalf("Wish asked for a card as it resolved: %+v", c)
	}
	if !effectstest.OutsideGraveyardHas(me, "Wish") || !rightIsLive(g, me) {
		t.Fatalf("Wish in the graveyard %v, right live %v", effectstest.OutsideGraveyardHas(me, "Wish"), rightIsLive(g, me))
	}

	// The card is chosen NOW, after Wish: unpaid it is refused (strict), paid it is played.
	bearID := effectstest.SideCardID(me, "Grizzly Bears")
	if err := effectstest.CastFromOutside(g, me, bearID, true); err == nil {
		t.Fatal("a Wish play was made for nothing")
	}
	effectstest.FundOutside(me, 2)
	if err := effectstest.CastFromOutside(g, me, bearID, true); err != nil {
		t.Fatalf("paid play from outside: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if card, ok := effectstest.BattlefieldCard(g, bearID); !ok || card.Controller != me.ID {
		t.Error("Grizzly Bears did not resolve onto the battlefield")
	}
	// One card, once.
	effectstest.FundOutside(me, 2)
	if err := effectstest.CastFromOutside(g, me, effectstest.SideCardID(me, "Runeclaw Bear"), true); err == nil {
		t.Error("Wish let a second card be played from outside the game")
	}
	if got := effectstest.SideNamed(me); !effectstest.EqualNames(got, []string{"Runeclaw Bear"}) {
		t.Errorf("sideboard %v, want the unplayed card to stay outside", got)
	}
}

// The database pool: the action opens the pick at the moment the player
// wants a card, and a land named there is played as the land drop.
func TestWishPlaysALandFromTheDatabaseAndItCountsAsTheLandDrop(t *testing.T) {
	effectstest.InstallOutsideCollection(t)
	g := effectstest.NewSideboardCatalogGame(t, 2) // no sideboard: the database is the pool
	me := g.Seats[0]
	resolveWish(t, g)
	if effectstest.LatestCollectionPickFor(g, me.ID) != nil {
		t.Fatal("Wish opened the pick as it resolved")
	}

	effectstest.NameWished(t, g, me, "Forest")
	forest := effectstest.SideCardID(me, "Forest")
	if forest == uuid.Nil {
		t.Fatal("the named card was not minted outside the game")
	}
	var left int
	g.WithWriteLock(func() { left = g.LandDropsRemainingLocked(me.ID) })
	if err := effectstest.CastFromOutside(g, me, forest, true); err != nil {
		t.Fatalf("play the land: %v", err)
	}
	if card, ok := effectstest.BattlefieldCard(g, forest); !ok || card.Controller != me.ID {
		t.Fatal("the land is not on the battlefield")
	}
	var after int
	g.WithWriteLock(func() { after = g.LandDropsRemainingLocked(me.ID) })
	if after != left-1 {
		t.Errorf("land drops left %d, want %d: a land played from outside is the land drop", after, left-1)
	}
	if rightIsLive(g, me) {
		t.Error("the land did not spend the right")
	}
}

// Declining the pick spends nothing: the same Wish can name a card again.
func TestWishCancellingTheNamingSpendsNothing(t *testing.T) {
	effectstest.InstallOutsideCollection(t)
	g := effectstest.NewSideboardCatalogGame(t, 2)
	me := g.Seats[0]
	resolveWish(t, g)

	if err := g.NameOutsideCardToPlay(me.ID); err != nil {
		t.Fatal(err)
	}
	c := effectstest.LatestCollectionPickFor(g, me.ID)
	if err := g.ResolveOutsideCollection(c.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if !rightIsLive(g, me) || len(me.SideboardCards()) != 0 {
		t.Fatalf("a cancelled pick: right live %v, %d cards minted", rightIsLive(g, me), len(me.SideboardCards()))
	}
	effectstest.NameWished(t, g, me, "Forest")
	if err := effectstest.CastFromOutside(g, me, effectstest.SideCardID(me, "Forest"), true); err != nil {
		t.Fatalf("play after cancelling once: %v", err)
	}
}

// The right is checked where the card is named: without Wish there is
// nothing to name a card for, and a sideboard pool is never asked.
func TestWishNamingNeedsTheRightAndTheDatabasePool(t *testing.T) {
	effectstest.InstallOutsideCollection(t)
	g := effectstest.NewSideboardCatalogGame(t, 2)
	me := g.Seats[0]
	if err := g.NameOutsideCardToPlay(me.ID); err == nil {
		t.Error("a card was named with no Wish")
	}
	g2 := effectstest.NewSideboardCatalogGame(t, 2, effectstest.OutsideCard("Forest", "Basic Land — Forest"))
	resolveWish(t, g2)
	if err := g2.NameOutsideCardToPlay(g2.Seats[0].ID); err == nil {
		t.Error("a deck with a sideboard was asked to name a database card")
	}
}

// The entry replacement of a land played from outside the game still
// applies: Temple of Abandon enters TAPPED, as a replacement (no tap
// event), and a shockland pauses its entry on the payment prompt and
// resumes it.
func TestWishedLandKeepsItsOwnEntryReplacement(t *testing.T) {
	effectstest.InstallOutsideCollection(t)
	g := effectstest.NewSideboardCatalogGame(t, 2)
	me := g.Seats[0]
	resolveWish(t, g)
	effectstest.NameWished(t, g, me, "Temple of Abandon")
	temple := effectstest.SideCardID(me, "Temple of Abandon")
	if err := effectstest.CastFromOutside(g, me, temple, true); err != nil {
		t.Fatalf("play the Temple: %v", err)
	}
	card, ok := effectstest.BattlefieldCard(g, temple)
	if !ok || !card.Tapped {
		t.Fatalf("Temple of Abandon on the battlefield %v, tapped %v: want it to enter tapped", ok, card.Tapped)
	}
	if n := effectstest.TapEventsFor(g, temple); n != 0 {
		t.Errorf("%d tap events; the land should have ENTERED tapped", n)
	}
}

func TestWishedShocklandPausesAndResumesItsEntry(t *testing.T) {
	effectstest.InstallOutsideCollection(t)
	g := effectstest.NewSideboardCatalogGame(t, 2)
	me := g.Seats[0]
	resolveWish(t, g)
	effectstest.NameWished(t, g, me, "Steam Vents")
	vents := effectstest.SideCardID(me, "Steam Vents")
	life := me.Life
	if err := effectstest.CastFromOutside(g, me, vents, true); err != nil {
		t.Fatalf("play the shockland: %v", err)
	}
	if effectstest.EntryPayLifeChoiceFor(g, me.ID) == nil {
		t.Fatal("the shockland asked nothing as it entered from outside the game")
	}
	if _, ok := effectstest.BattlefieldCard(g, vents); ok {
		t.Fatal("the land entered before the choice was made")
	}
	effectstest.AnswerEntryPayLife(t, g, me.ID, true)
	card, ok := effectstest.BattlefieldCard(g, vents)
	if !ok || card.Tapped {
		t.Fatalf("on the battlefield %v, tapped %v: want it untapped after paying", ok, card.Tapped)
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
	if got := effectstest.SideNamed(me); slices.Contains(got, "Steam Vents") {
		t.Errorf("the land is still outside the game: %v", got)
	}
	if rightIsLive(g, me) {
		t.Error("the paused land play did not spend the right")
	}
}
