package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// the_super_hero_civil_war_test.go — The Super Hero Civil War (Marvel
// Super Heroes, #4455, #4616), played through the real catalog entry:
// chapter I's "up to two target creatures with total mana value 6 or
// less" at announce and on resolution, and the other two chapters.

const superHeroCivilWarOracle = "11d74dc2-8c11-4d4b-b727-8216d2eb6b4e"

// civilWarCreature puts a creature with a mana cost onto owner's side.
func civilWarCreature(g *game.Game, owner uuid.UUID, name, cost string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Hero", ManaCost: cost,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

func civilWarController(t *testing.T, g *game.Game, id uuid.UUID) uuid.UUID {
	t.Helper()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("card %s is gone", id)
	}
	return c.Controller
}

// civilWarCast casts the Saga for the active seat and stops at chapter
// I's target prompt.
func civilWarCast(t *testing.T, g *game.Game) (uuid.UUID, *game.PendingChoice) {
	t.Helper()
	saga := castCatalogSpell(t, g, "The Super Hero Civil War", "Enchantment — Saga", superHeroCivilWarOracle, nil)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, g.Seats[g.Turn.ActiveSeat].ID)
	if p == nil {
		t.Fatal("chapter I asked for no targets")
	}
	return saga, p
}

func civilWarRefs(ids ...uuid.UUID) []game.TargetRef {
	out := make([]game.TargetRef, 0, len(ids))
	for _, id := range ids {
		out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return out
}

func TestTheSuperHeroCivilWarIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(superHeroCivilWarOracle)
	if !ok {
		t.Fatal("The Super Hero Civil War is not registered")
	}
	if spec.Name != "The Super Hero Civil War" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("spec = %q / %v / %v, want The Super Hero Civil War, Full, no caveats", spec.Name, spec.Completeness, spec.Caveats)
	}
	if len(spec.Triggered) != 3 {
		t.Fatalf("chapters = %d, want 3", len(spec.Triggered))
	}
	c := spec.Triggered[0].Targets
	if c == nil || c.Min != 0 || c.Max != 2 || c.Total == nil || c.Total.Max != 6 || c.Total.Stat != game.TargetTotalManaValue {
		t.Errorf("chapter I's clause = %+v, want up to two with total mana value 6 or less", c)
	}
}

// Chapter I, CR 601.2c via CR 603.3d: a pair over 6 is refused with the
// rule in the refusal and the prompt stays open; a pair at exactly 6 is
// taken, for as long as the Saga remains.
func TestTheSuperHeroCivilWarChapterIStealsUpToTwoWithinTheTotal(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	four := civilWarCreature(g, opp.ID, "Four", "{3}{R}", 4, 4)
	three := civilWarCreature(g, opp.ID, "Three", "{2}{R}", 3, 3)
	two := civilWarCreature(g, opp.ID, "Two", "{1}{R}", 2, 2)
	seven := civilWarCreature(g, opp.ID, "Seven", "{6}{R}", 7, 7)

	_, p := civilWarCast(t, g)
	if !hasID(p.PickTargetCards, seven) {
		t.Error("a mana value 7 creature is each-legal and offered; the total refuses it")
	}
	err := g.ResolvePickTargets(p.ID, me.ID, civilWarRefs(four, three))
	if !errors.Is(err, game.ErrIllegalTarget) || !strings.Contains(err.Error(), "have total mana value 6 or less") {
		t.Fatalf("mana values 4 and 3: err = %v, want the total refusal", err)
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, civilWarRefs(seven)); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("mana value 7 alone: err = %v, want ErrIllegalTarget", err)
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, civilWarRefs(four, two)); err != nil {
		t.Fatalf("mana values 4 and 2 total 6: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{four, two} {
		if civilWarController(t, g, id) != me.ID {
			t.Errorf("chapter I: %s is not yours", id)
		}
	}
	if civilWarController(t, g, three) != opp.ID || civilWarController(t, g, seven) != opp.ID {
		t.Error("chapter I took a creature it did not target")
	}
}

