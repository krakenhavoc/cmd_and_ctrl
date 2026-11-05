package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const hereComesANewHeroOracle = "c4fb12cb-05e1-4df2-b62e-3391621a601b"

// newHeroCast casts Here Comes a New Hero! strictly for X from a
// floated pool, at the player and (when not Nil) the creature.
func newHeroCast(t *testing.T, g *game.Game, me *game.Player, x int, pool string, player, creature uuid.UUID) error {
	t.Helper()
	id := socB20HandCard(me, "Here Comes a New Hero!", "Sorcery", "{X}{2}{U}", hereComesANewHeroOracle)
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	floatForTest(g, me, pool)
	targets := []game.TargetRef{{Kind: game.TargetPlayer, ID: player, Slot: 0}}
	if creature != uuid.Nil {
		targets = append(targets, game.TargetRef{Kind: game.TargetCard, ID: creature, Slot: 1})
	}
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: x, Targets: targets})
}

func newHeroCreature(g *game.Game, owner uuid.UUID, name, cost string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Mutant", ManaCost: cost,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

func newHeroTokensNamed(g *game.Game, controller uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name && c.IsToken() {
			out = append(out, c)
		}
	}
	return out
}

// X=2: the target player draws two, and the caster gets an untapped,
// counterless token copy of an opponent's mana value 2 creature.
func TestHereComesANewHeroDrawsXAndCopiesACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(opp, "A", "B", "C")
	bruiser := newHeroCreature(g, opp.ID, "Bruiser", "{1}{G}", 3, 3)
	putCounters(t, g, bruiser, game.CounterPlusOne, 2)
	hand := opp.Hand.Size()
	if err := newHeroCast(t, g, me, 2, "CCCCU", opp.ID, bruiser); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != hand+2 {
		t.Errorf("target player's hand %d, want %d", got, hand+2)
	}
	tokens := newHeroTokensNamed(g, me.ID, "Bruiser")
	if len(tokens) != 1 {
		t.Fatalf("the caster has %d Bruiser tokens, want 1", len(tokens))
	}
	if n := countersOn(g, tokens[0].InstanceID, game.CounterPlusOne); n != 0 {
		t.Errorf("the copy has %d counters; counters are not copied", n)
	}
	if tokens[0].Power != 3 || tokens[0].Toughness != 3 {
		t.Errorf("the copy is %d/%d, want the printed 3/3", tokens[0].Power, tokens[0].Toughness)
	}
	if len(newHeroTokensNamed(g, opp.ID, "Bruiser")) != 0 {
		t.Error("the copy went to the creature's controller")
	}
}

// A creature with mana value greater than X is not a legal target.
func TestHereComesANewHeroRefusesACreatureAboveX(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := newHeroCreature(g, opp.ID, "Big", "{2}{G}", 4, 4)
	if err := newHeroCast(t, g, me, 2, "CCCCU", me.ID, big); err == nil {
		t.Fatal("a mana value 3 creature was accepted at X=2")
	}
}

// A creature with {X} in its cost counts X as 0 on the battlefield
// (CR 202.3e): {X}{G} is mana value 1, legal at X=1.
func TestHereComesANewHeroReadsXInACopiedCostAsZero(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ooze := newHeroCreature(g, me.ID, "Ooze", "{X}{G}", 0, 0)
	putCounters(t, g, ooze, game.CounterPlusOne, 3)
	if err := newHeroCast(t, g, me, 1, "CCCU", me.ID, ooze); err != nil {
		t.Fatalf("cast at an {X}{G} creature for X=1: %v", err)
	}
	passPriorityAroundTable(t, g)
	if len(newHeroTokensNamed(g, me.ID, "Ooze")) != 1 {
		t.Error("no copy of the {X}{G} creature")
	}
}

// X=0 is a real cast: a mana value 0 creature (a token with no mana
// cost) is still copied, and a mana value 1 creature is refused.
func TestHereComesANewHeroAtXZeroCopiesAManaValueZeroCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spirit := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Spirit", TypeLine: "Token Creature — Spirit",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	elf := newHeroCreature(g, opp.ID, "Elf", "{G}", 1, 1)
	if err := newHeroCast(t, g, me, 0, "CCU", me.ID, elf); err == nil {
		t.Fatal("a mana value 1 creature was accepted at X=0")
	}
	if err := newHeroCast(t, g, me, 0, "CCU", me.ID, spirit); err != nil {
		t.Fatalf("cast at a mana value 0 token for X=0: %v", err)
	}
	passPriorityAroundTable(t, g)
	if len(newHeroTokensNamed(g, me.ID, "Spirit")) != 1 {
		t.Error("no copy of the mana value 0 token")
	}
}

// "Up to one": with no creature, the player still draws X.
func TestHereComesANewHeroWithNoCreatureStillDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C", "D")
	hand := me.Hand.Size()
	if err := newHeroCast(t, g, me, 3, "CCCCCU", me.ID, uuid.Nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	// The spell left the hand; three were drawn.
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand %d, want %d", got, hand+3)
	}
}

// The cost is {X}{2}{U}: X=2 with four mana does not pay.
func TestHereComesANewHeroRefusesAPoolThatCannotPayX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if err := newHeroCast(t, g, me, 2, "CCCU", me.ID, uuid.Nil); err == nil {
		t.Fatal("X=2 was paid with four mana")
	}
}

// The creature leaving before resolution leaves the draw in place and
// makes no copy (CR 608.2b).
func TestHereComesANewHeroCreatureGoneStillDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "A", "B")
	bear := newHeroCreature(g, opp.ID, "Bear", "{1}{G}", 2, 2)
	if err := newHeroCast(t, g, me, 2, "CCCCU", me.ID, bear); err != nil {
		t.Fatalf("cast: %v", err)
	}
	hand := me.Hand.Size()
	destroyNow(t, g, bear)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d, want %d", got, hand+2)
	}
	if len(newHeroTokensNamed(g, me.ID, "Bear")) != 0 {
		t.Error("a copy was made of a creature that left")
	}
}
