package protocol

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// foreign_library_grant_view_test.go — #1035. The view half of "a cast
// permission over another seat's LIBRARY top opens nothing". The
// engine's half is server/internal/game/foreign_library_cast_test.go;
// what is pinned here is the wire: the holder gets the announce
// surface on somebody else's library top, the library's owner and the
// bystanders get the card and nothing else, and a spectator gets no
// library at all.
//
// Xanathar, Guild Kingpin's clause is the fixture — "you may look at
// the top card of their library any time, you may play the top card of
// their library, and you may spend mana as though it were mana of any
// color" — because it is the one printed shape that names another
// seat's library, and its look clause is the only thing that makes the
// card visible to the holder at all.

// libraryTopCard seeds one card on top of a seat's library and returns
// it. No KnownBy: who may read a library card is decided by CR 401.5's
// position rule per frame (stampLibraryTop), never by a stamp on the
// card, and a fixture that set one would be testing the wrong rule.
func libraryTopCard(p *game.Player, name, oracle string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine = "Instant"
	c.ManaCost = "{1}{U}"
	c.OracleID = oracle
	p.Library.PushTop(c)
	return c.InstanceID
}

// xanatharOver gives `holder` Xanathar's clause over `victim`'s
// library: a stored ScopeStanding permission that names the seat.
func xanatharOver(t *testing.T, g *game.Game, holder, victim *game.Player) {
	t.Helper()
	g.WithWriteLock(func() {
		g.GrantCastPermissionForEffect(game.CastPermission{
			Player:           holder.ID,
			Zone:             game.ZoneLibrary,
			Scope:            game.ScopeStanding,
			ZoneOwner:        victim.ID,
			TopOfLibraryOnly: true,
			SeesLibraryTop:   true,
			AnyColor:         true,
			Label:            "Play the top card of their library",
		})
	})
}

// The holder gets the card and the whole announce surface, computed
// for THEM — out of a library that is otherwise wholesale hidden.
func TestAForeignLibraryGrantReachesItsHolder(t *testing.T) {
	g := buildActiveGame(t)
	victim, holder := g.Seats[0], g.Seats[1]
	const oracle = "test-view-foreign-library"
	withAltCostsByOracle(t, map[string][]game.AlternativeCost{oracle: nil})
	victim.Library.Cards = nil
	libraryTopCard(victim, "Buried", oracle)
	top := libraryTopCard(victim, "Their Top", oracle)
	xanatharOver(t, g, holder, victim)

	lib := ViewOfGameFor(g, holder.ID.String()).Seats[0].Library
	card := cardInZone(lib, top)
	if card == nil || card.Name == "" {
		t.Fatalf("the grant's look clause did not put the top card on the holder's wire: %+v", card)
	}
	if !card.CastableHere {
		t.Errorf("castable_here is clear for the permission's holder")
	}
	if card.ExilePlay == nil || card.ExilePlay.Player != holder.ID.String() {
		t.Errorf("exile_play = %+v, want the holder's own grant", card.ExilePlay)
	}
	// CR 401.5 opens one POSITION: the card under the top is not on
	// the wire at all, look clause or no look clause.
	if len(lib.Cards) != 1 {
		t.Errorf("the holder got %d library cards, want just the top", len(lib.Cards))
	}
	if lib.Count != 2 {
		t.Errorf("library count = %d, want the public 2", lib.Count)
	}
}

// The library's OWNER holds no permission and — under a look rather
// than a reveal — cannot even see the card. A bystander gets the same
// nothing, and a spectator gets the library it always got.
func TestTheLibraryOwnerAndBystandersGetNoForeignLibraryStamps(t *testing.T) {
	g := busyTable(t, 0)
	victim, holder, bystander := g.Seats[0], g.Seats[1], g.Seats[2]
	const oracle = "test-view-foreign-library-others"
	const revealer = "test-view-foreign-library-revealer"
	withAltCostsByOracle(t, map[string][]game.AlternativeCost{oracle: nil})
	// A "play with the top card of your library revealed" permanent
	// the VICTIM controls, so everybody can read the card and the
	// assertions below are about the STAMPS rather than about the
	// redaction that would otherwise hide them.
	withTopVisibility(t, revealer, game.LibraryTopRevealed)
	source := game.NewCard("Test Revealer", victim.ID)
	source.TypeLine = "Enchantment"
	source.OracleID = revealer
	source.Controller = victim.ID
	g.Battlefield.PushTop(source)

	victim.Library.Cards = nil
	top := libraryTopCard(victim, "Their Top", oracle)
	xanatharOver(t, g, holder, victim)

	for _, tc := range []struct {
		name string
		seat uuid.UUID
	}{
		{"the library's owner", victim.ID},
		{"a bystander", bystander.ID},
	} {
		card := cardInZone(ViewOfGameFor(g, tc.seat.String()).Seats[0].Library, top)
		if card == nil || card.Name == "" {
			t.Fatalf("%s lost a revealed top card: %+v", tc.name, card)
		}
		if card.CastableHere {
			t.Errorf("%s is told the top card is a cast surface", tc.name)
		}
		if len(card.AlternativeCosts) != 0 {
			t.Errorf("%s offers = %v, want none", tc.name, keysOf(card.AlternativeCosts))
		}
	}

	// And an unseated viewer gets the library wholesale-hidden, which
	// is the rule that has always covered a spectator.
	spectator := FilterViewFor(ViewOfGame(g), "").Seats[0].Library
	if len(spectator.Cards) != 0 {
		t.Errorf("a spectator was handed %d library cards", len(spectator.Cards))
	}
}

