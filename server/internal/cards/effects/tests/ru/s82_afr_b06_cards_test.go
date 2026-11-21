package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_afr_b06_cards_test.go — Adventures in the Forgotten Realms batch 6
// (#5499): simple Equipment and "is equipped" creatures.

const (
	afrB06LongbowOracle   = "4b9abe7f-050f-4d0a-bc01-f442164f95de"
	afrB06GreataxeOracle  = "843f3373-b8ef-416c-9163-aa1aa6b4a51d"
	afrB06MaceOracle      = "629fe1be-272d-465f-b9b1-2ce177410f13"
	afrB06BootsOracle     = "6f6601bc-cb12-4054-9027-e4a0ebdae1a3"
	afrB06ShieldOracle    = "dc548b03-ef5f-40df-8311-35d4cf26f3c5"
	afrB06RapierOracle    = "7ceec1fd-4af5-40d4-9d17-424e21889d64"
	afrB06LeatherOracle   = "7a6ce270-c261-422f-845f-11c22602a274"
	afrB06PlateOracle     = "36082b5e-760b-4245-af59-00d308470d1a"
	afrB06DaggerOracle    = "d895df97-e7e7-4f15-a79b-0ce8f1a252d3"
	afrB06ReaperOracle    = "bc2ea8c8-5360-4256-91c5-d821b5e59c0f"
	afrB06ChampionOracle  = "3db19689-73bb-48b2-a464-63615eb6bba0"
	afrB06VeteranOracle   = "2b51bbb3-1514-4a79-861f-b9cec9f0ec1b"
	afrB06TricksterOracle = "17469f58-286e-4410-b717-5908d24eecdd"
)

func TestAfrB06EveryCardIsFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		afrB06LongbowOracle: "Ranger's Longbow", afrB06GreataxeOracle: "Greataxe", afrB06MaceOracle: "+2 Mace",
		afrB06BootsOracle: "Boots of Speed", afrB06ShieldOracle: "Paladin's Shield", afrB06RapierOracle: "Dueling Rapier",
		afrB06LeatherOracle: "Leather Armor", afrB06PlateOracle: "Plate Armor", afrB06DaggerOracle: "Spare Dagger",
		afrB06ReaperOracle: "Reaper's Talisman", afrB06ChampionOracle: "Dwarfhold Champion",
		afrB06VeteranOracle: "Armory Veteran", afrB06TricksterOracle: "Trickster's Talisman",
	} {
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

func TestAfrB06PlainEquipmentStatsKeywordsAndEquipCosts(t *testing.T) {
	cases := []struct {
		name, oracle, equip string
		power, toughness    int
		keyword             string
	}{
		{"Ranger's Longbow", afrB06LongbowOracle, "{3}", 4, 3, "reach"},
		{"Greataxe", afrB06GreataxeOracle, "{5}", 6, 2, ""},
		{"+2 Mace", afrB06MaceOracle, "{3}", 4, 4, ""},
		{"Boots of Speed", afrB06BootsOracle, "{1}", 3, 2, "haste"},
	}
	for _, c := range cases {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		bear := effectstest.SeedBear(g, me.ID)
		eq := effectstest.SeedEquipment(g, me.ID, c.name, c.oracle)
		if got := effectstest.EquipPrice(t, g, me.ID, eq, bear); got != c.equip {
			t.Errorf("%s: equip costs %s, want %s", c.name, got, c.equip)
		}
		if err := effectstest.StrictEquip(t, g, me.ID, eq, bear, ""); err == nil {
			t.Errorf("%s: equipped with an empty pool", c.name)
		}
		g2 := effectstest.NewCatalogGame(t)
		me2 := g2.Seats[0]
		effectstest.AdvanceToMain(t, g2)
		bear2 := effectstest.SeedBear(g2, me2.ID)
		eq2 := effectstest.SeedEquipment(g2, me2.ID, c.name, c.oracle)
		effectstest.EquipTo(t, g2, me2.ID, eq2, bear2)
		if p, tt := effectstest.EffectivePower(t, g2, bear2), effectstest.EffectiveToughness(t, g2, bear2); p != c.power || tt != c.toughness {
			t.Errorf("%s: equipped Bear is %d/%d, want %d/%d", c.name, p, tt, c.power, c.toughness)
		}
		if c.keyword != "" && !effectstest.EffectiveAbilitiesContain(t, g2, bear2, c.keyword) {
			t.Errorf("%s: equipped Bear lacks %s", c.name, c.keyword)
		}
		other := effectstest.SeedBear(g2, me2.ID)
		if c.keyword != "" && effectstest.EffectiveAbilitiesContain(t, g2, other, c.keyword) {
			t.Errorf("%s: an unequipped Bear has %s", c.name, c.keyword)
		}
	}
}

func TestAfrB06FlashEquipmentAttachesOnEntry(t *testing.T) {
	for _, c := range []struct {
		name, oracle string
		power, tough int
	}{
		{"Paladin's Shield", afrB06ShieldOracle, 2, 4},
		{"Dueling Rapier", afrB06RapierOracle, 4, 2},
	} {
		spec, _ := effects.Lookup(c.oracle)
		if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
			t.Errorf("%s: printed keywords %v, want flash", c.name, spec.PrintedKeywords)
		}
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		bear := effectstest.SeedBear(g, me.ID)
		other := effectstest.SeedBear(g, me.ID)
		eq := effectstest.CastCatalogSpell(t, g, c.name, "Artifact — Equipment", c.oracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.CardRefs(bear)...)
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.AttachedTo(t, g, eq, bear) {
			t.Fatalf("%s did not attach to the chosen creature", c.name)
		}
		if p, tt := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != c.power || tt != c.tough {
			t.Errorf("%s: equipped Bear is %d/%d, want %d/%d", c.name, p, tt, c.power, c.tough)
		}
		if p := effectstest.EffectivePower(t, g, other); p != 2 {
			t.Errorf("%s: the other Bear has power %d", c.name, p)
		}
	}
}

