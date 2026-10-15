package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// locke_exiled_mill_test.go — CR 701.17c (#3846): a card a replacement
// put somewhere else instead of the graveyard is still one of the
// milled cards, and Mug's permission covers it wherever it went.

func lockeBoard(t *testing.T) (g *game.Game, me *game.Player, locke uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[0]
	locke = pushDiesCreatureForTest(g, me.ID, "Locke, Treasure Hunter", lockeOracle,
		"Legendary Creature — Human Rogue", 2, 3)
	return g, me, locke
}

func runLockeMug(t *testing.T, g *game.Game, locke uuid.UUID) {
	t.Helper()
	declareAttack(t, g, g.Seats[1].ID, locke)
	passPriorityAroundTable(t, g)
}

// Rest in Peace exiles every milled card. Each is still milled: the land
// makes the Treasure, and a spell can be cast from exile, once.
func TestLockeMugCoversCardsRestInPeaceExiled(t *testing.T) {
	g, me, locke := lockeBoard(t)
	pushTokenReplacementCard(g, drcRestInPeace, "Rest in Peace", "Enchantment", g.Seats[2].ID)
	land := topOfLibrary(me, "My Land", "Land", "")
	a := topOfLibrary(g.Seats[1], "Their Shock", "Instant", "{R}")
	b := topOfLibrary(g.Seats[2], "Their Opt", "Instant", "{U}")
	topOfLibrary(g.Seats[3], "Their Bear", "Creature — Bear", "{1}{G}")
	treasures := monTokens(g, me.ID, "Treasure")

	runLockeMug(t, g, locke)
	for _, id := range []uuid.UUID{land, a, b} {
		if z := e2Zone(g, id); z != game.ZoneExile {
			t.Fatalf("a milled card is in %q, want exile (Rest in Peace)", z)
		}
	}
	if got := monTokens(g, me.ID, "Treasure") - treasures; got != 1 {
		t.Errorf("%d Treasures, want 1: the exiled land was still milled", got)
	}
	if err := g.CastSpell(me.ID, a, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast an exiled milled card under Mug: %v", err)
	}
	if err := g.CastSpell(me.ID, b, game.CastSpellParams{FromZone: "exile"}); err == nil {
		t.Error("a second spell was cast from among the milled cards")
	}
}

// Leyline of the Void exiles only the cards bound for an opponent's
// graveyard, so the milled cards end up in two zones. Mug still allows
// ONE spell, from either zone, not one per zone.
func TestLockeMugSpansTwoZonesWithOneSpell(t *testing.T) {
	g, me, locke := lockeBoard(t)
	// Seat 2 controls the Leyline: seat 2's own card reaches its graveyard,
	// everyone else's goes to exile.
	pushTokenReplacementCard(g, leylineOfTheVoidOracle, "Leyline of the Void", "Enchantment", g.Seats[2].ID)
	mine := topOfLibrary(me, "My Bolt", "Instant", "{R}")
	a := topOfLibrary(g.Seats[1], "Their Shock", "Instant", "{R}")
	b := topOfLibrary(g.Seats[2], "Their Opt", "Instant", "{U}")
	topOfLibrary(g.Seats[3], "Their Bear", "Creature — Bear", "{1}{G}")

	runLockeMug(t, g, locke)
	if z := e2Zone(g, a); z != game.ZoneExile {
		t.Fatalf("an opponent's card is in %q, want exile (Leyline of the Void)", z)
	}
	if z := e2Zone(g, b); z != game.ZoneGraveyard {
		t.Fatalf("the Leyline controller's own card is in %q, want their graveyard", z)
	}
	if z := e2Zone(g, mine); z != game.ZoneExile {
		t.Fatalf("my card is in %q, want exile (I am the Leyline controller's opponent)", z)
	}
	if err := g.CastSpell(me.ID, b, game.CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast the graveyard half under Mug: %v", err)
	}
	if err := g.CastSpell(me.ID, a, game.CastSpellParams{FromZone: "exile"}); err == nil {
		t.Error("a second spell was cast from the exile half: the zones do not share one spell")
	}
}

// And the other order: the exile half spends the one spell, so the
// graveyard half is closed.
func TestLockeMugOneSpellFromExileClosesTheGraveyardHalf(t *testing.T) {
	g, me, locke := lockeBoard(t)
	pushTokenReplacementCard(g, leylineOfTheVoidOracle, "Leyline of the Void", "Enchantment", g.Seats[2].ID)
	a := topOfLibrary(g.Seats[1], "Their Shock", "Instant", "{R}")
	b := topOfLibrary(g.Seats[2], "Their Opt", "Instant", "{U}")
	topOfLibrary(g.Seats[3], "Their Bear", "Creature — Bear", "{1}{G}")

	runLockeMug(t, g, locke)
	if err := g.CastSpell(me.ID, a, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast the exile half under Mug: %v", err)
	}
	if err := g.CastSpell(me.ID, b, game.CastSpellParams{FromZone: "graveyard"}); err == nil {
		t.Error("a second spell was cast from the graveyard half")
	}
}

// A card in no zone a permission can open (a milled card a player
// returned to hand, here by hand) is simply not covered, and the grant
// over the rest still stands.
func TestGrantAcrossZonesSkipsOtherZonesAndNamesTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	inGrave := deck3655GraveyardCreature(me, "In Grave", 1)
	inExile := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: inExile, Name: "In Exile", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	inHand := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: inHand, Name: "In Hand", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})

	g.WithWriteLock(func() {
		var cards []game.Card
		for _, id := range []uuid.UUID{inGrave, inExile, inHand} {
			c, _ := g.LookupCardForEffect(id)
			cards = append(cards, c)
		}
		g.GrantCastPermissionAcrossZonesForEffect(game.CastPermission{
			Player: me.ID, CastOnly: true, CastsLeft: 1, SourceName: "test", Label: "test",
		}, cards)
	})
	var zones []game.ZoneKind
	var groups []uuid.UUID
	for _, p := range me.CastPermissions {
		zones = append(zones, p.Zone)
		groups = append(groups, p.BudgetGroup)
	}
	if len(zones) != 2 || zones[0] == zones[1] {
		t.Fatalf("permissions over zones %v, want one for the graveyard and one for exile", zones)
	}
	if groups[0] == uuid.Nil || groups[0] != groups[1] {
		t.Errorf("budget groups %v, want one shared non-nil group", groups)
	}
	if e2Permission(g, inHand) != nil {
		t.Error("a card in hand was granted a stored permission")
	}
}
