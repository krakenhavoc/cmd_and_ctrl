package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b747_test.go — S82 OLD batch 747 (Stronghold damage, #5627),
// through the real engine.

const (
	oldB747MobJustice = "0ff97869-f27c-4379-92fd-605062aa5418"
	oldB747Ring       = "29cb8f47-7bb5-4881-98ef-59fc858cf9d7"
	oldB747WallSouls  = "2c85c9ea-c3fd-4981-a65c-4405bd87764e"
	oldB747Heat       = "400285bb-ded9-4018-a7ea-3e6f4d24b900"
	oldB747Maniac     = "53b86704-4f16-498e-8c2c-e661cb85eb9c"
	oldB747Bombers    = "69812cd2-9885-45ed-9d79-658597bfeb8b"
	oldB747Angel      = "97140551-341a-40a1-93fe-0fcaa0c3be33"
	oldB747Acidic     = "a149eeb1-007c-435d-9d55-34af205090e0"
	oldB747FlameWave  = "cee8ec5f-e36d-4bca-9b19-cb33113835b1"
	oldB747Bullwhip   = "d4e7a9f5-5228-4e41-8f58-6e7453fe68dc"
	oldB747Basilisk   = "effe239e-ade0-4ee2-99d7-f93b2f5ef585"
)

func TestS82OldB747CardsAreFull(t *testing.T) {
	for _, id := range []string{oldB747MobJustice, oldB747Ring, oldB747WallSouls, oldB747Heat, oldB747Maniac, oldB747Bombers,
		oldB747Angel, oldB747Acidic, oldB747FlameWave, oldB747Bullwhip, oldB747Basilisk} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Fatalf("%s is not registered", id)
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB747Target(t *testing.T, g *game.Game, ref game.TargetRef) {
	t.Helper()
	c := effectstest.FirstPendingTarget(g)
	if c == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(c.ID, c.Chooser, ref); err != nil {
		t.Fatalf("picking the target: %v", err)
	}
}

func oldB747Settle(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 8; i++ {
		effectstest.PassPriorityAroundTable(t, g)
		if len(g.PendingChoices) > 0 {
			return
		}
	}
}

func TestS82OldB747MobJusticeCountsYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	}
	effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Mob Justice", "Sorcery", oldB747MobJustice, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d -> %d, want 3 damage for the three creatures I control", life, opp.Life)
	}
}

func TestS82OldB747MobJusticeCantTargetACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	err := effectstest.CastCatalogSpellErr(t, g, "Mob Justice", "Sorcery", oldB747MobJustice, effectstest.CardRefs(bear))
	if err == nil {
		t.Fatal("Mob Justice targeted a creature")
	}
}

func TestS82OldB747FlameWaveHitsThePlayerAndTheirCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	a := effectstest.B12Creature(g, opp.ID, "Theirs A", "Creature — Bear", 2, 4)
	b := effectstest.B12Creature(g, opp.ID, "Theirs B", "Creature — Bear", 2, 5)
	other := effectstest.B12Creature(g, third.ID, "Other", "Creature — Bear", 2, 2)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Flame Wave", "Sorcery", oldB747FlameWave, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("life %d -> %d, want 4 damage", life, opp.Life)
	}
	if g.Battlefield.Contains(a) {
		t.Error("the 2/4 survived 4 damage")
	}
	if !g.Battlefield.Contains(b) || effectstest.DamageMarkedOn(g, b) != 4 {
		t.Errorf("the 2/5 should have 4 damage marked (marked %d)", effectstest.DamageMarkedOn(g, b))
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(other) || effectstest.DamageMarkedOn(g, mine) != 0 || effectstest.DamageMarkedOn(g, other) != 0 {
		t.Error("a creature of another player took damage")
	}
}

