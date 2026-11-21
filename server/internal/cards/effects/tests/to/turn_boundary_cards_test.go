package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// turn_boundary_cards_test.go — #1009, the card side.
//
// Every "this turn" reader in the catalog now asks Game.TurnTally.
// The four that were still walking the event log back to
// EventBeginUpkeep are proved here, each with the case the walk got
// wrong: something that happened this turn BEFORE the upkeep event.
//
// An upkeep event that is not the first thing in the turn is
// reachable two ways — a permanent entering, an attack being
// declared or a Faerie connecting during the UNTAP step (#70,
// ADR 0070 makes untap-step choices and triggers real), and an EXTRA
// upkeep step (CR 500.8, ADR 0059 Decision 4). The tests spell it the
// second way because it needs no untap-step machinery: they do the
// thing, then put an upkeep event into the log behind it, which is
// precisely the shape the old walk answered "no" to.

// emitUpkeep puts an upkeep-began event into the log — the event the
// migrated walks used to stop at.
func emitUpkeep(g *game.Game, active uuid.UUID) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBeginUpkeep, Actor: active})
	})
}

// Oran-Rief, the Vastwood: "{T}: Put a +1/+1 counter on each green
// creature that entered this turn." The creature entered, an upkeep
// event landed behind it, and it still entered this turn.
func TestB06OranRiefGrowsACreatureThatEnteredBeforeTheUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.SeedPermanentWithOracle(g, me.ID, "Oran-Rief, the Vastwood", "Land", effectstest.B06OranRiefOracle)
	old := effectstest.SeedCreature(g, "Old Bear", me.ID) // no ETB event at all
	fresh := effectstest.B06GreenCreature(t, g, me.ID, "Fresh Elf")

	emitUpkeep(g, me.ID)

	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.CounterCount(g, fresh, "+1/+1") != 1 {
		t.Error("the upkeep is not the turn boundary: a creature that entered before it entered this turn")
	}
	if effectstest.CounterCount(g, old, "+1/+1") != 0 {
		t.Error("a creature that never entered this turn must not grow")
	}
}

// Éowyn, Shieldmaiden's "another Human entered this turn" reads the
// subtype tally for the count and the per-object cell to decide
// whether Éowyn herself is one of the entries it counted. With an
// upkeep event behind her own entry, the old walk said she was not,
// and one Human — her — satisfied "another".
func TestB43EowynCountsHerOwnEntryFromBeforeTheUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	eowyn := effectstest.B43Catalog(g, me.ID, "Éowyn, Shieldmaiden", "Legendary Creature — Human Knight", effectstest.B43EowynOracle, 5, 4)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: eowyn})
	})
	emitUpkeep(g, me.ID)

	source, ok := g.LookupCardForEffect(eowyn)
	if !ok {
		t.Fatal("setup: Éowyn is nowhere")
	}
	if effects.B43AnotherHumanEnteredThisTurn(g, &source) {
		t.Error("Éowyn is the only Human that entered this turn, so there is no ANOTHER")
	}

	other := effectstest.B12Creature(g, me.ID, "Human Soldier", "Creature — Human Soldier", 1, 1)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: other})
	})
	if !effects.B43AnotherHumanEnteredThisTurn(g, &source) {
		t.Error("a second Human entered, so there is an ANOTHER")
	}
}

// Chart a Course's "unless you attacked this turn" is the
// AttacksDeclared cell. An attack declared before an upkeep event is
// still an attack this turn.
func TestB18AttackedThisTurnSeesAnAttackBeforeTheUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if effects.B18AttackedThisTurn(g, me.ID) {
		t.Fatal("setup: nobody has attacked yet")
	}
	effectstest.AttackWith(t, g, opp.ID, bear)
	emitUpkeep(g, me.ID)

	if !effects.B18AttackedThisTurn(g, me.ID) {
		t.Error("the upkeep is not the turn boundary: the attack was declared this turn")
	}
	if effects.B18AttackedThisTurn(g, opp.ID) {
		t.Error("the defender did not attack")
	}
}
