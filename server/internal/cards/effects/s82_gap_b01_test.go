package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB01JeweledAmuletOracle = "e2ad4b4d-d5fe-4a64-9b40-296c8fc8bcce"
	gapB01DragonsSmileOracle  = "e7611ef9-e5fc-45eb-bc3f-4358a27107fd"
	gapB01EverforgerOracle    = "b8354ac6-5285-43dd-9ba5-93064def2e96"
	gapB01UramonOracle        = "92eeedc6-a331-445f-af56-c91dac69412b"
	gapB01IstvanOracle        = "5e277de5-793a-4364-923a-a55b53c0fa85"
	gapB01KurokiOracle        = "ce70f16c-ea45-4f1d-aafb-34a5d05d31e7"
	gapB01OlindaOracle        = "6ec8a9b8-233f-4de5-8f4c-3d677fbf2f30"
	gapB01SahirOracle         = "e1a8f5cf-9b59-46c6-8e41-6e44225ec0f4"
	gapB01ValkoOracle         = "8023d419-8fd2-4b92-b6ea-f25fd67e6410"
	gapB01DavvolOracle        = "e4aca279-b144-4517-acef-d925f4f552e2"
)

func TestGapB01EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB01JeweledAmuletOracle: "Jeweled Amulet",
		gapB01DragonsSmileOracle:  "Dragon's Smile",
		gapB01EverforgerOracle:    "The Everforger",
		gapB01UramonOracle:        "Boss Uramon, Shadow's Reach",
		gapB01IstvanOracle:        "Istvan, Butcher of Eln",
		gapB01KurokiOracle:        "Kuroki, Thief of Talents",
		gapB01OlindaOracle:        "Olinda the Oblivious",
		gapB01SahirOracle:         "Sahir, Visitor in Darkness",
		gapB01ValkoOracle:         "Valko Indorian",
		gapB01DavvolOracle:        "Davvol, Evincar of Rath",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func gapB01Enter(t *testing.T, g *game.Game, seat int, name, typeLine string) uuid.UUID {
	t.Helper()
	id := b13PlayAs(t, g, seat, name, typeLine, "")
	passPriorityAroundTable(t, g)
	return id
}

// --- Valko Indorian ---------------------------------------------------

func TestGapB01ValkoGivesYourPhyrexiansMenaceAndLifelinkOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	valko := b12Push(g, me.ID, "Valko Indorian", "Legendary Creature — Human Wizard", gapB01ValkoOracle, 2, 4)
	mine := b12Creature(g, me.ID, "Rager", "Creature — Phyrexian Horror", 2, 2)
	human := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Phyrexian", "Creature — Phyrexian Horror", 2, 2)
	for _, kw := range []string{"menace", "lifelink"} {
		if !hasEffectiveKeyword(t, g, mine, kw) {
			t.Errorf("my Phyrexian lacks %s", kw)
		}
		if hasEffectiveKeyword(t, g, human, kw) || hasEffectiveKeyword(t, g, theirs, kw) || hasEffectiveKeyword(t, g, valko, kw) {
			t.Errorf("a non-Phyrexian or an opponent's creature has %s", kw)
		}
	}
}

// --- Davvol, Evincar of Rath ----------------------------------------

func TestGapB01DavvolLosesALifeAndAddsBlackForAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Davvol, Evincar of Rath", "Legendary Creature — Phyrexian Noble", gapB01DavvolOracle, 2, 2)
	before := me.Life
	gapB01Enter(t, g, 0, "Bear", "Creature — Bear")
	if me.Life != before-1 {
		t.Errorf("life %d, want %d", me.Life, before-1)
	}
	if poolCount(me, "B") != 1 {
		t.Errorf("pool holds %d {B}, want 1", poolCount(me, "B"))
	}
}

func TestGapB01DavvolIgnoresItselfAndOpponentsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Davvol, Evincar of Rath", "Legendary Creature — Phyrexian Noble", gapB01DavvolOracle, 2, 2)
	before := me.Life
	gapB01Enter(t, g, 1, "Their Bear", "Creature — Bear")
	if me.Life != before || poolCount(me, "B") != 0 {
		t.Errorf("an opponent's creature triggered Davvol: life %d, {B} %d", me.Life, poolCount(me, "B"))
	}
}

// --- Istvan, Butcher of Eln -------------------------------------------

func TestGapB01IstvanGrowsByTwoWhenAnOpponentsCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	istvan := b12Push(g, me.ID, "Istvan, Butcher of Eln", "Legendary Creature — Human", gapB01IstvanOracle, 1, 3)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
	passPriorityAroundTable(t, g)
	if n := counterCount(g, istvan, "+1/+1"); n != 0 {
		t.Fatalf("my own creature dying put %d counters on Istvan", n)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	passPriorityAroundTable(t, g)
	if n := counterCount(g, istvan, "+1/+1"); n != 2 {
		t.Errorf("Istvan has %d +1/+1 counters, want 2", n)
	}
}

