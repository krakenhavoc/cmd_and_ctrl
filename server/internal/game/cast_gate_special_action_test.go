package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// cast_gate_special_action_test.go — where #760's gate meets #987's
// special actions (CR 116.2), and the two rules that decide the seam.
//
// CR 702.143c and CR 702.62c both say the same thing from opposite
// ends: a foretold or suspended card's LATER CAST is a cast like any
// other, and obeys "can't cast". The special ACTION that put it there
// is not a cast at all (CR 116.2), so nothing about a cast
// restriction touches it.
//
// Both halves fall out of where the gate sits rather than out of any
// rule written for them — `PerformSpecialAction` never calls
// `CastSpell`, and suspend's free cast is an ADR 0066 permission
// consumed BY `CastSpell` — which is exactly why they are worth
// pinning: a later refactor that moved either one would break a rule
// nothing else asserts.

const (
	gateRestrictionOracle = "test-gate-restriction"
	gateForetellOracle    = "test-gate-foretell"
)

// stubCastRestrictions wires CatalogCastRestrictions for one oracle
// ID and restores the previous hook at the end of the test.
func stubCastRestrictions(t *testing.T, oracleID string, rules []CastRestriction) {
	t.Helper()
	prev := CatalogCastRestrictions
	CatalogCastRestrictions = func(id string) []CastRestriction {
		if id == oracleID {
			return rules
		}
		if prev != nil {
			return prev(id)
		}
		return nil
	}
	t.Cleanup(func() { CatalogCastRestrictions = prev })
}

// seedTotalBan puts a permanent on the battlefield whose static
// forbids every cast by everyone. Blunter than any printed card on
// purpose: the question here is whether the gate is CONSULTED on a
// path, not what any particular clause says.
func seedTotalBan(t *testing.T, g *Game, controller uuid.UUID) {
	t.Helper()
	stubCastRestrictions(t, gateRestrictionOracle, []CastRestriction{{
		Label:   "Test Warden — players can't cast spells.",
		Forbids: func(CastQuery) bool { return true },
	}})
	g.Battlefield.PushTop(Card{
		InstanceID: uuid.New(),
		Name:       "Test Warden",
		TypeLine:   "Enchantment",
		OracleID:   gateRestrictionOracle,
		Owner:      controller,
		Controller: controller,
	})
}

// TestSuspendedFreeCastObeysTheCastGate is CR 702.62c: "a suspended
// card's cast obeys the rules for casting spells", which includes
// every "can't cast" effect. The free cast is an ADR 0066 permission
// and the permission is consumed by CastSpell, so the gate covers it
// with no rule of its own — this is the assertion that says so.
func TestSuspendedFreeCastObeysTheCastGate(t *testing.T) {
	g := newRestorableGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]

	id := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(Card{
			InstanceID: id,
			Name:       "Suspended Bolt",
			TypeLine:   "Sorcery",
			ManaCost:   "{2}{R}",
			Owner:      caster.ID,
			Controller: caster.ID,
		})
		// The permission grantSuspendedFreeCastLocked writes: free,
		// flash-timed, cast-only, hasty.
		g.GrantCastPermissionToCardsForEffect(CastPermission{
			Player:      caster.ID,
			Zone:        ZoneExile,
			Cost:        "{0}",
			Timing:      TimingFlash,
			CastOnly:    true,
			GrantsHaste: true,
			Duration:    g.UntilEndOfTurnDuration(),
			Label:       SuspendFreeCastLabel,
		}, []Card{g.Exile.Cards[len(g.Exile.Cards)-1]})
	})

	// Without a restriction the grant opens the cast — otherwise the
	// refusal below would prove nothing.
	if err := g.CastSpell(caster.ID, id, CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("the suspended free cast was refused with nothing restricting it: %v", err)
	}

	// Same setup, with the ban out.
	g2 := newRestorableGame(t)
	caster2 := g2.Seats[g2.Turn.ActiveSeat]
	id2 := uuid.New()
	g2.WithWriteLock(func() {
		seedTotalBan(t, g2, caster2.ID)
		g2.Exile.PushTop(Card{
			InstanceID: id2,
			Name:       "Suspended Bolt",
			TypeLine:   "Sorcery",
			ManaCost:   "{2}{R}",
			Owner:      caster2.ID,
			Controller: caster2.ID,
		})
		g2.GrantCastPermissionToCardsForEffect(CastPermission{
			Player:      caster2.ID,
			Zone:        ZoneExile,
			Cost:        "{0}",
			Timing:      TimingFlash,
			CastOnly:    true,
			GrantsHaste: true,
			Duration:    g2.UntilEndOfTurnDuration(),
			Label:       SuspendFreeCastLabel,
		}, []Card{g2.Exile.Cards[len(g2.Exile.Cards)-1]})
	})

	err := g2.CastSpell(caster2.ID, id2, CastSpellParams{FromZone: "exile"})
	if !errors.Is(err, ErrCantCast) {
		t.Errorf("a suspended card's free cast under a total ban = %v, want ErrCantCast (CR 702.62c)", err)
	}
	if g2.Exile.Contains(id2) == false {
		t.Errorf("a refused free cast did not leave the card in exile")
	}
}

// TestTheSpecialActionItselfIsNotGated is the other half, and the
// half that would be wrong if the gate were placed one function
// higher. Foretelling is a CR 116.2 special action: it does not use
// the stack, it is not a cast, and "players can't cast spells" has
// nothing to say about it.
func TestTheSpecialActionItselfIsNotGated(t *testing.T) {
	g := newRestorableGame(t)
	actor := g.Seats[g.Turn.ActiveSeat]

	prev := CatalogSpecialActions
	CatalogSpecialActions = func(id string) []SpecialAction {
		if id == gateForetellOracle {
			return []SpecialAction{{
				Kind:     SpecialActionForetell,
				Cost:     "{2}",
				CastCost: "{1}{U}",
				Label:    "Foretell {2}",
			}}
		}
		if prev != nil {
			return prev(id)
		}
		return nil
	}
	t.Cleanup(func() { CatalogSpecialActions = prev })

	id := uuid.New()
	g.WithWriteLock(func() {
		seedTotalBan(t, g, actor.ID)
		actor.Hand.PushTop(Card{
			InstanceID: id,
			Name:       "Foretellable",
			TypeLine:   "Instant",
			ManaCost:   "{3}{U}",
			OracleID:   gateForetellOracle,
			Owner:      actor.ID,
			Controller: actor.ID,
		})
	})

	// Permissive payment: the {2} is not what this test is about.
	if err := g.PerformSpecialAction(actor.ID, id, SpecialActionForetell, SpecialActionParams{}); err != nil {
		t.Fatalf("foretelling was refused under a cast restriction: %v — a special action is not a cast (CR 116.2)", err)
	}
	if actor.Hand.Contains(id) {
		t.Errorf("the foretold card is still in hand")
	}
}
