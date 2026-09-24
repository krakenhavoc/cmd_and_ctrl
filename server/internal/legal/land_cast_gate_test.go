package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// land_cast_gate_test.go — #1439: playing a land is a special action
// (CR 305.1, CR 116.2a), not a cast, so a "can't cast" restriction
// must not remove a land-play move from the enumeration, and the
// engine must actually accept the move the enumerator offers.
//
// legal.landPlayMove has never asked CastGateLocked, so the
// enumerator already agreed with the fixed engine; what these pin is
// the OTHER half of the #499/#618 agreement — that CastSpell, given
// the very move enumerated here, accepts it.
const oracleGrafdiggersCage1439 = "753cb2b2-24ce-484f-a2d0-be6fd2c67ebd"

// TestLandPlayIsStillEnumeratedUnderRuleOfLaw is Rule of Law's
// enumerator half. Without the fix in CastSpell, dispatching the
// offered move would fail with ErrCantCast even though the move
// itself was correctly offered.
func TestLandPlayIsStillEnumeratedUnderRuleOfLaw(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	forest := handCard(active, basic("Forest", "Forest"))
	battlefieldCard(g, active, game.Card{
		Name: "Rule of Law", TypeLine: "Enchantment", OracleID: oracleRuleOfLaw,
	})
	advanceTo(t, g, game.StepPrecombatMain)

	// One spell already cast this turn, which is what Rule of Law
	// reads (game.CastTallyFor).
	g.WithWriteLock(func() {
		if g.SpellsCastThisTurn == nil {
			g.SpellsCastThisTurn = make(map[uuid.UUID]game.CastTally)
		}
		g.SpellsCastThisTurn[active.ID] = game.CastTally{Total: 1}
	})

	moves := legal.EnumerateFor(g, active.ID)
	var landMove *legal.Move
	for i := range moves {
		if moves[i].Kind == legal.KindLand && moves[i].Source == forest {
			landMove = &moves[i]
		}
	}
	if landMove == nil {
		t.Fatalf("the land was not offered under Rule of Law: %v", labels(moves))
	}
	// #544/#1439: the offer is one the engine actually accepts.
	dispatchAll(t, g, active.ID, []legal.Move{*landMove})
}

// TestPermittedGraveyardLandPlayIsNotBlockedByGrafdiggersCage is the
// ZONE-restriction half: Grafdigger's Cage stops CASTING from a
// graveyard or library, but a land coming out of either is PLAYED,
// not cast, so a Crucible-of-Worlds-shaped permission still works
// with the Cage on the battlefield.
func TestPermittedGraveyardLandPlayIsNotBlockedByGrafdiggersCage(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)

	const crucibleOracle = "test-1439-crucible"
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
		Name: "Grafdigger's Cage", TypeLine: "Artifact", OracleID: oracleGrafdiggersCage1439,
	})

	forest := game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}
	forest.InstanceID = uuid.New()
	forest.Owner = active.ID
	forest.Controller = active.ID
	forest.KnownBy = map[uuid.UUID]bool{active.ID: true}
	active.Graveyard.PushTop(forest)

	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	var landMove *legal.Move
	for i := range moves {
		if moves[i].Kind == legal.KindLand && moves[i].Source == forest.InstanceID {
			landMove = &moves[i]
		}
	}
	if landMove == nil {
		t.Fatalf("the Crucible-permitted graveyard land was not offered under Grafdigger's Cage: %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, []legal.Move{*landMove})
}
