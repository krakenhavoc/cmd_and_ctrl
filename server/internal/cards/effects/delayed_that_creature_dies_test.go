package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// delayed_that_creature_dies_test.go — #2057: "when that creature dies
// this turn" as a delayed trigger with a target chosen as it triggers
// (CR 603.3d), about the creature as an object (CR 603.7c), reading its
// counters as it last existed (CR 608.2h). Heroic Sacrifice and Fatal
// Fissure.

const (
	hsHeroicSacrifice = "03c47f1c-02a7-428c-a126-9e85325ebc71"
	hsFatalFissure    = "202e45d6-8c7a-46db-85fb-634aa77b4097"
)

// hsDrain resolves everything: each target pick is answered with `pick`
// (nil chooses no target), and priority is passed until the stack and the
// pending triggers are empty. It reports how many picks it answered.
func hsDrain(t *testing.T, g *game.Game, pick []game.TargetRef) int {
	t.Helper()
	picks := 0
	for i := 0; i < 64; i++ {
		var open *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil {
				open = c
				break
			}
		}
		if open == nil {
			if len(g.StackMeta) == 0 && len(g.PendingTriggers) == 0 {
				return picks
			}
			if err := g.PassPriority(); err != nil {
				return picks
			}
			continue
		}
		if open.Kind != game.PendingChoicePickTarget {
			t.Fatalf("unexpected prompt %s", open.Kind)
		}
		picks++
		if err := g.ResolvePickTargets(open.ID, open.Chooser, pick); err != nil {
			t.Fatalf("pick target: %v", err)
		}
	}
	return picks
}

func hsCast(t *testing.T, g *game.Game, chosen uuid.UUID) uuid.UUID {
	t.Helper()
	id := castCatalogSpell(t, g, "Heroic Sacrifice", "Instant", hsHeroicSacrifice, pr6Card(chosen))
	passPriorityAroundTable(t, g)
	return id
}

func hsCounters(t *testing.T, g *game.Game, id uuid.UUID, counters map[string]int) {
	t.Helper()
	g.WithWriteLock(func() {
		for k, n := range counters {
			if err := g.AddCounterForEffect(id, k, n); err != nil {
				t.Fatalf("add counter: %v", err)
			}
		}
	})
}

func TestHeroicSacrificeAndFatalFissureAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{hsHeroicSacrifice: "Heroic Sacrifice", hsFatalFissure: "Fatal Fissure"} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name {
			t.Fatalf("%s is not registered", name)
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s: Completeness %v, want Full", name, spec.Completeness)
		}
	}
}

// Damage to you and to another creature you control is dealt to the
// chosen creature instead; damage to an opponent's creature is not.
func TestHeroicSacrificeRedirectsDamageToTheChosenCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	chosen := b12Creature(g, me.ID, "Chosen", "Creature — Soldier", 1, 20)
	other := b12Creature(g, me.ID, "Other", "Creature — Soldier", 1, 20)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 20)
	src := b12Creature(g, opp.ID, "Source", "Creature — Bear", 1, 1)
	hsCast(t, g, chosen)
	life := me.Life

	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(src, me.ID, 3)
		_ = g.DealDamageToCreatureForEffect(src, other, 2)
		_ = g.DealDamageToCreatureForEffect(src, chosen, 1)
		_ = g.DealDamageToCreatureForEffect(src, theirs, 4)
	})
	if me.Life != life {
		t.Errorf("you lost %d life, want 0", life-me.Life)
	}
	if got := pr6Marked(g, chosen); got != 6 {
		t.Errorf("chosen creature has %d damage, want 6", got)
	}
	if got := pr6Marked(g, other); got != 0 {
		t.Errorf("other creature has %d damage, want 0", got)
	}
	if got := pr6Marked(g, theirs); got != 4 {
		t.Errorf("opponent's creature has %d damage, want 4", got)
	}
}

