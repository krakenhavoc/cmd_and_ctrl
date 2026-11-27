package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b709_cards_test.go — OLD batch 709 (Portal Three Kingdoms, #5622):
// the ten horsemanship creatures. Each test plays the real catalog entry
// through the engine.

const (
	oldB709MaChao    = "023e26d8-b67d-43dd-89b1-8905b1183094"
	oldB709Zhao      = "22aa8e7f-b30d-49e2-b99e-9232e1a3bcf3"
	oldB709YuanShao  = "2fc5fffa-2eb1-4a3c-af47-3ff8c7df608d"
	oldB709SunCe     = "93b0185c-b86f-4137-bc6d-96212eee1708"
	oldB709CaoRen    = "99cbc054-2757-4c07-ad04-9ac67dad4abc"
	oldB709Raiders   = "d24ee48a-86a9-4a72-9311-79b6ea3f41f2"
	oldB709WuScout   = "ef577d7d-7a24-4bd3-b86b-f69e37bb2ce7"
	oldB709LiuBei    = "f46416cc-0c3d-4d90-bdcb-910c5d4a2014"
	oldB709GuanYu    = "f7ee3935-4510-485a-a863-ac38dffa1908"
	oldB709ZhangHe   = "fe35b809-7439-4bdc-bce2-42262095064f"
	oldB709ZhangFei  = "f3f815b8-75c6-40e0-9716-872f0f5c9d46"
	oldB709CreatureT = "Creature — Human Soldier"
)

func TestOldB709EveryCardIsRegisteredFullWithHorsemanship(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB709MaChao: "Ma Chao, Western Warrior", oldB709Zhao: "Zhao Zilong, Tiger General",
		oldB709YuanShao: "Yuan Shao, the Indecisive", oldB709SunCe: "Sun Ce, Young Conquerer",
		oldB709CaoRen: "Cao Ren, Wei Commander", oldB709Raiders: "Wei Night Raiders",
		oldB709WuScout: "Wu Scout", oldB709LiuBei: "Liu Bei, Lord of Shu",
		oldB709GuanYu: "Guan Yu, Sainted Warrior", oldB709ZhangHe: "Zhang He, Wei General",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name=%q completeness=%v caveats=%v, want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
		if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "horsemanship" {
			t.Errorf("%s: printed keywords %v, want [horsemanship]", name, spec.PrintedKeywords)
		}
	}
}

func oldB709Push(g *game.Game, owner uuid.UUID, name, oracle string, power, toughness int) uuid.UUID {
	return effectstest.B12Push(g, owner, name, oldB709CreatureT, oracle, power, toughness)
}

func TestOldB709MaChaoIsUnblockableOnlyWhenAttackingAlone(t *testing.T) {
	for _, alone := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		chao := oldB709Push(g, me.ID, "Ma Chao, Western Warrior", oldB709MaChao, 3, 3)
		other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		rider := effectstest.Pr7bCreature(g, opp.ID, "Rider", 2, "{2}", nil, "horsemanship")
		if alone {
			effectstest.BrAttack(t, g, chao)
		} else {
			effectstest.BrAttack(t, g, chao, other)
		}
		err := g.DeclareBlocker(rider, chao)
		if alone && err == nil {
			t.Error("Ma Chao attacking alone was blocked")
		}
		if !alone && err != nil {
			t.Errorf("Ma Chao attacking with another creature was unblockable: %v", err)
		}
	}
}

func TestOldB709HorsemanshipStillGatesTheBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chao := oldB709Push(g, me.ID, "Ma Chao, Western Warrior", oldB709MaChao, 3, 3)
	other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, opp.ID, "Plain", "Creature — Bear", 2, 2)
	effectstest.BrAttack(t, g, chao, other)
	if err := g.DeclareBlocker(plain, chao); err == nil {
		t.Error("a creature without horsemanship blocked a horsemanship creature")
	}
	if err := g.DeclareBlocker(plain, other); err != nil {
		t.Errorf("the plain blocker could not block the other attacker: %v", err)
	}
}

func TestOldB709ZhaoZilongGrowsWhenItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zhao := oldB709Push(g, opp.ID, "Zhao Zilong, Tiger General", oldB709Zhao, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.BrAttack(t, g, bear)
	if err := g.DeclareBlocker(zhao, bear); err != nil {
		t.Fatalf("block: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, zhao); p != 4 || tough != 4 {
		t.Errorf("Zhao Zilong after blocking is %d/%d, want 4/4", p, tough)
	}
}

func TestOldB709YuanShaoLimitsYourCreaturesToOneBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB709Push(g, me.ID, "Yuan Shao, the Indecisive", oldB709YuanShao, 2, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Attacker", "Creature — Bear", 2, 2)
	b1 := effectstest.B12Creature(g, opp.ID, "Blocker One", "Creature — Bear", 2, 2)
	b2 := effectstest.B12Creature(g, opp.ID, "Blocker Two", "Creature — Bear", 2, 2)
	effectstest.BrAttack(t, g, bear)
	if err := effectstest.ReqBlock(t, g, bear, b1, b2); err == nil {
		t.Error("two creatures blocked one of Yuan Shao's controller's creatures")
	}
	if err := effectstest.ReqBlock(t, g, bear, b1); err != nil {
		t.Errorf("a single block was refused: %v", err)
	}
	_ = theirs
}

func TestOldB709YuanShaoDoesNotBindOtherPlayersCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB709Push(g, opp.ID, "Yuan Shao, the Indecisive", oldB709YuanShao, 2, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b1 := effectstest.B12Creature(g, opp.ID, "Blocker One", "Creature — Bear", 2, 2)
	b2 := effectstest.B12Creature(g, opp.ID, "Blocker Two", "Creature — Bear", 2, 2)
	effectstest.BrAttack(t, g, bear)
	if err := effectstest.ReqBlock(t, g, bear, b1, b2); err != nil {
		t.Errorf("Yuan Shao limited a creature its controller doesn't control: %v", err)
	}
}

func TestOldB709SunCeMayBounceACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Sun Ce, Young Conquerer", "Legendary Creature — Human Soldier", oldB709SunCe, 3, 3)
	effectstest.AnswerPickTargetIfOpen(t, g, theirs)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the creature did not return to its owner's hand")
	}
}

func TestOldB709SunCeMayDeclineTheBounce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Sun Ce, Young Conquerer", "Legendary Creature — Human Soldier", oldB709SunCe, 3, 3)
	effectstest.AnswerPickTargetIfOpen(t, g, theirs)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) {
		t.Error("the creature was bounced although the controller declined")
	}
}

func TestOldB709CaoRenCostsThreeLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Cao Ren, Wei Commander", "Legendary Creature — Human Soldier Warrior", oldB709CaoRen, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - me.Life; got != 3 {
		t.Errorf("the caster lost %d life, want 3", got)
	}
}

func TestOldB709NightRaidersMakeAnOpponentDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	raiders := oldB709Push(g, me.ID, "Wei Night Raiders", oldB709Raiders, 2, 2)
	effectstest.HandCardFull(opp, "Card One", "Instant", "{U}", "", nil)
	effectstest.HandCardFull(opp, "Card Two", "Instant", "{U}", "", nil)
	hand := opp.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, raiders, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardFromHand(t, g, opp.ID)
	if got := opp.Hand.Size(); got != hand-1 {
		t.Errorf("the opponent's hand is %d, want %d", got, hand-1)
	}
}

func TestOldB709NightRaidersIgnoreDamageToYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	raiders := oldB709Push(g, me.ID, "Wei Night Raiders", oldB709Raiders, 2, 2)
	effectstest.HandCardFull(me, "My Card", "Instant", "{U}", "", nil)
	hand := me.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, raiders, me.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("damage to its own controller made them discard: hand %d, want %d", got, hand)
	}
}

func TestOldB709WuScoutLooksAtAnOpponentsHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	effectstest.HandCardFull(opp, "Secret", "Instant", "{U}", "", nil)
	effectstest.CastCatalogSpell(t, g, "Wu Scout", "Creature — Human Soldier Scout", oldB709WuScout, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	for i := range opp.Hand.Cards {
		if !opp.Hand.Cards[i].IsKnownTo(me.ID) {
			t.Error("the Scout's controller did not see the opponent's hand")
		}
		if opp.Hand.Cards[i].IsKnownTo(other.ID) {
			t.Error("\"look at\" showed the hand to another player")
		}
	}
}

