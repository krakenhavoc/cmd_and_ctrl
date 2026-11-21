package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b226_cards_test.go — Commander Legends: Battle for Baldur's Gate
// white creatures, Backgrounds and a Saga, GAP batch 226 (#5948). Each test
// plays the real catalog entry through the engine.

const (
	gap226DungeoneerOracle = "5ea675aa-88d6-4387-a14d-e606fa9c3aec"
	gap226SquadOracle      = "11f21b50-b4b6-449e-a66a-eedd2920ed93"
	gap226TabaxiOracle     = "36dd999e-53da-45c7-ae98-90f7aad78680"
	gap226ChangelingOracle = "65d9b8b6-5215-4037-8e50-e0ca80c5dc9e"
	gap226PatrolOracle     = "fad2d14f-8825-4b08-8804-010d9659f6dc"
	gap226PastOracle       = "54ad73d1-a17d-468c-a0ba-05c532a02cb2"
	gap226CultistOracle    = "bdc2a72a-f1c0-4e67-8e86-1121a4000537"
	gap226SigilOracle      = "ae7e14cf-a275-4260-8d78-324429a042b7"
	gap226FirjaOracle      = "fea32d4e-ceef-4954-a84d-e7e2509e4a13"
	gap226AcolyteOracle    = "38477748-bc57-449c-afe1-c5f34caf8eff"
)

func TestGapB226EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		gap226DungeoneerOracle: "Seasoned Dungeoneer",
		gap226SquadOracle:      "Squad Commander",
		gap226TabaxiOracle:     "Tabaxi Toucaneers",
		gap226ChangelingOracle: "Valiant Changeling",
		gap226PatrolOracle:     "Wyrm's Crossing Patrol",
		gap226PastOracle:       "Criminal Past",
		gap226CultistOracle:    "Cultist of the Absolute",
		gap226SigilOracle:      "Sigil of Myrkul",
		gap226FirjaOracle:      "Firja's Retribution",
		gap226AcolyteOracle:    "Acolyte of Bahamut",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Tabaxi Toucaneers, Wyrm's Crossing Patrol ---

func TestGapB226MyriadCreaturesCopyThemselvesForEachOtherOpponent(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		kw                     []string
	}{
		{"Tabaxi Toucaneers", "Creature — Cat Ranger", gap226TabaxiOracle, []string{"flying"}},
		{"Wyrm's Crossing Patrol", "Creature — Human Soldier", gap226PatrolOracle, nil},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
		effectstest.AssertKeywords(t, g, id, tc.kw...)
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(id, g.Seats[1].ID); err != nil {
			t.Fatalf("%s: DeclareAttacker: %v", tc.name, err)
		}
		effectstest.LockInAttacks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		for i := 0; i < 4; i++ {
			if effectstest.LatestConfirmFor(g, me.ID) == nil {
				break
			}
			effectstest.AnswerConfirmFor(t, g, me, true)
			effectstest.PassPriorityAroundTable(t, g)
		}
		if n := effectstest.CountOnBattlefield(g, tc.name, me.ID); n != 3 {
			t.Errorf("%s: %d on the battlefield after myriad, want 3 (the creature and a copy for each of two other opponents)", tc.name, n)
		}
	}
}

// --- Valiant Changeling ---

func TestGapB226ValiantChangelingCostsOneLessPerCreatureTypeUpToFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Creature(g, me.ID, "Elf Druid", "Creature — Elf Druid", 1, 1)
	effectstest.B12Creature(g, me.ID, "Elf Warrior", "Creature — Elf Warrior", 1, 1)
	// Elf, Druid, Warrior: three types, so {5}{W}{W} costs {2}{W}{W}.
	effectstest.B06AddMana(me, "W", "W", "C")
	if _, err := effectstest.CastFromHandErr(g, me, "Valiant Changeling", "Creature — Shapeshifter", "{5}{W}{W}", gap226ChangelingOracle,
		game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("three mana cast a spell that costs {2}{W}{W}")
	}
	effectstest.B06AddMana(me, "C")
	id, err := effectstest.CastFromHandErr(g, me, "Valiant Changeling", "Creature — Shapeshifter", "{5}{W}{W}", gap226ChangelingOracle,
		game.CastSpellParams{Strict: true})
	if err != nil {
		t.Fatalf("four mana should pay {2}{W}{W}: %v", err)
	}
	_ = id
	effectstest.PassPriorityAroundTable(t, g)
	got := effectstest.BattlefieldIDNamed(g, me.ID, "Valiant Changeling")
	effectstest.AssertKeywords(t, g, got, "double strike")
	if !effectstest.HasEffectiveKeyword(t, g, got, "changeling") {
		t.Error("no changeling")
	}
}

func TestGapB226ValiantChangelingCapIsFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Creature(g, me.ID, "Many", "Creature — Elf Druid Warrior Wizard Cleric Rogue Knight Soldier", 1, 1)
	effectstest.B06AddMana(me, "W", "W")
	if _, err := effectstest.CastFromHandErr(g, me, "Valiant Changeling", "Creature — Shapeshifter", "{5}{W}{W}", gap226ChangelingOracle,
		game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("{W}{W} should pay: eight types reduce {1} each but never more than {5}: %v", err)
	}
}

// --- Squad Commander ---

func TestGapB226SquadCommanderMakesATokenPerPartyMember(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, me.ID, "A Cleric", "Creature — Human Cleric", 1, 1)
	effectstest.B12Creature(g, me.ID, "A Wizard", "Creature — Human Wizard", 1, 1)
	effectstest.B12PlayFromHand(t, g, "Squad Commander", "Creature — Kor Warrior", gap226SquadOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	// Cleric, Wizard and the Commander's own Warrior: three.
	if n := effectstest.CountOnBattlefield(g, "Kor Warrior", me.ID); n != 3 {
		t.Errorf("%d Kor Warrior tokens, want 3", n)
	}
}

func TestGapB226SquadCommanderPumpsOnlyWithAFullParty(t *testing.T) {
	for _, full := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.B12Push(g, me.ID, "Squad Commander", "Creature — Kor Warrior", gap226SquadOracle, 3, 3)
		effectstest.B12Creature(g, me.ID, "A Cleric", "Creature — Human Cleric", 1, 1)
		bear := effectstest.B12Creature(g, me.ID, "A Bear", "Creature — Bear", 2, 2)
		if full {
			effectstest.B12Creature(g, me.ID, "A Rogue", "Creature — Human Rogue", 1, 1)
			effectstest.B12Creature(g, me.ID, "A Wizard", "Creature — Human Wizard", 1, 1)
		}
		effectstest.AdvanceTo(t, g, game.StepBeginCombat)
		effectstest.PassPriorityAroundTable(t, g)
		p, _ := effectstest.EffectivePTOf(t, g, bear)
		wantP, wantInd := 2, false
		if full {
			wantP, wantInd = 3, true
		}
		if p != wantP || effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") != wantInd {
			t.Errorf("full=%v: Bear has power %d, indestructible %v; want %d, %v", full, p, !wantInd, wantP, wantInd)
		}
	}
}

// --- Seasoned Dungeoneer ---

func TestGapB226SeasonedDungeoneerTakesTheInitiativeWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12PlayFromHand(t, g, "Seasoned Dungeoneer", "Creature — Human Warrior", gap226DungeoneerOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if !effects.YouHaveTheInitiative(g, me.ID) {
		t.Fatal("did not take the initiative")
	}
}

func TestGapB226SeasonedDungeoneerGivesAnAttackingPartyMemberProtectionAndExplores(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dungeoneer := effectstest.PushCatalogPermanent(g, me.ID, "Seasoned Dungeoneer", "Creature — Human Warrior", gap226DungeoneerOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "A Bear", "Creature — Bear", 2, 2)
	land := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: land, Name: "Explored Land", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(dungeoneer, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.HobAnswerTargets(t, g, me.ID, dungeoneer, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, dungeoneer, "protection from creatures") {
		t.Error("the attacking Warrior did not gain protection from creatures")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "protection from creatures") {
		t.Error("a non-attacking Bear gained protection")
	}
	if !me.Hand.Contains(land) {
		t.Error("the Dungeoneer's explore did not put the revealed land into hand")
	}
}

// --- Sigil of Myrkul ---

func TestGapB226SigilOfMyrkulBoostsOnlyWithFourCreatureCards(t *testing.T) {
	for _, creatures := range []int{2, 3} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.PushCatalogPermanent(g, me.ID, "Sigil of Myrkul", "Enchantment", gap226SigilOracle, false)
		bear := effectstest.B12Creature(g, me.ID, "A Bear", "Creature — Bear", 2, 2)
		for i := 0; i < creatures; i++ {
			effectstest.GraveyardCardOf(g, me, "Dead Bear", "Creature — Bear", "{1}", "")
		}
		// The milled card is a creature, making creatures+1 of them.
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Milled Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
		effectstest.AdvanceTo(t, g, game.StepBeginCombat)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.HobAnswerTargets(t, g, me.ID, bear)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if creatures+1 >= 4 {
			want = 1
		}
		if n := effectstest.BattlefieldPlusOnes(g, bear); n != want {
			t.Errorf("%d creature cards milled in total %d: Bear has %d counters, want %d", creatures+1, creatures+1, n, want)
		}
		if got := effectstest.HasEffectiveKeyword(t, g, bear, "deathtouch"); got != (want == 1) {
			t.Errorf("deathtouch = %v, want %v", got, want == 1)
		}
	}
}

