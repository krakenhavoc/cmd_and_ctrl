package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// damage_source_colorless_test.go — #6230: "Black and/or red permanents
// and spells are colorless sources of damage." (Ghostly Flame; CR 105.2c,
// CR 120.2, CR 702.16e).

const ghostlyFlameTestOracle = "test-6230-ghostly-flame"

func withGhostlyFlameRule(t *testing.T) {
	t.Helper()
	withStaticAbilities(t, func(key string) []StaticAbility {
		if key != ghostlyFlameTestOracle {
			return nil
		}
		return []StaticAbility{{Layer: Layer6Ability,
			AppliesTo: func(target *Card, _ *Game, source *Card) bool { return target.InstanceID == source.InstanceID },
			Apply:     func(c *Characteristic, _ *Card, _ *Game, _ *Card) { c.ColorlessDamageSourceColors = "BR" }}}
	})
}

// The acceptance: a black creature's damage gets through protection from
// black, and a red one's through protection from red; a green one's is
// still prevented by protection from green. Without the rule the black
// damage is prevented.
func TestColorlessDamageSourcesGetThroughProtection(t *testing.T) {
	withGhostlyFlameRule(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	proBlack := pushColouredCreature(g, opp, "Pro-Black", []string{"W"}, "protection from black")
	proRed := pushColouredCreature(g, opp, "Pro-Red", []string{"W"}, "protection from red")
	proGreen := pushColouredCreature(g, opp, "Pro-Green", []string{"W"}, "protection from green")
	black := pushLKISource(g, me.ID, []string{"B"})
	red := pushLKISource(g, me.ID, []string{"R"})
	green := pushLKISource(g, me.ID, []string{"G"})

	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(black, proBlack, 1); err != nil {
			t.Fatal(err)
		}
	})
	if got := damageOn(g, proBlack); got != 0 {
		t.Fatalf("without the rule pro-black took %d from a black source, want 0", got)
	}

	flame := pushTypedTestCard(g, Card{Name: "Ghostly Flame", TypeLine: "Enchantment", OracleID: ghostlyFlameTestOracle,
		Colors: []string{"B", "R"}, Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		g.layerVersion.Add(1)
		g.RecomputeLayersIfStaleLocked()
		for _, err := range []error{
			g.DealDamageToCreatureForEffect(black, proBlack, 1),
			g.DealDamageToCreatureForEffect(red, proRed, 1),
			g.DealDamageToCreatureForEffect(green, proGreen, 1),
			// The enchantment itself is a black and red permanent.
			g.DealDamageToCreatureForEffect(flame, proRed, 1),
		} {
			if err != nil {
				t.Fatal(err)
			}
		}
	})
	if got := damageOn(g, proBlack); got != 1 {
		t.Errorf("pro-black took %d from a black source under the rule, want 1", got)
	}
	if got := damageOn(g, proRed); got != 2 {
		t.Errorf("pro-red took %d from two red sources under the rule, want 2", got)
	}
	if got := damageOn(g, proGreen); got != 0 {
		t.Errorf("pro-green took %d from a green source, want 0: green is not named", got)
	}
	var cols []string
	g.ReadSnapshot(func() { cols = findBattlefieldCard(g, black).EffectiveColors() })
	if !slices.Equal(cols, []string{"B"}) {
		t.Errorf("the source's own colours changed to %v: only the damage source is colourless", cols)
	}
}

// A black permanent that has left the battlefield deals its damage as
// it last existed there, a permanent, so the rule still applies.
func TestColorlessDamageSourceDepartedPermanent(t *testing.T) {
	withGhostlyFlameRule(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	proBlack := pushColouredCreature(g, opp, "Pro-Black", []string{"W"}, "protection from black")
	black := pushLKISource(g, me.ID, []string{"B"})
	pushTypedTestCard(g, Card{Name: "Ghostly Flame", TypeLine: "Enchantment", OracleID: ghostlyFlameTestOracle,
		Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		g.layerVersion.Add(1)
		g.RecomputeLayersIfStaleLocked()
	})
	ref := refOf(t, g, black)
	destroy(t, g, black)
	g.WithWriteLock(func() {
		if err := g.DealDamageFromObjectForEffect(ref, proBlack, 2); err != nil {
			t.Fatal(err)
		}
	})
	if got := damageOn(g, proBlack); got != 2 {
		t.Errorf("pro-black took %d from a departed black permanent under the rule, want 2", got)
	}
}

// "A red source of your choice" can't choose a red permanent or spell
// under the rule: as a source of damage it is colourless.
func TestColorlessDamageSourceIsNoRedSourceToChoose(t *testing.T) {
	withGhostlyFlameRule(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	red := pushLKISource(g, me.ID, []string{"R"})
	white := pushLKISource(g, me.ID, []string{"W", "R"})
	q := []PermanentQuery{{Colors: []string{"R"}}}
	var before, after []uuid.UUID
	g.WithWriteLock(func() { before = g.DamageSourceCandidatesLocked(q) })
	if !slices.Contains(before, red) || !slices.Contains(before, white) {
		t.Fatalf("without the rule the red sources %v are not candidates: %v", []uuid.UUID{red, white}, before)
	}
	pushTypedTestCard(g, Card{Name: "Ghostly Flame", TypeLine: "Enchantment", OracleID: ghostlyFlameTestOracle,
		Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		g.layerVersion.Add(1)
		after = g.DamageSourceCandidatesLocked(q)
	})
	if slices.Contains(after, red) || slices.Contains(after, white) {
		t.Errorf("a red source of your choice offers %v under the rule, want neither red permanent", after)
	}
}
