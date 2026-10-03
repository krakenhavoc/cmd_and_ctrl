package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0108_pr8_statics_test.go — ADR 0108 PR 8 (#1906): the prevention
// statics whose additional effect mills, draws, exiles or strikes back.
// Each test proves the prevention, the additional effect's unit (per
// recipient or per source, in one damage instance) and what CR 615.12
// does to it, as the card's rulings say.

const (
	pr8GloomSurgeon     = "48b8ab2d-9f3f-42b2-9c5d-7497bceae43e"
	pr8AngelOfSuffering = "92380c57-0a92-48eb-b562-2be149b5792a"
	pr8ImmortalCoil     = "c85fc672-12d8-4ab0-b18d-ccd590ac063c"
	pr8Swans            = "b187aeeb-5cf5-4b73-a3ef-f39188d2ba33"
	pr8Vindicator       = "9b5cfbb7-21ed-491d-b77d-547e4d30a7da"
	pr8Mindskinner      = "2e3ec514-fc9e-4964-a493-2a6cc9e63630"
)

// pr8StaticsCreature puts a catalogued creature with real power and
// toughness onto the battlefield, sick-free.
func pr8StaticsCreature(g *game.Game, owner uuid.UUID, name, oracle string, power, toughness int) uuid.UUID {
	return apaPush(g, owner, owner, game.Card{Name: name, TypeLine: "Creature — Test", OracleID: oracle,
		Power: power, Toughness: toughness})
}

// Gloom Surgeon: combat damage to it is prevented and exiles that many
// from the top of your library; noncombat damage is not its business; and
// combat damage that can't be prevented (Frenzied Baloth) is dealt and
// still exiles that many (the ruling).
func TestADR0108PR8StaticsGloomSurgeon(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	surgeon := pr8StaticsCreature(g, me.ID, "Gloom Surgeon", pr8GloomSurgeon, 2, 3)
	attacker := pr7Creature(g, opp.ID, "Attacker", 2, "R")
	lib := me.Library.Size()
	if err := g.MarkCombatDamage(attacker, surgeon, 2); err != nil {
		t.Fatal(err)
	}
	g.RunStateChecksForTest()
	if damageMarkedOn(g, surgeon) != 0 || me.Library.Size() != lib-2 || g.Exile.Size() != 2 {
		t.Fatalf("combat 2: damage %d (want 0), library %d (want %d), exile %d (want 2)",
			damageMarkedOn(g, surgeon), me.Library.Size(), lib-2, g.Exile.Size())
	}
	pr6Damage(t, g, attacker, surgeon, 1)
	if damageMarkedOn(g, surgeon) != 1 || me.Library.Size() != lib-2 {
		t.Fatalf("noncombat 1: damage %d (want 1), library %d (want %d)", damageMarkedOn(g, surgeon), me.Library.Size(), lib-2)
	}
	pushCatalogPermanent(g, opp.ID, "Frenzied Baloth", "Creature — Beast", pr6BalothOracle, false)
	if err := g.MarkCombatDamage(attacker, surgeon, 1); err != nil {
		t.Fatal(err)
	}
	g.RunStateChecksForTest()
	if damageMarkedOn(g, surgeon) != 2 || me.Library.Size() != lib-3 {
		t.Fatalf("unpreventable combat 1: damage %d (want 2), library %d (want %d)",
			damageMarkedOn(g, surgeon), me.Library.Size(), lib-3)
	}
}

// Angel of Suffering: two sources at once are one application to you,
// milling twice the total; unpreventable damage is dealt and still mills
// twice that many (the ruling).
func TestADR0108PR8StaticsAngelOfSuffering(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pr8StaticsCreature(g, me.ID, "Angel of Suffering", pr8AngelOfSuffering, 5, 3)
	a := pr7Creature(g, opp.ID, "A", 2, "R")
	b := pr7Creature(g, opp.ID, "B", 1, "G")
	life, lib := me.Life, me.Library.Size()
	pr8Hit(t, g, me.ID, map[uuid.UUID]int{a: 2, b: 1})
	if me.Life != life || me.Library.Size() != lib-6 {
		t.Fatalf("life %d (want %d), library %d (want %d)", me.Life, life, me.Library.Size(), lib-6)
	}
	pr8Unpreventable(t, g, a, me.ID, 1)
	if me.Life != life-1 || me.Library.Size() != lib-8 {
		t.Fatalf("unpreventable: life %d (want %d), library %d (want %d)", me.Life, life-1, me.Library.Size(), lib-8)
	}
}

// Immortal Coil: the controller picks which cards when there is a choice;
// a count that covers the graveyard exiles all of it with no prompt, and
// then the empty graveyard loses the game; unpreventable damage exiles
// nothing (prevented this way is zero).
func TestADR0108PR8StaticsImmortalCoil(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Immortal Coil", "Artifact", pr8ImmortalCoil, false)
	var gy []uuid.UUID
	for _, n := range []string{"One", "Two", "Three", "Four"} {
		gy = append(gy, pushGraveyardCardForTest(me, n))
	}
	src := pr7Creature(g, opp.ID, "Src", 2, "R")
	life := me.Life
	pr8Unpreventable(t, g, src, me.ID, 1)
	if me.Life != life-1 || me.Graveyard.Size() != 4 {
		t.Fatalf("unpreventable: life %d (want %d), graveyard %d (want 4)", me.Life, life-1, me.Graveyard.Size())
	}
	pr7Hit(t, g, src, me.ID, 2)
	if me.Life != life-1 {
		t.Fatalf("life %d, want %d: the 2 is prevented", me.Life, life-1)
	}
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil || pick.ChooseMin != 2 || pick.ChooseMax != 2 {
		t.Fatalf("want a choose-two prompt, got %+v", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{gy[1], gy[3]}); err != nil {
		t.Fatal(err)
	}
	if me.Graveyard.Size() != 2 || !me.Graveyard.Contains(gy[0]) || !me.Graveyard.Contains(gy[2]) {
		t.Fatalf("graveyard after the choice: %d cards, want One and Three", me.Graveyard.Size())
	}
	pr7Hit(t, g, src, me.ID, 5)
	if me.Life != life-1 || me.Graveyard.Size() != 0 {
		t.Fatalf("5 against a 2-card graveyard: life %d (want %d), graveyard %d (want 0)", me.Life, life-1, me.Graveyard.Size())
	}
	passPriorityAroundTable(t, g)
	if !me.Eliminated {
		t.Fatal("no cards in the graveyard: Immortal Coil's controller is still in the game")
	}
}

// Swans of Bryn Argoll: one application per SOURCE — each source's
// controller draws its own source's damage; unpreventable damage draws
// nothing.
func TestADR0108PR8StaticsSwans(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	swans := pr8StaticsCreature(g, me.ID, "Swans of Bryn Argoll", pr8Swans, 4, 3)
	a := pr7Creature(g, opp.ID, "A", 2, "R")
	b := pr7Creature(g, opp.ID, "B", 3, "G")
	mine := pr7Creature(g, me.ID, "Mine", 1, "W")
	myHand, theirHand := me.Hand.Size(), opp.Hand.Size()
	pr8Hit(t, g, swans, map[uuid.UUID]int{a: 2, b: 3, mine: 1})
	if damageMarkedOn(g, swans) != 0 || opp.Hand.Size() != theirHand+5 || me.Hand.Size() != myHand+1 {
		t.Fatalf("damage %d (want 0), opponent drew %d (want 5), I drew %d (want 1)",
			damageMarkedOn(g, swans), opp.Hand.Size()-theirHand, me.Hand.Size()-myHand)
	}
	pr8Unpreventable(t, g, a, swans, 1)
	if damageMarkedOn(g, swans) != 1 || opp.Hand.Size() != theirHand+5 {
		t.Fatalf("unpreventable: damage %d (want 1), opponent drew %d more (want 0)",
			damageMarkedOn(g, swans), opp.Hand.Size()-theirHand-5)
	}
}

// Phyrexian Vindicator: the prevented damage comes back as a reflexive
// trigger at any OTHER target; unpreventable damage is dealt and triggers
// nothing.
func TestADR0108PR8StaticsPhyrexianVindicator(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vindicator := pr8StaticsCreature(g, me.ID, "Phyrexian Vindicator", pr8Vindicator, 5, 5)
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	pr6Damage(t, g, src, vindicator, 3)
	if damageMarkedOn(g, vindicator) != 0 {
		t.Fatalf("damage %d, want 0", damageMarkedOn(g, vindicator))
	}
	pick := passUntilPickTarget(t, g, me.ID)
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
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Fatalf("opponent life %d, want %d", opp.Life, life-3)
	}
	pr8Unpreventable(t, g, src, vindicator, 2)
	if damageMarkedOn(g, vindicator) != 2 || latestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID) != nil {
		t.Fatalf("unpreventable: damage %d (want 2), and no trigger", damageMarkedOn(g, vindicator))
	}
}

