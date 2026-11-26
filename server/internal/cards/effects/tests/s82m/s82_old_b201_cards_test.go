package testss82m

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b201_cards_test.go — Global Series Jiang Yanggu & Mu Yanling,
// batch 201 (S82 older sets, #5528). Every test plays the real catalog entry
// through the engine.

const (
	oldB201BreathOfFireOracle     = "046923d6-2d58-42ce-afb8-886f4932e76c"
	oldB201FireOmenCraneOracle    = "2306386d-5913-429e-8d8b-ba780841bf6e"
	oldB201ConfidenceOracle       = "341f75a5-df02-48e9-8997-1938148ad89a"
	oldB201CleansingScreechOracle = "4523c22b-363b-4bc6-ba23-b428054abc5d"
	oldB201DragonsPresenceOracle  = "4fb7aeb6-2fc3-459e-bad8-1c7ea92f7ae7"
	oldB201EarthshakingSiOracle   = "62c4d349-a18e-40a2-8cc3-244c219dec7f"

	oldB201BearType = "Creature — Bear"
)

func TestOldB201EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		oldB201BreathOfFireOracle, oldB201FireOmenCraneOracle, oldB201ConfidenceOracle,
		oldB201CleansingScreechOracle, oldB201DragonsPresenceOracle, oldB201EarthshakingSiOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func TestOldB201BreathOfFireDealsTwoToACreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.SeatsForCombat(g)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Breath of Fire", "Instant", oldB201BreathOfFireOracle,
		[]game.TargetRef{effectstest.PlayerTarget(opp.ID)}); err == nil {
		t.Fatal("a player was a legal target")
	}
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB201BearType, 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Breath of Fire", "Instant", oldB201BreathOfFireOracle, effectstest.CardTarget(bear))
	effectstest.CastCatalogSpell(t, g, "Breath of Fire", "Instant", oldB201BreathOfFireOracle, effectstest.CardTarget(big))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the 2/2 survived 2 damage")
	}
	if d := effectstest.DamageOn(t, g, big); d != 2 {
		t.Errorf("the 4/4 has %d damage, want 2", d)
	}
}

func TestOldB201CleansingScreechDealsFourToAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.SeatsForCombat(g)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Cleansing Screech", "Sorcery", oldB201CleansingScreechOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("life %d -> %d, want -4", life, opp.Life)
	}
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB201BearType, 3, 4)
	effectstest.CastCatalogSpell(t, g, "Cleansing Screech", "Sorcery", oldB201CleansingScreechOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the 3/4 survived 4 damage")
	}
}

func TestOldB201DragonsPresenceHitsOnlyACreatureInCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	idle := effectstest.B12Creature(g, opp.ID, "Idle Bear", oldB201BearType, 2, 2)
	attacker := effectstest.B12Creature(g, me.ID, "Attacking Giant", "Creature — Giant", 5, 5)
	effectstest.BrAttack(t, g, attacker)
	// Cast in the combat step itself: CastCatalogSpell would walk on to the
	// postcombat main phase, where nothing is attacking.
	spell := effectstest.PushCatalogHandCard(me, "Dragon's Presence", "Instant", oldB201DragonsPresenceOracle)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(idle)}); err == nil {
		t.Fatal("a creature not in combat was a legal target")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(attacker)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("the attacking 5/5 survived 5 damage")
	}
}

func TestOldB201ConfidenceFromStrengthPumpsAndGivesTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", oldB201BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Confidence from Strength", "Sorcery", oldB201ConfidenceOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 6 || tough != 6 {
		t.Errorf("the Bear is %d/%d, want 6/6", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("no trample")
	}
}

func TestOldB201EarthshakingSiTramples(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Earthshaking Si", "Creature — Beast", oldB201EarthshakingSiOracle, 5, 5)
	if !effectstest.HasEffectiveKeyword(t, g, id, "trample") {
		t.Error("no trample")
	}
	spec, _ := effects.Lookup(oldB201EarthshakingSiOracle)
	if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 || len(spec.Activated) != 0 {
		t.Error("Earthshaking Si declares more than its keyword")
	}
}

func TestOldB201FireOmenCraneShootsAnOpponentsCreatureOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	crane := effectstest.B12Push(g, me.ID, "Fire-Omen Crane", "Creature — Bird Spirit", oldB201FireOmenCraneOracle, 3, 3)
	if !effectstest.HasEffectiveKeyword(t, g, crane, "flying") {
		t.Error("no flying")
	}
	mine := effectstest.B12Creature(g, me.ID, "My Bear", oldB201BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB201BearType, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(crane, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	if c := effectstest.PickTargetPrompt(t, g); c != nil {
		for _, id := range c.PickTargetCards {
			if id == mine || id == crane {
				t.Errorf("my own creature %s was offered as a target", id)
			}
		}
	}
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if d := effectstest.DamageOn(t, g, theirs); d != 1 {
		t.Errorf("their creature has %d damage, want 1", d)
	}
	if d := effectstest.DamageOn(t, g, mine); d != 0 {
		t.Errorf("my creature has %d damage, want 0", d)
	}
}
