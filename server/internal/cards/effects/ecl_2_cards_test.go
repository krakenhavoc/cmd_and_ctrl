package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ecl_2_cards_test.go — Lorwyn Eclipsed, second batch (#4458): ten cards
// built from seams the catalog already has. Each test plays the real
// catalog entry through the engine and asserts what its text says.

const (
	ecl2AppealOracle       = "cea09261-41a2-43e5-bf43-c4097e6615c6"
	ecl2DefensesOracle     = "b9dee727-8ad8-42e0-93c6-5ef91d3f7309"
	ecl2PerfectionOracle   = "ecf574a5-6410-466d-a726-732134f904b2"
	ecl2BloodletterOracle  = "15febce7-8320-4591-9535-ef11c8adecc2"
	ecl2BileVialOracle     = "5a651557-ce08-46ba-aecd-0e4fbfdef3c0"
	ecl2BlightRotOracle    = "cf5b1cf9-dcd3-44d1-b8b2-338651ebb2a1"
	ecl2CursecrafterOracle = "5d58aefe-a3df-4e0a-b679-b967cd5fd2c6"
	ecl2PranksterOracle    = "1bcac664-835a-4474-8ca4-263801bde08e"
	ecl2AggressorOracle    = "ab32c671-7686-435b-a987-14a08f6790eb"
	ecl2BerserkerOracle    = "a08f69f9-de48-46d6-8b75-31886159cf45"

	ecl2Goblin = "Creature — Goblin Warrior"
)

func TestEcl2EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		ecl2AppealOracle:       "Appeal to Eirdu",
		ecl2DefensesOracle:     "Aquitect's Defenses",
		ecl2PerfectionOracle:   "Assert Perfection",
		ecl2BloodletterOracle:  "Barbed Bloodletter",
		ecl2BileVialOracle:     "Bile-Vial Boggart",
		ecl2BlightRotOracle:    "Blight Rot",
		ecl2CursecrafterOracle: "Boggart Cursecrafter",
		ecl2PranksterOracle:    "Boggart Prankster",
		ecl2AggressorOracle:    "Boldwyr Aggressor",
		ecl2BerserkerOracle:    "Boneclub Berserker",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// ecl2CastAt puts a catalog card in the active player's hand and casts it
// in the current step, with no mana check.
func ecl2CastAt(t *testing.T, g *game.Game, name, typeLine, oracle string, targets []game.TargetRef) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: me.ID, Controller: me.ID})
	return id, g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets})
}

// ecl2Bolt casts Lightning Bolt at a creature and lets it and whatever it
// triggers settle, short of a target prompt.
func ecl2Bolt(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetCard(id))
	passPriorityAroundTable(t, g)
}

// --- Appeal to Eirdu -----------------------------------------------------

func TestEcl2AppealToEirduConvokesAndPumpsTwoCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	tappers := pushTapCostSoldiers(g, me.ID, 4) // white: one pays the {W}
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	castFromHandForTest(t, g, me, "Appeal to Eirdu", "Instant", "{3}{W}", ecl2AppealOracle,
		game.CastSpellParams{Strict: true, TapIDs: tappers, Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b},
		}})
	for _, s := range tappers {
		if !tapCostTapped(g, s) {
			t.Error("a creature tapped to convoke is untapped")
		}
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if p, tt := ptOf(t, g, id); p != 4 || tt != 3 {
			t.Errorf("a target is %d/%d, want 4/3", p, tt)
		}
	}
	endTurn(t, g)
	if p, tt := ptOf(t, g, a); p != 2 || tt != 2 {
		t.Errorf("after the turn a target is %d/%d, want 2/2", p, tt)
	}
}

func TestEcl2AppealToEirduTakesOneTargetButNotThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	c := b12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Appeal to Eirdu", "Instant", ecl2AppealOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: c},
	}); err == nil {
		t.Fatal("three targets were accepted for one or two")
	}
	castCatalogSpell(t, g, "Appeal to Eirdu", "Instant", ecl2AppealOracle, b16TargetCard(c))
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, c); p != 4 || tt != 3 {
		t.Errorf("the one target is %d/%d, want 4/3", p, tt)
	}
}

// --- Aquitect's Defenses -------------------------------------------------