func TestAfrB06LeatherArmorWardsAndEquipsForFreeOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	other := effectstest.SeedBear(g, me.ID)
	armor := effectstest.SeedEquipment(g, me.ID, "Leather Armor", afrB06LeatherOracle)

	if got := effectstest.EquipPrice(t, g, me.ID, armor, bear); got != "{0}" {
		t.Fatalf("equip costs %s, want {0}", got)
	}
	if err := effectstest.StrictEquip(t, g, me.ID, armor, bear, ""); err != nil {
		t.Fatalf("free equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectiveToughness(t, g, bear); got != 3 {
		t.Fatalf("toughness %d, want 3", got)
	}
	if err := effectstest.StrictEquip(t, g, me.ID, armor, other, ""); err == nil {
		t.Fatal("equipped a second time in the same turn")
	}
	if !effectstest.AttachedTo(t, g, armor, bear) {
		t.Fatal("the refused equip moved the Armor")
	}

	effectstest.CastAtWardedCreature(t, g, opp, bear)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward prompt for a spell targeting the equipped creature")
	}
}

func TestAfrB06PlateArmorEquipCostsLessPerOtherEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	plate := effectstest.SeedEquipment(g, me.ID, "Plate Armor", afrB06PlateOracle)

	if got := effectstest.EquipPrice(t, g, me.ID, plate, bear); got != "{3}" {
		t.Fatalf("alone, equip costs %s, want {3}", got)
	}
	axe := effectstest.SeedEquipment(g, me.ID, "Greataxe", afrB06GreataxeOracle)
	if got := effectstest.EquipPrice(t, g, me.ID, plate, bear); got != "{2}" {
		t.Errorf("with one other Equipment, equip costs %s, want {2}", got)
	}
	if got := effectstest.EquipPrice(t, g, me.ID, axe, bear); got != "{5}" {
		t.Errorf("Plate Armor discounted another Equipment's equip: %s, want {5}", got)
	}
	effectstest.SeedEquipment(g, me.ID, "Greataxe", afrB06GreataxeOracle)
	effectstest.SeedEquipment(g, me.ID, "Greataxe", afrB06GreataxeOracle)
	effectstest.SeedEquipment(g, me.ID, "Greataxe", afrB06GreataxeOracle)
	if got := effectstest.EquipPrice(t, g, me.ID, plate, bear); got != "{0}" {
		t.Errorf("with four other Equipment, equip costs %s, want {0}", got)
	}
	opp := g.Seats[1]
	effectstest.SeedEquipment(g, opp.ID, "Greataxe", afrB06GreataxeOracle)

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	effectstest.AdvanceToMain(t, g2)
	bear2 := effectstest.SeedBear(g2, me2.ID)
	plate2 := effectstest.SeedEquipment(g2, me2.ID, "Plate Armor", afrB06PlateOracle)
	effectstest.SeedEquipment(g2, opp2.ID, "Greataxe", afrB06GreataxeOracle)
	if got := effectstest.EquipPrice(t, g2, me2.ID, plate2, bear2); got != "{3}" {
		t.Errorf("an opponent's Equipment discounted the equip: %s, want {3}", got)
	}
	if err := effectstest.StrictEquip(t, g2, me2.ID, plate2, bear2, "{C}{C}"); err == nil {
		t.Error("equipped for two mana with no discount")
	}
	g2.Seats[0].ManaPool = game.ManaPool{}
	if err := effectstest.StrictEquip(t, g2, me2.ID, plate2, bear2, "{C}{C}{C}"); err != nil {
		t.Fatalf("equip for {3}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if p, tt := effectstest.EffectivePower(t, g2, bear2), effectstest.EffectiveToughness(t, g2, bear2); p != 5 || tt != 5 {
		t.Errorf("equipped Bear is %d/%d, want 5/5", p, tt)
	}
}

func TestAfrB06PlateArmorWardsTheEquippedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	plate := effectstest.SeedEquipment(g, me.ID, "Plate Armor", afrB06PlateOracle)
	effectstest.EquipTo(t, g, me.ID, plate, bear)
	effectstest.CastAtWardedCreature(t, g, opp, bear)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward prompt for a spell targeting the equipped creature")
	}
}

func TestAfrB06EquippedCreaturesGainTheirBonus(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	champion := effectstest.B12Push(g, me.ID, "Dwarfhold Champion", "Creature — Dwarf Warrior", afrB06ChampionOracle, 3, 1)
	veteran := effectstest.B12Push(g, me.ID, "Armory Veteran", "Creature — Orc Warrior", afrB06VeteranOracle, 2, 2)
	if got := effectstest.EffectiveToughness(t, g, champion); got != 1 {
		t.Fatalf("unequipped Champion toughness %d, want 1", got)
	}
	if effectstest.EffectiveAbilitiesContain(t, g, veteran, "menace") {
		t.Fatal("unequipped Veteran has menace")
	}
	mace := effectstest.SeedEquipment(g, me.ID, "+2 Mace", afrB06MaceOracle)
	effectstest.EquipTo(t, g, me.ID, mace, champion)
	if p, tt := effectstest.EffectivePower(t, g, champion), effectstest.EffectiveToughness(t, g, champion); p != 5 || tt != 5 {
		t.Errorf("equipped Champion is %d/%d, want 5/5 (+2/+2 and +0/+2)", p, tt)
	}
	axe := effectstest.SeedEquipment(g, me.ID, "Greataxe", afrB06GreataxeOracle)
	effectstest.EquipTo(t, g, me.ID, axe, veteran)
	if !effectstest.EffectiveAbilitiesContain(t, g, veteran, "menace") {
		t.Error("equipped Veteran lacks menace")
	}
	effectstest.EquipTo(t, g, me.ID, axe, champion)
	if effectstest.EffectiveAbilitiesContain(t, g, veteran, "menace") {
		t.Error("Veteran kept menace after the Equipment moved away")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, champion, "menace") {
		t.Error("Champion gained menace")
	}
}

func afrB06ReaperScene(t *testing.T, withSecondAttacker bool) (g *game.Game, me, opp *game.Player, bear uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear = effectstest.SeedBear(g, me.ID)
	other := effectstest.SeedBear(g, me.ID)
	talisman := effectstest.SeedEquipment(g, me.ID, "Reaper's Talisman", afrB06ReaperOracle)
	effectstest.EquipTo(t, g, me.ID, talisman, bear)
	if effectstest.EffectiveAbilitiesContain(t, g, bear, "deathtouch") {
		t.Fatal("deathtouch before attacking")
	}
	attackers := []uuid.UUID{bear}
	if withSecondAttacker {
		attackers = append(attackers, other)
	}
	effectstest.DeclareAttack(t, g, opp.ID, attackers...)
	effectstest.SettleOrderingTriggers(t, g)
	return g, me, opp, bear
}

