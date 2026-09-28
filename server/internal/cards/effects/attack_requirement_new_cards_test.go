package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attack_requirement_new_cards_test.go — #1599's proof cards: Zurgo
// Helmsmasher's stale caveat cleared, and the five cards built on
// #1595's CR 508.1d machinery (Goblin Rabblemaster, Legion Warboss,
// Grand Melee, Kardur, Doomscourge, Disrupt Decorum). The six more
// "one file, same constructors" cards are in goad_cards_test.go.

const (
	goblinRabblemasterOracle = "24661b81-6bee-4ad8-b3ab-53cb65005c51"
	legionWarbossOracle      = "0ccb9af3-6902-4130-a59c-4c882dc3a2fd"
	grandMeleeOracle         = "1accf98a-0905-4a9d-9ab3-72e9f853f4ab"
	kardurDoomscourgeOracle  = "bc14356c-3a1a-47af-9a6e-2b449de0331f"
	disruptDecorumOracle     = "9b084cd4-cc21-4f3a-9ac5-1e1160a31aa6"
)

// --- Zurgo Helmsmasher: the cleared caveat -----------------------

// TestZurgoMustAttackEachCombat is the caveat's own replacement: a
// stale "not enforced" note becomes an enforced CR 508.1d requirement,
// and this pins both halves — the refusal when he sits home, and the
// pass once he's declared.
func TestZurgoMustAttackEachCombat(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	zurgo := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Zurgo Helmsmasher", TypeLine: "Legendary Creature — Orc Warrior",
		OracleID: zurgoOracle, Power: 7, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})

	advanceToDeclareAttackersOf(t, g, seat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != zurgo {
		t.Fatalf("pass with Zurgo home = %v, want a refusal naming him", err)
	}
	if err := g.DeclareAttacker(zurgo, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker(Zurgo): %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with Zurgo attacking: %v", err)
	}
}

// --- Goblin Rabblemaster ------------------------------------------

// TestGoblinRabblemasterTokenMustAttackAndPumpsTheLord covers all
// three lines: the token the beginning-of-combat trigger makes is
// required to attack (the lord's own static, "OTHER Goblin creatures
// you control"), Rabblemaster itself is never required by that line,
// and Rabblemaster's own attack pumps it +1/+0 for the token.
func TestGoblinRabblemasterTokenMustAttackAndPumpsTheLord(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	rabble := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin Rabblemaster", TypeLine: "Creature — Goblin Warrior",
		OracleID: goblinRabblemasterOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})

	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)

	token := findBattlefieldByName(g, "Goblin")
	if token == uuid.Nil {
		t.Fatal("beginning of combat did not create a Goblin token")
	}

	advanceToDeclareAttackersOf(t, g, seat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != token {
		t.Fatalf("pass with the token home = %v, want a refusal naming it", err)
	}
	// Rabblemaster is never required to attack by its own "other
	// Goblins" line.
	if err := g.DeclareAttacker(token, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker(token): %v", err)
	}
	if err := g.DeclareAttacker(rabble, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker(Rabblemaster): %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, rabble); got != 3 {
		t.Errorf("Rabblemaster's power = %d, want 3 (2 printed +1 for the token)", got)
	}
}

// --- Legion Warboss -------------------------------------------------

// TestLegionWarbossTokenAttacksAndMentorFires: the token has haste
// and is required to attack this combat (a per-token scoped record,
// not a lord static — see the card comment); Legion Warboss attacking
// alongside it fires Mentor, putting a +1/+1 counter on the smaller
// token.
func TestLegionWarbossTokenAttacksAndMentorFires(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	warboss := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Legion Warboss", TypeLine: "Creature — Goblin Soldier",
		OracleID: legionWarbossOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})

	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)

	token := findBattlefieldByName(g, "Goblin")
	if token == uuid.Nil {
		t.Fatal("beginning of combat did not create a Goblin token")
	}
	if !hasEffectiveKeyword(t, g, token, "haste") {
		t.Fatal("the token lacks the haste this ability grants")
	}

	advanceToDeclareAttackersOf(t, g, seat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != token {
		t.Fatalf("pass with the token home = %v, want a refusal naming it", err)
	}
	if err := g.DeclareAttacker(token, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker(token): %v", err)
	}
	if err := g.DeclareAttacker(warboss, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker(Warboss): %v", err)
	}
	lockInAttacks(t, g)
	pickCard(t, g, me.ID, token)
	passPriorityAroundTable(t, g)

	if got := allCountersOn(t, g, token)["+1/+1"]; got != 1 {
		t.Errorf("token's +1/+1 counters = %d, want 1 from Mentor", got)
	}
}