func TestGapB01IstvanPreventsDamageFromCreaturesButNotOtherSources(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	istvan := b12Push(g, me.ID, "Istvan, Butcher of Eln", "Legendary Creature — Human", gapB01IstvanOracle, 1, 3)
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	bolt := b12Permanent(g, opp.ID, "Pinger", "Artifact")
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(ogre, istvan, 5); err != nil {
			t.Fatalf("damage from a creature: %v", err)
		}
	})
	if got := gapB01Damage(g, istvan); got != 0 {
		t.Errorf("a creature dealt %d damage to Istvan, want it all prevented", got)
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(bolt, istvan, 2); err != nil {
			t.Fatalf("damage from an artifact: %v", err)
		}
	})
	if got := gapB01Damage(g, istvan); got != 2 {
		t.Errorf("an artifact dealt %d damage to Istvan, want 2 (only creatures are prevented)", got)
	}
}

func gapB01Damage(g *game.Game, id uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n = c.DamageMarked
			}
		}
	})
	return n
}

func gapB01CountNamed(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.Controller == controller {
				n++
			}
		}
	})
	return n
}

// --- Boss Uramon, Shadow's Reach --------------------------------------

func TestGapB01UramonGivesAttackingNinjasAndRoguesDeathtouchAndMakesRatsOnDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	uramon := b12Push(g, me.ID, "Boss Uramon, Shadow's Reach", "Legendary Creature — Human Rogue", gapB01UramonOracle, 2, 2)
	ninja := b12Creature(g, me.ID, "Ninja", "Creature — Human Ninja", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if hasEffectiveKeyword(t, g, ninja, "deathtouch") {
		t.Fatal("a Ninja that is not attacking has deathtouch")
	}
	attackWith(t, g, opp.ID, uramon, ninja, bear)
	for _, id := range []uuid.UUID{uramon, ninja} {
		if !hasEffectiveKeyword(t, g, id, "deathtouch") {
			t.Errorf("an attacking Rogue or Ninja lacks deathtouch")
		}
	}
	if hasEffectiveKeyword(t, g, bear, "deathtouch") {
		t.Error("an attacking Bear has deathtouch")
	}
	passPriorityAroundTable(t, g)
	if n := gapB01CountNamed(g, me.ID, "Rat Rogue"); n != 1 {
		t.Errorf("%d Rat Rogue tokens, want 1: two Ninjas and Rogues dealt damage to one player together", n)
	}
}

func TestGapB01UramonMakesNothingWhenOnlyAnOtherTypeDealsDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Boss Uramon, Shadow's Reach", "Legendary Creature — Human Rogue", gapB01UramonOracle, 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	attackWith(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if n := gapB01CountNamed(g, me.ID, "Rat Rogue"); n != 0 {
		t.Errorf("a Bear's damage made %d Rat Rogues", n)
	}
}

func TestGapB01UramonMakesOneRatPerDamagedPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	uramon := b12Push(g, me.ID, "Boss Uramon, Shadow's Reach", "Legendary Creature — Human Rogue", gapB01UramonOracle, 2, 2)
	rogue := b12Creature(g, me.ID, "Rogue", "Creature — Human Rogue", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(uramon, opp.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(rogue, third.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if n := gapB01CountNamed(g, me.ID, "Rat Rogue"); n != 2 {
		t.Errorf("%d Rat Rogues, want 2 (one per damaged player)", n)
	}
}

// --- Olinda the Oblivious -------------------------------------------------

func TestGapB01OlindaMarksOneCreaturePerOpponentAndTheyLoseLifeAtUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	a, b := g.Seats[1], g.Seats[2]
	ca := b12Creature(g, a.ID, "A's Bear", "Creature — Bear", 2, 2)
	ca2 := b12Creature(g, a.ID, "A's Other Bear", "Creature — Bear", 2, 2)
	cb := b12Creature(g, b.ID, "B's Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Olinda the Oblivious", "Legendary Creature — Skunk Rogue", gapB01OlindaOracle, nil)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, ca)
	answerPickTarget(t, g, cb)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{ca, cb} {
		if n := counterCount(g, id, "odor"); n != 1 {
			t.Errorf("%s has %d odor counters, want 1", id, n)
		}
	}
	if n := counterCount(g, ca2, "odor"); n != 0 {
		t.Errorf("the second creature of the same opponent got %d odor counters", n)
	}
	// The creature's controller loses 2 life at their own upkeep.
	before := a.Life
	advanceToUpkeepOfSeat(t, g, 1)
	passPriorityAroundTable(t, g)
	if a.Life != before-2 {
		t.Errorf("A's life %d, want %d", a.Life, before-2)
	}
}

func TestGapB01OlindaOnlyOffersCreaturesWithoutAnOdorCounter(t *testing.T) {
	g := newCatalogGame(t)
	a := g.Seats[1]
	ca := b12Creature(g, a.ID, "A's Smelly Bear", "Creature — Bear", 2, 2)
	fresh := b12Creature(g, a.ID, "A's Fresh Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ca, "odor", 1) })
	castCatalogSpell(t, g, "Olinda the Oblivious", "Legendary Creature — Skunk Rogue", gapB01OlindaOracle, nil)
	passPriorityAroundTable(t, g)
	ch := pickTargetPrompt(t, g)
	if slices.Contains(ch.PickTargetCards, ca) || !slices.Contains(ch.PickTargetCards, fresh) {
		t.Fatalf("offered %v: only the creature without an odor counter should be", ch.PickTargetCards)
	}
}
