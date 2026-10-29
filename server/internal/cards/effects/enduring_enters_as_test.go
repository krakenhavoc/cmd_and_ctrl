package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// enduring_enters_as_test.go — the Enduring cycle returns through ADR
// 0210's entry overwrite (#4512): a returning Glimmer is an enchantment
// from the moment it lands, so it is never a creature to "whenever a
// creature enters" (Soul Warden) or to an entry replacement (Authority
// of the Consuls).

var e4512Enduring = []struct{ name, typeLine, oracle string }{
	{"Enduring Curiosity", "Enchantment Creature — Cat Glimmer", enduringCuriosityOracle},
	{"Enduring Vitality", "Enchantment Creature — Elk Glimmer", enduringVitalityOracle},
	{"Enduring Innocence", "Enchantment Creature — Sheep Glimmer", enduringInnocenceOra},
	{"Enduring Tenacity", "Enchantment Creature — Snake Glimmer", e2EnduringTenacityOracle},
}

// e4512AssertEnchantment fails unless `id` is a plain, noncreature
// Enchantment on the battlefield.
func e4512AssertEnchantment(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Glimmer is not on the battlefield")
	}
	c := e4506View(t, g, id)
	if c.IsCreature() || !slices.Equal(c.Effective().Types, []string{"Enchantment"}) {
		t.Errorf("%s: types %v, want [Enchantment]", c.Name, c.Effective().Types)
	}
}

// Soul Warden ("whenever another creature enters, you may gain 1 life")
// never sees the returning Glimmer: it is an enchantment the moment it
// lands. The test reads the life total, so a creature entering would
// show as a gain.
func TestE4512ReturningGlimmerIsNeverACreatureToSoulWarden(t *testing.T) {
	for _, tc := range e4512Enduring {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			advanceToMain(t, g)
			pushCatalogPermanent(g, me.ID, "Soul Warden", "Creature — Human Cleric", e4506SoulWardenOracle, false)
			id := pushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			life := me.Life

			destroyNow(t, g, id)
			passPriorityAroundTable(t, g)

			e4512AssertEnchantment(t, g, id)
			if c := e4506View(t, g, id); c.Owner != me.ID || c.Controller != me.ID {
				t.Errorf("owner/controller %s/%s, want %s", c.Owner, c.Controller, me.ID)
			}
			if me.Life != life {
				t.Errorf("Soul Warden saw a creature enter: life %d, want %d", me.Life, life)
			}
		})
	}
}

// CR 614.12: an opponent's Authority of the Consuls taps "creatures your
// opponents control" as they enter, and gains its controller life when
// one does. A returning Glimmer is neither.
func TestE4512ReturningGlimmerIsNotACreatureToAuthorityOfTheConsuls(t *testing.T) {
	for _, tc := range e4512Enduring {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			advanceToMain(t, g)
			pushCatalogPermanent(g, opp.ID, "Authority of the Consuls", "Enchantment", e4506AuthorityOracle, false)
			id := pushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			oppLife := opp.Life

			destroyNow(t, g, id)
			passPriorityAroundTable(t, g)

			e4512AssertEnchantment(t, g, id)
			if c := e4506View(t, g, id); c.Tapped {
				t.Error("the Glimmer entered tapped: Authority of the Consuls judged it a creature")
			}
			if opp.Life != oppLife {
				t.Errorf("Authority of the Consuls gained life for it: %d, want %d", opp.Life, oppLife)
			}
		})
	}
}

// The control: a plain creature entering under the same Authority is
// tapped, so the test above fails for the reason it says.
func TestE4512AuthorityOfTheConsulsStillTapsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, opp.ID, "Authority of the Consuls", "Enchantment", e4506AuthorityOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	oppLife := opp.Life

	destroyNow(t, g, bear)
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardForEffect(bear, game.ZoneBattlefield); err != nil {
			t.Fatalf("reanimate: %v", err)
		}
	})
	passPriorityAroundTable(t, g)

	if c := e4506View(t, g, bear); !c.Tapped {
		t.Error("a creature entering under Authority of the Consuls should be tapped")
	}
	if opp.Life != oppLife+1 {
		t.Errorf("opponent life %d, want %d", opp.Life, oppLife+1)
	}
}

