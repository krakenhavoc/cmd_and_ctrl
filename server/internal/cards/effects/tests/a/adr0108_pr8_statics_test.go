package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gloom Surgeon: combat damage to it is prevented and exiles that many
// from the top of your library; noncombat damage is not its business; and
// combat damage that can't be prevented (Frenzied Baloth) is dealt and
// still exiles that many (the ruling).
func TestADR0108PR8StaticsGloomSurgeon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	surgeon := effectstest.Pr8StaticsCreature(g, me.ID, "Gloom Surgeon", effectstest.Pr8GloomSurgeon, 2, 3)
	attacker := effectstest.Pr7Creature(g, opp.ID, "Attacker", 2, "R")
	lib := me.Library.Size()
	if err := g.MarkCombatDamage(attacker, surgeon, 2); err != nil {
		t.Fatal(err)
	}
	g.RunStateChecksForTest()
	if effectstest.DamageMarkedOn(g, surgeon) != 0 || me.Library.Size() != lib-2 || g.Exile.Size() != 2 {
		t.Fatalf("combat 2: damage %d (want 0), library %d (want %d), exile %d (want 2)",
			effectstest.DamageMarkedOn(g, surgeon), me.Library.Size(), lib-2, g.Exile.Size())
	}
	effectstest.Pr6Damage(t, g, attacker, surgeon, 1)
	if effectstest.DamageMarkedOn(g, surgeon) != 1 || me.Library.Size() != lib-2 {
		t.Fatalf("noncombat 1: damage %d (want 1), library %d (want %d)", effectstest.DamageMarkedOn(g, surgeon), me.Library.Size(), lib-2)
	}
	effectstest.PushCatalogPermanent(g, opp.ID, "Frenzied Baloth", "Creature — Beast", effectstest.Pr6BalothOracle, false)
	if err := g.MarkCombatDamage(attacker, surgeon, 1); err != nil {
		t.Fatal(err)
	}
	g.RunStateChecksForTest()
	if effectstest.DamageMarkedOn(g, surgeon) != 2 || me.Library.Size() != lib-3 {
		t.Fatalf("unpreventable combat 1: damage %d (want 2), library %d (want %d)",
			effectstest.DamageMarkedOn(g, surgeon), me.Library.Size(), lib-3)
	}
}

// Angel of Suffering: two sources at once are one application to you,
// milling twice the total; unpreventable damage is dealt and still mills
// twice that many (the ruling).
func TestADR0108PR8StaticsAngelOfSuffering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Pr8StaticsCreature(g, me.ID, "Angel of Suffering", effectstest.Pr8AngelOfSuffering, 5, 3)
	a := effectstest.Pr7Creature(g, opp.ID, "A", 2, "R")
	b := effectstest.Pr7Creature(g, opp.ID, "B", 1, "G")
	life, lib := me.Life, me.Library.Size()
	effectstest.Pr8Hit(t, g, me.ID, map[uuid.UUID]int{a: 2, b: 1})
	if me.Life != life || me.Library.Size() != lib-6 {
		t.Fatalf("life %d (want %d), library %d (want %d)", me.Life, life, me.Library.Size(), lib-6)
	}
	effectstest.Pr8Unpreventable(t, g, a, me.ID, 1)
	if me.Life != life-1 || me.Library.Size() != lib-8 {
		t.Fatalf("unpreventable: life %d (want %d), library %d (want %d)", me.Life, life-1, me.Library.Size(), lib-8)
	}
}

