package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// the_ring_pool_c_test.go — ADR 0114 PR 5, the last pool.

const (
	boromirWardenOracle       = "d72b57e5-1c6e-4f02-968f-7e2dd24f0d3c"
	elrondLordOfRivendellOrac = "099ae8f2-5c84-43ab-aaa1-ea69f2bba784"
	samwiseStoutheartedOracle = "0e26b429-0469-474a-818a-2d4696c38a05"
	scrollOfIsildurOracle     = "aa9a45a5-0250-465b-962e-f2d45dea26f9"
	thereAndBackAgainOracle   = "de4e5120-b958-4dcb-ad78-d0f726b7a881"
	sauronNecromancerOracle   = "8dcd7e2b-75a9-476a-a0a0-798613e8dad6"
)

// ringAnswerPickIfAsked answers an open pick_target prompt with `id`,
// if one is open (a clause with one legal target may be filled for
// the player).
func ringAnswerPickIfAsked(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			answerPickTarget(t, g, id)
			return
		}
	}
}

// Boromir: an opponent's spell cast for no mana is countered (strict
// mana); yours is not; the sacrifice gives indestructible and tempts.
func TestBoromirWardenOfTheTower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Boromir, Warden of the Tower", "Legendary Creature — Human Soldier", boromirWardenOracle, 3, 3)

	aangAdvanceToMain(t, g, 1)
	free := uuid.New()
	g.WithWriteLock(func() {
		opp.Hand.PushTop(game.Card{InstanceID: free, Name: "Free Bear", TypeLine: "Creature — Bear", ManaCost: "{0}",
			Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	})
	if err := g.CastSpell(opp.ID, free, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the free bear: %v", err)
	}
	ringSettle(t, g)
	if g.Battlefield.Contains(free) || !opp.Graveyard.Contains(free) {
		t.Fatal("an opponent's spell cast for no mana was not countered")
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	boromir := b12Push(g2, me2.ID, "Boromir, Warden of the Tower", "Legendary Creature — Human Soldier", boromirWardenOracle, 3, 3)
	bear := b12Creature(g2, me2.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	advanceToMain(t, g2)
	if err := g2.ActivateCatalogAbility(me2.ID, boromir, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("sacrifice Boromir: %v", err)
	}
	ringSettle(t, g2)
	if !containsString(effectiveAbilities(t, g2, bear), "indestructible") {
		t.Fatal("the bear did not gain indestructible")
	}
	if ringCount(g2, me2.ID) != 1 || ringBearerOf(g2, me2.ID) != bear {
		t.Fatal("the Ring did not tempt onto the bear")
	}
}

// Elrond: scry 1 each time; the Ring tempts on the second resolution
// this turn only.
func TestElrondLordOfRivendellTemptsOnTheSecondResolution(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Elrond, Lord of Rivendell", "Legendary Creature — Elf Noble", elrondLordOfRivendellOrac, false)
	for i, want := range []int{0, 1, 1} {
		castRingCreature(t, g, "Grizzly Bears", "Creature — Bear", "", 2, 2)
		ringSettle(t, g)
		answerScryKeepAll(t, g, me.ID)
		ringSettle(t, g)
		if c := ringPrompt(g, me.ID); c != nil {
			answerRing(t, g, me.ID, c.ChooseCards[0])
			ringSettle(t, g)
		}
		if got := ringCount(g, me.ID); got != want {
			t.Fatalf("after creature %d: tempted %d times, want %d", i+1, got, want)
		}
	}
}

// Samwise: only a permanent card that went to your graveyard from the
// battlefield this turn may be returned; then the tempt.
func TestSamwiseTheStoutheartedReturnsThisTurnsDead(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	old := pushGraveyardCardTyped(me, "Old Bear", "Creature — Bear")
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	ringSettle(t, g)
	samwise := castRingCreature(t, g, "Samwise the Stouthearted", "Legendary Creature — Halfling Peasant", samwiseStoutheartedOracle, 2, 1)
	passPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget && slices.Contains(c.PickTargetCards, old) {
			t.Fatal("a card that was already in the graveyard is offered")
		}
	}
	ringAnswerPickIfAsked(t, g, bear)
	ringSettle(t, g)
	if !me.Hand.Contains(bear) {
		t.Fatal("the creature that died this turn did not return to hand")
	}
	if ringCount(g, me.ID) != 1 || ringBearerOf(g, me.ID) != samwise {
		t.Fatal("the Ring did not tempt onto Samwise")
	}
}

// Scroll of Isildur: chapter I takes an artifact for as long as you
// control the Saga, then tempts.
func TestScrollOfIsildurChapterOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	castCatalogSpell(t, g, "Scroll of Isildur", "Enchantment — Saga", scrollOfIsildurOracle, nil)
	passPriorityAroundTable(t, g)
	ringAnswerPickIfAsked(t, g, rock)
	ringSettle(t, g)
	if c, ok := battlefieldCard(g, rock); !ok || c.Controller != me.ID {
		t.Fatal("chapter I did not take the artifact")
	}
	if ringCount(g, me.ID) != 1 {
		t.Fatal("chapter I did not tempt")
	}
}

// There and Back Again: chapter I stops a creature blocking while you
// control the Saga; chapter III makes Smaug, whose death makes fourteen
// Treasures.
func TestThereAndBackAgain(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	wall := typedCreature(g, opp.ID, "Their Wall", "Creature — Wall", "", 0, 4)
	castCatalogSpell(t, g, "There and Back Again", "Enchantment — Saga", thereAndBackAgainOracle, nil)
	passPriorityAroundTable(t, g)
	ringAnswerPickIfAsked(t, g, wall)
	ringSettle(t, g)
	if !ringBFCard(t, g, wall).Effective().Restrictions.Has(game.CantBlock) {
		t.Fatal("chapter I: the wall can still block")
	}
	if ringCount(g, me.ID) != 1 {
		t.Fatal("chapter I did not tempt")
	}

	advanceToPrecombatMainOf(t, g, seat)
	ringSettle(t, g)
	advanceToPrecombatMainOf(t, g, seat)
	ringSettle(t, g)
	var smaug uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Smaug" {
			smaug = c.InstanceID
		}
	}
	if smaug == uuid.Nil {
		t.Fatal("chapter III made no Smaug")
	}
	sm := ringBFCard(t, g, smaug)
	if !sm.IsLegendary() || sm.Controller != me.ID {
		t.Fatal("Smaug is not a legendary token of yours")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(smaug) })
	ringSettle(t, g)
	if n := onBattlefieldNamed(g, "Treasure"); n != 14 {
		t.Fatalf("%d Treasures after Smaug died, want 14", n)
	}
}