// CR 608.2b: the survivors are summed again as the chapter resolves. A
// target whose mana value grew so the pair now totals 8 makes the whole
// selection illegal: nothing is taken (the printed rulings for this
// family).
func TestTheSuperHeroCivilWarChapterITakesNothingWhenTheTotalGrows(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := civilWarCreature(g, opp.ID, "A", "{2}{R}", 3, 3)
	b := civilWarCreature(g, opp.ID, "B", "{2}{R}", 3, 3)

	_, p := civilWarCast(t, g)
	if err := g.ResolvePickTargets(p.ID, me.ID, civilWarRefs(a, b)); err != nil {
		t.Fatalf("mana values 3 and 3: %v", err)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == b {
				g.Battlefield.Cards[i].ManaCost = "{4}{R}"
			}
		}
	})
	passPriorityAroundTable(t, g)
	if civilWarController(t, g, a) != opp.ID || civilWarController(t, g, b) != opp.ID {
		t.Error("a selection totalling 8 on resolution is illegal: neither creature is taken")
	}
}

// CR 608.2b: a target that leaves in response costs only itself; the
// other is summed alone and taken.
func TestTheSuperHeroCivilWarChapterIKeepsTheSurvivorWhenOneLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := civilWarCreature(g, opp.ID, "A", "{2}{R}", 3, 3)
	b := civilWarCreature(g, opp.ID, "B", "{2}{R}", 3, 3)

	_, p := civilWarCast(t, g)
	if err := g.ResolvePickTargets(p.ID, me.ID, civilWarRefs(a, b)); err != nil {
		t.Fatalf("mana values 3 and 3: %v", err)
	}
	g.WithWriteLock(func() { g.Battlefield.Remove(b) })
	passPriorityAroundTable(t, g)
	if civilWarController(t, g, a) != me.ID {
		t.Error("the survivor is within the total and is taken")
	}
}

// Chapters II and III, and the stolen creatures going home: II gives
// your creatures +1/+1 and vigilance until end of turn; III has one of
// yours fight up to one other; the Saga is then sacrificed and the
// creatures chapter I took return to their controller.
func TestTheSuperHeroCivilWarChaptersIIAndIII(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	stolen := civilWarCreature(g, opp.ID, "Stolen", "{3}{R}", 4, 4)
	victim := civilWarCreature(g, opp.ID, "Victim", "{1}{R}", 2, 2)
	mine := civilWarCreature(g, me.ID, "Mine", "{1}{W}", 2, 2)

	saga, p := civilWarCast(t, g)
	if err := g.ResolvePickTargets(p.ID, me.ID, civilWarRefs(stolen)); err != nil {
		t.Fatalf("chapter I: %v", err)
	}
	passPriorityAroundTable(t, g)
	if civilWarController(t, g, stolen) != me.ID {
		t.Fatal("chapter I: the targeted creature is not yours")
	}

	// Chapter II at your next precombat main.
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if got := loreCountersOn(g, saga); got != 2 {
		t.Fatalf("lore counters = %d, want 2", got)
	}
	for _, id := range []uuid.UUID{mine, stolen} {
		c := ceOnBattlefield(t, g, id)
		if !game.HasKeyword(&c, "vigilance") {
			t.Errorf("chapter II: %s has no vigilance", c.Name)
		}
	}
	if c := ceOnBattlefield(t, g, mine); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("chapter II: Mine is %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	if c := ceOnBattlefield(t, g, victim); c.CurrentPower() != 2 {
		t.Errorf("chapter II pumped an opponent's creature to power %d", c.CurrentPower())
	}

	// Chapter III: the stolen 4/4 fights the 2/2, which dies; the Saga is
	// sacrificed and the 4/4 goes home.
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, stolen)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("chapter III: the 2/2 survived a fight with a 4/4")
	}
	if c := ceOnBattlefield(t, g, stolen); c.DamageMarked != 2 {
		t.Errorf("chapter III: the 4/4 took %d damage, want 2", c.DamageMarked)
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
	if civilWarController(t, g, stolen) != opp.ID {
		t.Error("the stolen creature did not go home with the Saga")
	}
}

// Chapter III's second target is optional: with none, nothing fights.
func TestTheSuperHeroCivilWarChapterIIIWithNoSecondTargetDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	mine := civilWarCreature(g, me.ID, "Mine", "{1}{W}", 2, 2)
	saga := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Super Hero Civil War", TypeLine: "Enchantment — Saga",
		OracleID: superHeroCivilWarOracle, ManaCost: "{3}{R}{W}", Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == saga {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterLore: 2}
			}
		}
	})
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, mine)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
			t.Fatalf("declining the second target: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(saga) {
		t.Fatal("chapter III did not resolve: the Saga is still on the battlefield")
	}
	if c := ceOnBattlefield(t, g, mine); c.DamageMarked != 0 {
		t.Errorf("with no second target nothing fights, but Mine took %d damage", c.DamageMarked)
	}
}
