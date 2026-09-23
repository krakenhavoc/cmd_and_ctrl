package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// granted_permissions_view_test.go — S42 / ADR 0066. The client's cast
// buttons and the CR 401.5 top card both come off this projection, so
// the facts to pin are the ones that turn into a wrong button or a
// leak:
//
//   - `castable_here` marks a graveyard card a PERMISSION opens, not
//     just one whose own text does, and it carries the price that
//     permission names.
//   - the top of a library reaches its owner under "you may look", the
//     whole table under "play with the top card revealed", and nobody
//     otherwise — and only ever the top card.

// withStandingPermission stubs the standing-permission hook for one
// oracle ID.
func withStandingPermission(t *testing.T, oracle string, perm game.CastPermission) {
	t.Helper()
	prev := game.CatalogCastPermissions
	game.CatalogCastPermissions = func(id string) []game.CastPermission {
		if id == oracle {
			return []game.CastPermission{perm}
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogCastPermissions = prev })
}

// withGatedStandingPermission stubs the GATED standing-permission
// hook (#1314) for one oracle ID.
func withGatedStandingPermission(t *testing.T, oracle string, gate game.CastPermissionGate) {
	t.Helper()
	prev := game.CatalogGatedCastPermissions
	game.CatalogGatedCastPermissions = func(id string) []game.CastPermissionGate {
		if id == oracle {
			return []game.CastPermissionGate{gate}
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogGatedCastPermissions = prev })
}

// withTopVisibility stubs the CR 401.5 visibility hook.
func withTopVisibility(t *testing.T, oracle string, v game.LibraryTopVisibility) {
	t.Helper()
	prev := game.CatalogLibraryTopVisible
	game.CatalogLibraryTopVisible = func(id string) game.LibraryTopVisibility {
		if id == oracle {
			return v
		}
		return game.LibraryTopHidden
	}
	t.Cleanup(func() { game.CatalogLibraryTopVisible = prev })
}

// cardInZone finds a projected card by instance ID.
func cardInZone(z ZoneView, id uuid.UUID) *CardView {
	for i := range z.Cards {
		if z.Cards[i].InstanceID == id.String() {
			return &z.Cards[i]
		}
	}
	return nil
}

// A card a standing permission opens is a cast surface, and the offer
// it is shown is the one the cast path accepts.
func TestCastableHereStampedOnAGrantedGraveyardCard(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	const oracle = "test-view-breach"
	withStandingPermission(t, oracle, game.CastPermission{
		Zone:                    game.ZoneGraveyard,
		Scope:                   game.ScopeStanding,
		Duration:                game.WhileInZoneDuration(),
		Filter:                  game.PermissionFilter{NonLandOnly: true},
		AltCostKey:              "escape",
		ExileOtherFromGraveyard: 3,
		Label:                   "Escape — its mana cost, exile three",
	})

	seen := map[uuid.UUID]bool{me.ID: true, g.Seats[1].ID: true}
	breach := game.NewCard("Test Breach", me.ID)
	breach.TypeLine = "Enchantment"
	breach.OracleID = oracle
	breach.Controller = me.ID
	breach.KnownBy = seen
	g.Battlefield.PushTop(breach)

	spell := game.NewCard("Dead Spell", me.ID)
	spell.TypeLine = "Instant"
	spell.ManaCost = "{1}{U}"
	spell.KnownBy = seen
	me.Graveyard.PushTop(spell)
	// #695: the offer is shown only when every component of it is
	// payable, and escape's is "exile three OTHER cards from your
	// graveyard". Three fuel cards is the board the cast path would
	// accept; the sibling test below holds the case with two.
	for _, name := range []string{"Fuel A", "Fuel B", "Fuel C"} {
		fuel := game.NewCard(name, me.ID)
		fuel.TypeLine = "Instant"
		fuel.KnownBy = seen
		me.Graveyard.PushTop(fuel)
	}

	v := ViewOfGameFor(g, me.ID.String())
	got := cardInZone(v.Seats[0].Graveyard, spell.InstanceID)
	if got == nil {
		t.Fatalf("the graveyard card is missing from the view")
	}
	if !got.CastableHere {
		t.Errorf("castable_here not stamped on a card a permission opens")
	}
	if got.AlternativeCosts == nil || len(got.AlternativeCosts) != 1 {
		t.Fatalf("alternative costs = %+v, want the one granted offer", got.AlternativeCosts)
	}
	if got.AlternativeCosts[0].Key != "escape" {
		t.Errorf("offered key = %q, want escape", got.AlternativeCosts[0].Key)
	}
	// The stamp is the VIEWER's own answer since #1055 — this view is
	// built for `me`, who holds the Breach, so the bit above is theirs
	// and not a public statement about the pile's owner. What must NOT
	// be stamped is a card no permission reaches: the opponent's own
	// graveyard, where this Breach grants nothing.
	them := g.Seats[1]
	theirSpell := game.NewCard("Their Dead Spell", them.ID)
	theirSpell.TypeLine = "Instant"
	theirSpell.ManaCost = "{1}{U}"
	theirSpell.KnownBy = seen
	them.Graveyard.PushTop(theirSpell)

	v = ViewOfGameFor(g, me.ID.String())
	if other := cardInZone(v.Seats[1].Graveyard, theirSpell.InstanceID); other == nil || other.CastableHere {
		t.Errorf("a card in a graveyard no permission reaches was stamped castable: %+v", other)
	}
}