// Immortal Coil: the controller picks which cards when there is a choice;
// a count that covers the graveyard exiles all of it with no prompt, and
// then the empty graveyard loses the game; unpreventable damage exiles
// nothing (prevented this way is zero).
func TestADR0108PR8StaticsImmortalCoil(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Immortal Coil", "Artifact", effectstest.Pr8ImmortalCoil, false)
	var gy []uuid.UUID
	for _, n := range []string{"One", "Two", "Three", "Four"} {
		gy = append(gy, effectstest.PushGraveyardCardForTest(me, n))
	}
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 2, "R")
	life := me.Life
	effectstest.Pr8Unpreventable(t, g, src, me.ID, 1)
	if me.Life != life-1 || me.Graveyard.Size() != 4 {
		t.Fatalf("unpreventable: life %d (want %d), graveyard %d (want 4)", me.Life, life-1, me.Graveyard.Size())
	}
	effectstest.Pr7Hit(t, g, src, me.ID, 2)
	if me.Life != life-1 {
		t.Fatalf("life %d, want %d: the 2 is prevented", me.Life, life-1)
	}
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil || pick.ChooseMin != 2 || pick.ChooseMax != 2 {
		t.Fatalf("want a choose-two prompt, got %+v", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{gy[1], gy[3]}); err != nil {
		t.Fatal(err)
	}
	if me.Graveyard.Size() != 2 || !me.Graveyard.Contains(gy[0]) || !me.Graveyard.Contains(gy[2]) {
		t.Fatalf("graveyard after the choice: %d cards, want One and Three", me.Graveyard.Size())
	}
	effectstest.Pr7Hit(t, g, src, me.ID, 5)
	if me.Life != life-1 || me.Graveyard.Size() != 0 {
		t.Fatalf("5 against a 2-card graveyard: life %d (want %d), graveyard %d (want 0)", me.Life, life-1, me.Graveyard.Size())
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Eliminated {
		t.Fatal("no cards in the graveyard: Immortal Coil's controller is still in the game")
	}
}

// Swans of Bryn Argoll: one application per SOURCE — each source's
// controller draws its own source's damage; unpreventable damage draws
// nothing.
func TestADR0108PR8StaticsSwans(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	swans := effectstest.Pr8StaticsCreature(g, me.ID, "Swans of Bryn Argoll", effectstest.Pr8Swans, 4, 3)
	a := effectstest.Pr7Creature(g, opp.ID, "A", 2, "R")
	b := effectstest.Pr7Creature(g, opp.ID, "B", 3, "G")
	mine := effectstest.Pr7Creature(g, me.ID, "Mine", 1, "W")
	myHand, theirHand := me.Hand.Size(), opp.Hand.Size()
	effectstest.Pr8Hit(t, g, swans, map[uuid.UUID]int{a: 2, b: 3, mine: 1})
	if effectstest.DamageMarkedOn(g, swans) != 0 || opp.Hand.Size() != theirHand+5 || me.Hand.Size() != myHand+1 {
		t.Fatalf("damage %d (want 0), opponent drew %d (want 5), I drew %d (want 1)",
			effectstest.DamageMarkedOn(g, swans), opp.Hand.Size()-theirHand, me.Hand.Size()-myHand)
	}
	effectstest.Pr8Unpreventable(t, g, a, swans, 1)
	if effectstest.DamageMarkedOn(g, swans) != 1 || opp.Hand.Size() != theirHand+5 {
		t.Fatalf("unpreventable: damage %d (want 1), opponent drew %d more (want 0)",
			effectstest.DamageMarkedOn(g, swans), opp.Hand.Size()-theirHand-5)
	}
}

// Phyrexian Vindicator: the prevented damage comes back as a reflexive
// trigger at any OTHER target; unpreventable damage is dealt and triggers
// nothing.
func TestADR0108PR8StaticsPhyrexianVindicator(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vindicator := effectstest.Pr8StaticsCreature(g, me.ID, "Phyrexian Vindicator", effectstest.Pr8Vindicator, 5, 5)
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
	effectstest.Pr6Damage(t, g, src, vindicator, 3)
	if effectstest.DamageMarkedOn(g, vindicator) != 0 {
		t.Fatalf("damage %d, want 0", effectstest.DamageMarkedOn(g, vindicator))
	}
	pick := effectstest.PassUntilPickTarget(t, g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the reflexive trigger")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: vindicator}); err == nil {
		t.Fatal("the Vindicator was accepted as its own target: it is \"any other target\"")
	}
	life := opp.Life
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Fatalf("opponent life %d, want %d", opp.Life, life-3)
	}
	effectstest.Pr8Unpreventable(t, g, src, vindicator, 2)
	if effectstest.DamageMarkedOn(g, vindicator) != 2 || effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID) != nil {
		t.Fatalf("unpreventable: damage %d (want 2), and no trigger", effectstest.DamageMarkedOn(g, vindicator))
	}
}

// The Mindskinner: a source you control dealing damage to an opponent is
// prevented and each opponent mills that many — once per source; damage
// to you is not its business; unpreventable damage still mills; and two
// Mindskinners mill once (the ruling).
func TestADR0108PR8StaticsMindskinner(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	skinner := effectstest.Pr8StaticsCreature(g, me.ID, "The Mindskinner", effectstest.Pr8Mindskinner, 10, 1)
	mine := effectstest.Pr7Creature(g, me.ID, "Mine", 2, "U")
	theirs := effectstest.Pr7Creature(g, opp.ID, "Theirs", 2, "R")
	life, lib, lib3 := opp.Life, opp.Library.Size(), third.Library.Size()
	effectstest.Pr8Hit(t, g, opp.ID, map[uuid.UUID]int{skinner: 3, mine: 2})
	if opp.Life != life || opp.Library.Size() != lib-5 || third.Library.Size() != lib3-5 {
		t.Fatalf("life %d (want %d), opponent milled %d (want 5), third player milled %d (want 5)",
			opp.Life, life, lib-opp.Library.Size(), lib3-third.Library.Size())
	}
	mineLife := me.Life
	effectstest.Pr8Hit(t, g, me.ID, map[uuid.UUID]int{theirs: 2})
	if me.Life != mineLife-2 {
		t.Fatalf("damage to me: life %d, want %d", me.Life, mineLife-2)
	}
	effectstest.Pr8Unpreventable(t, g, mine, opp.ID, 1)
	if opp.Life != life-1 || opp.Library.Size() != lib-6 {
		t.Fatalf("unpreventable: life %d (want %d), milled %d (want 6)", opp.Life, life-1, lib-opp.Library.Size())
	}
	effectstest.Pr8StaticsCreature(g, me.ID, "The Mindskinner", effectstest.Pr8Mindskinner, 10, 1)
	effectstest.Pr8Hit(t, g, opp.ID, map[uuid.UUID]int{mine: 1})
	if opp.Library.Size() != lib-7 {
		t.Fatalf("two Mindskinners: milled %d in all, want 7 (once, not twice)", lib-opp.Library.Size())
	}
}
