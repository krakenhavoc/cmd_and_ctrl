package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// weeping_angel_test.go — #2041: Weeping Angel, the card that waited on
// CR 205.1a's "losing a card type loses its subtypes".

const weepingAngelOracle = "1a89c623-c9c3-443c-bbd3-834767b6bdb8"

func pushWeepingAngel(g *game.Game, owner uuid.UUID) uuid.UUID {
	return apaPush(g, owner, owner, game.Card{Name: "Weeping Angel", TypeLine: "Artifact Creature — Alien Angel",
		OracleID: weepingAngelOracle, ManaCost: "{1}{U}{B}", Power: 2, Toughness: 2})
}

// weepingAngelOpponentCastsCreature has `opp` cast a vanilla creature
// with flash (so it may at any time), and lets the stack settle.
func weepingAngelOpponentCastsCreature(t *testing.T, g *game.Game, opp *game.Player) {
	t.Helper()
	id := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Flash Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Keywords: []string{"flash"}, Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent casts a creature: %v", err)
	}
	passPriorityAroundTable(t, g)
}

// "Whenever an opponent casts a creature spell, this creature isn't a
// creature until end of turn": it is still an artifact, and it is no
// longer an Alien or an Angel (the ruling, CR 205.1a). Its controller's
// own creature spell does nothing. The next turn it is all three again.
func TestWeepingAngelStopsBeingAnAlienAngelWhileItIsntACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	angel := pushWeepingAngel(g, me.ID)
	if !containsString(effectiveSubtypes(t, g, angel), "Angel") {
		t.Fatal("control: Weeping Angel is printed an Alien Angel")
	}
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	weepingAngelOpponentCastsCreature(t, g, me)
	if !containsString(effectiveTypes(t, g, angel), "Creature") {
		t.Fatal("its controller's own creature spell turned it off")
	}

	weepingAngelOpponentCastsCreature(t, g, opp)
	types, subtypes := effectiveTypes(t, g, angel), effectiveSubtypes(t, g, angel)
	if containsString(types, "Creature") || !containsString(types, "Artifact") {
		t.Fatalf("types = %v, want an Artifact that isn't a creature", types)
	}
	if len(subtypes) != 0 {
		t.Errorf("subtypes = %v, want none: it stops being an Alien and an Angel (CR 205.1a)", subtypes)
	}
	var angelType bool
	g.ReadSnapshot(func() {
		if c := findBattlefieldCardForTest(g, angel); c != nil {
			angelType = c.HasSubtype("Angel")
		}
	})
	if angelType {
		t.Error("HasSubtype(\"Angel\") still answers for a noncreature Weeping Angel")
	}

	if err := g.PassTurn(); err != nil {
		t.Fatal(err)
	}
	if !containsString(effectiveTypes(t, g, angel), "Creature") ||
		!containsString(effectiveSubtypes(t, g, angel), "Alien") || !containsString(effectiveSubtypes(t, g, angel), "Angel") {
		t.Errorf("next turn: types %v subtypes %v, want an Alien Angel artifact creature again",
			effectiveTypes(t, g, angel), effectiveSubtypes(t, g, angel))
	}
}

// "If this creature would deal combat damage to a creature, prevent that
// damage and that creature's owner shuffles it into their library":
// once per creature, into its OWNER's library (a stolen creature goes
// home); noncombat damage is not its business; damage that can't be prevented is dealt and the creature still goes
// (CR 615.12).
func TestWeepingAngelShufflesTheCreatureItWouldDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	angel := pushWeepingAngel(g, me.ID)
	blocker := pr7Creature(g, opp.ID, "Blocker", 2)
	stolen := apaPush(g, third.ID, opp.ID, game.Card{Name: "Stolen", TypeLine: "Creature — Test", Power: 2, Toughness: 4})
	oppLib, thirdLib := opp.Library.Size(), third.Library.Size()

	pr8Hit(t, g, blocker, map[uuid.UUID]int{angel: 1})
	if damageMarkedOn(g, blocker) != 1 {
		t.Fatalf("noncombat damage: blocker has %d damage, want 1 (only combat damage is prevented)", damageMarkedOn(g, blocker))
	}

	if err := g.MarkCombatDamage(angel, blocker, 2); err != nil {
		t.Fatal(err)
	}
	if err := g.MarkCombatDamage(angel, stolen, 1); err != nil {
		t.Fatal(err)
	}
	g.RunStateChecksForTest()
	if findBattlefieldCardForTest(g, blocker) != nil || opp.Library.Size() != oppLib+1 {
		t.Errorf("blocker: on battlefield %v, opponent's library %d (want %d)",
			findBattlefieldCardForTest(g, blocker) != nil, opp.Library.Size(), oppLib+1)
	}
	if findBattlefieldCardForTest(g, stolen) != nil || third.Library.Size() != thirdLib+1 || opp.Library.Size() != oppLib+1 {
		t.Errorf("stolen creature: on battlefield %v, owner's library %d (want %d), controller's %d (want %d)",
			findBattlefieldCardForTest(g, stolen) != nil, third.Library.Size(), thirdLib+1, opp.Library.Size(), oppLib+1)
	}

	target := pr7Creature(g, opp.ID, "Target", 2)
	oppLib = opp.Library.Size()
	pushCatalogPermanent(g, opp.ID, "Frenzied Baloth", "Creature — Beast", pr6BalothOracle, false)
	if err := g.MarkCombatDamage(angel, target, 2); err != nil {
		t.Fatal(err)
	}
	g.RunStateChecksForTest()
	if findBattlefieldCardForTest(g, target) != nil || opp.Library.Size() != oppLib+1 {
		t.Errorf("unpreventable: on battlefield %v, library %d (want %d); the creature is still shuffled in (CR 615.12)",
			findBattlefieldCardForTest(g, target) != nil, opp.Library.Size(), oppLib+1)
	}
}