func TestEcl2AquitectsDefensesFlashesOnAndGrantsHexproofForTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepEnd)
	if _, err := ecl2CastAt(t, g, "Aquitect's Defenses", "Enchantment — Aura", ecl2DefensesOracle, b16TargetCard(theirs)); err == nil {
		t.Fatal("enchant creature you control: an opponent's creature was accepted")
	}
	aura, err := ecl2CastAt(t, g, "Aquitect's Defenses", "Enchantment — Aura", ecl2DefensesOracle, b16TargetCard(bear))
	if err != nil {
		t.Fatalf("flash: cast in the end step: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Fatal("the Aura did not resolve")
	}
	if p, tt := ptOf(t, g, bear); p != 3 || tt != 4 {
		t.Errorf("enchanted creature is %d/%d, want 3/4", p, tt)
	}
	if !hasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("the enters trigger gives hexproof")
	}
	endTurn(t, g)
	if hasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("hexproof lasts only until end of turn")
	}
	if p, tt := ptOf(t, g, bear); p != 3 || tt != 4 {
		t.Errorf("the +1/+2 stays: %d/%d, want 3/4", p, tt)
	}
}

// --- Assert Perfection ---------------------------------------------------

func TestEcl2AssertPerfectionPumpsThenDealsItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	castCatalogSpell(t, g, "Assert Perfection", "Sorcery", ecl2PerfectionOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: theirs, Slot: 1},
	})
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, mine); p != 3 || tt != 2 {
		t.Errorf("your creature is %d/%d, want 3/2", p, tt)
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the 3/3 took 3 (2 power plus 1) and should have died")
	}
	if damageMarkedOn(g, mine) != 0 {
		t.Error("it is one-sided: your creature takes nothing")
	}
}

func TestEcl2AssertPerfectionWithoutASecondTargetOnlyPumps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	if err := castCatalogSpellErr(t, g, "Assert Perfection", "Sorcery", ecl2PerfectionOracle, b16TargetCard(theirs)); err == nil {
		t.Fatal("the first target must be a creature you control")
	}
	castCatalogSpell(t, g, "Assert Perfection", "Sorcery", ecl2PerfectionOracle, b16TargetCard(mine))
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, mine); p != 3 {
		t.Errorf("your creature has power %d, want 3", p)
	}
	if damageMarkedOn(g, theirs) != 0 {
		t.Error("no second target: nothing is dealt damage")
	}
}

// --- Barbed Bloodletter --------------------------------------------------

func TestEcl2BarbedBloodletterAttachesAndItsCreatureWithers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	other := pushVanillaCreature(g, me.ID, "Other Bear", 2, 2)
	wall := pushVanillaCreature(g, opp.ID, "Wall", 0, 6)
	blade := castCatalogSpell(t, g, "Barbed Bloodletter", "Artifact — Equipment", ecl2BloodletterOracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if hasID(pick.PickTargetCards, wall) {
		t.Error("an opponent's creature is not a creature you control")
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if c := findBattlefieldCardForTest(g, blade); c == nil || c.AttachedTo.ID != bear {
		t.Fatal("the Bloodletter is not attached to the target")
	}
	if p, tt := ptOf(t, g, bear); p != 3 || tt != 4 {
		t.Errorf("equipped creature is %d/%d, want 3/4", p, tt)
	}
	if !hasEffectiveKeyword(t, g, bear, game.KeywordWither) {
		t.Fatal("the target gains wither")
	}
	if hasEffectiveKeyword(t, g, other, game.KeywordWither) {
		t.Error("only the target gains wither")
	}

	pr7aBlockedCombat(t, g, opp.ID, bear, wall)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, wall, game.CounterMinusOne); got != 3 {
		t.Errorf("wither: the blocker has %d -1/-1 counters, want 3", got)
	}
	if damageMarkedOn(g, wall) != 0 {
		t.Error("wither deals its damage as counters, not marked damage")
	}

	endTurn(t, g)
	if hasEffectiveKeyword(t, g, bear, game.KeywordWither) {
		t.Error("wither lasts only until end of turn")
	}
	if p, tt := ptOf(t, g, bear); p != 3 || tt != 4 {
		t.Errorf("the +1/+2 stays while equipped: %d/%d, want 3/4", p, tt)
	}
}

