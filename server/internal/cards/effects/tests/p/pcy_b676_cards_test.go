package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pcy_b676_cards_test.go — Prophecy, S82 OLD batch 676 (#5617): Hollow Warrior,
// Blessed Wind and Coffin Puppets, each through the real cast, declaration and
// activation paths. (Rhystic Syphon was already on develop.)

const (
	pcy676Warrior = "18ca77d9-4740-4be4-913a-e21c9ed9e43f"
	pcy676Wind    = "93da4a35-d2d8-476c-ad03-318101c0118b"
	pcy676Puppets = "e49771ab-cabb-4a5f-aef3-a0ac8147e6bf"
)

func TestPcy676EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		pcy676Warrior: "Hollow Warrior", pcy676Wind: "Blessed Wind", pcy676Puppets: "Coffin Puppets",
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

// --- Hollow Warrior ---

func TestPcy676HollowWarriorAttacksOnlyByTappingAnotherUntappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warrior := effectstest.B12Push(g, me.ID, "Hollow Warrior", "Artifact Creature — Golem Warrior", pcy676Warrior, 4, 4)
	helper := effectstest.B12Creature(g, me.ID, "Helper", "Creature — Bear", 1, 1)
	tapped := effectstest.B12Creature(g, me.ID, "Tired", "Creature — Bear", 1, 1)
	effectstest.TapForTest(t, g, tapped)
	effectstest.B64ReadyToAttack(g, warrior)
	effectstest.AdvanceToMain(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttackerWith(warrior, opp.ID, game.DeclareAttackersParams{}); err == nil {
		t.Fatal("attacked without tapping a creature")
	}
	if err := g.DeclareAttackerWith(warrior, opp.ID, game.DeclareAttackersParams{Pay: game.CombatCostPayment{TapIDs: []uuid.UUID{tapped}}}); err == nil {
		t.Fatal("paid with a creature that was already tapped")
	}
	if err := g.DeclareAttackerWith(warrior, opp.ID, game.DeclareAttackersParams{Pay: game.CombatCostPayment{TapIDs: []uuid.UUID{warrior}}}); err == nil {
		t.Fatal("paid with the Warrior itself")
	}
	if err := g.DeclareAttackerWith(warrior, opp.ID, game.DeclareAttackersParams{Pay: game.CombatCostPayment{TapIDs: []uuid.UUID{helper}}}); err != nil {
		t.Fatalf("attack with a creature tapped: %v", err)
	}
	if !effectstest.IsTapped(g, helper) {
		t.Error("the named creature was not tapped")
	}
}

func TestPcy676HollowWarriorCannotPayWithACreatureAlreadyAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warrior := effectstest.B12Push(g, me.ID, "Hollow Warrior", "Artifact Creature — Golem Warrior", pcy676Warrior, 4, 4)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 1, 1)
	effectstest.B64ReadyToAttack(g, warrior)
	effectstest.B64ReadyToAttack(g, other)
	effectstest.AdvanceToMain(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(other, opp.ID); err != nil {
		t.Fatalf("the other creature could not attack: %v", err)
	}
	if err := g.DeclareAttackerWith(warrior, opp.ID, game.DeclareAttackersParams{Pay: game.CombatCostPayment{TapIDs: []uuid.UUID{other}}}); err == nil {
		t.Error("an attacking creature paid the cost")
	}
}

func TestPcy676HollowWarriorBlocksOnlyByTappingAnotherUntappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	warrior := effectstest.B12Push(g, opp.ID, "Hollow Warrior", "Artifact Creature — Golem Warrior", pcy676Warrior, 4, 4)
	helper := effectstest.B12Creature(g, opp.ID, "Helper", "Creature — Bear", 1, 1)
	effectstest.B64ReadyToAttack(g, attacker)
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(warrior, attacker); err != nil {
		t.Fatalf("staging the block: %v", err)
	}
	if effectstest.IsTapped(g, helper) {
		t.Fatal("staging the block paid the cost")
	}
	if err := g.FinishBlocksWith(opp.ID, game.DeclareBlockersParams{}); err == nil {
		t.Fatal("finished the block without paying")
	}
	if err := g.FinishBlocksWith(opp.ID, game.DeclareBlockersParams{Pay: game.CombatCostPayment{TapIDs: []uuid.UUID{helper}}}); err != nil {
		t.Fatalf("finishing with a creature tapped: %v", err)
	}
	if !effectstest.IsTapped(g, helper) {
		t.Error("the named creature was not tapped")
	}
}

