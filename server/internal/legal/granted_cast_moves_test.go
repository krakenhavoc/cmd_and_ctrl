package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// granted_cast_moves_test.go — S42 / ADR 0066, the enumerator half of
// #673 for GRANTED permissions.
//
// The contract is the one every other enumerator test has: the
// enumerator and the engine read the same function, so every move
// offered is a move CastSpell accepts. dispatchAll is what proves it —
// it applies each move against the live game and fails on a rejection.

// graveyardCard drops a card into a seat's graveyard.
func graveyardCard(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Graveyard.PushTop(c)
	return c.InstanceID
}

// A card a permission opens in the graveyard is enumerated, at the
// price the permission names — which is the shape Snapcaster Mage and
// Underworld Breach both produce.
func TestEnumeratorOffersAGrantedGraveyardCast(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	mana(g, seat, 4)

	granted := graveyardCard(seat, game.Card{
		Name: "Granted Sorcery", TypeLine: "Sorcery", ManaCost: "{1}{U}", Layout: "normal",
	})
	plain := graveyardCard(seat, game.Card{
		Name: "Plain Sorcery", TypeLine: "Sorcery", ManaCost: "{1}{U}", Layout: "normal",
	})
	g.WithWriteLock(func() {
		g.GrantCastPermissionOverCardForEffect(granted, game.CastPermission{
			Player: seat.ID, Zone: game.ZoneGraveyard,
			AltCostKey: "flashback", ExileOnResolution: true,
		})
	})

	moves := legal.EnumerateFor(g, seat.ID)
	got := castMovesFor(moves, granted)
	if len(got) == 0 {
		t.Fatalf("no granted graveyard cast offered: %v", labels(moves))
	}
	if n := len(castMovesFor(moves, plain)); n != 0 {
		t.Errorf("a graveyard card no permission opens was offered %d casts", n)
	}
	// Every move offered is one the engine accepts.
	dispatchAll(t, g, seat.ID, moves)
}

// CR 401.5 for the bot: the top card of the library is enumerated when
// a permission opens it, the card below it never is, and a land off
// the top is a land PLAY rather than a cast.
func TestEnumeratorOffersALibraryTopPlay(t *testing.T) {
	const oracle = "test-enum-oracle"
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)

	prevPerms := game.CatalogCastPermissions
	game.CatalogCastPermissions = func(id string) []game.CastPermission {
		if id != oracle {
			return nil
		}
		return []game.CastPermission{{
			Zone: game.ZoneLibrary, Scope: game.ScopeStanding, Duration: game.WhileInZoneDuration(),
			TopOfLibraryOnly: true, Filter: game.PermissionFilter{LandsOnly: true},
		}}
	}
	t.Cleanup(func() { game.CatalogCastPermissions = prevPerms })

	prevVis := game.CatalogLibraryTopVisible
	game.CatalogLibraryTopVisible = func(id string) game.LibraryTopVisibility {
		if id == oracle {
			return game.LibraryTopRevealed
		}
		return game.LibraryTopHidden
	}
	t.Cleanup(func() { game.CatalogLibraryTopVisible = prevVis })

	source := game.NewCard("Test Oracle", seat.ID)
	source.TypeLine = "Creature — Elf"
	source.OracleID = oracle
	source.Controller = seat.ID
	g.Battlefield.PushTop(source)

	seat.Library.Cards = nil
	buried := game.NewCard("Buried Forest", seat.ID)
	buried.TypeLine = "Basic Land — Forest"
	seat.Library.PushTop(buried)
	top := game.NewCard("Top Forest", seat.ID)
	top.TypeLine = "Basic Land — Forest"
	seat.Library.PushTop(top)

	moves := legal.EnumerateFor(g, seat.ID)
	var offeredTop, offeredBuried int
	for _, m := range moves {
		switch m.Source {
		case top.InstanceID:
			offeredTop++
			if m.Kind != legal.KindLand {
				t.Errorf("a land off the top was offered as %v, want a land play", m.Kind)
			}
		case buried.InstanceID:
			offeredBuried++
		}
	}
	if offeredTop != 1 {
		t.Fatalf("library-top plays offered = %d, want 1: %v", offeredTop, labels(moves))
	}
	if offeredBuried != 0 {
		t.Errorf("the card below the top was offered %d moves (CR 401.5)", offeredBuried)
	}
	dispatchAll(t, g, seat.ID, moves)
}