func TestOldB709LiuBeiGrowsWithGuanYuOrZhangFei(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	liu := oldB709Push(g, me.ID, "Liu Bei, Lord of Shu", oldB709LiuBei, 2, 4)
	recompute := func() { g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() }) }
	if p, tough := effectstest.EffectivePTOf(t, g, liu); p != 2 || tough != 4 {
		t.Fatalf("alone: %d/%d, want 2/4", p, tough)
	}
	theirs := oldB709Push(g, opp.ID, "Zhang Fei, Fierce Warrior", oldB709ZhangFei, 4, 4)
	recompute()
	if p, _ := effectstest.EffectivePTOf(t, g, liu); p != 2 {
		t.Fatalf("an opponent's Zhang Fei counted: power %d, want 2", p)
	}
	_ = theirs
	zf := oldB709Push(g, me.ID, "Zhang Fei, Fierce Warrior", oldB709ZhangFei, 4, 4)
	recompute()
	if p, tough := effectstest.EffectivePTOf(t, g, liu); p != 4 || tough != 6 {
		t.Fatalf("with your Zhang Fei: %d/%d, want 4/6", p, tough)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(zf) })
	recompute()
	if p, _ := effectstest.EffectivePTOf(t, g, liu); p != 2 {
		t.Fatalf("Zhang Fei left: power %d, want 2", p)
	}
	oldB709Push(g, me.ID, "Guan Yu, Sainted Warrior", oldB709GuanYu, 3, 5)
	recompute()
	if p, tough := effectstest.EffectivePTOf(t, g, liu); p != 4 || tough != 6 {
		t.Fatalf("with your Guan Yu: %d/%d, want 4/6", p, tough)
	}
	oldB709Push(g, me.ID, "Zhang Fei, Fierce Warrior", oldB709ZhangFei, 4, 4)
	recompute()
	if p, tough := effectstest.EffectivePTOf(t, g, liu); p != 4 || tough != 6 {
		t.Fatalf("both named permanents give the bonus once: %d/%d, want 4/6", p, tough)
	}
}

func TestOldB709GuanYuMayShuffleItselfIntoTheLibrary(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		id := oldB709Push(g, me.ID, "Guan Yu, Sainted Warrior", oldB709GuanYu, 3, 5)
		effectstest.DestroyNow(t, g, id)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.B21SettleStack(t, g)
		if yes && (!me.Library.Contains(id) || me.Graveyard.Contains(id)) {
			t.Errorf("accepted: in library %v, in graveyard %v; want it shuffled in", me.Library.Contains(id), me.Graveyard.Contains(id))
		}
		if !yes && (me.Library.Contains(id) || !me.Graveyard.Contains(id)) {
			t.Errorf("declined: in library %v, in graveyard %v; want it to stay in the graveyard", me.Library.Contains(id), me.Graveyard.Contains(id))
		}
	}
}

func TestOldB709ZhangHePumpsTheOthersWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	he := oldB709Push(g, me.ID, "Zhang He, Wei General", oldB709ZhangHe, 4, 2)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.BrAttack(t, g, he)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, mine); p != 3 {
		t.Errorf("my other creature has power %d, want 3", p)
	}
	if p, _ := effectstest.PtOf(t, g, he); p != 4 {
		t.Errorf("Zhang He itself has power %d, want 4", p)
	}
	if p, _ := effectstest.PtOf(t, g, theirs); p != 2 {
		t.Errorf("an opposing creature has power %d, want 2", p)
	}
}

func TestOldB709GuanYuDoesNotTriggerWhenItDiesUnderAnotherPlayersControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.ApaPush(g, opp.ID, me.ID, game.Card{Name: "Guan Yu, Sainted Warrior", TypeLine: "Legendary Creature — Human Soldier Warrior",
		OracleID: oldB709GuanYu, Power: 3, Toughness: 5})
	effectstest.DestroyNow(t, g, id)
	effectstest.B21SettleStack(t, g)
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil || effectstest.OwedMays(g, me.ID) > 0 {
		t.Error("the thief was asked about shuffling Guan Yu into a library")
	}
	if !opp.Graveyard.Contains(id) || opp.Library.Contains(id) || me.Library.Contains(id) {
		t.Errorf("a stolen Guan Yu is in its owner's graveyard %v, owner's library %v, thief's library %v; want the first only",
			opp.Graveyard.Contains(id), opp.Library.Contains(id), me.Library.Contains(id))
	}
}
