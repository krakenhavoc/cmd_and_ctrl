package game

import (
	"testing"

	"github.com/google/uuid"
)

// A spell's own conditional riders — "If X is 5 or more, this spell
// can't be countered and the damage can't be prevented" (Banefire) —
// are judged as they are asked, off the spell's stack item (ADR 0107
// §5).
func TestSpellConditionsAreJudgedOffTheItem(t *testing.T) {
	const oracle = "test-banefire-riders"
	big := func(_ *Game, item *StackItem) bool { return item.XValue >= 5 }
	oldCounter, oldDamage := CatalogCantBeCounteredIf, CatalogSpellDamageUnpreventable
	CatalogCantBeCounteredIf = func(key string) SpellCondition {
		if key == oracle {
			return big
		}
		return nil
	}
	CatalogSpellDamageUnpreventable = CatalogCantBeCounteredIf
	t.Cleanup(func() { CatalogCantBeCounteredIf, CatalogSpellDamageUnpreventable = oldCounter, oldDamage })

	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushColouredCreature(g, opp, "Pro-Red", []string{"W"}, "protection from red")
	spell := uuid.New()
	pushStackSpell(t, g, Card{InstanceID: spell, Name: "Banefire", TypeLine: "Sorcery", Colors: []string{"R"},
		OracleID: oracle, Owner: me.ID, Controller: me.ID})

	g.StackMeta[spell].XValue = 4
	g.WithWriteLock(func() {
		if g.SpellCantBeCounteredForEffect(spell) || g.SpellDamageCantBePreventedForEffect(spell) {
			t.Error("X=4: neither rider applies")
		}
		_ = g.DealDamageToCreatureForEffect(spell, victim, 4)
	})
	if got := damageOn(g, victim); got != 0 {
		t.Fatalf("X=4 damage got through protection: %d", got)
	}
	g.StackMeta[spell].XValue = 5
	g.WithWriteLock(func() {
		if !g.SpellCantBeCounteredForEffect(spell) || !g.SpellDamageCantBePreventedForEffect(spell) {
			t.Error("X=5: both riders apply")
		}
		_ = g.DealDamageToCreatureForEffect(spell, victim, 5)
	})
	if got := damageOn(g, victim); got != 5 {
		t.Errorf("X=5 damage: victim has %d, want 5 through protection", got)
	}
}
