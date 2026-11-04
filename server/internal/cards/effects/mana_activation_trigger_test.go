package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mana_activation_trigger_test.go — #4928. A damage trigger printed beside
// a mana ability ("Whenever you tap this creature for mana, it deals 1
// damage to each opponent") is not a mana ability, because it could not
// add mana (CR 605.1b). The mana lands at once (CR 605.3b) and the damage
// goes on the stack as an ordinary triggered ability, where it can be
// responded to and countered. A painland's "This land deals 1 damage to
// you" is part of the mana ability's own effect and stays inside it.

// druidTable is a main phase with an untapped, unsick Zhur-Taa Druid for
// the active player.
func druidTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g, me, _ := spendTable(t)
	druid := pushCatalogPermanent(g, me.ID, "Zhur-Taa Druid", "Creature — Human Druid", b40ZhurTaaDruidOracle, false)
	return g, me, druid
}

func druidTriggers(g *game.Game, druid uuid.UUID) int {
	n := 0
	g.WithWriteLock(func() { n = triggersOnStackFrom(g, druid) })
	return n
}

func TestZhurTaaDruidManaLandsFirstAndTheDamageUsesTheStack(t *testing.T) {
	g, me, druid := druidTable(t)
	before, mine := opponentLife(g, me), me.Life

	if err := g.ActivateManaAbility(me.ID, druid, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "G" {
		t.Fatalf("CR 605.3b: the {G} is in the pool at once, got %v", got)
	}
	if after := opponentLife(g, me); !sameLife(before, after) {
		t.Fatalf("the damage was dealt inside the mana ability: %v → %v", before, after)
	}
	if n := druidTriggers(g, druid); n != 1 {
		t.Fatalf("CR 605.1b / 603.3: want the damage trigger on the stack, found %d", n)
	}
	item := triggerOnStack(g, druid)
	corpusRequireTriggeredStamp(t, item, "own:")
	if g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Fatalf("the activator should hold priority with the trigger on the stack, holder %d", g.Turn.PriorityHolder)
	}

	passPriorityAroundTable(t, g)
	after := opponentLife(g, me)
	for i := range after {
		if after[i] != before[i]-1 {
			t.Errorf("opponent %d life %d → %d, want -1", i+1, before[i], after[i])
		}
	}
	if me.Life != mine {
		t.Errorf("the controller's life moved: %d → %d", mine, me.Life)
	}
}

// Stifle counters it: the mana is kept and nobody takes damage.
func TestZhurTaaDruidDamageTriggerCanBeCountered(t *testing.T) {
	g, me, druid := druidTable(t)
	before, mine := opponentLife(g, me), me.Life

	if err := g.ActivateManaAbility(me.ID, druid, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	item := triggerOnStack(g, druid)
	if item == nil {
		t.Fatal("no damage trigger on the stack to counter")
	}
	if !taLegalTargets(g, me.ID, stifleOracle)[item.ID] {
		t.Fatal("Stifle must be able to target the Druid's trigger")
	}
	castCatalogSpell(t, g, "Stifle", "Instant", stifleOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: item.ID}})
	passPriorityAroundTable(t, g)

	if after := opponentLife(g, me); !sameLife(before, after) {
		t.Errorf("a countered trigger dealt damage: %v → %v", before, after)
	}
	if me.Life != mine {
		t.Errorf("the controller's life moved: %d → %d", mine, me.Life)
	}
	if druidTriggers(g, druid) != 0 {
		t.Error("the countered trigger is still on the stack")
	}
}

// "Whenever you tap this creature FOR MANA": an opponent's tap effect and
// an attack tap the Druid without activating its mana ability.
func TestZhurTaaDruidTappedAnotherWayTriggersNothing(t *testing.T) {
	t.Run("an effect taps it", func(t *testing.T) {
		g, me, druid := druidTable(t)
		before := opponentLife(g, me)
		var err error
		g.WithWriteLock(func() { err = g.TapTargetForEffect(druid) })
		if err != nil {
			t.Fatalf("TapTargetForEffect: %v", err)
		}
		if !isTapped(g, druid) {
			t.Fatal("setup: the Druid did not tap")
		}
		if n := druidTriggers(g, druid); n != 0 {
			t.Fatalf("a tap effect triggered the Druid %d times", n)
		}
		passPriorityAroundTable(t, g)
		if after := opponentLife(g, me); !sameLife(before, after) {
			t.Errorf("opponents' life %v → %v", before, after)
		}
	})
	t.Run("it attacks", func(t *testing.T) {
		g, me, druid := druidTable(t)
		before := opponentLife(g, me)
		declareAttack(t, g, g.Seats[1].ID, druid)
		if !isTapped(g, druid) {
			t.Fatal("setup: the attacking Druid did not tap")
		}
		if n := druidTriggers(g, druid); n != 0 {
			t.Fatalf("attacking triggered the Druid %d times", n)
		}
		if after := opponentLife(g, me); !sameLife(before, after) {
			t.Errorf("opponents' life %v → %v before combat damage", before, after)
		}
	})
}

// A table with the damage waiting is a restore point, and the restored
// trigger still deals it.
func TestZhurTaaDruidTriggerSurvivesASnapshotRoundTrip(t *testing.T) {
	g, me, druid := druidTable(t)
	before := opponentLife(g, me)
	if err := g.ActivateManaAbility(me.ID, druid, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	corpusRequireTriggeredStamp(t, triggerOnStack(g, druid), "own:")

	restored := restoreThroughJSON(t, g)
	if triggerOnStack(restored, druid) == nil {
		t.Fatal("the restored game lost the trigger")
	}
	passPriorityAroundTable(t, restored)
	rMe := restored.Seats[0]
	after := opponentLife(restored, rMe)
	for i := range after {
		if after[i] != before[i]-1 {
			t.Errorf("restored: opponent %d life %d → %d, want -1", i+1, before[i], after[i])
		}
	}
}

// A painland's damage is the ability's own effect: dealt inside the mana
// ability, nothing on the stack.
func TestPainlandDamageStaysInsideTheManaAbility(t *testing.T) {
	g, me, _ := spendTable(t)
	land := pushCatalogPermanent(g, me.ID, "Shivan Reef", "Land", shivanReefOracle, false)
	mine := me.Life
	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.Life != mine-1 {
		t.Errorf("life %d → %d, want the point dealt inside the mana ability", mine, me.Life)
	}
	var n int
	g.WithWriteLock(func() { n = triggersOnStackFrom(g, land) })
	if n != 0 {
		t.Errorf("a painland put %d triggers on the stack", n)
	}
}
