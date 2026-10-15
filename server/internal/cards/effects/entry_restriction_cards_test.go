package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// entry_restriction_cards_test.go — #1985, the card half: each "… can't
// enter the battlefield" shape refuses exactly the cards it names and no
// others, and the cast clause each card prints beside it. The engine's
// per-path proof (land play, search, reanimation, exile return, hand put,
// token, permanent spell) is internal/game/entry_gate_test.go.

const (
	erKunorosOracle   = "d297ce39-da98-4351-be9d-24f4d09bee7f"
	erJailerOracle    = "7866ab01-22f3-4d87-bdb0-c0ae0e91ffc9"
	erRunestoneOracle = "4386ae4e-560a-4251-942c-0f84ba7448fa"
	erTestWormsOracle = "test-1985-lands-cant-enter"
)

// erGraveyardCard puts a card in p's graveyard.
func erGraveyardCard(p *game.Player, name, typeLine string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine = typeLine
	p.Graveyard.PushTop(c)
	return c.InstanceID
}

// erReturn reanimates a graveyard card through the engine's own path.
func erReturn(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardForEffect(id, game.ZoneBattlefield); err != nil {
			t.Errorf("return: %v", err)
		}
	})
}

// erFromLibrary puts a card on top of p's library and then onto the
// battlefield.
func erFromLibrary(t *testing.T, g *game.Game, p *game.Player, name, typeLine string) uuid.UUID {
	t.Helper()
	c := game.NewCard(name, p.ID)
	c.TypeLine = typeLine
	p.Library.PushTop(c)
	g.WithWriteLock(func() {
		if _, err := g.PutFromLibraryOntoBattlefieldForEffect(c.InstanceID, game.LibraryEntryOptions{}); err != nil {
			t.Errorf("library put: %v", err)
		}
	})
	return c.InstanceID
}

func erOnBattlefield(g *game.Game, id uuid.UUID) bool {
	_, ok := battlefieldCard(g, id)
	return ok
}

// Kunoros: creature cards leaving a graveyard stay there, its own
// controller's included; a noncreature card does not; spells can't be
// cast from graveyards; and its three keywords are printed.
func TestKunorosStopsCreatureCardsInGraveyards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kunoros := pushCatalogPermanent(g, me.ID, "Kunoros, Hound of Athreos", "Legendary Creature — Dog", erKunorosOracle, false)
	bear := erGraveyardCard(me, "Bear", "Creature — Bear")
	rock := erGraveyardCard(me, "Stone Idol", "Artifact")
	erReturn(t, g, bear)
	erReturn(t, g, rock)
	if erOnBattlefield(g, bear) || !me.Graveyard.Contains(bear) {
		t.Error("a creature card left the graveyard for the battlefield under Kunoros")
	}
	if !erOnBattlefield(g, rock) {
		t.Error("a noncreature card was refused by Kunoros")
	}
	// A creature card from a library is not named.
	if lib := erFromLibrary(t, g, me, "Elk", "Creature — Elk"); !erOnBattlefield(g, lib) {
		t.Error("a creature card from a library was refused by Kunoros")
	}
	for _, kw := range []string{"vigilance", "menace", "lifelink"} {
		if !effectiveAbilitiesContain(t, g, kunoros, kw) {
			t.Errorf("Kunoros lacks %s", kw)
		}
	}
	_, err := castFlashbackFromGraveyard(t, g, faithlessLootingOracle)
	assertCantCast(t, err, "Players can't cast spells from graveyards.")
}

// Soulless Jailer: every permanent card in a graveyard stays there, a land
// included (and a land played from a graveyard is refused before it is
// made); a card from a library is not named; noncreature spells can't be
// cast from a graveyard.
func TestSoullessJailerStopsPermanentCardsInGraveyards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, g.Seats[1].ID, "Soulless Jailer", "Artifact Creature — Phyrexian Golem", erJailerOracle, false)
	forest := erGraveyardCard(me, "Forest", "Basic Land — Forest")
	bear := erGraveyardCard(me, "Bear", "Creature — Bear")
	erReturn(t, g, forest)
	erReturn(t, g, bear)
	if erOnBattlefield(g, forest) || erOnBattlefield(g, bear) {
		t.Error("a permanent card left a graveyard for the battlefield under Soulless Jailer")
	}
	var playErr error
	g.WithWriteLock(func() {
		playErr = g.LandPlayGateLocked(me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}, game.ZoneGraveyard)
	})
	if playErr == nil {
		t.Error("a land play from a graveyard was allowed under Soulless Jailer")
	}
	if lib := erFromLibrary(t, g, me, "Elk", "Creature — Elk"); !erOnBattlefield(g, lib) {
		t.Error("a card from a library was refused by Soulless Jailer")
	}
	_, err := castFlashbackFromGraveyard(t, g, faithlessLootingOracle)
	assertCantCast(t, err, "noncreature spells from graveyards or exile")
}

// Weathered Runestone: nonland permanent cards from a graveyard or a
// library stay put; a land card does not.
func TestWeatheredRunestoneSparesLandCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, g.Seats[1].ID, "Weathered Runestone", "Artifact", erRunestoneOracle, false)
	forest := erGraveyardCard(me, "Forest", "Basic Land — Forest")
	bear := erGraveyardCard(me, "Bear", "Creature — Bear")
	erReturn(t, g, forest)
	erReturn(t, g, bear)
	if !erOnBattlefield(g, forest) {
		t.Error("a land card was refused by Weathered Runestone")
	}
	if erOnBattlefield(g, bear) || !me.Graveyard.Contains(bear) {
		t.Error("a creature card left the graveyard under Weathered Runestone")
	}
	if lib := erFromLibrary(t, g, me, "Elk", "Creature — Elk"); erOnBattlefield(g, lib) || !me.Library.Contains(lib) {
		t.Error("a creature card left the library under Weathered Runestone")
	}
	_, err := castFlashbackFromGraveyard(t, g, faithlessLootingOracle)
	assertCantCast(t, err, "graveyards or libraries")
}

// LandsCantEnter, the constructor #3576's Worms of the Earth declares,
// through the catalog: a land from anywhere is refused, a land token is
// never created, and the land play is refused with the clause.
func TestLandsCantEnterThroughTheCatalog(t *testing.T) {
	registerForTest(t, Spec{
		OracleID:          erTestWormsOracle,
		Name:              "Test Worms",
		EntryRestrictions: []game.EntryRestriction{LandsCantEnter("Lands can't enter the battlefield.")},
	})
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, g.Seats[1].ID, "Test Worms", "Enchantment", erTestWormsOracle, false)
	forest := erGraveyardCard(me, "Forest", "Basic Land — Forest")
	erReturn(t, g, forest)
	if erOnBattlefield(g, forest) {
		t.Error("a reanimated land entered")
	}
	if lib := erFromLibrary(t, g, me, "Island", "Basic Land — Island"); erOnBattlefield(g, lib) {
		t.Error("a land from the library entered")
	}
	var made []uuid.UUID
	g.WithWriteLock(func() {
		made, _ = g.CreateTokensForEffect(me.ID, game.Card{Name: "Land", TypeLine: "Token Land"}, 1, game.TokenEntryOptions{})
	})
	if len(made) != 0 {
		t.Error("a land token was created")
	}
	land := lpHandLand(me, "Forest")
	lpAssertRefused(t, g, me, land, lpPlay(t, g, me, land), "Lands can't enter the battlefield.")
}
