package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b755_cards_test.go — Tempest, batch 755 (S82 older sets, #5628):
// pump and combat tricks. Every test plays the real catalog entry through the
// engine.

const (
	oldB755Stalker  = "0103b2f2-05dd-40af-954d-303c144edb04"
	oldB755BattleCr = "12ef651c-5a86-49c4-a512-292e9d135a76"
	oldB755Wyvern   = "13459bf3-fcdc-4838-9669-918f97113054"
	oldB755Raider   = "240746f5-caab-46b9-914f-c2c7ecce6ca9"
	oldB755PitImp   = "3decbc45-3e79-455f-9d62-878ee31b458d"
	oldB755Squad    = "4dac142a-c8f8-45c8-85d3-40e170dee112"
	oldB755Vampire  = "58214e60-3258-4159-bf5a-b4069d80186d"
	oldB755Cannon   = "710a51e1-a305-49b5-a429-0ff1a10244fa"
	oldB755Frenzy   = "8d3d5e87-8b12-4549-a46e-70a103dbcf44"
)

func TestOldB755EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		oldB755Stalker, oldB755BattleCr, oldB755Wyvern, oldB755Raider, oldB755PitImp,
		oldB755Squad, oldB755Vampire, oldB755Cannon, oldB755Frenzy,
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

func oldB755Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	return g, me, opp
}

func oldB755Activate(g *game.Game, me *game.Player, id uuid.UUID, idx int, params game.ActivateAbilityParams) error {
	params.Strict = true
	return g.ActivateCatalogAbility(me.ID, id, idx, params)
}

func oldB755Target(ids ...uuid.UUID) game.ActivateAbilityParams {
	return game.ActivateAbilityParams{Targets: effectstest.CardRefs(ids...)}
}

func oldB755PT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	return effectstest.EffectivePTOf(t, g, id)
}

