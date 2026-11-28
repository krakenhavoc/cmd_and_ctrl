package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB835CelestialSword  = "8f64c3d4-8ed6-43be-9583-33de4b7761fa"
	oldB835WalkingWall     = "909201e5-5599-442a-9e4f-ff61f187e957"
	oldB835ChubToad        = "a0daace2-92fc-4c7d-b497-94b4ba775559"
	oldB835EliteGuard      = "a1f6df2a-fc97-44d4-a0d1-902d8df94002"
	oldB835JohtullWurm     = "ab5a8242-4e42-43e6-a6ee-eeb0a8db3d11"
	oldB835BattleCry       = "b92afdba-5aea-42c8-bdd8-6ba5ff6cba4f"
	oldB835GrizzledWolf    = "da4ed552-a85a-47e1-ab67-5c93a9a20048"
	oldB835OrderWhiteShld  = "e0b3cfc1-68d4-46a8-9ae4-71bdb9c87889"
	oldB835GoblinSkiPatrol = "e98396d0-6952-4755-992f-813f565b9189"
	oldB835KarplusanGiant  = "f0fac3bc-a566-448e-a9a2-10927a49f182"
	oldB835FanaticalFever  = "f1231def-96fb-4408-a0c9-25d6932d7ead"
)

var oldB835Cards = map[string]string{
	oldB835CelestialSword: "Celestial Sword", oldB835WalkingWall: "Walking Wall", oldB835ChubToad: "Chub Toad",
	oldB835EliteGuard: "Kjeldoran Elite Guard", oldB835JohtullWurm: "Johtull Wurm", oldB835BattleCry: "Battle Cry",
	oldB835GrizzledWolf: "Grizzled Wolverine", oldB835OrderWhiteShld: "Order of the White Shield",
	oldB835GoblinSkiPatrol: "Goblin Ski Patrol", oldB835KarplusanGiant: "Karplusan Giant",
	oldB835FanaticalFever: "Fanatical Fever",
}

func TestOldB835EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range oldB835Cards {
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

func oldB835Creature(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int, colors ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner, Colors: colors,
	})
}

func oldB835PT(t *testing.T, g *game.Game, id uuid.UUID, wantP, wantT int, what string) {
	t.Helper()
	if p, tough := effectstest.PtOf(t, g, id); p != wantP || tough != wantT {
		t.Errorf("%s is %d/%d, want %d/%d", what, p, tough, wantP, wantT)
	}
}

func oldB835Cast(t *testing.T, g *game.Game, who *game.Player, name, oracle, mana string, targets []game.TargetRef) {
	t.Helper()
	id := uuid.New()
	who.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", OracleID: oracle, Owner: who.ID, Controller: who.ID})
	effectstest.FloatMana(t, g, who, mana)
	if err := g.CastSpell(who.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB835CelestialSwordPumpsThenSacrificesAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	sword := effectstest.B12Push(g, me.ID, "Celestial Sword", "Artifact", oldB835CelestialSword, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{G}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, sword, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err == nil {
		t.Fatal("an opponent's creature was a legal target")
	}
	effectstest.B16Activate(t, g, me.ID, sword, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	oldB835PT(t, g, bear, 5, 5, "the Bear")
	effectstest.AdvanceToStepInTurn(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, bear) {
		t.Error("the Bear was not sacrificed at the beginning of the end step")
	}
}

func TestOldB835WalkingWallAttacksOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wall := effectstest.TypedCreature(g, me.ID, "Walking Wall", "Artifact Creature — Wall", oldB835WalkingWall, 0, 6, "defender")
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}")
	effectstest.B16Activate(t, g, me.ID, wall, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	oldB835PT(t, g, wall, 3, 5, "the Wall")
	if err := g.ActivateCatalogAbility(me.ID, wall, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated a second time in one turn")
	}
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, wall)
}

func TestOldB835ChubToadAndJohtullWurm(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toad := effectstest.TypedCreature(g, me.ID, "Chub Toad", "Creature — Frog", oldB835ChubToad, 1, 1, "haste")
	wurm := effectstest.TypedCreature(g, me.ID, "Johtull Wurm", "Creature — Wurm", oldB835JohtullWurm, 6, 6, "haste")
	b1 := effectstest.B12Creature(g, opp.ID, "B1", "Creature — Bear", 0, 9)
	b2 := effectstest.B12Creature(g, opp.ID, "B2", "Creature — Bear", 0, 9)
	b3 := effectstest.B12Creature(g, opp.ID, "B3", "Creature — Bear", 0, 9)
	effectstest.DeclareAttack(t, g, opp.ID, toad, wurm)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	for _, p := range [][2]uuid.UUID{{b1, toad}, {b2, wurm}, {b3, wurm}} {
		if err := g.DeclareBlocker(p[0], p[1]); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
	}
	effectstest.LockInBlocks(t, g)
	effectstest.AnswerTriggerOrderInOfferedOrder(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	oldB835PT(t, g, toad, 3, 3, "the blocked Chub Toad")
	oldB835PT(t, g, wurm, 4, 5, "Johtull Wurm blocked by two")
}

func TestOldB835ChubToadBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := effectstest.B12Creature(g, me.ID, "Atk", "Creature — Bear", 1, 9)
	toad := effectstest.B12Push(g, opp.ID, "Chub Toad", "Creature — Frog", oldB835ChubToad, 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, atk)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(toad, atk); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	oldB835PT(t, g, toad, 3, 3, "the blocking Chub Toad")
}

func TestOldB835EliteGuardSacrificesItselfWhenTheTargetLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guard := effectstest.B12Push(g, me.ID, "Kjeldoran Elite Guard", "Creature — Human Soldier", oldB835EliteGuard, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	if err := g.ActivateCatalogAbility(me.ID, guard, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("activate during combat: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	oldB835PT(t, g, bear, 4, 4, "the target")
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, guard) {
		t.Error("the Guard was not sacrificed")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	guard2 := effectstest.B12Push(g2, me2.ID, "Kjeldoran Elite Guard", "Creature — Human Soldier", oldB835EliteGuard, 2, 2)
	bear2 := effectstest.B12Creature(g2, me2.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g2)
	if err := g2.ActivateCatalogAbility(g2.Seats[g2.Turn.ActiveSeat].ID, guard2, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear2}}}); err == nil {
		t.Error("activated outside combat")
	}
}