// --- Firja's Retribution ---

func TestGapB226FirjasRetributionMakesAnAngelThenArmsAngels(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Firja's Retribution", "Enchantment — Saga", gap226FirjaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	angel := effectstest.BattlefieldIDNamed(g, me.ID, "Angel Warrior")
	if angel == uuid.Nil {
		t.Fatal("chapter I made no Angel Warrior")
	}
	effectstest.AssertKeywords(t, g, angel, "flying", "vigilance")
	if p, tough := effectstest.EffectivePTOf(t, g, angel); p != 4 || tough != 4 {
		t.Errorf("Angel is %d/%d, want 4/4", p, tough)
	}
	small := effectstest.PushVanillaCreature(g, opp.ID, "Small", 2, 2)
	big := effectstest.PushVanillaCreature(g, opp.ID, "Big", 5, 5)
	// Advance to my next turn: lore counter 2.
	effectstest.AdvanceToNextSeatsTurn(t, g)
	for i := 0; i < 3 && g.Turn.ActiveSeat != 0; i++ {
		effectstest.AdvanceToNextSeatsTurn(t, g)
	}
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, angel, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(big), Strict: true}); err == nil {
		t.Error("destroyed a creature with power not less than the Angel's")
	}
	if err := g.ActivateCatalogAbility(me.ID, angel, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(small), Strict: true}); err != nil {
		t.Fatalf("chapter II's granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the small creature was not destroyed")
	}
}

// --- Backgrounds ---

func TestGapB226CriminalPastGivesCommandersMenaceAndPumpPerCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B21Push(g, me.ID, "Criminal Past", "Legendary Enchantment — Background", gap226PastOracle, 0, 0, "B")
	commander := effectstest.B21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B21Commander(g, opp.ID, "Their Commander", "Legendary Creature — Human", 3, 3)
	effectstest.GraveyardCardOf(g, me, "Dead A", "Creature — Bear", "{1}", "")
	effectstest.GraveyardCardOf(g, me, "Dead B", "Creature — Bear", "{1}", "")
	effectstest.GraveyardCardOf(g, me, "Not a creature", "Sorcery", "{1}", "")
	if !effectstest.HasEffectiveKeyword(t, g, commander, "menace") {
		t.Error("your commander lacks menace")
	}
	if p, _ := effectstest.EffectivePTOf(t, g, commander); p != 5 {
		t.Errorf("your commander has power %d, want 5 (3, +2 for two creature cards)", p)
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "menace") || effectstest.HasEffectiveKeyword(t, g, theirs, "menace") {
		t.Error("a non-commander or an opponent's commander has menace")
	}
	if p, _ := effectstest.EffectivePTOf(t, g, theirs); p != 3 {
		t.Errorf("an opponent's commander has power %d, want 3", p)
	}
}

func TestGapB226CultistOfTheAbsoluteMakesAFlyingDeathtouchWardedCommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B21Push(g, me.ID, "Cultist of the Absolute", "Legendary Enchantment — Background", gap226CultistOracle, 0, 0, "B")
	commander := effectstest.B21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AssertKeywords(t, g, commander, "flying", "deathtouch")
	if p, tough := effectstest.EffectivePTOf(t, g, commander); p != 6 || tough != 6 {
		t.Errorf("commander is %d/%d, want 6/6", p, tough)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 2 {
		t.Errorf("a Bear is %d/x, not a commander", p)
	}
}

func TestGapB226CultistOfTheAbsoluteCommanderSacrificesACreatureEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B21Push(g, me.ID, "Cultist of the Absolute", "Legendary Enchantment — Background", gap226CultistOracle, 0, 0, "B")
	effectstest.B21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	for i := 0; i < 4 && g.Turn.ActiveSeat != 0; i++ {
		effectstest.AdvanceToNextSeatsTurn(t, g)
	}
	effectstest.AdvanceTo(t, g, game.StepUpkeep)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear was not sacrificed")
	}
}

func TestGapB226AcolyteOfBahamutDiscountsTheFirstDragonSpellOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B21Push(g, me.ID, "Acolyte of Bahamut", "Legendary Enchantment — Background", gap226AcolyteOracle, 0, 0, "G")
	effectstest.B21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
	effectstest.B06AddMana(me, "R", "C")
	if _, err := effectstest.CastFromHandErr(g, me, "Fire Dragon", "Creature — Dragon", "{3}{R}", "",
		game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the first Dragon spell should cost {1}{R}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AddMana(me, "R", "C")
	if _, err := effectstest.CastFromHandErr(g, me, "Another Dragon", "Creature — Dragon", "{3}{R}", "",
		game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a second Dragon spell was discounted too")
	}
}
