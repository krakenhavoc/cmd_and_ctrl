package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b696_test.go — the eight Mercadian Masques enters-and-dies
// creatures (#5619), each played through the real catalog and the engine.

const (
	mm6Cutpurse   = "147e0e74-f76b-4f0a-951f-d39afdd72c02"
	mm6Griffin    = "6a58f2a6-ff5e-4823-b3a7-49df14d64ffe"
	mm6HiredGiant = "79c9a963-a428-45f1-a647-5a91ef64286e"
	mm6Chameleon  = "9ae547df-d280-41fb-81b4-24e52d60c25b"
	mm6Footpad    = "a59f147b-3e03-4e82-aa34-48c220ea32e6"
	mm6TimidDrake = "aa8c5467-190a-4c6a-872c-56b8f8d01a2b"
	mm6Brigand    = "ae2db472-c699-4c39-aa35-c359ac42df4e"
	mm6Megather   = "b98123b2-c913-41b9-bb24-10a287d24bca"
)

func TestMm6B696Registered(t *testing.T) {
	for oracle, name := range map[string]string{
		mm6Cutpurse: "Rishadan Cutpurse", mm6Griffin: "Charmed Griffin",
		mm6HiredGiant: "Hired Giant", mm6Chameleon: "Chameleon Spirit",
		mm6Footpad: "Rishadan Footpad", mm6TimidDrake: "Timid Drake",
		mm6Brigand: "Rishadan Brigand", mm6Megather: "Megatherium",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

func mm6Opponents(g *game.Game) (*game.Player, []*game.Player) {
	me := g.Seats[g.Turn.ActiveSeat]
	var opps []*game.Player
	for _, p := range g.Seats {
		if p != me {
			opps = append(opps, p)
		}
	}
	return me, opps
}

func TestMm6RishadanPiratesMakeEachOpponentPayOrSacrifice(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
		power, toughness       int
		cost                   int
	}{
		{"Rishadan Cutpurse", mm6Cutpurse, "Creature — Human Pirate", 1, 1, 1},
		{"Rishadan Footpad", mm6Footpad, "Creature — Human Pirate", 2, 2, 2},
		{"Rishadan Brigand", mm6Brigand, "Creature — Human Pirate", 3, 2, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me, opps := mm6Opponents(g)
			payer, decliner, empty := opps[0], opps[1], opps[2]
			keepA := effectstest.B12Creature(g, payer.ID, "Payer's Bear", "Creature — Bear", 2, 2)
			loseB := effectstest.B12Creature(g, decliner.ID, "Decliner's Bear", "Creature — Bear", 2, 2)
			mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
			pool := make([]string, tc.cost)
			for i := range pool {
				pool[i] = "C"
			}
			effectstest.ApaMana(payer, pool...)
			effectstest.CastRingCreature(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerPayUnless(t, g, payer.ID, true)
			effectstest.AnswerPayUnless(t, g, decliner.ID, false)
			effectstest.AnswerPayUnless(t, g, empty.ID, false)
			effectstest.AnswerSacrifice(t, g, decliner.ID, loseB)
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(keepA) {
				t.Error("an opponent who paid sacrificed anyway")
			}
			if g.Battlefield.Contains(loseB) {
				t.Error("an opponent who did not pay kept every permanent")
			}
			if !g.Battlefield.Contains(mine) {
				t.Error("the pirate's controller sacrificed a permanent")
			}
		})
	}
}

func TestMm6RishadanBrigandIsAFlyerThatBlocksOnlyFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opps := mm6Opponents(g)
	defender := opps[0]
	brigand := effectstest.PushCatalogPermanent(g, defender.ID, "Rishadan Brigand", "Creature — Human Pirate", mm6Brigand, false)
	effectstest.AssertKeywords(t, g, brigand, "flying")
	ground := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Ground Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Sky Bird", TypeLine: "Creature — Bird", Keywords: []string{"flying"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	refused := effectstest.Msc13BlockRefused(t, g, defender.ID, brigand, ground, flyer)
	if !refused[ground] {
		t.Error("the Brigand blocked a creature without flying")
	}
	if refused[flyer] {
		t.Error("the Brigand could not block a creature with flying")
	}
}

func TestMm6CharmedGriffinLetsEachOtherPlayerPutAnArtifactOrEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opps := mm6Opponents(g)
	opp := opps[0]
	art := effectstest.PushCatalogHandCard(opp, "Hand Rock", "Artifact", "")
	effectstest.PushCatalogHandCard(opp, "Hand Bear", "Creature — Bear", "")
	mine := effectstest.PushCatalogHandCard(me, "My Aura", "Enchantment", "")
	effectstest.CastRingCreature(t, g, "Charmed Griffin", "Creature — Griffin", mm6Griffin, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, art)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, art)
	if !ok || c.Controller != opp.ID {
		t.Fatal("the opponent's artifact card did not enter under their control")
	}
	if g.Battlefield.Contains(mine) {
		t.Error("the Griffin's controller put an enchantment in too")
	}
}

