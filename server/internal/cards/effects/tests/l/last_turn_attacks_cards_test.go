package testsl

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

// Goblin Rock Sled: it attacked, so it stays tapped through its
// controller's next untap step; having sat that turn out, it untaps
// the turn after.
func TestGoblinRockSledStaysTappedAfterAttackingLastTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, b := g.Seats[0], g.Seats[1]
	effectstest.PushTestPermanent(g, b.ID, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	sled := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Goblin Rock Sled", OracleID: goblinRockSledOracle,
		TypeLine: "Creature — Goblin", Power: 3, Toughness: 1})
	effectstest.AdvanceToMainOf(t, g, 0)

	effectstest.AttackWith(t, g, b.ID, sled)
	if !effectstest.IsTapped(g, sled) {
		t.Fatal("setup: the Sled did not tap to attack")
	}
	effectstest.AroundTheTableToMain(t, g)
	if !effectstest.IsTapped(g, sled) {
		t.Error("the Sled attacked during its controller's last turn, so it doesn't untap")
	}
	effectstest.AroundTheTableToMain(t, g)
	if effectstest.IsTapped(g, sled) {
		t.Error("the Sled did not attack during its controller's last turn, so it untaps")
	}
}

// Another player's turn is not "your last turn": a Sled that attacked,
// then sat through its controller's next turn, is free again.
func TestGoblinRockSledUntapsIfItDidNotAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sled := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Goblin Rock Sled", OracleID: goblinRockSledOracle,
		TypeLine: "Creature — Goblin", Power: 3, Toughness: 1, Tapped: true})
	effectstest.AdvanceToMainOf(t, g, 0)
	effectstest.AroundTheTableToMain(t, g)
	if effectstest.IsTapped(g, sled) {
		t.Error("a Sled that never attacked untaps as usual")
	}
}

// It may attack only a player who controls a Mountain.
func TestGoblinRockSledNeedsTheDefenderToControlAMountain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.PushTestPermanent(g, b.ID, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	sled := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Goblin Rock Sled", OracleID: goblinRockSledOracle,
		TypeLine: "Creature — Goblin", Power: 3, Toughness: 1})
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := effectstest.DeclareResult(g, sled, b.ID); err != nil {
		t.Errorf("attack the player with a Mountain: %v", err)
	}
	effectstest.WantAttackRefused(t, g, sled, c.ID, "")
}

// Giant Turtle can't attack if it attacked during your last turn, and
// the three turns in between are not its controller's.
func TestGiantTurtleCantAttackTwiceInARow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, b := g.Seats[0], g.Seats[1]
	turtle := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Giant Turtle", OracleID: giantTurtleOracle,
		TypeLine: "Creature — Turtle", Power: 2, Toughness: 4})
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(turtle, b.ID); err != nil {
		t.Fatalf("the Turtle's first attack: %v", err)
	}
	effectstest.AroundTheTableToMain(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	err := effectstest.DeclareResult(g, turtle, b.ID)
	if !errors.Is(err, game.ErrCantAttack) {
		t.Fatalf("the Turtle attacked last turn: err = %v, want ErrCantAttack", err)
	}
	// It sat that turn out, so its last turn had no attack.
	effectstest.AroundTheTableToMain(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := effectstest.DeclareResult(g, turtle, b.ID); err != nil {
		t.Errorf("the Turtle did not attack last turn: %v", err)
	}
}

// An opponent's turn does not reset it: the Turtle attacks, and when it
// is its controller's turn again the restriction holds even though the
// other three players attacked in between.
func TestGiantTurtleIgnoresOpponentsTurns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, b := g.Seats[0], g.Seats[1]
	turtle := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Giant Turtle", OracleID: giantTurtleOracle,
		TypeLine: "Creature — Turtle", Power: 2, Toughness: 4})
	bear := effectstest.PushVanillaCreature(g, b.ID, "Bear", 2, 2)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(turtle, b.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.AttackWith(t, g, me.ID, bear)
	effectstest.AroundTheTableToMainOf(t, g, 0)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := effectstest.DeclareResult(g, turtle, b.ID); !errors.Is(err, game.ErrCantAttack) {
		t.Errorf("the Turtle attacked during its controller's last turn: err = %v", err)
	}
}

// Tangle Kelp taps the creature as it enters, and the creature stays
// tapped through its controller's untap step only if it attacked during
// that controller's last turn.
func TestTangleKelpHoldsAnAttackerDown(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, b := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, b.ID, "Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Tangle Kelp", "Enchantment — Aura", tangleKelpOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsTapped(g, bear) {
		t.Fatal("the entry trigger taps the enchanted creature")
	}
	// Its controller has taken no turn yet: it untaps.
	effectstest.AdvanceToMainOf(t, g, 1)
	if effectstest.IsTapped(g, bear) {
		t.Fatal("the Bear did not attack during its controller's last turn, so it untaps")
	}
	effectstest.AttackWith(t, g, me.ID, bear)
	// Its controller's next untap step: it attacked last turn.
	effectstest.AroundTheTableToMainOf(t, g, 1)
	if !effectstest.IsTapped(g, bear) {
		t.Error("the Bear attacked during its controller's last turn, so it doesn't untap")
	}
	effectstest.AroundTheTableToMainOf(t, g, 1)
	if effectstest.IsTapped(g, bear) {
		t.Error("the Bear sat out its last turn, so it untaps")
	}
}

