package game

import "testing"

// colorless_damage_sources_test.go — #6230, Ghostly Flame: "Black and/or
// red permanents and spells are colorless sources of damage." CR 105.2c:
// a colorless object has no color, so CR 702.16e's "damage that would be
// dealt by sources that have the stated quality" no longer reaches it.

func withGhostlyFlame(t *testing.T) {
	t.Helper()
	prev := CatalogColorlessDamageSources
	CatalogColorlessDamageSources = func(key string) []string {
		if key == "test-ghostly-flame" {
			return []string{"B", "R"}
		}
		return nil
	}
	t.Cleanup(func() { CatalogColorlessDamageSources = prev })
}

// A black creature's damage gets through protection from black while
// the static is on the battlefield, and is prevented again once it is
// gone.
func TestColorlessDamageSourceGetsThroughProtection(t *testing.T) {
	withGhostlyFlame(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knight := pushColouredCreature(g, opp, "Pro-Black Knight", []string{"W"}, "protection from black")
	shade := pushColouredCreature(g, me, "Shade", []string{"B"})
	flame := pushBattlefieldForTest(g, me.ID, "Ghostly Flame", "Enchantment", "test-ghostly-flame")

	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(shade, knight, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	if got := damageOn(g, knight); got != 1 {
		t.Fatalf("under Ghostly Flame the black source's damage was prevented: %d damage, want 1", got)
	}

	g.WithWriteLock(func() {
		if _, err := g.Battlefield.Remove(flame); err != nil {
			t.Fatal(err)
		}
		if err := g.DealDamageToCreatureForEffect(shade, knight, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	if got := damageOn(g, knight); got != 1 {
		t.Errorf("without Ghostly Flame protection from black let damage through: %d damage, want still 1", got)
	}
}

// A red spell on the stack is a colourless source too; a green one is
// untouched, and the static does not change the object's colour for
// anything else.
func TestColorlessDamageSourceCoversSpellsAndOnlyItsColours(t *testing.T) {
	withGhostlyFlame(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	proRed := pushColouredCreature(g, opp, "Pro-Red Bear", []string{"W"}, "protection from red")
	proGreen := pushColouredCreature(g, opp, "Pro-Green Bear", []string{"W"}, "protection from green")
	pushBattlefieldForTest(g, me.ID, "Ghostly Flame", "Enchantment", "test-ghostly-flame")

	bolt := NewCard("Lightning Bolt", me.ID)
	bolt.TypeLine = "Instant"
	bolt.Colors = []string{"R"}
	growth := NewCard("Test Green Burn", me.ID)
	growth.TypeLine = "Instant"
	growth.Colors = []string{"G"}
	g.WithWriteLock(func() {
		g.Stack.PushTop(bolt)
		g.Stack.PushTop(growth)
		if err := g.DealDamageToCreatureForEffect(bolt.InstanceID, proRed, 1); err != nil {
			t.Fatal(err)
		}
		if err := g.DealDamageToCreatureForEffect(growth.InstanceID, proGreen, 1); err != nil {
			t.Fatal(err)
		}
	})
	if got := damageOn(g, proRed); got != 1 {
		t.Errorf("the red spell's damage was prevented: %d, want 1", got)
	}
	if got := damageOn(g, proGreen); got != 0 {
		t.Errorf("a green source became colourless: %d damage, want 0", got)
	}
	if c, _ := g.LookupCardForEffect(bolt.InstanceID); len(c.Colors) != 1 || c.Colors[0] != "R" {
		t.Errorf("the spell itself lost its colour: %v", c.Colors)
	}
}
