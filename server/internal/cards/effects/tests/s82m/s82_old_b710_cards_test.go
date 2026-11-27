package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b710_cards_test.go — OLD batch 710 (Portal Three Kingdoms, #5622):
// pump and combat tricks. Each test plays the real catalog entry through the
// engine.

const (
	oldB710ShuDefender = "0b6625cb-4177-45bd-ba99-2af43dfd6f46"
	oldB710WarStand    = "1f08072d-7dec-4f0d-b465-dda62eb17630"
	oldB710Ambush      = "2d1446f4-cf71-4798-b921-8f6824350319"
	oldB710Admiral     = "578c5b92-9f7e-4645-9624-f2c1a16b6aaf"
	oldB710DiluHorse   = "b7162f90-8430-4c1d-9f9a-a912bf1e9f50"
	oldB710Tiger       = "c07a86de-ec29-478c-b6cb-057eb9d3e126"
	oldB710Cheetah     = "dad3b459-725e-4bf5-bf7a-bd89866ea8c4"
	oldB710RedHare     = "ddd45d52-20c3-4731-8860-142af1ccaaa0"
	oldB710MengHuo     = "ea7fdff3-9d45-4a4a-b7fe-84b334f04291"
)

func TestOldB710EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB710ShuDefender: "Shu Defender", oldB710WarStand: "Warrior's Stand", oldB710Ambush: "Wei Ambush Force",
		oldB710Admiral: "Wu Admiral", oldB710DiluHorse: "Riding the Dilu Horse", oldB710Tiger: "Slashing Tiger",
		oldB710Cheetah: "Trained Cheetah", oldB710RedHare: "Riding Red Hare", oldB710MengHuo: "Meng Huo, Barbarian King",
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

func TestOldB710ShuDefenderGainsToughnessWhenItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	def := effectstest.B12Push(g, opp.ID, "Shu Defender", "Creature — Human Soldier", oldB710ShuDefender, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.BrAttack(t, g, bear)
	if err := g.DeclareBlocker(def, bear); err != nil {
		t.Fatalf("block: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, def); p != 2 || tough != 4 {
		t.Errorf("Shu Defender after blocking is %d/%d, want 2/4", p, tough)
	}
}

func TestOldB710WeiAmbushForceGainsPowerWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	force := effectstest.B12Push(g, me.ID, "Wei Ambush Force", "Creature — Human Soldier", oldB710Ambush, 1, 1)
	if p, _ := effectstest.PtOf(t, g, force); p != 1 {
		t.Fatalf("before attacking: power %d, want 1", p)
	}
	effectstest.BrAttack(t, g, force)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, force); p != 3 || tough != 1 {
		t.Errorf("after attacking: %d/%d, want 3/1", p, tough)
	}
}

func TestOldB710WarriorsStandNeedsYouToBeAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if err := effectstest.CastCatalogSpellErr(t, g, "Warrior's Stand", "Instant", oldB710WarStand, nil); err == nil {
		t.Fatal("Warrior's Stand was cast outside the declare attackers step")
	}
	att := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 3, 3)
	mine := effectstest.B12Creature(g, opp.ID, "Defender", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, me.ID, "Home", "Creature — Bear", 1, 1)
	effectstest.SfAttackHold(t, g, opp.ID, att)
	effectstest.SfCast(t, g, opp, game.Card{Name: "Warrior's Stand", TypeLine: "Instant", OracleID: oldB710WarStand}, game.CastSpellParams{})
	if p, tough := effectstest.PtOf(t, g, mine); p != 4 || tough != 4 {
		t.Errorf("the defender's creature is %d/%d, want 4/4", p, tough)
	}
	if p, _ := effectstest.PtOf(t, g, att); p != 3 {
		t.Errorf("the attacker's creature changed: power %d, want 3", p)
	}
	_ = theirs
}

func TestOldB710WarriorsStandRefusesAPlayerNobodyAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, third := g.Seats[0], g.Seats[2]
	att := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 3, 3)
	effectstest.SfAttackHold(t, g, third.ID, att)
	opp := g.Seats[1]
	card := game.Card{InstanceID: uuid.New(), Name: "Warrior's Stand", TypeLine: "Instant", OracleID: oldB710WarStand, Owner: opp.ID, Controller: opp.ID}
	opp.Hand.PushTop(card)
	if err := g.CastSpell(opp.ID, card.InstanceID, game.CastSpellParams{}); err == nil {
		t.Fatal("Warrior's Stand was cast by a player nobody attacked")
	}
}

