package effects

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// last_turn_attacks_cards_test.go — ADR 0108 §6 (#1882): Goblin Rock Sled,
// Giant Turtle, Tangle Kelp, Avenge and O-Kagachi, driven through whole
// turns. The engine half is pinned in game/last_turn_attacks_test.go.

const (
	goblinRockSledOracle = "ada3247e-ec5e-499d-bc15-1d9dd80a59ae"
	giantTurtleOracle    = "9297c0a6-1a8e-4e6e-99d6-f0877b2ec46c"
	tangleKelpOracle     = "06ba313f-50b4-4e84-a8e0-b6dda47225f2"
	avengeOracle         = "65a64759-3a93-4542-84fc-3af5ccb741f4"
	oKagachiOracle       = "300715a0-4f95-4212-9e13-558434c9d1a4"
)

// aroundTheTableToMain walks from seat 0's turn through the other three
// seats' turns to seat 0's next precombat main step, which is after its
// untap step.
func aroundTheTableToMain(t *testing.T, g *game.Game) {
	t.Helper()
	for _, seat := range []int{1, 2, 3, 0} {
		advanceToMainOf(t, g, seat)
	}
}

// Goblin Rock Sled: it attacked, so it stays tapped through its
// controller's next untap step; having sat that turn out, it untaps
// the turn after.
func TestGoblinRockSledStaysTappedAfterAttackingLastTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, b := g.Seats[0], g.Seats[1]
	pushTestPermanent(g, b.ID, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	sled := pushTestPermanent(g, me.ID, game.Card{Name: "Goblin Rock Sled", OracleID: goblinRockSledOracle,
		TypeLine: "Creature — Goblin", Power: 3, Toughness: 1})
	advanceToMainOf(t, g, 0)

	attackWith(t, g, b.ID, sled)
	if !isTapped(g, sled) {
		t.Fatal("setup: the Sled did not tap to attack")
	}
	aroundTheTableToMain(t, g)
	if !isTapped(g, sled) {
		t.Error("the Sled attacked during its controller's last turn, so it doesn't untap")
	}
	aroundTheTableToMain(t, g)
	if isTapped(g, sled) {
		t.Error("the Sled did not attack during its controller's last turn, so it untaps")
	}
}

// Another player's turn is not "your last turn": a Sled that attacked,
// then sat through its controller's next turn, is free again.
func TestGoblinRockSledUntapsIfItDidNotAttack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sled := pushTestPermanent(g, me.ID, game.Card{Name: "Goblin Rock Sled", OracleID: goblinRockSledOracle,
		TypeLine: "Creature — Goblin", Power: 3, Toughness: 1, Tapped: true})
	advanceToMainOf(t, g, 0)
	aroundTheTableToMain(t, g)
	if isTapped(g, sled) {
		t.Error("a Sled that never attacked untaps as usual")
	}
}