// The holder's grant is the only key, and the position rule still
// holds on the wire: no grant, no stamps, however visible the card.
func TestARevealedLibraryTopIsNotACastSurfaceWithoutAGrant(t *testing.T) {
	g := buildActiveGame(t)
	victim, other := g.Seats[0], g.Seats[1]
	const revealer = "test-view-library-revealer-only"
	withTopVisibility(t, revealer, game.LibraryTopRevealed)
	source := game.NewCard("Test Revealer", victim.ID)
	source.TypeLine = "Enchantment"
	source.OracleID = revealer
	source.Controller = victim.ID
	g.Battlefield.PushTop(source)
	victim.Library.Cards = nil
	top := libraryTopCard(victim, "Their Top", "")

	card := cardInZone(ViewOfGameFor(g, other.ID.String()).Seats[0].Library, top)
	if card == nil || card.Name == "" {
		t.Fatalf("the revealed top card is missing from an opponent's view")
	}
	if card.CastableHere || card.ExilePlay != nil {
		t.Errorf("a revealed top card nobody may play was marked playable: castable=%v grant=%+v",
			card.CastableHere, card.ExilePlay)
	}
}

// A standing "your library" permission does not reach an opponent's,
// on the wire or in the move list. The engine's copy of this is
// TestAStandingLibraryPermissionStopsAtItsHoldersOwnLibrary; this one
// is here because it is the surface a Courser controller would see
// across the table if the scoping came off.
//
// The coherence property for a foreign library top rides #1024's
// TestViewAndEnumeratorOfferTheSamePrices, which seeds one.
func TestAStandingLibraryGrantStopsAtItsHoldersOwnLibraryOnTheWire(t *testing.T) {
	g := buildActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	const oracle = "test-view-standing-courser"
	withStandingPermission(t, oracle, game.CastPermission{
		Zone:             game.ZoneLibrary,
		Scope:            game.ScopeStanding,
		Duration:         game.WhileInZoneDuration(),
		TopOfLibraryOnly: true,
		Label:            "Play the top card of your library",
	})
	// Both seats' top cards are REVEALED, so the cross-seat read is
	// refused by the ownership rule rather than by the redaction: me
	// has the Courser (reveal plus permission) and them has a bare
	// revealer (reveal and no permission at all), which is what makes
	// the assertion below about the SCOPE of me's standing grant.
	const revealer = "test-view-standing-revealer"
	prev := game.CatalogLibraryTopVisible
	game.CatalogLibraryTopVisible = func(id string) game.LibraryTopVisibility {
		if id == oracle || id == revealer {
			return game.LibraryTopRevealed
		}
		return game.LibraryTopHidden
	}
	t.Cleanup(func() { game.CatalogLibraryTopVisible = prev })
	for _, tc := range []struct {
		seat   *game.Player
		oracle string
	}{{me, oracle}, {them, revealer}} {
		source := game.NewCard("Test Courser", tc.seat.ID)
		source.TypeLine = "Creature — Centaur"
		source.OracleID = tc.oracle
		source.Controller = tc.seat.ID
		g.Battlefield.PushTop(source)
		tc.seat.Library.Cards = nil
	}
	mine := libraryTopCard(me, "My Top", "")
	theirs := libraryTopCard(them, "Their Top", "")

	v := ViewOfGameFor(g, me.ID.String())
	if own := cardInZone(v.Seats[0].Library, mine); own == nil || !own.CastableHere {
		t.Errorf("the Courser controller's OWN library top is not a cast surface: %+v", own)
	}
	if other := cardInZone(v.Seats[1].Library, theirs); other == nil || other.CastableHere {
		t.Errorf("a standing \"your library\" permission reached an opponent's library: %+v", other)
	}
	// And the engine agrees, which is where it matters.
	for _, m := range legal.EnumerateLocked(g, me.ID, legal.Options{}) {
		if m.Type == legal.TypeCastSpell && strings.Contains(string(m.Params), theirs.String()) {
			t.Errorf("the enumerator offered a cast off an opponent's library under a standing grant: %s", m.Params)
		}
	}
}
