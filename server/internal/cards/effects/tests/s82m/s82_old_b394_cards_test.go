package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB394Boulderfall = "2160aed4-5a7a-4eca-96b5-707204fa73bb"
	oldB394Wheel       = "58f70225-17b5-4aa6-a62a-89c57e31ecc1"
	oldB394Triton      = "8a327626-ce42-4ef8-8869-3468960c671e"
	oldB394Medomai     = "b6b91ea7-a22c-4835-8c62-3360b0fd3a8b"
	oldB394Satyr       = "bb107b33-ae2e-4d45-b537-1e6e25c66ec8"
	oldB394Jolt        = "f24634f3-49aa-4362-8faa-93598b9324d9"
)

func TestOldB394AreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB394Boulderfall: "Boulderfall", oldB394Wheel: "Flamecast Wheel", oldB394Triton: "Crackling Triton",
		oldB394Medomai: "Medomai the Ageless", oldB394Satyr: "Firedrinker Satyr", oldB394Jolt: "Spark Jolt",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB394BoulderfallDividesFiveAmongAnyTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	effectstest.AdvanceToMain(t, g)
	card := effectstest.HandCardOf(g, me, "Boulderfall", "Instant", "", oldB394Boulderfall)
	refs := append(effectstest.CardRefs(wall), game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Targets: refs, Distribution: map[uuid.UUID]int{wall: 2, opp.ID: 2}, Strict: true}); err == nil {
		t.Fatal("a division of 4 was accepted")
	}
	life := opp.Life
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Targets: refs, Distribution: map[uuid.UUID]int{wall: 2, opp.ID: 3}, Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if d := effectstest.E2Card(t, g, wall).DamageMarked; d != 2 {
		t.Errorf("the wall has %d damage, want 2", d)
	}
	if opp.Life != life-3 {
		t.Errorf("life %d, want %d", opp.Life, life-3)
	}
}

func TestOldB394FlamecastWheelSacrificesToHitACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wheel := effectstest.PushCatalogPermanent(g, me.ID, "Flamecast Wheel", "Artifact", oldB394Wheel, false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 9)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPool(me, 5)
	if err := g.ActivateCatalogAbility(me.ID, wheel, 0, game.ActivateAbilityParams{Targets: effectstest.Pr6Player(opp.ID), Strict: true}); err == nil {
		t.Fatal("targeted a player")
	}
	if err := g.ActivateCatalogAbility(me.ID, wheel, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(wheel) {
		t.Error("the Wheel is sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if d := effectstest.E2Card(t, g, bear).DamageMarked; d != 3 {
		t.Errorf("damage %d, want 3", d)
	}
}

func TestOldB394CracklingTritonSacrificesToDealTwoToAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	triton := effectstest.PushCatalogPermanent(g, me.ID, "Crackling Triton", "Creature — Merfolk Wizard", oldB394Triton, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "R", 3)
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, triton, 0, game.ActivateAbilityParams{Targets: effectstest.Pr6Player(opp.ID), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(triton) {
		t.Error("the Triton is sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d", opp.Life, life-2)
	}
}

func TestOldB394SparkJoltDealsOneAndScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Spark Jolt", "Instant", oldB394Jolt, effectstest.Pr6Player(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d, want %d", opp.Life, life-1)
	}
	if len(g.PendingChoices) == 0 {
		t.Fatal("no scry prompt")
	}
	effectstest.B26AnswerScryAllTop(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d prompts left after the scry", len(g.PendingChoices))
	}
}

func TestOldB394FiredrinkerSatyrReflectsTheDamageItIsDealt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	satyr := effectstest.PushCatalogPermanent(g, me.ID, "Firedrinker Satyr", "Creature — Satyr Shaman", oldB394Satyr, false)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, satyr)
		c.Power, c.Toughness = 2, 5
	})
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.CardTarget(satyr))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-3 {
		t.Errorf("life %d, want %d (3 damage reflected to its controller)", me.Life, life-3)
	}
}

func TestOldB394FiredrinkerSatyrPumpsAndHurtsItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	satyr := effectstest.PushCatalogPermanent(g, me.ID, "Firedrinker Satyr", "Creature — Satyr Shaman", oldB394Satyr, false)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, satyr)
		c.Power, c.Toughness = 2, 1
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "R", 2)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, satyr, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, satyr); p != 3 {
		t.Errorf("power %d, want 3", p)
	}
	// The Satyr was dealt 1 damage by itself: 1 to you from the ability, and
	// the reflection of the damage it was dealt does not apply (it deals the
	// damage, it is not dealt any).
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
}

func TestOldB394MedomaiTakesAnExtraTurnButCantAttackInOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	medomai := effectstest.PushCatalogPermanent(g, me.ID, "Medomai the Ageless", "Legendary Creature — Sphinx", oldB394Medomai, false)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, medomai)
		c.Power, c.Toughness = 4, 4
		c.Keywords = []string{"flying"}
	})
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	g.WithWriteLock(func() { g.Turn.Extra = true })
	effectstest.WantAttackRefused(t, g, medomai, opp.ID, "extra turn")
	g.WithWriteLock(func() { g.Turn.Extra = false })
	if err := effectstest.DeclareResult(g, medomai, opp.ID); err != nil {
		t.Fatalf("attack on a normal turn: %v", err)
	}
	if err := g.DeclareAttacker(medomai, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(g.ExtraTurnsQueued()); n != 1 {
		t.Errorf("%d extra turns queued, want 1", n)
	}
}
