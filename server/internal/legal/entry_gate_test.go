package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// entry_gate_test.go — #1985: a land that can't ENTER the battlefield
// can't be played either (CR 305.1), so the enumerator must not offer the
// play. Otherwise a bot would be handed a move the engine refuses, every
// priority, forever.

const oracleSoullessJailer1985 = "7866ab01-22f3-4d87-bdb0-c0ae0e91ffc9"

func landMoveFor(moves []legal.Move, id uuid.UUID) bool {
	for _, m := range moves {
		if m.Kind == legal.KindLand && m.Source == id {
			return true
		}
	}
	return false
}

// A hand land under "Lands can't enter the battlefield" is not offered,
// and is once the source has gone.
func TestLandThatCantEnterIsNotOffered(t *testing.T) {
	const wormsOracle = "test-1985-worms"
	prev := game.CatalogEntryRestrictions
	game.CatalogEntryRestrictions = func(id string) []game.EntryRestriction {
		if id == wormsOracle {
			return []game.EntryRestriction{{
				Label:   "Lands can't enter the battlefield.",
				Forbids: func(q game.EntryQuery) bool { return q.WouldBe().IsLand() },
			}}
		}
		if prev != nil {
			return prev(id)
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogEntryRestrictions = prev })

	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	forest := handCard(active, basic("Forest", "Forest"))
	worms := battlefieldCard(g, active, game.Card{Name: "Test Worms", TypeLine: "Enchantment", OracleID: wormsOracle})
	advanceTo(t, g, game.StepPrecombatMain)

	if moves := legal.EnumerateFor(g, active.ID); landMoveFor(moves, forest) {
		t.Fatalf("a land that can't enter was offered: %v", labels(moves))
	}
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(worms) })
	if moves := legal.EnumerateFor(g, active.ID); !landMoveFor(moves, forest) {
		t.Fatalf("the land was not offered once the prohibition left: %v", labels(moves))
	}
}

// Soulless Jailer: a Crucible-permitted graveyard land play is not
// offered (permanent cards in graveyards can't enter), where Grafdigger's
// Cage, which names only creature cards, lets it through
// (TestPermittedGraveyardLandPlayIsNotBlockedByGrafdiggersCage).
func TestGraveyardLandPlayIsNotOfferedUnderSoullessJailer(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)

	const crucibleOracle = "test-1985-crucible"
	prev := game.CatalogCastPermissions
	game.CatalogCastPermissions = func(id string) []game.CastPermission {
		if id == crucibleOracle {
			return []game.CastPermission{{
				Zone:     game.ZoneGraveyard,
				Scope:    game.ScopeStanding,
				Duration: game.WhileInZoneDuration(),
				Filter:   game.PermissionFilter{LandsOnly: true},
				Label:    "Test Crucible — play lands from your graveyard",
			}}
		}
		if prev != nil {
			return prev(id)
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogCastPermissions = prev })

	battlefieldCard(g, active, game.Card{Name: "Test Crucible", TypeLine: "Artifact", OracleID: crucibleOracle})
	battlefieldCard(g, active, game.Card{
		Name: "Soulless Jailer", TypeLine: "Artifact Creature — Phyrexian Golem", OracleID: oracleSoullessJailer1985,
	})
	forest := game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}
	forest.InstanceID = uuid.New()
	forest.Owner = active.ID
	forest.Controller = active.ID
	forest.KnownBy = map[uuid.UUID]bool{active.ID: true}
	active.Graveyard.PushTop(forest)
	advanceTo(t, g, game.StepPrecombatMain)

	if moves := legal.EnumerateFor(g, active.ID); landMoveFor(moves, forest.InstanceID) {
		t.Fatalf("a graveyard land play was offered under Soulless Jailer: %v", labels(moves))
	}
}