func TestS82OldB747JinxedRingDamagesYouForYourNontokenPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Jinxed Ring", "Artifact", oldB747Ring, 0, 0)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	tokenID := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Spirit", TypeLine: "Token Creature — Spirit",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	life := me.Life
	effectstest.P1Destroy(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("an opponent's creature dying cost me %d life", life-me.Life)
	}
	effectstest.P1Destroy(t, g, tokenID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("a token dying cost me %d life", life-me.Life)
	}
	effectstest.P1Destroy(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life %d -> %d, want 1 damage for my nontoken creature", life, me.Life)
	}
}

func TestS82OldB747JinxedRingChangesHandsForACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ring := effectstest.B12Push(g, me.ID, "Jinxed Ring", "Artifact", oldB747Ring, 0, 0)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, ring, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}, SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err == nil {
		t.Fatal("targeted myself")
	}
	if !g.Battlefield.Contains(fodder) {
		t.Fatal("a refused activation sacrificed the creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, ring, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, ring); !ok || c.Controller != opp.ID {
		t.Errorf("the Ring is controlled by %v, want the opponent", c.Controller)
	}
}

func TestS82OldB747WallOfSoulsReflectsCombatDamageOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Ogre", 3, 3)
	wall := effectstest.B12Push(g, opp.ID, "Wall of Souls", "Creature — Wall", oldB747WallSouls, 0, 4)
	pinger := effectstest.B12Creature(g, me.ID, "Pinger", "Creature — Goblin", 1, 1)
	if !effectstest.HasKeywordOnBattlefield(t, g, wall, "defender") {
		t.Error("no defender")
	}
	effectstest.Cmdr2B08Damage(g, pinger, wall, 2)
	oldB747Settle(t, g)
	if effectstest.FirstPendingTarget(g) != nil {
		t.Fatal("noncombat damage triggered the Wall")
	}
	life := me.Life
	effectstest.Pr7bAttack(t, g, opp.ID, atk)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{wall: atk})
	effectstest.Pr7bDamageStep(t, g)
	oldB747Settle(t, g)
	oldB747Target(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	oldB747Settle(t, g)
	if me.Life != life-3 {
		t.Errorf("life %d -> %d, want the 3 combat damage reflected", life, me.Life)
	}
}

func TestS82OldB747MoggManiacReflectsLethalDamageToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	maniac := effectstest.B12Push(g, opp.ID, "Mogg Maniac", "Creature — Goblin", oldB747Maniac, 1, 1)
	pinger := effectstest.B12Creature(g, me.ID, "Pinger", "Creature — Goblin", 1, 1)
	life := me.Life
	effectstest.Cmdr2B08Damage(g, pinger, maniac, 1)
	oldB747Target(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	oldB747Settle(t, g)
	if g.Battlefield.Contains(maniac) {
		t.Fatal("the 1/1 Maniac survived 1 damage")
	}
	if me.Life != life-1 {
		t.Errorf("life %d -> %d, want the 1 damage reflected by the dead Maniac", life, me.Life)
	}
}

func TestS82OldB747MoggManiacReflectsAnyDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	maniac := effectstest.B12Push(g, opp.ID, "Mogg Maniac", "Creature — Goblin", oldB747Maniac, 1, 3)
	pinger := effectstest.B12Creature(g, me.ID, "Pinger", "Creature — Goblin", 1, 1)
	life := me.Life
	effectstest.Cmdr2B08Damage(g, pinger, maniac, 1)
	oldB747Target(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	oldB747Settle(t, g)
	if me.Life != life-1 {
		t.Errorf("life %d -> %d, want the 1 damage reflected", life, me.Life)
	}
}

func TestS82OldB747HeatOfBattleHurtsTheBlockersController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Heat of Battle", "Enchantment", oldB747Heat, 0, 0)
	atk := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Ogre", 3, 3)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	life, mine := opp.Life, me.Life
	effectstest.Pr7bAttack(t, g, opp.ID, atk)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{wall: atk})
	oldB747Settle(t, g)
	if opp.Life != life-1 {
		t.Errorf("the blocker's controller: life %d -> %d, want 1 damage", life, opp.Life)
	}
	if me.Life != mine {
		t.Errorf("the enchantment's controller lost %d life", mine-me.Life)
	}
}