// "If it's still on the battlefield": once the chosen creature has died,
// damage to you is dealt to you.
func TestHeroicSacrificeStopsRedirectingOnceTheCreatureIsGone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	chosen := b12Creature(g, me.ID, "Chosen", "Creature — Soldier", 1, 1)
	src := b12Creature(g, opp.ID, "Source", "Creature — Bear", 1, 1)
	hsCast(t, g, chosen)
	p1Destroy(t, g, chosen)
	g.RunStateChecksForTest()
	hsDrain(t, g, nil)
	life := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 3) })
	if lost := life - me.Life; lost != 3 {
		t.Errorf("you lost %d life, want 3", lost)
	}
}

// When the chosen creature dies, every kind of counter it had as it last
// existed goes on the target, and you draw a card.
func TestHeroicSacrificeMovesItsCountersAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	chosen := b12Creature(g, me.ID, "Chosen", "Creature — Soldier", 1, 1)
	heir := b12Creature(g, me.ID, "Heir", "Creature — Soldier", 1, 1)
	hsCounters(t, g, chosen, map[string]int{game.CounterPlusOne: 2, "stun": 1})
	hsCast(t, g, chosen)
	hand := len(me.Hand.Cards)

	p1Destroy(t, g, chosen)
	g.RunStateChecksForTest()
	if picks := hsDrain(t, g, pr6Card(heir)); picks != 1 {
		t.Fatalf("%d target picks, want 1", picks)
	}
	c := layeredCard(t, g, heir)
	if c.Counters[game.CounterPlusOne] != 2 || c.Counters["stun"] != 1 {
		t.Errorf("heir counters %v, want two +1/+1 and one stun", c.Counters)
	}
	if got := len(me.Hand.Cards) - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("%d delayed triggers left, want 0 (it triggers once)", len(g.DelayedTriggers))
	}
}

// "Up to one": choosing no target still draws the card.
func TestHeroicSacrificeNoTargetStillDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	chosen := b12Creature(g, me.ID, "Chosen", "Creature — Soldier", 1, 1)
	heir := b12Creature(g, me.ID, "Heir", "Creature — Soldier", 1, 1)
	hsCounters(t, g, chosen, map[string]int{game.CounterPlusOne: 3})
	hsCast(t, g, chosen)
	hand := len(me.Hand.Cards)

	p1Destroy(t, g, chosen)
	g.RunStateChecksForTest()
	hsDrain(t, g, nil)
	if n := layeredCard(t, g, heir).Counters[game.CounterPlusOne]; n != 0 {
		t.Errorf("heir got %d +1/+1 counters with no target chosen", n)
	}
	if got := len(me.Hand.Cards) - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

// The counters go through the counter replacements as a new placement
// (CR 614.16): Doubling Season doubles them.
func TestHeroicSacrificeCountersMeetDoublingSeason(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	chosen := b12Creature(g, me.ID, "Chosen", "Creature — Soldier", 1, 1)
	heir := b12Creature(g, me.ID, "Heir", "Creature — Soldier", 1, 1)
	hsCounters(t, g, chosen, map[string]int{game.CounterPlusOne: 2})
	hsCast(t, g, chosen)
	seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)

	p1Destroy(t, g, chosen)
	g.RunStateChecksForTest()
	hsDrain(t, g, pr6Card(heir))
	if n := layeredCard(t, g, heir).Counters[game.CounterPlusOne]; n != 4 {
		t.Errorf("heir has %d +1/+1 counters, want 4", n)
	}
}

