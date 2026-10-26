package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stickers_sheets3_test.go — the ability stickers that needed a seam of
// their own (ADR 0208, #4307): protection from a new quality, hellbent in a
// bundle, a divided sacrifice, a pairwise "can't block", heroic.

// stkProtectedFrom asks the engine's one protection predicate: does the
// host have protection from the source?
func stkProtectedFrom(s *stkScene, source uuid.UUID) bool {
	var protected bool
	s.g.WithWriteLock(func() {
		s.g.RecomputeLayersIfStaleLocked()
		host, _ := s.g.LookupCardForEffect(s.host)
		src, ok := s.g.LookupCardForEffect(source)
		if !ok {
			return
		}
		protected = game.ProtectedFrom(&host, game.SourceCharacteristics(&src))
	})
	return protected
}

// Notorious Sliver War.
func TestStkNotoriousSliverWarPumpsYourCreatures(t *testing.T) {
	s := newStkScene(t, "Notorious Sliver War", "Creatures you control get")
	other := p2Push(s.g, s.me.ID, "Other", "Creature — Bear", "", 1, 1)
	theirs := p2Push(s.g, s.opp.ID, "Theirs", "Creature — Bear", "", 1, 1)
	s.activate(0, "RRRRR")
	if p, _ := p2PT(s.g, other); p != 2 {
		t.Errorf("my other creature's power = %d, want 2", p)
	}
	if p, _ := p2PT(s.g, theirs); p != 1 {
		t.Errorf("an opponent's creature got the pump: power %d", p)
	}
}

func TestStkNotoriousSliverWarProtectsFromMultitypedCreatures(t *testing.T) {
	s := newStkScene(t, "Notorious Sliver War", "Protection from creatures")
	one := p2Push(s.g, s.opp.ID, "One", "Creature — Elf", "", 1, 1)
	two := p2Push(s.g, s.opp.ID, "Two", "Creature — Elf Warrior", "", 1, 1)
	if stkProtectedFrom(s, one) {
		t.Error("protected from a creature with one type")
	}
	if !stkProtectedFrom(s, two) {
		t.Error("not protected from a creature with two types")
	}
	// It stops the two-type creature from blocking the host.
	var r game.BlockRefusal
	s.g.ReadSnapshot(func() {
		host, _ := s.g.LookupCardForEffect(s.host)
		blocker, _ := s.g.LookupCardForEffect(two)
		r = s.g.BlockPairRefusalLocked(&host, &blocker)
	})
	if r.Reason != game.BlockReasonProtection {
		t.Errorf("block refusal = %q, want protection", r.Reason)
	}
}

// Weird Angel Flame.
func TestStkWeirdAngelFlameHeroic(t *testing.T) {
	s := newStkScene(t, "Weird Angel Flame", "Heroic")
	castCatalogSpell(t, s.g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{p2Card(s.host)})
	s.settle()
	if got := countersOn(s.g, s.host, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	// A spell that targets something else does not trigger it.
	other := p2Push(s.g, s.me.ID, "Other", "Creature — Bear", "", 1, 5)
	castCatalogSpell(t, s.g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{p2Card(other)})
	s.settle()
	if got := countersOn(s.g, s.host, game.CounterPlusOne); got != 2 {
		t.Errorf("a spell aimed elsewhere changed the counters: %d", got)
	}
}

func TestStkWeirdAngelFlameProtectsFromEvenManaValues(t *testing.T) {
	s := newStkScene(t, "Weird Angel Flame", "Protection from even")
	even := pushMVCreature2(s.g, s.opp, "Even", "{1}{G}")
	odd := pushMVCreature2(s.g, s.opp, "Odd", "{G}")
	if !stkProtectedFrom(s, even) {
		t.Error("not protected from a mana value 2 source")
	}
	if stkProtectedFrom(s, odd) {
		t.Error("protected from a mana value 1 source")
	}
}

func pushMVCreature2(g *game.Game, owner *game.Player, name, cost string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test", ManaCost: cost,
		Power: 1, Toughness: 1, Owner: owner.ID, Controller: owner.ID,
	})
}