// It keeps its abilities: Enduring Curiosity, back as an enchantment,
// still draws when a creature you control connects.
func TestE4512ReturnedCuriosityKeepsItsAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	cat := pushCatalogPermanent(g, me.ID, "Enduring Curiosity", "Enchantment Creature — Cat Glimmer", enduringCuriosityOracle, false)
	destroyNow(t, g, cat)
	passPriorityAroundTable(t, g)
	e4512AssertEnchantment(t, g, cat)

	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	before := me.Hand.Size()
	dealCombatDamageToPlayer(g, bear, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d: the returned Curiosity lost its draw trigger", got, before+1)
	}
}

// The record belongs to the object (CR 400.7). A flicker returns the
// printed Enchantment Creature, which Soul Warden sees; it dies as a
// creature, so it returns as an enchantment once more. A reanimation by
// another effect after it died as an enchantment is a creature again.
func TestE4512ANewObjectAfterwardsEntersAsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Soul Warden", "Creature — Human Cleric", e4506SoulWardenOracle, false)
	cat := pushCatalogPermanent(g, me.ID, "Enduring Curiosity", "Enchantment Creature — Cat Glimmer", enduringCuriosityOracle, false)
	destroyNow(t, g, cat)
	passPriorityAroundTable(t, g)
	e4512AssertEnchantment(t, g, cat)
	if n := e4512Records(g); n != 1 {
		t.Fatalf("%d entry records, want 1", n)
	}

	// The flicker.
	life := me.Life
	g.WithWriteLock(func() {
		if err := (Flicker{Target: cat}).Apply(ctxFor(g, &game.StackItem{Controller: me.ID, SourceCardID: uuid.New()})); err != nil {
			t.Fatalf("flicker: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	// The flickered card is a new object with a new ID.
	cat = findBattlefieldByName(g, "Enduring Curiosity")
	if cat == uuid.Nil {
		t.Fatal("the flickered Enduring Curiosity did not come back")
	}
	c := e4506View(t, g, cat)
	if !c.IsCreature() || !c.IsEnchantment() {
		t.Fatalf("the flickered card is %v, want the Enchantment Creature it is", c.Effective().Types)
	}
	if me.Life <= life {
		t.Errorf("Soul Warden did not see the flickered creature enter: life %d", me.Life)
	}
	if n := e4512Records(g); n != 0 {
		t.Errorf("%d entry records after the flicker, want 0: the old object's record was not swept", n)
	}

	// It died as a creature, so the trigger returns it once more.
	destroyNow(t, g, cat)
	passPriorityAroundTable(t, g)
	e4512AssertEnchantment(t, g, cat)

	// It dies as an enchantment: it stays, and another effect's
	// reanimation makes it a creature again.
	destroyNow(t, g, cat)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cat) || !me.Graveyard.Contains(cat) {
		t.Fatal("dying as an enchantment should leave it in the graveyard")
	}
	if n := e4512Records(g); n != 0 {
		t.Errorf("%d entry records after it died, want 0", n)
	}
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardForEffect(cat, game.ZoneBattlefield); err != nil {
			t.Fatalf("reanimate: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if c := e4506View(t, g, cat); !c.IsCreature() {
		t.Errorf("reanimated as %v, want a creature", c.Effective().Types)
	}
}

// e4512Records counts the live Enduring Curiosity entry records.
func e4512Records(g *game.Game) int {
	n := 0
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		for _, e := range g.ScopedEffects {
			if e.Label == "Enduring Curiosity — it's an enchantment (it's not a creature)" {
				n++
			}
		}
	})
	return n
}

// A table holding a returned Glimmer is a restore point, and it is still
// an enchantment after the round trip. Dying again in the restored game
// leaves it in the graveyard.
func TestE4512ReturnedGlimmerSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	cat := pushCatalogPermanent(g, me.ID, "Enduring Curiosity", "Enchantment Creature — Cat Glimmer", enduringCuriosityOracle, false)
	destroyNow(t, g, cat)
	passPriorityAroundTable(t, g)
	e4512AssertEnchantment(t, g, cat)

	r := restoreRoundTrip(t, g, true)
	e4512AssertEnchantment(t, r, cat)
	if n := e4512Records(r); n != 1 {
		t.Errorf("%d entry records after the restore, want 1", n)
	}
	destroyNow(t, r, cat)
	passPriorityAroundTable(t, r)
	if r.Battlefield.Contains(cat) {
		t.Error("dying as a restored enchantment returned it again")
	}
}