// Sauron, the Necromancer: the attack exiles a creature card and makes
// a tapped and attacking 3/3 black Wraith copy with menace, which is
// exiled at the next end step unless Sauron is your Ring-bearer.
func TestSauronTheNecromancer(t *testing.T) {
	for _, bearer := range []bool{false, true} {
		g := newCatalogGame(t)
		seat := g.Turn.ActiveSeat
		me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
		sauron := typedCreature(g, me.ID, "Sauron, the Necromancer", "Legendary Creature — Avatar Horror", sauronNecromancerOracle, 4, 4, "haste")
		dead := pushGraveyardCardTyped(me, "Dead Giant", "Creature — Giant Warrior")
		if bearer {
			ringTempt(t, g, me.ID)
		}
		declareAttack(t, g, opp.ID, sauron)
		ringAnswerPickIfAsked(t, g, dead)
		ringSettle(t, g)
		if !g.Exile.Contains(dead) {
			t.Fatal("the creature card was not exiled")
		}
		var wraith *game.Card
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].Name == "Dead Giant" {
				wraith = &g.Battlefield.Cards[i]
			}
		}
		if wraith == nil {
			t.Fatal("no token copy")
		}
		id := wraith.InstanceID
		w := ringBFCard(t, g, id)
		if !w.Tapped || w.AttackingTarget != opp.ID || !w.HasSubtype("Wraith") || w.HasSubtype("Giant") ||
			!slices.Equal(w.Colors, []string{"B"}) || !containsString(effectiveAbilities(t, g, id), "menace") {
			t.Fatalf("the token is not a tapped, attacking 3/3 black Wraith with menace: %+v", *w)
		}
		if p, tt := ptOf(t, g, id); p != 3 || tt != 3 {
			t.Fatalf("the token is %d/%d, want 3/3", p, tt)
		}
		advanceTo(t, g, game.StepEnd)
		ringSettle(t, g)
		if g.Battlefield.Contains(id) == !bearer {
			t.Errorf("Ring-bearer %v: the token is on the battlefield = %v at the end step", bearer, g.Battlefield.Contains(id))
		}
	}
}
