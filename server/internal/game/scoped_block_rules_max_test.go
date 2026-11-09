package game

import "testing"

// cantBeBlockedByMoreThan (Full Steam Ahead's granted "This creature can't be
// blocked by more than one creature", CR 509.1b): a count rule on the pinned
// attacker, the tightest maximum winning. The card-level behaviour (two
// blockers refused, one allowed, a later arrival unaffected) is in
// cards/effects/std_b34_a_cards_test.go; this pins the registration check and
// the adapter.
func TestCantBeBlockedByMoreThanNeedsAMaximumOfAtLeastOne(t *testing.T) {
	if blockRuleModProblem(CantBeBlockedByMoreThanMod(0)) == "" {
		t.Error("a maximum of 0 was accepted")
	}
	if p := blockRuleModProblem(CantBeBlockedByMoreThanMod(1)); p != "" {
		t.Errorf("a maximum of 1 was refused: %s", p)
	}
}

func TestCantBeBlockedByMoreThanBecomesACountRuleOnThePinnedAttacker(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	attacker := pushCombatant(t, g, me, "Attacker", 2, 2)
	other := pushCombatant(t, g, me, "Other Attacker", 2, 2)
	g.WithWriteLock(func() {
		if !g.RegisterScopedEffectForEffect(me.ID, g.PinnedObjectsLocked(attacker),
			[]Mod{CantBeBlockedByMoreThanMod(1)}, IndefiniteDuration(), "test") {
			t.Fatal("setup: registered nothing")
		}
		g.RecomputeLayersIfStaleLocked()
		var min, max int
		found := false
		g.forEachScopedBlockRuleLocked(func(rule BlockRule, _ *Card) bool {
			if rule.Count != nil {
				min, max = rule.Count(g, findCard(g, attacker), nil)
				found = true
			}
			return true
		})
		if !found || min != 0 || max != 1 {
			t.Errorf("count rule for the pinned attacker = (%d, %d, found %v), want (0, 1)", min, max, found)
		}
		g.forEachScopedBlockRuleLocked(func(rule BlockRule, _ *Card) bool {
			if rule.Count != nil {
				if min, max := rule.Count(g, findCard(g, other), nil); min != 0 || max != 0 {
					t.Errorf("an unpinned attacker is bound: (%d, %d)", min, max)
				}
			}
			return true
		})
	})
}

// MinSpellManaValue is MaxSpellManaValue's floor (Glarb, Calamity's Augur's
// "cast spells with mana value 4 or greater from the top of your library").
func TestMinSpellManaValueIsAFloorOnTheSpellCast(t *testing.T) {
	four := 4
	perm := &CastPermission{MinSpellManaValue: &four}
	cheap := Card{Name: "Cheap", TypeLine: "Creature — Bear", ManaCost: "{2}{G}"}
	dear := Card{Name: "Dear", TypeLine: "Creature — Bear", ManaCost: "{3}{G}"}
	if spellManaValueWithinCap(cheap, perm) {
		t.Error("a mana value 3 spell passed a floor of 4")
	}
	if !spellManaValueWithinCap(dear, perm) {
		t.Error("a mana value 4 spell failed a floor of 4")
	}
	if !spellManaValueWithinCap(cheap, nil) {
		t.Error("no permission is no bound")
	}
	two := 2
	both := &CastPermission{MinSpellManaValue: &four, MaxSpellManaValue: &two}
	if spellManaValueWithinCap(dear, both) {
		t.Error("a floor above the cap admits nothing")
	}
}
