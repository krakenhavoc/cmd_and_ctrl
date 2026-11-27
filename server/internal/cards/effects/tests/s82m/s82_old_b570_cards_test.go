package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b570_cards_test.go — Guildpact's Auras, OLD batch 570 (#5599). Each
// test plays the real catalog entry.

const (
	oldB570HypervoltOracle    = "269a2b53-52d7-4032-ba62-7b7742e7a159"
	oldB570FencersOracle      = "5f08f79e-bbbe-4af0-ade6-9ad006a5a0d3"
	oldB570PilloryOracle      = "74776491-29c3-46e6-989b-cf9a00bbfcef"
	oldB570GuardiansOracle    = "95348ba4-038d-4f86-8be0-eb1a68a1f231"
	oldB570SinstrikersOracle  = "98e7b75b-f24d-4317-8b45-9397a4742081"
	oldB570InfiltratorsOracle = "df82b67f-f685-4278-a108-f3689a3a121b"
	oldB570NecromancersOracle = "dffa7c06-096a-47f8-9645-1ba7306ba5b0"
	oldB570ShadowLanceOracle  = "f1a2a78c-2794-4bc9-9b06-49306378ef69"
)

func TestOldB570AurasAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB570HypervoltOracle:    "Hypervolt Grasp",
		oldB570FencersOracle:      "Fencer's Magemark",
		oldB570PilloryOracle:      "Pillory of the Sleepless",
		oldB570GuardiansOracle:    "Guardian's Magemark",
		oldB570SinstrikersOracle:  "Sinstriker's Will",
		oldB570InfiltratorsOracle: "Infiltrator's Magemark",
		oldB570NecromancersOracle: "Necromancer's Magemark",
		oldB570ShadowLanceOracle:  "Shadow Lance",
	} {
		def, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Name != name || def.Completeness != effects.CompletenessFull || len(def.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, def.Name, def.Completeness, def.Caveats)
		}
	}
}

func oldB570Enchant(t *testing.T, g *game.Game, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	return effectstest.AuraCast(t, g, name, oracle, host)
}

func oldB570Wants(t *testing.T, g *game.Game, id uuid.UUID, what string, p, tough int) {
	t.Helper()
	if gp, gt := effectstest.EffectivePTOf(t, g, id); gp != p || gt != tough {
		t.Errorf("%s is %d/%d, want %d/%d", what, gp, gt, p, tough)
	}
}

// --- the Magemarks ------------------------------------------------------

func TestOldB570GuardiansMagemarkPumpsEveryCreatureYouControlThatIsEnchanted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.AuraBear(g, me.ID)
	plain := effectstest.AuraBear(g, me.ID)
	theirs := effectstest.AuraBear(g, opp.ID)
	oldB570Enchant(t, g, "Guardian's Magemark", oldB570GuardiansOracle, mine)
	oldB570Wants(t, g, mine, "the enchanted bear", 3, 3)
	oldB570Wants(t, g, plain, "an unenchanted bear", 2, 2)

	// The Aura's controller's creatures that are enchanted by ANY Aura get it.
	oldB570Enchant(t, g, "Guardian's Magemark", oldB570GuardiansOracle, plain)
	oldB570Wants(t, g, plain, "the second enchanted bear", 4, 4) // two Magemarks apply to it
	oldB570Wants(t, g, mine, "the first enchanted bear", 4, 4)

	// On an opponent's creature it pumps only your own enchanted creatures.
	oldB570Enchant(t, g, "Guardian's Magemark", oldB570GuardiansOracle, theirs)
	oldB570Wants(t, g, theirs, "the opponent's enchanted bear", 2, 2)
}

func TestOldB570GuardiansMagemarkHasFlash(t *testing.T) {
	def, _ := effects.Lookup(oldB570GuardiansOracle)
	if !effects.KeywordSliceContains(def.PrintedKeywords, "flash") {
		t.Errorf("printed keywords %v, want flash", def.PrintedKeywords)
	}
}

func TestOldB570FencersMagemarkPumpsAndGivesFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	plain := effectstest.AuraBear(g, me.ID)
	oldB570Enchant(t, g, "Fencer's Magemark", oldB570FencersOracle, bear)
	oldB570Wants(t, g, bear, "the enchanted bear", 3, 3)
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "first strike") {
		t.Error("the enchanted bear has no first strike")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, plain, "first strike") {
		t.Error("an unenchanted bear has first strike")
	}
}