func TestAfrB06ReapersTalismanGivesDeathtouchAndDrainsWhenAlone(t *testing.T) {
	g, me, opp, bear := afrB06ReaperScene(t, false)
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "deathtouch") {
		t.Error("the attacker did not gain deathtouch")
	}
	if opp.Life != 38 || me.Life != 42 {
		t.Errorf("life: opponent %d, you %d; want 38 and 42", opp.Life, me.Life)
	}
}

func TestAfrB06ReapersTalismanDoesNotDrainWhenNotAlone(t *testing.T) {
	g, me, opp, bear := afrB06ReaperScene(t, true)
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "deathtouch") {
		t.Error("the attacker did not gain deathtouch")
	}
	if opp.Life != 40 || me.Life != 40 {
		t.Errorf("life: opponent %d, you %d; want 40 and 40", opp.Life, me.Life)
	}
}

func afrB06DaggerScene(t *testing.T) (g *game.Game, me, opp *game.Player, bear, dagger uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear = effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	dagger = effectstest.SeedEquipment(g, me.ID, "Spare Dagger", afrB06DaggerOracle)
	effectstest.EquipTo(t, g, me.ID, dagger, bear)
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Fatalf("equipped power %d, want 3", got)
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, opp, bear, dagger
}

func TestAfrB06SpareDaggerSacrificesItselfToDealOne(t *testing.T) {
	g, me, opp, bear, dagger := afrB06DaggerScene(t)
	life := opp.Life
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	if g.Battlefield.Contains(dagger) {
		t.Fatal("the Dagger was not sacrificed")
	}
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 1 {
		t.Errorf("the opponent lost %d life, want 1", got)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("power %d once the Dagger is gone, want 2", got)
	}
}

func TestAfrB06SpareDaggerDecliningKeepsIt(t *testing.T) {
	g, me, opp, _, dagger := afrB06DaggerScene(t)
	life := opp.Life
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dagger) {
		t.Error("the Dagger left after declining")
	}
	if opp.Life != life {
		t.Errorf("life %d -> %d after declining", life, opp.Life)
	}
}

func afrB06TricksterScene(t *testing.T) (g *game.Game, me *game.Player, bear, talisman uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear = effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	talisman = effectstest.SeedEquipment(g, me.ID, "Trickster's Talisman", afrB06TricksterOracle)
	effectstest.EquipTo(t, g, me.ID, talisman, bear)
	if p, tt := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tt != 3 {
		t.Fatalf("equipped Bear is %d/%d, want 3/3", p, tt)
	}
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, bear, talisman
}

func TestAfrB06TricksterTalismanSacrificesForATokenCopy(t *testing.T) {
	g, me, bear, talisman := afrB06TricksterScene(t)
	if n := effectstest.CountNamed(g, "Bear"); n != 1 {
		t.Fatalf("%d Bears before the answer, want 1", n)
	}
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(talisman) {
		t.Fatal("the Talisman was not sacrificed")
	}
	if n := effectstest.CountNamed(g, "Bear"); n != 2 {
		t.Fatalf("%d Bears, want the original and one token copy", n)
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Bear" && c.InstanceID != bear {
			if !c.IsToken() {
				t.Error("the copy is not a token")
			}
			if got := effectstest.EffectivePower(t, g, c.InstanceID); got != 2 {
				t.Errorf("the copy has power %d, want the printed 2", got)
			}
		}
	}
}

func TestAfrB06TricksterTalismanDecliningMakesNoCopy(t *testing.T) {
	g, me, _, talisman := afrB06TricksterScene(t)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(talisman) {
		t.Error("the Talisman left after declining")
	}
	if n := effectstest.CountNamed(g, "Bear"); n != 1 {
		t.Errorf("%d Bears after declining, want 1", n)
	}
}
