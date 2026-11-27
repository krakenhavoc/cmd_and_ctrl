package testss82m

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b712_cards_test.go — OLD batch 712 (Portal Three Kingdoms, #5622):
// the other creatures. Each test plays the real catalog entry through the
// engine.

const (
	oldB712Flanking = "47bbb630-2bf7-47a2-93bb-266e491803c3"
	oldB712Huang    = "7d6c4290-d46d-4b98-805c-3f537462c4c8"
	oldB712Infantry = "8ad91015-db7f-43ea-97c9-bf2bf5656858"
	oldB712Dragon   = "e5c2cf0f-c1c0-4fd6-9b5f-7d4aa25246a7"
	oldB712Mystic   = "e705ec38-b8f7-4f99-8cb9-8448dc3e50ec"
	oldB712WolfPack = "ec58a82f-8c78-4f88-90d5-e32d83be7a48"
	oldB712SimaYi   = "eefa9a94-9915-4dcd-b5dc-7f8e9aa621a7"
)

func TestOldB712EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB712Flanking: "Flanking Troops", oldB712Huang: "Huang Zhong, Shu General", oldB712Infantry: "Yuan Shao's Infantry",
		oldB712Dragon: "Zodiac Dragon", oldB712Mystic: "Taoist Mystic", oldB712WolfPack: "Wolf Pack",
		oldB712SimaYi: "Sima Yi, Wei Field Marshal",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name=%q completeness=%v caveats=%v, want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB712FlankingTroopsMayTapACreatureWhenItAttacks(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		troops := effectstest.B12Push(g, me.ID, "Flanking Troops", "Creature — Human Soldier", oldB712Flanking, 2, 2)
		blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(troops, opp.ID); err != nil {
			t.Fatal(err)
		}
		effectstest.LockInAttacks(t, g)
		effectstest.AnswerPickTargetIfOpen(t, g, blocker)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.HmlTapped(t, g, blocker); got != yes {
			t.Errorf("accepted=%v: the target is tapped=%v", yes, got)
		}
	}
}

func TestOldB712HuangZhongCantBeBlockedByMoreThanOneCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	huang := effectstest.B12Push(g, me.ID, "Huang Zhong, Shu General", "Legendary Creature — Human Soldier", oldB712Huang, 2, 3)
	other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b1 := effectstest.B12Creature(g, opp.ID, "Blocker One", "Creature — Bear", 2, 2)
	b2 := effectstest.B12Creature(g, opp.ID, "Blocker Two", "Creature — Bear", 2, 2)
	effectstest.BrAttack(t, g, huang, other)
	if err := effectstest.ReqBlock(t, g, huang, b1, b2); err == nil {
		t.Error("two creatures blocked Huang Zhong")
	}
	if err := effectstest.ReqBlock(t, g, other, b1, b2); err != nil {
		t.Errorf("two creatures could not block a different attacker: %v", err)
	}
	if err := effectstest.ReqBlock(t, g, huang, b1); err != nil {
		t.Errorf("a single block on Huang Zhong was refused: %v", err)
	}
}

func TestOldB712YuanShaosInfantryIsUnblockableOnlyWhenAttackingAlone(t *testing.T) {
	for _, alone := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		inf := effectstest.B12Push(g, me.ID, "Yuan Shao's Infantry", "Creature — Human Soldier", oldB712Infantry, 2, 2)
		other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
		if alone {
			effectstest.BrAttack(t, g, inf)
		} else {
			effectstest.BrAttack(t, g, inf, other)
		}
		err := g.DeclareBlocker(blocker, inf)
		if alone && err == nil {
			t.Error("the Infantry attacking alone was blocked")
		}
		if !alone && err != nil {
			t.Errorf("the Infantry attacking with another creature was unblockable: %v", err)
		}
	}
}

func TestOldB712ZodiacDragonMayReturnToHandWhenItDies(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		id := effectstest.B12Push(g, me.ID, "Zodiac Dragon", "Creature — Dragon", oldB712Dragon, 8, 8)
		effectstest.DestroyNow(t, g, id)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.B21SettleStack(t, g)
		if yes && (!me.Hand.Contains(id) || me.Graveyard.Contains(id)) {
			t.Errorf("accepted: in hand %v, in graveyard %v; want it in hand", me.Hand.Contains(id), me.Graveyard.Contains(id))
		}
		if !yes && (me.Hand.Contains(id) || !me.Graveyard.Contains(id)) {
			t.Errorf("declined: in hand %v, in graveyard %v; want it to stay in the graveyard", me.Hand.Contains(id), me.Graveyard.Contains(id))
		}
	}
}

func TestOldB712ZodiacDragonDoesNotTriggerForAStolenDragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.ApaPush(g, opp.ID, me.ID, game.Card{Name: "Zodiac Dragon", TypeLine: "Creature — Dragon", OracleID: oldB712Dragon, Power: 8, Toughness: 8})
	effectstest.DestroyNow(t, g, id)
	effectstest.B21SettleStack(t, g)
	if !opp.Graveyard.Contains(id) || opp.Hand.Contains(id) || me.Hand.Contains(id) {
		t.Errorf("a stolen Dragon is in its owner's graveyard %v, owner's hand %v, thief's hand %v; want the first only",
			opp.Graveyard.Contains(id), opp.Hand.Contains(id), me.Hand.Contains(id))
	}
}

func TestOldB712TaoistMysticCantBeBlockedByHorsemanship(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mystic := effectstest.B12Push(g, me.ID, "Taoist Mystic", "Creature — Human Mystic", oldB712Mystic, 2, 2)
	rider := effectstest.Pr7bCreature(g, opp.ID, "Rider", 2, "{2}", nil, "horsemanship")
	plain := effectstest.B12Creature(g, opp.ID, "Plain", "Creature — Bear", 2, 2)
	effectstest.BrAttack(t, g, mystic)
	if err := g.DeclareBlocker(rider, mystic); err == nil {
		t.Error("a creature with horsemanship blocked the Mystic")
	}
	if err := g.DeclareBlocker(plain, mystic); err != nil {
		t.Errorf("a creature without horsemanship could not block: %v", err)
	}
}

func TestOldB712WolfPackMayAssignAsThoughUnblocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wolf := effectstest.B12Push(g, me.ID, "Wolf Pack", "Creature — Wolf", oldB712WolfPack, 7, 6)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !effectstest.ApaLive(g, wolf).Effective().AssignsAsThoughUnblocked {
		t.Error("Wolf Pack cannot assign its damage as though unblocked")
	}
	if effectstest.ApaLive(g, bear).Effective().AssignsAsThoughUnblocked {
		t.Error("another creature got the ability")
	}
}

func TestOldB712SimaYiCountsYourSwamps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sima := effectstest.B12Push(g, me.ID, "Sima Yi, Wei Field Marshal", "Legendary Creature — Human Soldier", oldB712SimaYi, 0, 4)
	recompute := func() { g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() }) }
	if p, tough := effectstest.EffectivePTOf(t, g, sima); p != 0 || tough != 4 {
		t.Fatalf("no Swamps: %d/%d, want 0/4", p, tough)
	}
	effectstest.B12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	effectstest.B12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.B12Permanent(g, opp.ID, "Swamp", "Basic Land — Swamp")
	recompute()
	if p, tough := effectstest.EffectivePTOf(t, g, sima); p != 2 || tough != 4 {
		t.Errorf("two Swamps (the Forest and the opponent's Swamp don't count): %d/%d, want 2/4", p, tough)
	}
}