func TestS82OldB747MoggBombersSacrificesItselfWhenAnotherCreatureEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bombers := effectstest.B12Push(g, me.ID, "Mogg Bombers", "Creature — Goblin", oldB747Bombers, 3, 4)
	oldB747Settle(t, g)
	if effectstest.FirstPendingTarget(g) != nil || !g.Battlefield.Contains(bombers) {
		t.Fatal("the Bombers triggered off its own entry")
	}
	life := opp.Life
	effectstest.AdvanceToMain(t, g)
	effectstest.CastHoldCreature(t, g, "Newcomer", "Creature — Bear", "", 1, 1)
	oldB747Target(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	oldB747Settle(t, g)
	if g.Battlefield.Contains(bombers) || !me.Graveyard.Contains(bombers) {
		t.Error("the Bombers was not sacrificed")
	}
	if opp.Life != life-3 {
		t.Errorf("life %d -> %d, want 3 damage", life, opp.Life)
	}
}

func TestS82OldB747WarriorAngelGainsLifeForDamageDealt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	angel := effectstest.B12Push(g, me.ID, "Warrior Angel", "Creature — Angel Warrior", oldB747Angel, 3, 4)
	if !effectstest.HasKeywordOnBattlefield(t, g, angel, "flying") {
		t.Error("no flying")
	}
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 5)
	life := me.Life
	effectstest.Cmdr2B08Damage(g, angel, bear, 3)
	oldB747Settle(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want +3", life, me.Life)
	}
}

func TestS82OldB747AcidicSliverGivesEverySliverASacrificeAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Acidic Sliver", "Creature — Sliver", oldB747Acidic, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Plain Sliver", "Creature — Sliver", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Sliver", "Creature — Sliver", 1, 1)
	notSliver := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.RavB2Activate(t, g, me.ID, notSliver, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}, "C", "C"); err == nil {
		t.Fatal("a non-Sliver had the ability")
	}
	if err := effectstest.RavB2Activate(t, g, me.ID, theirs, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}, "C", "C"); err == nil {
		t.Fatal("I activated an opponent's Sliver's ability")
	}
	life := opp.Life
	effectstest.RavB2MustActivate(t, g, me.ID, mine, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}, "C", "C")
	if opp.Life != life-2 {
		t.Errorf("life %d -> %d, want 2 damage", life, opp.Life)
	}
	if g.Battlefield.Contains(mine) {
		t.Error("the Sliver was not sacrificed")
	}
	oppLife := me.Life
	effectstest.RavB2MustActivate(t, g, opp.ID, theirs, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}, "C", "C")
	if me.Life != oppLife-2 {
		t.Errorf("the opponent's Sliver: life %d -> %d, want 2 damage", oppLife, me.Life)
	}
}

func TestS82OldB747BullwhipPingsAndForcesTheAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	whip := effectstest.B12Push(g, opp.ID, "Bullwhip", "Artifact", oldB747Bullwhip, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.RavB2Activate(t, g, opp.ID, whip, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}, "C"); err == nil {
		t.Fatal("activated for {1}")
	}
	effectstest.RavB2MustActivate(t, g, opp.ID, whip, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}, "C", "C")
	if got := effectstest.DamageMarkedOn(g, bear); got != 1 {
		t.Errorf("damage marked %d, want 1", got)
	}
	if !effectstest.HasRequirementRecord(g) {
		t.Error("no attack requirement was recorded")
	}
}

func TestS82OldB747LowlandBasiliskDestroysWhatItDamagedAtEndOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bas := effectstest.B12Push(g, me.ID, "Lowland Basilisk", "Creature — Basilisk", oldB747Basilisk, 1, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 5)
	effectstest.Pr7bAttack(t, g, opp.ID, bas)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{bear: bas})
	effectstest.Pr7bDamageStep(t, g)
	oldB747Settle(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the Bear was destroyed before the end of combat")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	oldB747Settle(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the damaged creature survived the end of combat")
	}
}