func TestMm6HiredGiantLetsEachOtherPlayerFetchALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opps := mm6Opponents(g)
	opp := opps[0]
	land := effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Test Land", TypeLine: "Land", Owner: opp.ID, Controller: opp.ID})
	mine := effectstest.PushLibraryCardForTest(me, game.Card{Name: "My Land", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	effectstest.CastRingCreature(t, g, "Hired Giant", "Creature — Giant", mm6HiredGiant, 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Fatal("the Giant's controller was offered a search")
	}
	effectstest.AnswerSearchNamed(t, g, opp.ID, "Test Land")
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, land)
	if !ok || c.Controller != opp.ID {
		t.Fatal("the fetched land did not enter under its owner's control")
	}
	if c.Tapped {
		t.Error("the land entered tapped")
	}
	if g.Battlefield.Contains(mine) {
		t.Error("the Giant's controller's land entered")
	}
}

func TestMm6ChameleonSpiritCountsTheChosenColourAmongOpponentsPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opps := mm6Opponents(g)
	opp := opps[0]
	red := func(owner uuid.UUID, name string) uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Goblin", Colors: []string{"R"}, Power: 1, Toughness: 1, Owner: owner, Controller: owner})
	}
	a, b := red(opp.ID, "Red A"), red(opps[1].ID, "Red B")
	red(me.ID, "My Red")
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Elf", TypeLine: "Creature — Elf", Colors: []string{"G"}, Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	spirit := effectstest.PushChosenColorPermanent(t, g, me.ID, "Chameleon Spirit", "Creature — Illusion Spirit", mm6Chameleon, "R")
	if p, tough := effectstest.EffectivePTOf(t, g, spirit); p != 2 || tough != 2 {
		t.Fatalf("Chameleon Spirit is %d/%d, want 2/2 (two red permanents opponents control)", p, tough)
	}
	effectstest.B14Kill(g, a)
	if p, tough := effectstest.EffectivePTOf(t, g, spirit); p != 1 || tough != 1 {
		t.Errorf("after one dies: %d/%d, want 1/1", p, tough)
	}
	effectstest.B14Kill(g, b)
	if p, tough := effectstest.EffectivePTOf(t, g, spirit); p != 0 || tough != 0 {
		t.Errorf("with none left: %d/%d, want 0/0", p, tough)
	}
}

func TestMm6TimidDrakeReturnsWhenAnotherCreatureEntersButNotItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	drake := effectstest.CastRingCreature(t, g, "Timid Drake", "Creature — Drake", mm6TimidDrake, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(drake) {
		t.Fatal("the Drake bounced itself on entering")
	}
	// Another player's creature entering returns it too.
	g2 := effectstest.NewCatalogGame(t)
	_, opps2 := mm6Opponents(g2)
	drake2 := effectstest.PushCatalogPermanent(g2, opps2[0].ID, "Timid Drake", "Creature — Drake", mm6TimidDrake, false)
	effectstest.CastRingCreature(t, g2, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(drake2) || !opps2[0].Hand.Contains(drake2) {
		t.Error("an opponent's creature entering did not return the Drake to its owner's hand")
	}
}

func TestMm6MegatheriumNeedsOneManaPerCardInHand(t *testing.T) {
	for _, tc := range []struct {
		name  string
		hand  int
		pay   bool
		stays bool
	}{
		{"pays for a full hand", 3, true, true},
		{"declines and is sacrificed", 3, false, false},
		{"empty hand costs nothing", 0, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me, _ := mm6Opponents(g)
			me.Hand.Cards = nil
			for i := 0; i < tc.hand; i++ {
				effectstest.PushCatalogHandCard(me, "Hand Card", "Sorcery", "")
			}
			if tc.pay {
				effectstest.ApaMana(me, "C", "C", "C")
			}
			id := effectstest.CastRingCreature(t, g, "Megatherium", "Creature — Beast", mm6Megather, 4, 4)
			effectstest.PassPriorityAroundTable(t, g)
			if tc.hand > 0 {
				effectstest.AnswerPayUnless(t, g, me.ID, tc.pay)
				effectstest.PassPriorityAroundTable(t, g)
			} else if effectstest.HasPayUnlessFor(g, me.ID) {
				t.Fatal("an empty hand was asked to pay")
			}
			if got := g.Battlefield.Contains(id); got != tc.stays {
				t.Errorf("on the battlefield = %v, want %v", got, tc.stays)
			}
		})
	}
}
