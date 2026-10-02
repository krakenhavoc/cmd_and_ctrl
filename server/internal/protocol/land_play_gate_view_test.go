package protocol

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// land_play_gate_view_test.go — ADR 0109 §4 (#1895): the view reads the
// engine's land-play gate. A land the gate refuses carries the refusing
// clause on its `cant_cast`, is never `castable_here` out of a graveyard,
// and the seat carries `cant_play_lands`; and in each case the bot
// enumerator and CastSpell give the same answer.

const oracleLandBan1895 = "test-1895-land-ban"

// landBanOnBoard puts a permanent under p whose static forbids a land
// named `only` (every land when it is empty), from the hand or a graveyard.
func landBanOnBoard(t *testing.T, g *game.Game, p *game.Player, label, only string) {
	t.Helper()
	prev := game.CatalogLandPlayRestrictions
	game.CatalogLandPlayRestrictions = func(id string) []game.LandPlayRestriction {
		if id == oracleLandBan1895 {
			return []game.LandPlayRestriction{{
				Label: label,
				Forbids: func(q game.LandPlayQuery) bool {
					return only == "" || q.Card.Name == only
				},
			}}
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogLandPlayRestrictions = prev })
	g.Battlefield.PushTop(publicCard(g, p.ID, "Test Dispute", "Enchantment", "{4}{R}{R}", oracleLandBan1895))
}

// ownHandCard is viewer's own copy of a card in their hand.
func ownHandCard(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID) *CardView {
	t.Helper()
	v := ViewOfGameFor(g, p.ID.String())
	for i := range v.Seats {
		if v.Seats[i].ID == p.ID.String() {
			if c := cardInZone(v.Seats[i].Hand, id); c != nil {
				return c
			}
		}
	}
	t.Fatalf("card %s is not in %s's hand view", id, p.Name)
	return nil
}

func seatView(t *testing.T, g *game.Game, viewer, seat *game.Player) PlayerView {
	t.Helper()
	v := ViewOfGameFor(g, viewer.ID.String())
	for _, s := range v.Seats {
		if s.ID == seat.ID.String() {
			return s
		}
	}
	t.Fatalf("no seat view for %s", seat.Name)
	return PlayerView{}
}

// A land in hand under "players can't play lands" says why, a land that
// the ban does not name does not, and the enumerator and the engine agree
// with the view for both.
func TestHandLandCarriesTheRefusingClause(t *testing.T) {
	g, me, _ := stripTable(t)
	banned := publicCard(g, me.ID, "Desert", "Land", "", "")
	free := publicCard(g, me.ID, "Forest", "Basic Land — Forest", "", "")
	me.Hand.PushTop(banned)
	me.Hand.PushTop(free)
	landBanOnBoard(t, g, me, "Players can't play lands with a name originally printed in the Arabian Nights expansion.", "Desert")

	if got := ownHandCard(t, g, me, banned.InstanceID).CantCast; !strings.Contains(got, "Arabian Nights") || !strings.Contains(got, "Test Dispute") {
		t.Errorf("the refused land's cant_cast = %q, want the clause and the card", got)
	}
	if got := ownHandCard(t, g, me, free.InstanceID).CantCast; got != "" {
		t.Errorf("a land the ban does not name carries cant_cast = %q", got)
	}
	if landPlayOffered(g, me.ID, banned.InstanceID) {
		t.Error("the enumerator offers the land the view refuses")
	}
	if !landPlayOffered(g, me.ID, free.InstanceID) {
		t.Error("the enumerator refuses a land the view does not")
	}
	if err := g.CastSpell(me.ID, banned.InstanceID, game.CastSpellParams{}); !errors.Is(err, game.ErrCantPlayLand) {
		t.Errorf("CastSpell on the refused land: %v, want ErrCantPlayLand", err)
	}
}

// A graveyard land a Crucible would let you play is not castable_here under
// a ban, with the engine and the enumerator in step.
func TestGraveyardLandIsNotCastableUnderALandBan(t *testing.T) {
	g, me, _ := stripTable(t)
	crucibleOnBoard(t, g, me)
	land := publicCard(g, me.ID, "Dead Forest", "Basic Land — Forest", "", "")
	me.Graveyard.PushTop(land)
	assertLandBit(t, g, me.ID, ownGraveyardCard(t, g, me, land.InstanceID), land.InstanceID, true, "before the ban")

	landBanOnBoard(t, g, me, "Players can't play lands.", "")
	c := ownGraveyardCard(t, g, me, land.InstanceID)
	assertLandBit(t, g, me.ID, c, land.InstanceID, false, "under the ban")
	if !strings.Contains(c.CantCast, "Players can't play lands.") {
		t.Errorf("the graveyard land's cant_cast = %q", c.CantCast)
	}
	if err := g.CastSpell(me.ID, land.InstanceID, game.CastSpellParams{FromZone: "graveyard"}); !errors.Is(err, game.ErrCantPlayLand) {
		t.Errorf("playing the graveyard land under the ban: %v, want ErrCantPlayLand", err)
	}
}

// The seat's banner is the blanket ban only, and is public.
func TestSeatCarriesAnyBlanketLandBan(t *testing.T) {
	g, me, opp := stripTable(t)
	if got := seatView(t, g, opp, me).CantPlayLands; got != "" {
		t.Fatalf("a seat with no ban carries %q", got)
	}
	g.WithWriteLock(func() { g.CantPlayLandsThisTurnForEffect(uuid.Nil, me.ID, "Turf Wound") })
	got := seatView(t, g, opp, me).CantPlayLands
	if !strings.Contains(got, "You can't play lands this turn") {
		t.Errorf("the banned seat's banner, as an opponent sees it = %q", got)
	}
	if other := seatView(t, g, me, opp).CantPlayLands; other != "" {
		t.Errorf("the other seat carries %q", other)
	}
}