// Wild Ogre Bupkis.
func TestStkWildOgreBupkisAttackGrows(t *testing.T) {
	s := newStkScene(t, "Wild Ogre Bupkis", "put a +1/+1 counter")
	s.attack()
	if got := countersOn(s.g, s.host, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
}

func TestStkWildOgreBupkisMetalcraft(t *testing.T) {
	s := newStkScene(t, "Wild Ogre Bupkis", "Metalcraft")
	rock := p2Push(s.g, s.opp.ID, "Their Rock", "Artifact", "", 0, 0)
	p2Push(s.g, s.me.ID, "Rock 1", "Artifact", "", 0, 0)
	p2Push(s.g, s.me.ID, "Rock 2", "Artifact", "", 0, 0)
	if stkProtectedFrom(s, rock) {
		t.Fatal("protected with two artifacts")
	}
	third := p2Push(s.g, s.me.ID, "Rock 3", "Artifact", "", 0, 0)
	if !stkProtectedFrom(s, rock) {
		t.Error("not protected from a noncreature permanent with three artifacts")
	}
	creature := p2Push(s.g, s.opp.ID, "Their Bear", "Creature — Bear", "", 1, 1)
	if stkProtectedFrom(s, creature) {
		t.Error("protected from a creature")
	}
	b25Destroy(s.g, third)
	if stkProtectedFrom(s, rock) {
		t.Error("still protected after an artifact left")
	}
}

// Vampire Champion Fury.
func TestStkVampireChampionFuryHellbent(t *testing.T) {
	s := newStkScene(t, "Vampire Champion Fury", "Hellbent")
	pushHandCard(s.g, s.me)
	if p, tough := p2PT(s.g, s.host); p != 2 || tough != 2 {
		t.Fatalf("with a card in hand: %d/%d, want 2/2", p, tough)
	}
	// Emptying the hand re-reads the static at once.
	s.g.WithWriteLock(func() {
		for s.me.Hand.Size() > 0 {
			c, err := s.me.Hand.PopTop()
			if err != nil {
				t.Fatalf("pop: %v", err)
			}
			s.me.Graveyard.PushTop(c)
			s.g.EmitEvent(game.Event{Kind: game.EventDiscardCard, CardID: c.InstanceID, OldZone: game.ZoneHand, NewZone: game.ZoneGraveyard})
		}
	})
	if p, tough := p2PT(s.g, s.host); p != 5 || tough != 5 {
		t.Errorf("with an empty hand: %d/%d, want 5/5", p, tough)
	}
	pushHandCard(s.g, s.me)
	s.g.WithWriteLock(func() {
		s.g.EmitEvent(game.Event{Kind: game.EventDrawCard, Actor: s.me.ID, OldZone: game.ZoneLibrary, NewZone: game.ZoneHand})
	})
	if p, _ := p2PT(s.g, s.host); p != 2 {
		t.Errorf("a card back in hand: power %d, want 2", p)
	}
}

func TestStkVampireChampionFurySacrificesForDividedDamage(t *testing.T) {
	s := newStkSceneOn(t, "Vampire Champion Fury", "divided as you choose", "Creature — Vampire", 4, 4)
	a := p2Push(s.g, s.opp.ID, "Bear A", "Creature — Bear", "", 3, 3)
	b := p2Push(s.g, s.opp.ID, "Bear B", "Creature — Bear", "", 3, 3)
	floatForTest(s.g, s.me, "RR")
	if err := s.g.ActivateCatalogAbility(s.me.ID, s.host, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{p2Card(a), p2Card(b)},
		Distribution: map[uuid.UUID]int{a: 3, b: 1},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if onBattlefield(s.g, s.host) {
		t.Error("the creature was not sacrificed as the cost")
	}
	s.settle()
	if onBattlefield(s.g, a) {
		t.Error("Bear A took 3 and survived")
	}
	if !onBattlefield(s.g, b) {
		t.Error("Bear B took 1 and died")
	}
}

// A division that does not add up to the creature's power (4) is refused.
func TestStkVampireChampionFuryRefusesADivisionThatIsNotItsPower(t *testing.T) {
	s := newStkSceneOn(t, "Vampire Champion Fury", "divided as you choose", "Creature — Vampire", 4, 4)
	a := p2Push(s.g, s.opp.ID, "Bear A", "Creature — Bear", "", 3, 3)
	floatForTest(s.g, s.me, "RR")
	if err := s.g.ActivateCatalogAbility(s.me.ID, s.host, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{p2Card(a)},
		Distribution: map[uuid.UUID]int{a: 3},
	}); err == nil {
		t.Error("a division of 3 for a 4-power creature was accepted")
	}
}

// Wrinkly Monkey Shenanigans.
func TestStkWrinklyMonkeyStopsATargetBlocking(t *testing.T) {
	s := newStkScene(t, "Wrinkly Monkey Shenanigans", "can't block this creature")
	a := b12Creature(s.g, s.opp.ID, "Wall A", "Creature — Wall", 0, 4)
	b := b12Creature(s.g, s.opp.ID, "Wall B", "Creature — Wall", 0, 4)
	s.activate(0, "R", p2Card(a))
	advanceTo(t, s.g, game.StepDeclareAttackers)
	if err := s.g.DeclareAttacker(s.host, s.opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, s.g)
	advanceTo(t, s.g, game.StepDeclareBlockers)
	if err := s.g.DeclareBlocker(a, s.host); err == nil {
		t.Error("the barred creature blocked")
	}
	if err := s.g.DeclareBlocker(b, s.host); err != nil {
		t.Errorf("another creature was barred too: %v", err)
	}
}

func TestStkWrinklyMonkeyMorbid(t *testing.T) {
	s := newStkScene(t, "Wrinkly Monkey Shenanigans", "Morbid")
	seedLibrary(s.me, "A", "B")
	// No creature died: nothing happens at the end step.
	hand := s.me.Hand.Size()
	advanceToEndStepOf(t, s.g, 0)
	s.settle()
	if s.me.Hand.Size() != hand || countersOn(s.g, s.host, game.CounterMinusOne) != 0 {
		t.Fatal("morbid triggered with no creature having died")
	}
}

func TestStkWrinklyMonkeyMorbidAfterADeath(t *testing.T) {
	s := newStkScene(t, "Wrinkly Monkey Shenanigans", "Morbid")
	seedLibrary(s.me, "A", "B")
	victim := p2Push(s.g, s.opp.ID, "Victim", "Creature — Bear", "", 1, 1)
	b25Destroy(s.g, victim)
	hand := s.me.Hand.Size()
	advanceToEndStepOf(t, s.g, 0)
	s.settle()
	if got := countersOn(s.g, s.host, game.CounterMinusOne); got != 1 {
		t.Errorf("-1/-1 counters = %d, want 1", got)
	}
	if got := s.me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}