// It may attack only a player who controls a Mountain.
func TestGoblinRockSledNeedsTheDefenderToControlAMountain(t *testing.T) {
	g := newCatalogGame(t)
	me, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	pushTestPermanent(g, b.ID, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	sled := pushTestPermanent(g, me.ID, game.Card{Name: "Goblin Rock Sled", OracleID: goblinRockSledOracle,
		TypeLine: "Creature — Goblin", Power: 3, Toughness: 1})
	advanceToDeclareAttackersOf(t, g, 0)
	if err := declareResult(g, sled, b.ID); err != nil {
		t.Errorf("attack the player with a Mountain: %v", err)
	}
	wantAttackRefused(t, g, sled, c.ID, "")
}

// Giant Turtle can't attack if it attacked during your last turn, and
// the three turns in between are not its controller's.
func TestGiantTurtleCantAttackTwiceInARow(t *testing.T) {
	g := newCatalogGame(t)
	me, b := g.Seats[0], g.Seats[1]
	turtle := pushTestPermanent(g, me.ID, game.Card{Name: "Giant Turtle", OracleID: giantTurtleOracle,
		TypeLine: "Creature — Turtle", Power: 2, Toughness: 4})
	advanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(turtle, b.ID); err != nil {
		t.Fatalf("the Turtle's first attack: %v", err)
	}
	aroundTheTableToMain(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	err := declareResult(g, turtle, b.ID)
	if !errors.Is(err, game.ErrCantAttack) {
		t.Fatalf("the Turtle attacked last turn: err = %v, want ErrCantAttack", err)
	}
	// It sat that turn out, so its last turn had no attack.
	aroundTheTableToMain(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := declareResult(g, turtle, b.ID); err != nil {
		t.Errorf("the Turtle did not attack last turn: %v", err)
	}
}

// An opponent's turn does not reset it: the Turtle attacks, and when it
// is its controller's turn again the restriction holds even though the
// other three players attacked in between.
func TestGiantTurtleIgnoresOpponentsTurns(t *testing.T) {
	g := newCatalogGame(t)
	me, b := g.Seats[0], g.Seats[1]
	turtle := pushTestPermanent(g, me.ID, game.Card{Name: "Giant Turtle", OracleID: giantTurtleOracle,
		TypeLine: "Creature — Turtle", Power: 2, Toughness: 4})
	bear := pushVanillaCreature(g, b.ID, "Bear", 2, 2)
	advanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(turtle, b.ID); err != nil {
		t.Fatal(err)
	}
	advanceToMainOf(t, g, 1)
	attackWith(t, g, me.ID, bear)
	aroundTheTableToMainOf(t, g, 0)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := declareResult(g, turtle, b.ID); !errors.Is(err, game.ErrCantAttack) {
		t.Errorf("the Turtle attacked during its controller's last turn: err = %v", err)
	}
}

// Tangle Kelp taps the creature as it enters, and the creature stays
// tapped through its controller's untap step only if it attacked during
// that controller's last turn.
func TestTangleKelpHoldsAnAttackerDown(t *testing.T) {
	g := newCatalogGame(t)
	me, b := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, b.ID, "Bear", 2, 2)
	castCatalogSpell(t, g, "Tangle Kelp", "Enchantment — Aura", tangleKelpOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if !isTapped(g, bear) {
		t.Fatal("the entry trigger taps the enchanted creature")
	}
	// Its controller has taken no turn yet: it untaps.
	advanceToMainOf(t, g, 1)
	if isTapped(g, bear) {
		t.Fatal("the Bear did not attack during its controller's last turn, so it untaps")
	}
	attackWith(t, g, me.ID, bear)
	// Its controller's next untap step: it attacked last turn.
	aroundTheTableToMainOf(t, g, 1)
	if !isTapped(g, bear) {
		t.Error("the Bear attacked during its controller's last turn, so it doesn't untap")
	}
	aroundTheTableToMainOf(t, g, 1)
	if isTapped(g, bear) {
		t.Error("the Bear sat out its last turn, so it untaps")
	}
}

// aroundTheTableToMainOf walks to the next precombat main step of `seat`,
// passing the turns of the others.
func aroundTheTableToMainOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 1; i <= len(g.Seats); i++ {
		advanceToMainOf(t, g, (g.Turn.ActiveSeat+i)%len(g.Seats))
		if g.Turn.ActiveSeat == seat {
			return
		}
	}
}

// Avenge costs {2} less only when some player attacked YOU during their
// own last turn: an attack on someone else does not count, and neither
// does an attack on one of your planeswalkers.
func TestAvengeCostsLessAfterBeingAttacked(t *testing.T) {
	g := newCatalogGame(t)
	me, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	bear := pushVanillaCreature(g, b.ID, "Bear", 2, 2)
	if got := selfPricedMV(t, g, me, avengeOracle, "Sorcery", "{4}{W}{W}"); got != 6 {
		t.Fatalf("Avenge with nobody having attacked: %d, want 6", got)
	}
	advanceToMainOf(t, g, 1)
	attackWith(t, g, c.ID, bear)
	if got := selfPricedMV(t, g, me, avengeOracle, "Sorcery", "{4}{W}{W}"); got != 6 {
		t.Errorf("an attack made this turn is not an attack during their LAST turn: %d, want 6", got)
	}
	advanceToMainOf(t, g, 2)
	if got := selfPricedMV(t, g, me, avengeOracle, "Sorcery", "{4}{W}{W}"); got != 6 {
		t.Errorf("seat 1 attacked seat 2, not me: %d, want 6", got)
	}
	if got := selfPricedMV(t, g, c, avengeOracle, "Sorcery", "{4}{W}{W}"); got != 4 {
		t.Errorf("seat 1 attacked seat 2 during its last turn: %d, want 4", got)
	}
}

func TestAvengeIgnoresAnAttackOnYourPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me, b := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, b.ID, "Bear", 2, 2)
	walker := pushWalkerForTest(g, me.ID, "Walker", "", 4)
	advanceToMainOf(t, g, 1)
	attackWith(t, g, walker, bear)
	aroundTheTableToMainOf(t, g, 0)
	if got := selfPricedMV(t, g, me, avengeOracle, "Sorcery", "{4}{W}{W}"); got != 6 {
		t.Errorf("an attack on a planeswalker is not an attack on me: %d, want 6", got)
	}
}

