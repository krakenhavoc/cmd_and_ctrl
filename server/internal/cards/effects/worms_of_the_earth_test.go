package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// worms_of_the_earth_test.go — #3576, ADR 0187 slice 8: the real card.

const wormsOracle = "9638ec0c-c8bf-4aa5-a7ca-8677e4b9f241"

// wormsToOffer puts Worms under seat 1, walks to the next upkeep (the
// trigger is "each upkeep", so whoever's it is) and lets the trigger
// resolve, leaving the first link open.
func wormsToOffer(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	worms := pushCatalogPermanent(g, g.Seats[1].ID, "Worms of the Earth", "Enchantment", wormsOracle, false)
	for i := 0; i == 0 || g.Turn.Step != game.StepUpkeep; i++ {
		if i > 40 {
			t.Fatal("never reached an upkeep")
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	return worms
}

// wormsApnap is the seats in turn order from the active player.
func wormsApnap(g *game.Game) []uuid.UUID {
	var out []uuid.UUID
	for i := range g.Seats {
		out = append(out, g.Seats[(g.Turn.ActiveSeat+i)%len(g.Seats)].ID)
	}
	return out
}

// wormsAnswer answers each link in turn with `answer` (option nil is a
// decline) and returns the seats asked.
func wormsAnswer(t *testing.T, g *game.Game, answer func(*game.PendingChoice) (*int, []uuid.UUID)) []uuid.UUID {
	t.Helper()
	var asked []uuid.UUID
	for i := 0; i < 8; i++ {
		link := appOpenLink(g)
		if link == nil {
			passPriorityAroundTable(t, g)
			return asked
		}
		asked = append(asked, link.Chooser)
		opt, cards := answer(link)
		if err := g.ResolvePayUnlessWithOption(link.ID, link.Chooser, opt != nil, opt, cards); err != nil {
			t.Fatalf("answer link %d: %v", i, err)
		}
	}
	t.Fatal("the offer never ended")
	return nil
}

func TestWormsOfTheEarthStaysWhenNobodyPays(t *testing.T) {
	g := newCatalogGame(t)
	worms := wormsToOffer(t, g)
	asked := wormsAnswer(t, g, func(*game.PendingChoice) (*int, []uuid.UUID) { return nil, nil })
	want := wormsApnap(g)
	if len(asked) != 4 {
		t.Fatalf("asked %v, want every seat once", asked)
	}
	for i := range want {
		if asked[i] != want[i] {
			t.Errorf("asked[%d] = %v, want APNAP from the active player", i, asked[i])
		}
	}
	if !appInZone(g, worms, game.ZoneBattlefield) {
		t.Error("Worms left the battlefield with no payer")
	}
}

func TestWormsOfTheEarthDamageBranchDestroysIt(t *testing.T) {
	g := newCatalogGame(t)
	worms := wormsToOffer(t, g)
	taker := g.Seats[(g.Turn.ActiveSeat+2)%4]
	one := 1
	asked := wormsAnswer(t, g, func(c *game.PendingChoice) (*int, []uuid.UUID) {
		if c.Chooser == taker.ID {
			return &one, nil
		}
		return nil, nil
	})
	if len(asked) != 3 {
		t.Errorf("asked %d seats, want the first payer to end the offer after 3", len(asked))
	}
	if taker.Life != 35 {
		t.Errorf("life %d, want 40 - 5", taker.Life)
	}
	if !appInZone(g, worms, game.ZoneGraveyard) {
		t.Error("Worms was not destroyed")
	}
}

func TestWormsOfTheEarthSacrificeBranchDestroysIt(t *testing.T) {
	g := newCatalogGame(t)
	lands := map[uuid.UUID][2]uuid.UUID{}
	for _, s := range g.Seats {
		lands[s.ID] = [2]uuid.UUID{pushLand(g, s.ID, "Island A"), pushLand(g, s.ID, "Island B")}
	}
	worms := wormsToOffer(t, g)
	payer := g.Seats[(g.Turn.ActiveSeat+3)%4]
	a, b := lands[payer.ID][0], lands[payer.ID][1]
	zero := 0
	wormsAnswer(t, g, func(c *game.PendingChoice) (*int, []uuid.UUID) {
		if c.Chooser == payer.ID {
			return &zero, []uuid.UUID{a, b}
		}
		return nil, nil
	})
	if payer.Life != 40 {
		t.Errorf("life %d, a sacrifice must not hurt", payer.Life)
	}
	if appInZone(g, a, game.ZoneBattlefield) || appInZone(g, b, game.ZoneBattlefield) {
		t.Error("the sacrificed lands are still on the battlefield")
	}
	if !appInZone(g, worms, game.ZoneGraveyard) {
		t.Error("Worms was not destroyed")
	}
}

func TestWormsOfTheEarthStopsLandsPlayedAndEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, g.Seats[1].ID, "Worms of the Earth", "Enchantment", wormsOracle, false)
	land := lpHandLand(me, "Forest")
	lpAssertRefused(t, g, me, land, lpPlay(t, g, me, land), "Players can't play lands.")
	forest := erGraveyardCard(me, "Forest", "Basic Land — Forest")
	erReturn(t, g, forest)
	if erOnBattlefield(g, forest) {
		t.Error("a reanimated land entered")
	}
}