// --- Blessed Wind ---

func TestPcy676BlessedWindSetsTheTargetsLifeTotalToTwenty(t *testing.T) {
	for _, tc := range []struct{ life int }{{5}, {20}, {47}} {
		g := effectstest.NewCatalogGame(t)
		_, victim := effectstest.SeatsForCombat(g)
		victim.Life = tc.life
		effectstest.CastCatalogSpell(t, g, "Blessed Wind", "Sorcery", pcy676Wind, []game.TargetRef{effectstest.PlayerTarget(victim.ID)})
		effectstest.PassPriorityAroundTable(t, g)
		if victim.Life != 20 {
			t.Errorf("a life total of %d became %d, want 20", tc.life, victim.Life)
		}
	}
}

func TestPcy676BlessedWindCanTargetYouAndLeavesOthersAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	me.Life, opp.Life = 3, 33
	effectstest.CastCatalogSpell(t, g, "Blessed Wind", "Sorcery", pcy676Wind, []game.TargetRef{effectstest.PlayerTarget(me.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 20 || opp.Life != 33 {
		t.Errorf("life: you %d (want 20), opponent %d (want 33)", me.Life, opp.Life)
	}
}

// --- Coffin Puppets ---

func pcy676PuppetsTable(t *testing.T, swamp bool, lands int) (*game.Game, *game.Player, uuid.UUID, []uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogGraveyardCard(me, "Coffin Puppets", "Creature — Zombie", pcy676Puppets, 3, 3)
	var ids []uuid.UUID
	for i := range lands {
		name, typeLine := "Forest", "Basic Land — Forest"
		if swamp && i == 0 {
			name, typeLine = "Swamp", "Basic Land — Swamp"
		}
		ids = append(ids, effectstest.SeedLand(g, me.ID, name, typeLine, ""))
	}
	return g, me, id, ids
}

func TestPcy676CoffinPuppetsReturnsDuringYourUpkeepForTwoLandsWithASwamp(t *testing.T) {
	g, me, id, lands := pcy676PuppetsTable(t, true, 3)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	if g.Turn.Step != game.StepUpkeep {
		t.Skipf("the upkeep ended before the activation (step %s)", g.Turn.Step)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: lands[:1]}); err == nil {
		t.Fatal("activated with only one land sacrificed")
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: lands[1:3]}); err != nil {
		t.Fatalf("activate during upkeep: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Errorf("Coffin Puppets did not return (zone %s)", effectstest.B12ZoneOf(g, id))
	}
	if g.Battlefield.Contains(lands[1]) || g.Battlefield.Contains(lands[2]) || !g.Battlefield.Contains(lands[0]) {
		t.Error("exactly the two named lands should be sacrificed")
	}
}

func TestPcy676CoffinPuppetsNeedsASwampAndYourUpkeep(t *testing.T) {
	t.Run("no Swamp", func(t *testing.T) {
		g, me, id, lands := pcy676PuppetsTable(t, false, 3)
		effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
		if g.Turn.Step != game.StepUpkeep {
			t.Skipf("the upkeep ended before the activation (step %s)", g.Turn.Step)
		}
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: lands[:2]}); err == nil {
			t.Error("activated without controlling a Swamp")
		}
	})
	t.Run("main phase", func(t *testing.T) {
		g, me, id, lands := pcy676PuppetsTable(t, true, 3)
		effectstest.AdvanceToMain(t, g)
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: lands[:2]}); err == nil {
			t.Error("activated outside the upkeep")
		}
	})
	t.Run("an opponent's upkeep", func(t *testing.T) {
		g, me, id, lands := pcy676PuppetsTable(t, true, 3)
		effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
		if g.Turn.Step != game.StepUpkeep {
			t.Skipf("the upkeep ended before the activation (step %s)", g.Turn.Step)
		}
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: lands[:2]}); err == nil {
			t.Error("activated during an opponent's upkeep")
		}
	})
}