// The Mindskinner: a source you control dealing damage to an opponent is
// prevented and each opponent mills that many — once per source; damage
// to you is not its business; unpreventable damage still mills; and two
// Mindskinners mill once (the ruling).
func TestADR0108PR8StaticsMindskinner(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	skinner := pr8StaticsCreature(g, me.ID, "The Mindskinner", pr8Mindskinner, 10, 1)
	mine := pr7Creature(g, me.ID, "Mine", 2, "U")
	theirs := pr7Creature(g, opp.ID, "Theirs", 2, "R")
	life, lib, lib3 := opp.Life, opp.Library.Size(), third.Library.Size()
	pr8Hit(t, g, opp.ID, map[uuid.UUID]int{skinner: 3, mine: 2})
	if opp.Life != life || opp.Library.Size() != lib-5 || third.Library.Size() != lib3-5 {
		t.Fatalf("life %d (want %d), opponent milled %d (want 5), third player milled %d (want 5)",
			opp.Life, life, lib-opp.Library.Size(), lib3-third.Library.Size())
	}
	mineLife := me.Life
	pr8Hit(t, g, me.ID, map[uuid.UUID]int{theirs: 2})
	if me.Life != mineLife-2 {
		t.Fatalf("damage to me: life %d, want %d", me.Life, mineLife-2)
	}
	pr8Unpreventable(t, g, mine, opp.ID, 1)
	if opp.Life != life-1 || opp.Library.Size() != lib-6 {
		t.Fatalf("unpreventable: life %d (want %d), milled %d (want 6)", opp.Life, life-1, lib-opp.Library.Size())
	}
	pr8StaticsCreature(g, me.ID, "The Mindskinner", pr8Mindskinner, 10, 1)
	pr8Hit(t, g, opp.ID, map[uuid.UUID]int{mine: 1})
	if opp.Library.Size() != lib-7 {
		t.Fatalf("two Mindskinners: milled %d in all, want 7 (once, not twice)", lib-opp.Library.Size())
	}
}
