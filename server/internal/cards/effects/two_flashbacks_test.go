package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// two_flashbacks_test.go is #3938 on the two cards it was filed for
// (ADR 0066 decision 3 as amended 2026-10-16). A card that prints
// flashback and is granted flashback has two flashback abilities (CR
// 113.2c), and its owner casts it with either (CR 702.34a): the printed
// one under "flashback", the granted one under
// game.GrantedInstanceKey("flashback").

var grantedFlashback = game.GrantedInstanceKey("flashback")

// flashbackOffers is the cast offers for a card in seat's graveyard, by
// claim key, each with its mana cost.
func flashbackOffers(t *testing.T, g *game.Game, seat, id uuid.UUID) map[string]string {
	t.Helper()
	out := map[string]string{}
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			t.Fatal("the card is gone")
		}
		for _, ac := range g.CastOffersForLocked(seat, c, game.ZoneGraveyard, g.CastPermissionForLocked(seat, c, game.ZoneGraveyard)) {
			if ac == nil {
				out[""] = c.ManaCost
				continue
			}
			out[ac.Key] = ac.ManaCost
		}
	})
	return out
}

// setManaCost gives a seeded card its printed mana cost, which the
// catalog game does not fill in.
func setManaCost(g *game.Game, seat *game.Player, id uuid.UUID, cost string) {
	g.WithWriteLock(func() {
		for i := range seat.Graveyard.Cards {
			if seat.Graveyard.Cards[i].InstanceID == id {
				seat.Graveyard.Cards[i].ManaCost = cost
			}
		}
	})
}

func addRed(g *game.Game, seat *game.Player, n int) {
	g.WithWriteLock(func() {
		for range n {
			seat.ManaPool.AddMana(game.ManaToken{Color: "R"})
		}
	})
}

// Lier on a Faithless Looting: its printed {2}{R} or its {R} mana cost.
// Both are offered, the bot is offered both, the {R} one is cast for {R},
// and the spell is a flashback spell, exiled as it leaves the stack.
func TestLierOffersAFlashbackCardBothFlashbacks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	lier := pushCatalogPermanent(g, me.ID, "Lier, Disciple of the Drowned", "Legendary Creature — Human Wizard", csLierOracle, false)
	looting := seedGraveyardCard(t, g, "Faithless Looting", "Sorcery", faithlessLootingOracle)
	setManaCost(g, me, looting, "{R}")

	got := flashbackOffers(t, g, me.ID, looting)
	if len(got) != 2 || got["flashback"] != "{2}{R}" || got[grantedFlashback] != "{R}" {
		t.Fatalf("offers = %v, want the printed flashback {2}{R} and Lier's {R}", got)
	}

	// The bot enumerator offers both, from the same list, when the mana
	// for both is there.
	addRed(g, me, 3)
	claims := map[string]bool{}
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Kind == legal.KindCast && m.Source == looting {
			for _, k := range []string{"flashback", grantedFlashback} {
				if strings.Contains(string(m.Params), `"alternative_cost":"`+k+`"`) {
					claims[k] = true
				}
			}
		}
	}
	if !claims["flashback"] || !claims[grantedFlashback] {
		t.Errorf("the enumerator offered %v, want both flashbacks", claims)
	}

	if err := g.CastSpell(me.ID, looting, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: grantedFlashback}); err != nil {
		t.Fatalf("Lier's flashback for {R}: %v", err)
	}
	var item *game.StackItem
	g.ReadSnapshot(func() { item = g.StackMeta[looting] })
	if item == nil || item.AltCost != "flashback" {
		t.Fatalf("stack item = %+v, want a flashback spell", item)
	}
	if n := me.ManaPool.Total(); n != 2 {
		t.Errorf("Lier's flashback left %d of three red, want 2: it costs {R}", n)
	}
	passPriorityAroundTable(t, g)
	discardFromHand(t, g, me.ID)
	if !g.Exile.Contains(looting) {
		t.Error("the card cast with Lier's flashback is not in exile (CR 702.34a)")
	}

	// Without Lier the card has its printed flashback alone.
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(lier) })
	again := seedGraveyardCard(t, g, "Faithless Looting", "Sorcery", faithlessLootingOracle)
	setManaCost(g, me, again, "{R}")
	if got := flashbackOffers(t, g, me.ID, again); len(got) != 1 || got["flashback"] != "{2}{R}" {
		t.Errorf("offers without Lier = %v, want the printed flashback alone", got)
	}
}

// Past in Flames under a second Past in Flames: its printed {4}{R} or
// its {3}{R} mana cost. The declared Full holds only because both are
// offered.
func TestPastInFlamesUnderASecondOneOffersItsManaCost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	first := seedGraveyardCard(t, g, "Past in Flames", "Sorcery", pastInFlamesOracle)
	setManaCost(g, me, first, "{3}{R}")

	if got := flashbackOffers(t, g, me.ID, first); len(got) != 1 || got["flashback"] != "{4}{R}" {
		t.Fatalf("offers before the second one = %v, want its printed flashback alone", got)
	}

	castCatalogSpell(t, g, "Past in Flames", "Sorcery", pastInFlamesOracle, nil)
	passPriorityAroundTable(t, g)

	got := flashbackOffers(t, g, me.ID, first)
	if len(got) != 2 || got["flashback"] != "{4}{R}" || got[grantedFlashback] != "{3}{R}" {
		t.Fatalf("offers = %v, want the printed flashback {4}{R} and the granted {3}{R}", got)
	}

	addRed(g, me, 4)
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: grantedFlashback}); err != nil {
		t.Fatalf("the granted flashback for {3}{R}: %v", err)
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("the granted flashback left %d mana: it should cost {3}{R}", n)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(first) {
		t.Error("Past in Flames cast with flashback is not in exile (CR 702.34a)")
	}
}