func TestOldB835BattleCryUntapsWhiteAndGivesBlockersToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	white := oldB835Creature(g, me.ID, "White Bear", "Creature — Bear", "", 2, 2, "W")
	green := oldB835Creature(g, me.ID, "Green Bear", "Creature — Bear", "", 2, 2, "G")
	atk := oldB835Creature(g, me.ID, "Atk", "Creature — Bear", "", 1, 1, "R")
	blocker := oldB835Creature(g, opp.ID, "Blocker", "Creature — Bear", "", 1, 6, "G")
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, white).Tapped = true
		effectstest.FindBattlefieldCardForTest(g, green).Tapped = true
	})
	oldB835Cast(t, g, me, "Battle Cry", oldB835BattleCry, "{W}{W}{W}", nil)
	if effectstest.IsTapped(g, white) {
		t.Error("the white creature stayed tapped")
	}
	if !effectstest.IsTapped(g, green) {
		t.Error("the green creature was untapped")
	}
	effectstest.DeclareAttack(t, g, opp.ID, atk)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, atk); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	oldB835PT(t, g, blocker, 1, 7, "the blocker")
}

func TestOldB835GrizzledWolverineNeedsABlockerAndOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wolf := effectstest.B12Push(g, me.ID, "Grizzled Wolverine", "Creature — Wolverine", oldB835GrizzledWolf, 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 0, 9)
	effectstest.FloatMana(t, g, me, "{R}{R}")
	effectstest.DeclareAttack(t, g, opp.ID, wolf)
	if err := g.ActivateCatalogAbility(me.ID, wolf, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated before blockers")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, wolf); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.B16Activate(t, g, me.ID, wolf, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	oldB835PT(t, g, wolf, 4, 2, "the Wolverine")
	if err := g.ActivateCatalogAbility(me.ID, wolf, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated twice in one turn")
	}
}

func TestOldB835OrderOfTheWhiteShieldAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	k := effectstest.TypedCreature(g, me.ID, "Order of the White Shield", "Creature — Human Knight", oldB835OrderWhiteShld, 2, 1, "protection from black")
	effectstest.FloatMana(t, g, me, "{W}{W}{W}")
	effectstest.B16Activate(t, g, me.ID, k, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B16Activate(t, g, me.ID, k, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, k, "first strike") {
		t.Error("no first strike")
	}
	oldB835PT(t, g, k, 3, 1, "the Knight")
}

func TestOldB835GoblinSkiPatrolNeedsASnowMountainAndGoesOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	gob := effectstest.B12Push(g, me.ID, "Goblin Ski Patrol", "Creature — Goblin", oldB835GoblinSkiPatrol, 1, 1)
	effectstest.FloatMana(t, g, me, "{R}{R}{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, gob, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without a snow Mountain")
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Snow-Covered Mountain", "Basic Snow Land — Mountain", "", false)
	effectstest.B16Activate(t, g, me.ID, gob, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	oldB835PT(t, g, gob, 3, 1, "the Goblin")
	if !effectstest.HasEffectiveKeyword(t, g, gob, "flying") {
		t.Error("no flying")
	}
	if err := g.ActivateCatalogAbility(me.ID, gob, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated a second time")
	}
	effectstest.AdvanceToStepInTurn(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, gob) {
		t.Error("the Goblin was not sacrificed")
	}
}

func TestOldB835KarplusanGiantTapsSnowLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	giant := effectstest.B12Push(g, me.ID, "Karplusan Giant", "Creature — Giant", oldB835KarplusanGiant, 3, 3)
	snow := effectstest.PushCatalogPermanent(g, me.ID, "Snow-Covered Mountain", "Basic Snow Land — Mountain", "", false)
	plain := effectstest.PushCatalogPermanent(g, me.ID, "Mountain", "Basic Land — Mountain", "", false)
	_ = plain
	if err := g.ActivateCatalogAbility(me.ID, giant, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{plain}}); err == nil {
		t.Fatal("a nonsnow land paid the cost")
	}
	effectstest.B16Activate(t, g, me.ID, giant, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{snow}})
	effectstest.PassPriorityAroundTable(t, g)
	oldB835PT(t, g, giant, 4, 4, "the Giant")
	if !effectstest.IsTapped(g, snow) {
		t.Error("the snow land was not tapped")
	}
	if effectstest.IsTapped(g, plain) {
		t.Error("the nonsnow land was tapped")
	}
}

func TestOldB835FanaticalFever(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	oldB835Cast(t, g, me, "Fanatical Fever", oldB835FanaticalFever, "{G}{G}{G}{G}", []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	oldB835PT(t, g, bear, 5, 2, "the Bear")
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("no trample")
	}
}