func TestEcl2BarbedBloodletterHasFlashAndEquipTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	advanceTo(t, g, game.StepEnd)
	blade, err := ecl2CastAt(t, g, "Barbed Bloodletter", "Artifact — Equipment", ecl2BloodletterOracle, nil)
	if err != nil {
		t.Fatalf("flash: cast in the end step: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(blade) {
		t.Fatal("the Bloodletter did not resolve in the end step")
	}
	if latestPickTarget(g, me.ID) != nil {
		pickCard(t, g, me.ID, bear)
	}
	passPriorityAroundTable(t, g)
	if c := findBattlefieldCardForTest(g, blade); c == nil || c.AttachedTo.ID != bear {
		t.Fatal("the creature you control was not attached to")
	}

	spec, _ := Lookup(ecl2BloodletterOracle)
	if len(spec.Activated) != 1 || spec.Activated[0].Label != "Equip {2}" || !spec.Activated[0].Equip {
		t.Fatalf("activated abilities %+v, want Equip {2}", spec.Activated)
	}
}

// --- Bile-Vial Boggart ---------------------------------------------------

func TestEcl2BileVialBoggartPutsACounterWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	boggart := b12Push(g, me.ID, "Bile-Vial Boggart", "Creature — Goblin Assassin", ecl2BileVialOracle, 1, 1)
	theirs := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	b12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1)
	ecl2Bolt(t, g, boggart)
	if g.Battlefield.Contains(boggart) {
		t.Fatal("the Boggart survived a Bolt")
	}
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, theirs, game.CounterMinusOne); got != 1 {
		t.Errorf("-1/-1 counters on the target = %d, want 1", got)
	}
	if p, tt := ptOf(t, g, theirs); p != 2 || tt != 2 {
		t.Errorf("the target is %d/%d, want 2/2", p, tt)
	}
}

func TestEcl2BileVialBoggartMayTargetNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	boggart := b12Push(g, me.ID, "Bile-Vial Boggart", "Creature — Goblin Assassin", ecl2BileVialOracle, 1, 1)
	a := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	b := b12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1)
	ecl2Bolt(t, g, boggart)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the dies trigger")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
		t.Fatalf("up to one: choosing no target was refused: %v", err)
	}
	passPriorityAroundTable(t, g)
	if countersOn(g, a, game.CounterMinusOne) != 0 || countersOn(g, b, game.CounterMinusOne) != 0 {
		t.Error("no target was chosen, yet a creature got a counter")
	}
}

// --- Blight Rot ----------------------------------------------------------

func TestEcl2BlightRotPutsFourCountersOnTargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	giant := b12Creature(g, opp.ID, "Giant", "Creature — Giant", 6, 6)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Blight Rot", "Instant", ecl2BlightRotOracle, b16TargetCard(giant))
	passPriorityAroundTable(t, g)
	if got := countersOn(g, giant, game.CounterMinusOne); got != 4 {
		t.Errorf("-1/-1 counters = %d, want 4", got)
	}
	if p, tt := ptOf(t, g, giant); p != 2 || tt != 2 {
		t.Errorf("the 6/6 is %d/%d, want 2/2", p, tt)
	}
	castCatalogSpell(t, g, "Blight Rot", "Instant", ecl2BlightRotOracle, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("a 2/2 with four -1/-1 counters dies")
	}
	endTurn(t, g)
	if p, tt := ptOf(t, g, giant); p != 2 || tt != 2 {
		t.Errorf("the counters stay after the turn: %d/%d, want 2/2", p, tt)
	}
}

// --- Boggart Cursecrafter ------------------------------------------------

func TestEcl2BoggartCursecrafterPingsEachOpponentWhenAnotherGoblinDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crafter := b12Push(g, me.ID, "Boggart Cursecrafter", "Creature — Goblin Warlock", ecl2CursecrafterOracle, 2, 3)
	if !hasEffectiveKeyword(t, g, crafter, "deathtouch") {
		t.Error("deathtouch is printed")
	}
	goblin := b12Creature(g, me.ID, "Goblin", ecl2Goblin, 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirGoblin := b12Creature(g, opp.ID, "Their Goblin", ecl2Goblin, 1, 1)
	lives := func() []int {
		out := make([]int, len(g.Seats))
		for i, p := range g.Seats {
			out[i] = p.Life
		}
		return out
	}

	start := lives()
	ecl2Bolt(t, g, bear)
	ecl2Bolt(t, g, theirGoblin)
	if got := lives(); got[1] != start[1] || got[2] != start[2] {
		t.Fatalf("a non-Goblin of yours and an opponent's Goblin died: lives %v → %v", start, got)
	}

	ecl2Bolt(t, g, goblin)
	after := lives()
	for i := 1; i < len(g.Seats); i++ {
		if after[i] != start[i]-1 {
			t.Errorf("opponent at seat %d: life %d → %d, want 1 damage", i, start[i], after[i])
		}
	}
	if after[0] != start[0] {
		t.Errorf("you took damage: %d → %d", start[0], after[0])
	}

	ecl2Bolt(t, g, crafter)
	if got := lives(); got[1] != after[1] {
		t.Errorf("its own death is not another Goblin's: %d → %d", after[1], got[1])
	}
}