func TestOldB710WuAdmiralGrowsWhileAnOpponentControlsAnIsland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	admiral := effectstest.B12Push(g, me.ID, "Wu Admiral", "Creature — Human Soldier", oldB710Admiral, 3, 3)
	recompute := func() { g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() }) }
	if p, tough := effectstest.EffectivePTOf(t, g, admiral); p != 3 || tough != 3 {
		t.Fatalf("no Island: %d/%d, want 3/3", p, tough)
	}
	effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	recompute()
	if p, _ := effectstest.EffectivePTOf(t, g, admiral); p != 3 {
		t.Fatalf("your own Island counted: power %d, want 3", p)
	}
	effectstest.B12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain")
	recompute()
	if p, _ := effectstest.EffectivePTOf(t, g, admiral); p != 3 {
		t.Fatalf("an opposing Mountain counted: power %d, want 3", p)
	}
	island := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	recompute()
	if p, tough := effectstest.EffectivePTOf(t, g, admiral); p != 4 || tough != 4 {
		t.Fatalf("an opposing Island: %d/%d, want 4/4", p, tough)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(island) })
	recompute()
	if p, _ := effectstest.EffectivePTOf(t, g, admiral); p != 3 {
		t.Fatalf("the Island left: power %d, want 3", p)
	}
}

func TestOldB710RidingTheDiluHorseLastsIndefinitely(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Riding the Dilu Horse", "Sorcery", oldB710DiluHorse, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Fatalf("after the spell: %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "horsemanship") {
		t.Fatal("the creature lacks horsemanship")
	}
	effectstest.AdvancePastCleanupForTest(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("after the turn ended: %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "horsemanship") {
		t.Error("horsemanship ended with the turn")
	}
}

func TestOldB710RidingRedHareEndsWithTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Riding Red Hare", "Sorcery", oldB710RedHare, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 5 {
		t.Fatalf("after the spell: %d/%d, want 5/5", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "horsemanship") {
		t.Fatal("the creature lacks horsemanship")
	}
	effectstest.AdvancePastCleanupForTest(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("after the turn ended: %d/%d, want 2/2", p, tough)
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "horsemanship") {
		t.Error("horsemanship outlived the turn")
	}
}

func TestOldB710BlockedCatsGrowOncePerBlock(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		base, bonus  int
	}{
		{"Slashing Tiger", oldB710Tiger, 3, 2},
		{"Trained Cheetah", oldB710Cheetah, 2, 1},
	} {
		for _, blockers := range []int{0, 1, 2} {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			cat := effectstest.B12Push(g, me.ID, tc.name, "Creature — Cat", tc.oracle, tc.base, tc.base)
			var decls []game.BlockDeclaration
			for i := 0; i < blockers; i++ {
				b := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
				decls = append(decls, game.BlockDeclaration{Blocker: b, Attacker: cat})
			}
			effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
			if err := g.DeclareAttacker(cat, opp.ID); err != nil {
				t.Fatal(err)
			}
			effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
			if blockers > 0 {
				if err := g.DeclareBlockers(decls); err != nil {
					t.Fatal(err)
				}
			}
			effectstest.LockInBlocks(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			want := tc.base
			if blockers > 0 {
				want += tc.bonus
			}
			if p, tough := effectstest.PtOf(t, g, cat); p != want || tough != want {
				t.Errorf("%s with %d blocker(s): %d/%d, want %d/%d", tc.name, blockers, p, tough, want, want)
			}
		}
	}
}

func TestOldB710MengHuoPumpsOnlyOtherGreenCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	huo := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Meng Huo, Barbarian King", TypeLine: "Legendary Creature — Human Barbarian Soldier",
		OracleID: oldB710MengHuo, Colors: []string{"G"}, Power: 4, Toughness: 4})
	green := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Green Bear", TypeLine: "Creature — Bear", Colors: []string{"G"}, Power: 2, Toughness: 2})
	red := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Red Bear", TypeLine: "Creature — Bear", Colors: []string{"R"}, Power: 2, Toughness: 2})
	theirs := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Their Green Bear", TypeLine: "Creature — Bear", Colors: []string{"G"}, Power: 2, Toughness: 2})
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	for id, want := range map[uuid.UUID]int{huo: 4, green: 3, red: 2, theirs: 2} {
		if p, _ := effectstest.EffectivePTOf(t, g, id); p != want {
			t.Errorf("power %d, want %d", p, want)
		}
	}
}