// --- Grand Melee ------------------------------------------------

// TestGrandMeleeRequiresAThirdPartysCreatureToo: no "you control", no
// "other" — the attack half reaches every creature at the table,
// including ones the enchantment's controller doesn't own.
func TestGrandMeleeRequiresAThirdPartysCreatureToo(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	thirdSeat := (seat + 2) % len(g.Seats)
	third := g.Seats[thirdSeat]
	pushPermanentForTest(g, me.ID, "Grand Melee", grandMeleeOracle, "Enchantment")
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: third.ID, Controller: third.ID, Keywords: []string{"haste"},
	})

	advanceToDeclareAttackersOf(t, g, thirdSeat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != theirs {
		t.Fatalf("pass with their Bear home = %v, want a refusal naming it", err)
	}
	if err := g.DeclareAttacker(theirs, me.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with their Bear attacking: %v", err)
	}
}

// --- Kardur, Doomscourge ------------------------------------------

// TestKardurEntersMakesOpponentsAttackOtherThanHim: the ETB
// requirement reaches an opponent's creature and refuses a pass
// until it both attacks AND attacks someone other than Kardur's
// controller.
func TestKardurEntersMakesOpponentsAttackOtherThanHim(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	oppSeat := (seat + 1) % len(g.Seats)
	opp := g.Seats[oppSeat]
	third := g.Seats[(seat+2)%len(g.Seats)]
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID, Keywords: []string{"haste"},
	})

	castCatalogSpell(t, g, "Kardur, Doomscourge", "Legendary Creature — Demon Berserker", kardurDoomscourgeOracle, nil)
	passPriorityAroundTable(t, g)
	if !hasRequirementRecord(g) {
		t.Fatal("Kardur's ETB registered no requirement")
	}

	advanceToDeclareAttackersOf(t, g, oppSeat)
	if err := g.PassPriority(); err == nil {
		t.Fatal("pass with their Bear home succeeded, want a refusal")
	}
	if err := g.DeclareAttacker(bear, me.ID); err == nil {
		t.Fatal("attacking Kardur's controller succeeded, want ErrAttackRequirement's \"other than you\" half refused")
	}
	if err := g.DeclareAttacker(bear, third.ID); err != nil {
		t.Fatalf("DeclareAttacker(their Bear, third player): %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with their Bear attacking someone else: %v", err)
	}
}

// --- Disrupt Decorum ------------------------------------------------

// TestDisruptDecorumGoadsEveryOpponentsCreature: goads every creature
// the caster doesn't control and refuses a pass until each attacks
// someone other than the caster.
func TestDisruptDecorumGoadsEveryOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	oppSeat := (seat + 1) % len(g.Seats)
	opp := g.Seats[oppSeat]
	third := g.Seats[(seat+2)%len(g.Seats)]
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID, Keywords: []string{"haste"},
	})
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)

	castCatalogSpell(t, g, "Disrupt Decorum", "Sorcery", disruptDecorumOracle, nil)
	passPriorityAroundTable(t, g)

	if c, ok := g.LookupCardForEffect(bear); !ok || c.GoadedBy != me.ID {
		t.Fatal("Disrupt Decorum did not goad the opponent's creature")
	}
	if c, ok := g.LookupCardForEffect(mine); !ok || c.GoadedBy != uuid.Nil {
		t.Fatal("Disrupt Decorum goaded the caster's own creature")
	}

	advanceToDeclareAttackersOf(t, g, oppSeat)
	if err := g.DeclareAttacker(bear, me.ID); err == nil {
		t.Fatal("attacking the caster succeeded, want the goad's \"other than you\" half refused")
	}
	if err := g.PassPriority(); err == nil {
		t.Fatal("pass with the goaded Bear home succeeded, want a refusal")
	}
	if err := g.DeclareAttacker(bear, third.ID); err != nil {
		t.Fatalf("DeclareAttacker(goaded Bear, third player): %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the goaded Bear attacking someone else: %v", err)
	}
}