// TestEnumeratorOffersAGatedLibraryTopPlayOnlyWhenBothHalvesHold is
// #1314's enumerator half: a standing permission gated by BOTH an ADR
// 0071 designation and a card Condition is offered only once both are
// satisfied, and the engine accepts the move the enumerator offers —
// the same #544 agreement TestEnumeratorOffersALibraryTopPlay pins
// for the ungated shape.
func TestEnumeratorOffersAGatedLibraryTopPlayOnlyWhenBothHalvesHold(t *testing.T) {
	const oracle = "test-enum-gated-oracle"
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	mana(g, seat, 4)

	prevGated := game.CatalogGatedCastPermissions
	game.CatalogGatedCastPermissions = func(id string) []game.CastPermissionGate {
		if id != oracle {
			return nil
		}
		return []game.CastPermissionGate{{
			Permission: game.CastPermission{Zone: game.ZoneLibrary, TopOfLibraryOnly: true},
			ActiveWhen: game.ClassLevel(2),
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return g.CastTallyFor(controller).Total > 0
			},
		}}
	}
	t.Cleanup(func() { game.CatalogGatedCastPermissions = prevGated })

	prevVis := game.CatalogLibraryTopVisible
	game.CatalogLibraryTopVisible = func(id string) game.LibraryTopVisibility {
		if id == oracle {
			return game.LibraryTopOwner
		}
		return game.LibraryTopHidden
	}
	t.Cleanup(func() { game.CatalogLibraryTopVisible = prevVis })

	source := game.NewCard("Test Talent", seat.ID)
	source.TypeLine = "Enchantment — Class"
	source.OracleID = oracle
	source.Controller = seat.ID
	g.Battlefield.PushTop(source)

	seat.Library.Cards = nil
	top := game.NewCard("Top Instant", seat.ID)
	top.TypeLine, top.ManaCost = "Instant", "{U}"
	seat.Library.PushTop(top)

	if n := len(castMovesFor(legal.EnumerateFor(g, seat.ID), top.InstanceID)); n != 0 {
		t.Errorf("level 1: got %d moves for the library-top card, want 0", n)
	}

	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == source.InstanceID {
				g.Battlefield.Cards[i].ClassLevel = 2
			}
		}
	})
	if n := len(castMovesFor(legal.EnumerateFor(g, seat.ID), top.InstanceID)); n != 0 {
		t.Errorf("level 2 with nothing cast: got %d moves, want 0", n)
	}

	g.WithWriteLock(func() {
		if g.SpellsCastThisTurn == nil {
			g.SpellsCastThisTurn = make(map[uuid.UUID]game.CastTally)
		}
		g.SpellsCastThisTurn[seat.ID] = game.CastTally{Total: 1}
	})
	moves := castMovesFor(legal.EnumerateFor(g, seat.ID), top.InstanceID)
	if len(moves) == 0 {
		t.Fatal("level 2 after casting a spell this turn: the library-top card was not offered")
	}
	dispatchAll(t, g, seat.ID, moves[:1])
}

// The negative that keeps the gate honest: with nothing granting
// anything, the three extra zones produce no moves at all — which is
// also what the AnyCastPermissionsForEffect fast path relies on.
func TestEnumeratorOffersNothingFromTheseZonesWithoutAPermission(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	mana(g, seat, 4)

	inYard := graveyardCard(seat, game.Card{
		Name: "Dead Sorcery", TypeLine: "Sorcery", ManaCost: "{1}{U}", Layout: "normal",
	})
	moves := legal.EnumerateFor(g, seat.ID)
	if n := len(castMovesFor(moves, inYard)); n != 0 {
		t.Errorf("an ungranted graveyard card was offered %d casts", n)
	}
	for _, m := range moves {
		if m.Source == seat.Library.Cards[len(seat.Library.Cards)-1].InstanceID {
			t.Errorf("the top of an unopened library was offered a move: %q", m.Label)
		}
	}
}
