package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// library_card_grant_view_test.go — Sunbird's Invocation's free cast is
// a ScopeCards grant over ONE card in its owner's library: the card it
// revealed and left on top. No permanent makes the library top visible,
// so what puts the card on the owner's wire is the reveal's knowledge,
// and what makes it a cast surface is the grant. Pinned here: the owner
// sees it castable, and a known card beneath it is not stamped.
func TestALibraryCardGrantStampsTheRevealedTopForItsOwner(t *testing.T) {
	g := buildActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	const oracle = "test-view-library-card-grant"
	withAltCostsByOracle(t, map[string][]game.AlternativeCost{oracle: nil})
	withTopVisibility(t, "unused", game.LibraryTopHidden)
	seen := map[uuid.UUID]bool{me.ID: true, them.ID: true}

	me.Library.Cards = nil
	below := game.NewCard("Known Below", me.ID)
	below.TypeLine = "Instant"
	below.ManaCost = "{U}"
	below.OracleID = oracle
	below.KnownBy = seen
	me.Library.PushTop(below)
	top := game.NewCard("Revealed Pick", me.ID)
	top.TypeLine = "Instant"
	top.ManaCost = "{1}{U}"
	top.OracleID = oracle
	top.KnownBy = seen
	me.Library.PushTop(top)

	limit := 3
	g.WithWriteLock(func() {
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:            me.ID,
			Zone:              game.ZoneLibrary,
			Duration:          g.UntilEndOfTurnDuration(),
			Cost:              "{0}",
			Timing:            game.TimingFlash,
			CastOnly:          true,
			MaxSpellManaValue: &limit,
			LapseOnPass:       game.LapseToLibraryBottom,
			Label:             "Cast it without paying its mana cost",
		}, []game.Card{top})
	})

	lib := ViewOfGameFor(g, me.ID.String()).Seats[0].Library
	card := cardInZone(lib, top.InstanceID)
	if card == nil || card.Name == "" {
		t.Fatalf("the revealed pick is not readable in its owner's library: %+v", card)
	}
	if !card.CastableHere {
		t.Errorf("castable_here is clear on the card the grant names")
	}
	if under := cardInZone(lib, below.InstanceID); under != nil && under.CastableHere {
		t.Errorf("a card the grant does not name was stamped castable: %+v", under)
	}
}