func TestAvengeCostsLessWhenAPlayerAttackedYou(t *testing.T) {
	g := newCatalogGame(t)
	me, b := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, b.ID, "Bear", 2, 2)
	advanceToMainOf(t, g, 1)
	attackWith(t, g, me.ID, bear)
	aroundTheTableToMainOf(t, g, 0)
	if got := selfPricedMV(t, g, me, avengeOracle, "Sorcery", "{4}{W}{W}"); got != 4 {
		t.Errorf("Avenge after being attacked: %d, want 4", got)
	}
}

func TestAvengeDestroysEveryCreatureAndGainsLifeForEach(t *testing.T) {
	g := newCatalogGame(t)
	me, b := g.Seats[0], g.Seats[1]
	pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	pushVanillaCreature(g, b.ID, "Their Bear", 2, 2)
	pushVanillaCreature(g, b.ID, "Their Ox", 3, 3)
	life := me.Life
	castCatalogSpell(t, g, "Avenge", "Sorcery", avengeOracle, nil)
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.IsCreature() {
			t.Errorf("%s survived Avenge", c.Name)
		}
	}
	if me.Life != life+3 {
		t.Errorf("life = %d, want %d (one for each of three creatures)", me.Life, life+3)
	}
}

// O-Kagachi exiles a nonland permanent of the player it hit, but only
// if that player attacked its controller during their last turn.
func TestOKagachiExilesFromAPlayerWhoAttackedYou(t *testing.T) {
	g := newCatalogGame(t)
	me, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	kagachi := b12Push(g, me.ID, "O-Kagachi, Vengeful Kami", "Legendary Creature — Dragon Spirit", oKagachiOracle, 6, 6)
	bear := pushVanillaCreature(g, b.ID, "Their Bear", 2, 2)
	rock := b12Permanent(g, b.ID, "Their Rock", "Artifact")
	land := b12Permanent(g, b.ID, "Their Land", "Basic Land — Island")
	otherRock := b12Permanent(g, c.ID, "Other Rock", "Artifact")

	// Seat 1 attacks me during its turn.
	advanceToMainOf(t, g, 1)
	attackWith(t, g, me.ID, bear)
	aroundTheTableToMainOf(t, g, 0)
	attackWith(t, g, b.ID, kagachi)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetCards, rock) || !hasID(p.PickTargetCards, bear) {
		t.Error("the hit player's nonland permanents are offered")
	}
	if hasID(p.PickTargetCards, land) || hasID(p.PickTargetCards, otherRock) {
		t.Error("a land and another player's permanent are not offered")
	}
	pickCard(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the chosen permanent is exiled")
	}
	if !g.Exile.Contains(rock) {
		t.Error("the chosen permanent went to exile")
	}
}

func TestOKagachiDoesNothingToAPlayerWhoDidNotAttackYou(t *testing.T) {
	g := newCatalogGame(t)
	me, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	kagachi := b12Push(g, me.ID, "O-Kagachi, Vengeful Kami", "Legendary Creature — Dragon Spirit", oKagachiOracle, 6, 6)
	bear := pushVanillaCreature(g, b.ID, "Their Bear", 2, 2)
	rock := b12Permanent(g, b.ID, "Their Rock", "Artifact")

	// Seat 1 attacks seat 2 during its turn, not me.
	advanceToMainOf(t, g, 1)
	attackWith(t, g, c.ID, bear)
	aroundTheTableToMainOf(t, g, 0)
	attackWith(t, g, b.ID, kagachi)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("seat 1 did not attack me during its last turn: no trigger")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("nothing is exiled")
	}
}