// --- Boggart Prankster ---------------------------------------------------

func TestEcl2BoggartPranksterPumpsAnAttackingGoblinYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	prankster := b12Push(g, me.ID, "Boggart Prankster", "Creature — Goblin Warrior", ecl2PranksterOracle, 1, 3)
	a := b12Creature(g, me.ID, "Goblin A", ecl2Goblin, 1, 1)
	b := b12Creature(g, me.ID, "Goblin B", ecl2Goblin, 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, a, b, bear)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt: whenever you attack")
	}
	if hasID(pick.PickTargetCards, bear) || hasID(pick.PickTargetCards, prankster) {
		t.Error("a non-Goblin attacker and a Goblin that isn't attacking are not legal targets")
	}
	if !hasID(pick.PickTargetCards, a) || !hasID(pick.PickTargetCards, b) {
		t.Error("both attacking Goblins are legal targets")
	}
	pickCard(t, g, me.ID, a)
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, a); p != 2 {
		t.Errorf("the target has power %d, want 2", p)
	}
	if p, _ := ptOf(t, g, b); p != 1 {
		t.Errorf("one trigger per attack, not per attacker: the other Goblin has power %d", p)
	}
	if triggersOnStackFrom(g, prankster) != 0 || latestPickTarget(g, me.ID) != nil {
		t.Error("a second trigger was raised for the same attack")
	}
}

func TestEcl2BoggartPranksterDoesNothingWithoutAnAttackingGoblin(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Boggart Prankster", "Creature — Goblin Warrior", ecl2PranksterOracle, 1, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, bear)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("no attacking Goblin: the trigger has no target and is removed")
	}
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, bear); p != 2 {
		t.Errorf("the Bear has power %d, want 2", p)
	}
}

// --- Boldwyr Aggressor ---------------------------------------------------

func TestEcl2BoldwyrAggressorGivesOtherGiantsYouControlDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aggressor := b12Push(g, me.ID, "Boldwyr Aggressor", "Creature — Giant Warrior", ecl2AggressorOracle, 2, 5)
	giant := b12Creature(g, me.ID, "Giant", "Creature — Giant", 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirGiant := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 3, 3)
	if !hasEffectiveKeyword(t, g, aggressor, "double strike") {
		t.Error("its own double strike is printed")
	}
	if !hasEffectiveKeyword(t, g, giant, "double strike") {
		t.Error("another Giant you control has double strike")
	}
	if hasEffectiveKeyword(t, g, bear, "double strike") || hasEffectiveKeyword(t, g, theirGiant, "double strike") {
		t.Error("a non-Giant and an opponent's Giant get nothing")
	}
	before := opp.Life
	attackWith(t, g, opp.ID, giant)
	if opp.Life != before-6 {
		t.Errorf("a 3/3 Giant with double strike deals 6 unblocked: %d → %d", before, opp.Life)
	}
}

// --- Boneclub Berserker --------------------------------------------------

func TestEcl2BoneclubBerserkerGetsTwoPowerPerOtherGoblinYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	berserker := b12Push(g, me.ID, "Boneclub Berserker", "Creature — Goblin Berserker", ecl2BerserkerOracle, 2, 4)
	if p, tt := ptOf(t, g, berserker); p != 2 || tt != 4 {
		t.Fatalf("alone it is %d/%d, want 2/4", p, tt)
	}
	a := b12Creature(g, me.ID, "Goblin A", ecl2Goblin, 1, 1)
	b12Creature(g, me.ID, "Goblin B", ecl2Goblin, 1, 1)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Goblin", ecl2Goblin, 1, 1)
	if p, tt := ptOf(t, g, berserker); p != 6 || tt != 4 {
		t.Errorf("two other Goblins of yours: %d/%d, want 6/4", p, tt)
	}
	ecl2Bolt(t, g, a)
	if p, _ := ptOf(t, g, berserker); p != 4 {
		t.Errorf("one other Goblin left: power %d, want 4", p)
	}
}