// A creature that left and came back is a new object (CR 400.7): its
// death does not fire the trigger, and it is gone at end of turn.
func TestHeroicSacrificeFlickeredCreatureIsANewObject(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	chosen := b12Creature(g, me.ID, "Chosen", "Creature — Soldier", 1, 1)
	b12Creature(g, me.ID, "Heir", "Creature — Soldier", 1, 1)
	hsCast(t, g, chosen)
	flickerForTest(t, g, me.ID, chosen)
	hand := len(me.Hand.Cards)

	p1Destroy(t, g, crewTestIDByName(t, g, "Chosen"))
	g.RunStateChecksForTest()
	if picks := hsDrain(t, g, nil); picks != 0 {
		t.Errorf("%d target picks after the new object died, want 0", picks)
	}
	if got := len(me.Hand.Cards) - hand; got != 0 {
		t.Errorf("drew %d cards, want 0", got)
	}
	advanceToUpkeepOf(t, g, 1)
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("%d delayed triggers survived the turn, want 0", len(g.DelayedTriggers))
	}
}

// The condition is about the object: the creature's death to a graveyard
// with the epoch it was created about, and nothing else (CR 603.7c, 700.4).
func TestThatCreatureDiedKeysOnTheObject(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	c := b12Creature(g, me.ID, "Chosen", "Creature — Soldier", 1, 1)
	var ref game.ObjectRef
	g.WithWriteLock(func() { ref, _ = g.PermanentRefForEffect(c) })
	p1Destroy(t, g, c)
	died := game.Event{Kind: game.EventLTB, CardID: c, OldZone: game.ZoneBattlefield, NewZone: game.ZoneGraveyard}
	exiled := died
	exiled.NewZone = game.ZoneExile
	stale := ref
	stale.Epoch++
	other := ref
	other.ID = uuid.New()
	g.WithWriteLock(func() {
		if got := thatCreatureDied(died, nil, g, game.EffectParams{Object: ref}); len(got) != 1 || got[0] != c {
			t.Errorf("its death matched %v, want [%v]", got, c)
		}
		if got := thatCreatureDied(exiled, nil, g, game.EffectParams{Object: ref}); len(got) != 0 {
			t.Errorf("an exile matched %v, want nothing", got)
		}
		if got := thatCreatureDied(died, nil, g, game.EffectParams{Object: stale}); len(got) != 0 {
			t.Errorf("another epoch matched %v, want nothing", got)
		}
		if got := thatCreatureDied(died, nil, g, game.EffectParams{Object: other}); len(got) != 0 {
			t.Errorf("another creature matched %v, want nothing", got)
		}
	})
}

// Fatal Fissure: when the opponent's creature dies, you earthbend 4 a land
// you control, chosen as the trigger goes on the stack.
func TestFatalFissureEarthbendsWhenThatCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	bystander := b12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 2)
	land := aangPushLand(g, me.ID, "Forest", false)
	castCatalogSpell(t, g, "Fatal Fissure", "Instant", hsFatalFissure, pr6Card(victim))
	passPriorityAroundTable(t, g)

	p1Destroy(t, g, bystander)
	g.RunStateChecksForTest()
	if picks := hsDrain(t, g, pr6Card(land)); picks != 0 {
		t.Fatalf("another creature's death fired it (%d picks)", picks)
	}
	p1Destroy(t, g, victim)
	g.RunStateChecksForTest()
	if picks := hsDrain(t, g, pr6Card(land)); picks != 1 {
		t.Fatalf("%d target picks, want 1", picks)
	}
	c := layeredCard(t, g, land)
	if !c.IsCreature() || c.Counters[game.CounterPlusOne] != 4 {
		t.Errorf("land is creature=%v with %d +1/+1 counters, want an earthbent 4", c.IsCreature(), c.Counters[game.CounterPlusOne])
	}
}

// With no land to choose, the trigger is removed (CR 603.3d).
func TestFatalFissureWithNoLandDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := cmcSeats(t, g)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Fatal Fissure", "Instant", hsFatalFissure, pr6Card(victim))
	passPriorityAroundTable(t, g)
	p1Destroy(t, g, victim)
	g.RunStateChecksForTest()
	if picks := hsDrain(t, g, nil); picks != 0 {
		t.Errorf("%d target picks with no land, want 0", picks)
	}
	if len(g.StackMeta) != 0 {
		t.Errorf("%d stack items left, want 0", len(g.StackMeta))
	}
}