// Avenge costs {2} less only when some player attacked YOU during their
// own last turn: an attack on someone else does not count, and neither
// does an attack on one of your planeswalkers.
func TestAvengeCostsLessAfterBeingAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	bear := effectstest.PushVanillaCreature(g, b.ID, "Bear", 2, 2)
	if got := effectstest.SelfPricedMV(t, g, me, avengeOracle, "Sorcery", "{4}{W}{W}"); got != 6 {
		t.Fatalf("Avenge with nobody having attacked: %d, want 6", got)
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.AttackWith(t, g, c.ID, bear)
	if got := effectstest.SelfPricedMV(t, g, me, avengeOracle, "Sorcery", "{4}{W}{W}"); got != 6 {
		t.Errorf("an attack made this turn is not an attack during their LAST turn: %d, want 6", got)
	}
	effectstest.AdvanceToMainOf(t, g, 2)
	if got := effectstest.SelfPricedMV(t, g, me, avengeOracle, "Sorcery", "{4}{W}{W}"); got != 6 {
		t.Errorf("seat 1 attacked seat 2, not me: %d, want 6", got)
	}
	if got := effectstest.SelfPricedMV(t, g, c, avengeOracle, "Sorcery", "{4}{W}{W}"); got != 4 {
		t.Errorf("seat 1 attacked seat 2 during its last turn: %d, want 4", got)
	}
}

func TestAvengeIgnoresAnAttackOnYourPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, b := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, b.ID, "Bear", 2, 2)
	walker := effectstest.PushWalkerForTest(g, me.ID, "Walker", "", 4)
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.AttackWith(t, g, walker, bear)
	effectstest.AroundTheTableToMainOf(t, g, 0)
	if got := effectstest.SelfPricedMV(t, g, me, avengeOracle, "Sorcery", "{4}{W}{W}"); got != 6 {
		t.Errorf("an attack on a planeswalker is not an attack on me: %d, want 6", got)
	}
}

func TestAvengeCostsLessWhenAPlayerAttackedYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, b := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, b.ID, "Bear", 2, 2)
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.AttackWith(t, g, me.ID, bear)
	effectstest.AroundTheTableToMainOf(t, g, 0)
	if got := effectstest.SelfPricedMV(t, g, me, avengeOracle, "Sorcery", "{4}{W}{W}"); got != 4 {
		t.Errorf("Avenge after being attacked: %d, want 4", got)
	}
}

func TestAvengeDestroysEveryCreatureAndGainsLifeForEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, b := g.Seats[0], g.Seats[1]
	effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	effectstest.PushVanillaCreature(g, b.ID, "Their Bear", 2, 2)
	effectstest.PushVanillaCreature(g, b.ID, "Their Ox", 3, 3)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Avenge", "Sorcery", avengeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	kagachi := effectstest.B12Push(g, me.ID, "O-Kagachi, Vengeful Kami", "Legendary Creature — Dragon Spirit", oKagachiOracle, 6, 6)
	bear := effectstest.PushVanillaCreature(g, b.ID, "Their Bear", 2, 2)
	rock := effectstest.B12Permanent(g, b.ID, "Their Rock", "Artifact")
	land := effectstest.B12Permanent(g, b.ID, "Their Land", "Basic Land — Island")
	otherRock := effectstest.B12Permanent(g, c.ID, "Other Rock", "Artifact")

	// Seat 1 attacks me during its turn.
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.AttackWith(t, g, me.ID, bear)
	effectstest.AroundTheTableToMainOf(t, g, 0)
	effectstest.AttackWith(t, g, b.ID, kagachi)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(p.PickTargetCards, rock) || !effectstest.HasID(p.PickTargetCards, bear) {
		t.Error("the hit player's nonland permanents are offered")
	}
	if effectstest.HasID(p.PickTargetCards, land) || effectstest.HasID(p.PickTargetCards, otherRock) {
		t.Error("a land and another player's permanent are not offered")
	}
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the chosen permanent is exiled")
	}
	if !g.Exile.Contains(rock) {
		t.Error("the chosen permanent went to exile")
	}
}

func TestOKagachiDoesNothingToAPlayerWhoDidNotAttackYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	kagachi := effectstest.B12Push(g, me.ID, "O-Kagachi, Vengeful Kami", "Legendary Creature — Dragon Spirit", oKagachiOracle, 6, 6)
	bear := effectstest.PushVanillaCreature(g, b.ID, "Their Bear", 2, 2)
	rock := effectstest.B12Permanent(g, b.ID, "Their Rock", "Artifact")

	// Seat 1 attacks seat 2 during its turn, not me.
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.AttackWith(t, g, c.ID, bear)
	effectstest.AroundTheTableToMainOf(t, g, 0)
	effectstest.AttackWith(t, g, b.ID, kagachi)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("seat 1 did not attack me during its last turn: no trigger")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("nothing is exiled")
	}
}