// TestCastableHereWaitsOnAGatedPermission is #1314's view half: the
// same `castable_here` stamp, for a permission that is not there yet
// because its designation gate has not been satisfied. The fast
// negative (AnyCastPermissionsForEffect) has to see the GATED hook too
// — before #1314 it only walked CatalogCastPermissions, and a card
// gated entirely behind CatalogGatedCastPermissions would have made
// the whole library walk skip itself, whether or not the gate was
// open.
func TestCastableHereWaitsOnAGatedPermission(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	const oracle = "test-view-talent"
	withGatedStandingPermission(t, oracle, game.CastPermissionGate{
		Permission: game.CastPermission{Zone: game.ZoneLibrary, TopOfLibraryOnly: true},
		ActiveWhen: game.ClassLevel(2),
		Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
			return g.CastTallyFor(controller).Total > 0
		},
	})
	withTopVisibility(t, oracle, game.LibraryTopOwner)

	seen := map[uuid.UUID]bool{me.ID: true}
	talent := game.NewCard("Test Talent", me.ID)
	talent.TypeLine = "Enchantment — Class"
	talent.OracleID = oracle
	talent.Controller = me.ID
	talent.KnownBy = seen
	g.Battlefield.PushTop(talent)

	top := game.NewCard("Top Instant", me.ID)
	top.TypeLine = "Instant"
	top.ManaCost = "{U}"
	top.KnownBy = seen
	me.Library.PushTop(top)

	// Level 1, nothing cast: the gate is shut, and the card is not a
	// cast surface at all.
	v := ViewOfGameFor(g, me.ID.String())
	got := cardInZone(v.Seats[0].Library, top.InstanceID)
	if got == nil {
		t.Fatalf("the library-top card is missing from the view")
	}
	if got.CastableHere {
		t.Error("castable_here stamped at level 1 with nothing cast — the gate is shut")
	}

	// Level 2, a spell cast this turn: both halves open, and the card
	// is a cast surface with the permission's own price.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == talent.InstanceID {
				g.Battlefield.Cards[i].ClassLevel = 2
			}
		}
		if g.SpellsCastThisTurn == nil {
			g.SpellsCastThisTurn = make(map[uuid.UUID]game.CastTally)
		}
		g.SpellsCastThisTurn[me.ID] = game.CastTally{Total: 1}
	})
	v = ViewOfGameFor(g, me.ID.String())
	got = cardInZone(v.Seats[0].Library, top.InstanceID)
	if got == nil || !got.CastableHere {
		t.Error("castable_here not stamped once both the level and the condition are satisfied")
	}
}

// CR 401.5's two strengths, and the redaction that separates them.
func TestLibraryTopVisibilityOnTheWire(t *testing.T) {
	const oracle = "test-view-top"
	for _, tc := range []struct {
		name      string
		vis       game.LibraryTopVisibility
		ownerSees bool
		tableSees bool
	}{
		{"hidden", game.LibraryTopHidden, false, false},
		{"you may look", game.LibraryTopOwner, true, false},
		{"played revealed", game.LibraryTopRevealed, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := buildActiveGame(t)
			me, them := g.Seats[0], g.Seats[1]
			withTopVisibility(t, oracle, tc.vis)

			source := game.NewCard("Test Oracle", me.ID)
			source.TypeLine = "Creature — Elf"
			source.OracleID = oracle
			source.Controller = me.ID
			g.Battlefield.PushTop(source)

			me.Library.Cards = nil
			buried := game.NewCard("Buried Card", me.ID)
			buried.TypeLine = "Instant"
			me.Library.PushTop(buried)
			top := game.NewCard("Top Card", me.ID)
			top.TypeLine = "Instant"
			me.Library.PushTop(top)

			mine := ViewOfGameFor(g, me.ID.String())
			own := cardInZone(mine.Seats[0].Library, top.InstanceID)
			if own == nil {
				t.Fatalf("the owner's own library card is missing from their view")
			}
			if gotName := own.Name != ""; gotName != tc.ownerSees {
				t.Errorf("owner sees the top card = %v, want %v", gotName, tc.ownerSees)
			}
			// Never the card below the top, whatever the rule says.
			if below := cardInZone(mine.Seats[0].Library, buried.InstanceID); below != nil && below.Name != "" {
				t.Errorf("a card below the top is readable: %q", below.Name)
			}

			theirs := ViewOfGameFor(g, them.ID.String())
			other := cardInZone(theirs.Seats[0].Library, top.InstanceID)
			if tc.tableSees {
				if other == nil || other.Name == "" {
					t.Errorf("the table cannot read a revealed top card: %+v", other)
				}
			} else if other != nil {
				t.Errorf("an opponent's library leaked a card: %+v", other)
			}
			// The count is public either way — it always was.
			if theirs.Seats[0].Library.Count != 2 {
				t.Errorf("opponent library count = %d, want 2", theirs.Seats[0].Library.Count)
			}
		})
	}
}

// A one-shot reveal makes cards KNOWN without making them visible in
// the zone. This is the distinction the projection keys on, and
// getting it wrong would turn every "reveal the top two cards" into a
// permanent window into the library.
func TestAOneShotRevealDoesNotOpenTheLibraryProjection(t *testing.T) {
	g := buildActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	withTopVisibility(t, "unused", game.LibraryTopHidden)

	me.Library.Cards = nil
	top := game.NewCard("Revealed Card", me.ID)
	top.TypeLine = "Instant"
	top.KnownBy = map[uuid.UUID]bool{me.ID: true, them.ID: true}
	me.Library.PushTop(top)

	theirs := ViewOfGameFor(g, them.ID.String())
	if got := cardInZone(theirs.Seats[0].Library, top.InstanceID); got != nil {
		t.Errorf("a once-revealed library card is on the wire in the zone: %+v", got)
	}
}
