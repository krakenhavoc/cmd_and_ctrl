package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// weeping_angel_test.go — #2041: Weeping Angel, the card that waited on
// CR 205.1a's "losing a card type loses its subtypes".

const weepingAngelOracle = "1a89c623-c9c3-443c-bbd3-834767b6bdb8"

func pushWeepingAngel(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.ApaPush(g, owner, owner, game.Card{Name: "Weeping Angel", TypeLine: "Artifact Creature — Alien Angel",
		OracleID: weepingAngelOracle, ManaCost: "{1}{U}{B}", Power: 2, Toughness: 2})
}

// "Whenever an opponent casts a creature spell, this creature isn't a
// creature until end of turn": it is still an artifact, and it is no
// longer an Alien or an Angel (the ruling, CR 205.1a). Its controller's
// own creature spell does nothing. The next turn it is all three again.
func TestWeepingAngelStopsBeingAnAlienAngelWhileItIsntACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	angel := pushWeepingAngel(g, me.ID)
	if !effectstest.ContainsString(effectstest.EffectiveSubtypes(t, g, angel), "Angel") {
		t.Fatal("control: Weeping Angel is printed an Alien Angel")
	}
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.WeepingAngelOpponentCastsCreature(t, g, me)
	if !effectstest.ContainsString(effectstest.EffectiveTypes(t, g, angel), "Creature") {
		t.Fatal("its controller's own creature spell turned it off")
	}

	effectstest.WeepingAngelOpponentCastsCreature(t, g, opp)
	types, subtypes := effectstest.EffectiveTypes(t, g, angel), effectstest.EffectiveSubtypes(t, g, angel)
	if effectstest.ContainsString(types, "Creature") || !effectstest.ContainsString(types, "Artifact") {
		t.Fatalf("types = %v, want an Artifact that isn't a creature", types)
	}
	if len(subtypes) != 0 {
		t.Errorf("subtypes = %v, want none: it stops being an Alien and an Angel (CR 205.1a)", subtypes)
	}
	var angelType bool
	g.ReadSnapshot(func() {
		if c := effectstest.FindBattlefieldCardForTest(g, angel); c != nil {
			angelType = c.HasSubtype("Angel")
		}
	})
	if angelType {
		t.Error("HasSubtype(\"Angel\") still answers for a noncreature Weeping Angel")
	}

	if err := g.PassTurn(); err != nil {
		t.Fatal(err)
	}
	if !effectstest.ContainsString(effectstest.EffectiveTypes(t, g, angel), "Creature") ||
		!effectstest.ContainsString(effectstest.EffectiveSubtypes(t, g, angel), "Alien") || !effectstest.ContainsString(effectstest.EffectiveSubtypes(t, g, angel), "Angel") {
		t.Errorf("next turn: types %v subtypes %v, want an Alien Angel artifact creature again",
			effectstest.EffectiveTypes(t, g, angel), effectstest.EffectiveSubtypes(t, g, angel))
	}
}

// "If this creature would deal combat damage to a creature, prevent that
// damage and that creature's owner shuffles it into their library":
// once per creature, into its OWNER's library (a stolen creature goes
// home); noncombat damage is not its business; damage that can't be prevented is dealt and the creature still goes
// (CR 615.12).
func TestWeepingAngelShufflesTheCreatureItWouldDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	angel := pushWeepingAngel(g, me.ID)
	blocker := effectstest.Pr7Creature(g, opp.ID, "Blocker", 2)
	stolen := effectstest.ApaPush(g, third.ID, opp.ID, game.Card{Name: "Stolen", TypeLine: "Creature — Test", Power: 2, Toughness: 4})
	oppLib, thirdLib := opp.Library.Size(), third.Library.Size()

	effectstest.Pr8Hit(t, g, blocker, map[uuid.UUID]int{angel: 1})
	if effectstest.DamageMarkedOn(g, blocker) != 1 {
		t.Fatalf("noncombat damage: blocker has %d damage, want 1 (only combat damage is prevented)", effectstest.DamageMarkedOn(g, blocker))
	}

	if err := g.MarkCombatDamage(angel, blocker, 2); err != nil {
		t.Fatal(err)
	}
	if err := g.MarkCombatDamage(angel, stolen, 1); err != nil {
		t.Fatal(err)
	}
	g.RunStateChecksForTest()
	if effectstest.FindBattlefieldCardForTest(g, blocker) != nil || opp.Library.Size() != oppLib+1 {
		t.Errorf("blocker: on battlefield %v, opponent's library %d (want %d)",
			effectstest.FindBattlefieldCardForTest(g, blocker) != nil, opp.Library.Size(), oppLib+1)
	}
	if effectstest.FindBattlefieldCardForTest(g, stolen) != nil || third.Library.Size() != thirdLib+1 || opp.Library.Size() != oppLib+1 {
		t.Errorf("stolen creature: on battlefield %v, owner's library %d (want %d), controller's %d (want %d)",
			effectstest.FindBattlefieldCardForTest(g, stolen) != nil, third.Library.Size(), thirdLib+1, opp.Library.Size(), oppLib+1)
	}

	target := effectstest.Pr7Creature(g, opp.ID, "Target", 2)
	oppLib = opp.Library.Size()
	effectstest.PushCatalogPermanent(g, opp.ID, "Frenzied Baloth", "Creature — Beast", effectstest.Pr6BalothOracle, false)
	if err := g.MarkCombatDamage(angel, target, 2); err != nil {
		t.Fatal(err)
	}
	g.RunStateChecksForTest()
	if effectstest.FindBattlefieldCardForTest(g, target) != nil || opp.Library.Size() != oppLib+1 {
		t.Errorf("unpreventable: on battlefield %v, library %d (want %d); the creature is still shuffled in (CR 615.12)",
			effectstest.FindBattlefieldCardForTest(g, target) != nil, opp.Library.Size(), oppLib+1)
	}
}