func TestOldB570InfiltratorsMagemarkIsBlockableOnlyByDefenders(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.AuraBear(g, me.ID)
	plain := effectstest.AuraBear(g, me.ID)
	oldB570Enchant(t, g, "Infiltrator's Magemark", oldB570InfiltratorsOracle, bear)
	oldB570Wants(t, g, bear, "the enchanted bear", 3, 3)
	wall := effectstest.B12Push(g, opp.ID, "Wall", "Creature — Wall", "", 0, 4)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == wall {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "defender")
			}
		}
	})
	plainBlocker := effectstest.B12Creature(g, opp.ID, "Plain Blocker", "Creature — Bear", 2, 2)
	effectstest.Pr7bAttack(t, g, opp.ID, bear, plain)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(plainBlocker, bear); err == nil {
		t.Error("a creature without defender blocked the enchanted creature")
	}
	if err := g.DeclareBlocker(plainBlocker, plain); err != nil {
		t.Errorf("an unenchanted attacker could not be blocked: %v", err)
	}
	if err := g.DeclareBlocker(wall, bear); err != nil {
		t.Errorf("a creature with defender could not block the enchanted creature: %v", err)
	}
}

func TestOldB570NecromancersMagemarkReturnsAnEnchantedCreatureToHandInsteadOfDying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	plain := effectstest.AuraBear(g, me.ID)
	oldB570Enchant(t, g, "Necromancer's Magemark", oldB570NecromancersOracle, bear)
	oldB570Wants(t, g, bear, "the enchanted bear", 3, 3)

	effectstest.KillCreature(t, g, me.ID, plain)
	g.RunStateChecksForTest()
	if !me.Graveyard.Contains(plain) {
		t.Error("an unenchanted creature did not die")
	}

	effectstest.KillCreature(t, g, me.ID, bear)
	g.RunStateChecksForTest()
	if me.Graveyard.Contains(bear) {
		t.Error("the enchanted creature died")
	}
	if !me.Hand.Contains(bear) {
		t.Error("the enchanted creature was not returned to its owner's hand")
	}
}

// --- the other Auras ------------------------------------------------------

func TestOldB570HypervoltGraspPingsAnyTargetAndReturnsToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.AuraBear(g, me.ID)
	aura := oldB570Enchant(t, g, "Hypervolt Grasp", oldB570HypervoltOracle, bear)
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d→%d, want 1 damage", life, opp.Life)
	}
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Error("the enchanted creature tapped twice")
	}

	// Not enough mana: the Aura stays.
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("returned the Aura with no mana")
	}
	effectstest.ApaMana(me, "U", "C")
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("return the Aura: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(aura) {
		t.Error("the Aura is not back in its owner's hand")
	}
}

func TestOldB570PilloryOfTheSleeplessStopsCombatAndDrainsTheControllerEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	oldB570Enchant(t, g, "Pillory of the Sleepless", oldB570PilloryOracle, bear)
	if r := effectstest.AuraRestrictions(t, g, bear); r&game.CantAttackOrBlock != game.CantAttackOrBlock {
		t.Fatalf("restrictions %v, want can't attack or block", r)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, g.Seats[1].ID); err == nil {
		t.Error("the enchanted creature attacked")
	}
	life := me.Life
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life %d→%d at my upkeep, want -1", life, me.Life)
	}
}

func TestOldB570SinstrikersWillDamagesAnAttackingOrBlockingCreatureEqualToPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Ogre", 3, 3)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 3)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 3)
	oldB570Enchant(t, g, "Sinstriker's Will", oldB570SinstrikersOracle, host)
	act := func(target uuid.UUID) error {
		return g.ActivateCatalogAbility(me.ID, host, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(target)})
	}
	if err := act(idle); err == nil {
		t.Error("targeted a creature that is neither attacking nor blocking")
	}
	effectstest.Pr7bAttack(t, g, opp.ID, attacker)
	if err := act(attacker); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(attacker) {
		t.Error("3 damage did not destroy the 2/3 attacker")
	}
}

func TestOldB570ShadowLanceGivesFirstStrikeAndPumpsTheEnchantedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	lance := oldB570Enchant(t, g, "Shadow Lance", oldB570ShadowLanceOracle, bear)
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "first strike") {
		t.Error("the enchanted bear has no first strike")
	}
	if err := g.ActivateCatalogAbility(me.ID, lance, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("pumped with no mana")
	}
	effectstest.ApaMana(me, "B", "C")
	if err := g.ActivateCatalogAbility(me.ID, lance, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	oldB570Wants(t, g, bear, "the pumped bear", 4, 4)
}
