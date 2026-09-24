package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attack_requirement_cards_test.go — the proof cards of #1571 (CR
// 508.1d attack requirements): Bident of Thassa's activated ability and
// The Akroan War. Alela's goad is pinned in batch33_test.go.

const akroanWarOracle = "8c0b9596-c44f-44fb-89c9-056da0db33d3"

// advanceToDeclareAttackersOf walks AdvanceStep to `seat`'s
// declare-attackers step.
func advanceToDeclareAttackersOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 300; i++ {
		if g.Turn.Step == game.StepDeclareAttackers && g.Turn.ActiveSeat == seat {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep toward declare attackers of seat %d: %v", seat, err)
		}
	}
	t.Fatalf("never reached declare attackers of seat %d", seat)
}

func hasRequirementRecord(g *game.Game) bool {
	for _, e := range g.ScopedEffects {
		for _, m := range e.Mods {
			if m.Kind == game.ModAddAttackRequirement {
				return true
			}
		}
	}
	return false
}

// TestBidentMakesOpponentsCreaturesAttackThisTurn — "{1}{U}, {T}:
// Creatures your opponents control attack this turn if able." Activated
// in an opponent's upkeep, it reaches a creature that opponent gains
// AFTER it resolved (CR 611.2c: a requirement does not lock its set),
// the opponent's pass is refused until both attack, and it is gone by
// the next turn.
func TestBidentMakesOpponentsCreaturesAttackThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	bident := pushPermanentForTest(g, me.ID, "Bident of Thassa", bidentOracle, "Legendary Enchantment Artifact")
	early := pushVanillaCreature(g, opp.ID, "Early Bear", 2, 2)
	advanceToUpkeepOf(t, g, 1)

	me.ManaPool.AddMana(game.ManaToken{Color: "U"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, bident, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Bident: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasRequirementRecord(g) {
		t.Fatal("the ability registered no requirement")
	}
	// A hasty creature arriving after the ability resolved: it can
	// attack, so it must.
	late := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Late Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID, Keywords: []string{"haste"},
	})
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)

	advanceToDeclareAttackersOf(t, g, 1)
	if err := g.PassPriority(); !errors.Is(err, game.ErrAttackRequirement) {
		t.Fatalf("opponent's pass with both Bears home = %v, want ErrAttackRequirement", err)
	}
	if err := g.DeclareAttacker(early, me.ID); err != nil {
		t.Fatal(err)
	}
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != late {
		t.Fatalf("pass with the late Bear home = %v, want a refusal naming it", err)
	}
	if got := re.Sentence(opp.ID); got != "Late Bear must attack this combat if able (Bident of Thassa)." {
		t.Errorf("sentence = %q", got)
	}
	if err := g.DeclareAttacker(late, third.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with both Bears attacking: %v", err)
	}
	_ = mine // the Bident's controller's own creature is never affected

	advanceToUpkeepOf(t, g, 2)
	if hasRequirementRecord(g) {
		t.Error("\"this turn\" outlived the turn")
	}
}

// TestTheAkroanWarChapters — I steals a creature for as long as the
// Saga remains; II makes every opponent's creature attack until your
// next turn; III has each tapped creature deal damage to itself; and
// the stolen creature goes home when the Saga is sacrificed.
func TestTheAkroanWarChapters(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	stolen := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	wolf := b12Creature(g, opp.ID, "Their Wolf", "Creature — Wolf", 2, 2)

	saga := castCatalogSpell(t, g, "The Akroan War", "Enchantment — Saga", akroanWarOracle, nil)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, stolen)
	passPriorityAroundTable(t, g)
	if c, ok := g.LookupCardForEffect(stolen); !ok || c.Controller != me.ID {
		t.Fatal("chapter I: the targeted creature is not yours")
	}

	// Chapter II at your next precombat main.
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if got := loreCountersOn(g, saga); got != 2 {
		t.Fatalf("lore counters = %d, want 2", got)
	}
	advanceToDeclareAttackersOf(t, g, (seat+1)%len(g.Seats))
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != wolf {
		t.Fatalf("opponent's pass with the Wolf home = %v, want a refusal naming it", err)
	}
	if err := g.DeclareAttacker(wolf, me.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Wolf attacking: %v", err)
	}

	// Chapter III: the Wolf is still tapped from attacking, deals 2 to
	// itself and dies; the Saga is sacrificed and the Bear goes home.
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if hasRequirementRecord(g) {
		t.Error("chapter II's requirement outlived \"until your next turn\"")
	}
	if g.Battlefield.Contains(wolf) {
		t.Error("chapter III: the tapped Wolf survived 2 damage from itself")
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
	if c, ok := g.LookupCardForEffect(stolen); !ok || c.Controller != opp.ID {
		t.Error("the stolen creature did not go home with the Saga")
	}
}