func TestOldB755StalkerRegeneratesAndPumps(t *testing.T) {
	g, me, _ := oldB755Table(t)
	stalker := effectstest.B12Push(g, me.ID, "Darkling Stalker", "Creature — Shade Spirit", oldB755Stalker, 1, 1)
	effectstest.FloatForTest(g, me, "B")
	if err := oldB755Activate(g, me, stalker, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("pump: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := oldB755PT(t, g, stalker); p != 2 || tough != 2 {
		t.Errorf("the Stalker is %d/%d, want 2/2", p, tough)
	}
	effectstest.FloatForTest(g, me, "B")
	if err := oldB755Activate(g, me, stalker, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("regenerate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.UsgB3Destroy(t, g, stalker)
	if !g.Battlefield.Contains(stalker) {
		t.Error("the regenerated Stalker was destroyed")
	}
}

func TestOldB755BattleCryPumpsOnlyYourCreatures(t *testing.T) {
	g, me, opp := oldB755Table(t)
	cry := effectstest.B12Push(g, me.ID, "Gerrard's Battle Cry", "Enchantment", oldB755BattleCr, 0, 0)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "CCW")
	if err := oldB755Activate(g, me, cry, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := oldB755PT(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("my Bear is %d/%d, want 3/3", p, tough)
	}
	if p, tough := oldB755PT(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("their Bear is %d/%d, want 2/2", p, tough)
	}
}

func TestOldB755WyvernGainsPowerAndLosesToughness(t *testing.T) {
	g, me, _ := oldB755Table(t)
	wyvern := effectstest.B12Push(g, me.ID, "Flowstone Wyvern", "Creature — Drake", oldB755Wyvern, 3, 3)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, wyvern), "flying") {
		t.Error("the Wyvern does not fly")
	}
	effectstest.FloatForTest(g, me, "RR")
	for i := 0; i < 2; i++ {
		if err := oldB755Activate(g, me, wyvern, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(wyvern) {
		t.Error("the Wyvern survived at -1 toughness")
	}
}

func TestOldB755WyvernPumpsOnce(t *testing.T) {
	g, me, _ := oldB755Table(t)
	wyvern := effectstest.B12Push(g, me.ID, "Flowstone Wyvern", "Creature — Drake", oldB755Wyvern, 3, 3)
	effectstest.FloatForTest(g, me, "R")
	if err := oldB755Activate(g, me, wyvern, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := oldB755PT(t, g, wyvern); p != 5 || tough != 1 {
		t.Errorf("the Wyvern is %d/%d, want 5/1", p, tough)
	}
}

func TestOldB755RaiderSacrificesAGoblinToPumpATarget(t *testing.T) {
	g, me, _ := oldB755Table(t)
	raider := effectstest.B12Push(g, me.ID, "Mogg Raider", "Creature — Goblin", oldB755Raider, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	p := oldB755Target(bear)
	p.SacrificeIDs = []uuid.UUID{bear}
	if err := oldB755Activate(g, me, raider, 0, p); err == nil {
		t.Fatal("the Raider sacrificed a Bear")
	}
	goblin := effectstest.B12Creature(g, me.ID, "Goblin", "Creature — Goblin", 1, 1)
	p.SacrificeIDs = []uuid.UUID{goblin}
	if err := oldB755Activate(g, me, raider, 0, p); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(goblin) {
		t.Error("the Goblin was not sacrificed")
	}
	if pw, tough := oldB755PT(t, g, bear); pw != 3 || tough != 3 {
		t.Errorf("the Bear is %d/%d, want 3/3", pw, tough)
	}
}

func TestOldB755PitImpPumpsAtMostTwiceEachTurn(t *testing.T) {
	g, me, _ := oldB755Table(t)
	imp := effectstest.B12Push(g, me.ID, "Pit Imp", "Creature — Imp", oldB755PitImp, 0, 1)
	effectstest.FloatForTest(g, me, "BBB")
	for i := 0; i < 2; i++ {
		if err := oldB755Activate(g, me, imp, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate %d: %v", i, err)
		}
	}
	if err := oldB755Activate(g, me, imp, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a third activation in one turn was accepted")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := oldB755PT(t, g, imp); p != 2 {
		t.Errorf("the Imp has %d power, want 2", p)
	}
}

func TestOldB755SquadShrinksForEveryOtherCreature(t *testing.T) {
	g, me, opp := oldB755Table(t)
	squad := effectstest.B12Push(g, me.ID, "Mogg Squad", "Creature — Goblin", oldB755Squad, 3, 3)
	if p, tough := oldB755PT(t, g, squad); p != 3 || tough != 3 {
		t.Fatalf("alone the Squad is %d/%d, want 3/3", p, tough)
	}
	effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if p, tough := oldB755PT(t, g, squad); p != 1 || tough != 1 {
		t.Errorf("with two others the Squad is %d/%d, want 1/1", p, tough)
	}
	effectstest.B12Creature(g, opp.ID, "Another Bear", "Creature — Bear", 2, 2)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(squad) {
		t.Error("the Squad survived at 0 toughness")
	}
}

func TestOldB755VampireDiscardsACreatureCardForAPump(t *testing.T) {
	g, me, _ := oldB755Table(t)
	vampire := effectstest.B12Push(g, me.ID, "Skyshroud Vampire", "Creature — Vampire", oldB755Vampire, 3, 3)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, vampire), "flying") {
		t.Error("the Vampire does not fly")
	}
	land := effectstest.HandCardForTest(me, "Forest", "Basic Land — Forest", "")
	if err := oldB755Activate(g, me, vampire, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("discarded a land to the Vampire")
	}
	bear := effectstest.HandCardForTest(me, "Bear", "Creature — Bear", "")
	if err := oldB755Activate(g, me, vampire, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(bear) {
		t.Error("the creature card was not discarded")
	}
	if p, tough := oldB755PT(t, g, vampire); p != 5 || tough != 5 {
		t.Errorf("the Vampire is %d/%d, want 5/5", p, tough)
	}
}

func TestOldB755CannonLaunchesACreatureThatDiesAtTheEndStep(t *testing.T) {
	g, me, opp := oldB755Table(t)
	cannon := effectstest.B12Push(g, me.ID, "Mogg Cannon", "Artifact", oldB755Cannon, 0, 0)
	foe := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := oldB755Activate(g, me, cannon, 0, oldB755Target(foe)); err == nil {
		t.Fatal("the Cannon targeted an opponent's creature")
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := oldB755Activate(g, me, cannon, 0, oldB755Target(bear)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := oldB755PT(t, g, bear); p != 3 {
		t.Errorf("the Bear has %d power, want 3", p)
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, bear), "flying") {
		t.Error("the Bear did not gain flying")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear was not destroyed at the beginning of the end step")
	}
}

func TestOldB755BloodFrenzyPumpsAnAttackerThenDestroysIt(t *testing.T) {
	g, me, _ := oldB755Table(t)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	spell := game.Card{Name: "Blood Frenzy", TypeLine: "Instant", OracleID: oldB755Frenzy, ManaCost: "{1}{R}"}
	spell.InstanceID = uuid.New()
	spell.Owner, spell.Controller = me.ID, me.ID
	me.Hand.PushTop(spell)
	if err := g.CastSpell(me.ID, spell.InstanceID, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("cast on a creature that is not attacking or blocking")
	}
	effectstest.AttackAndTrigger(t, g, bear)
	effectstest.FloatForTest(g, me, "CR")
	if err := g.CastSpell(me.ID, spell.InstanceID, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("cast on an attacker: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := oldB755PT(t, g, bear); p != 6 {
		t.Errorf("the attacker has %d power, want 6", p)
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature was not destroyed at the beginning of the end step")
	}
}

func TestOldB755BloodFrenzyWindowEndsAtTheCombatDamageStep(t *testing.T) {
	g, me, _ := oldB755Table(t)
	for step, want := range map[game.Step]bool{
		game.StepUpkeep: true, game.StepPrecombatMain: true, game.StepBeginCombat: true,
		game.StepDeclareAttackers: true, game.StepDeclareBlockers: true,
		game.StepFirstStrikeDamage: false, game.StepCombatDamage: false, game.StepEndCombat: false,
		game.StepPostcombatMain: false, game.StepEnd: false,
	} {
		g.Turn.Step = step
		if got := effects.OldB755BeforeTheCombatDamageStep(g, me.ID, game.Card{}); got != want {
			t.Errorf("step %s: window open = %v, want %v", step, got, want)
		}
	}
}
