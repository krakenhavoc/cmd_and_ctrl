package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b744_test.go — Exodus, S82 OLD batch 744 (#5626).

const (
	oldB744KeeperDead = "50ddd513-e06e-4728-b6cd-981839d0fba6"
	oldB744WallNets   = "cd112962-b297-403c-bede-c8b587f379b3"
	oldB744Plague     = "f8cc1dc4-740d-4c5b-ac51-7355730a6595"
	oldB744Warlord    = "42bb6537-fa54-43b7-84f2-dea41adc35fd"
	oldB744Visionary  = "7971f5c2-97b7-441f-ad77-954f43ede0a1"
	oldB744Jackal     = "8847b057-8896-4c72-a11f-3c183b1cc60e"
	oldB744Limited    = "5b287dcc-f853-4df1-8ddf-5b8b2adc2873"
	oldB744WarBeast   = "c6ad434b-6e34-439b-8046-f0047c06e371"
	oldB744Boars      = "a4c59feb-e76a-4e08-b296-21c3e61cfc64"
	oldB744Herd       = "bcc8ce61-a881-42be-ab2c-64351d17f9f8"
)

func TestOldB744CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB744KeeperDead: "Keeper of the Dead", oldB744WallNets: "Wall of Nets", oldB744Plague: "Plaguebearer",
		oldB744Warlord: "Dauthi Warlord", oldB744Visionary: "Soltari Visionary", oldB744Jackal: "Dauthi Jackal",
		oldB744Limited: "Limited Resources", oldB744WarBeast: "Skyshroud War Beast", oldB744Boars: "Crashing Boars",
		oldB744Herd: "Jackalope Herd",
	} {
		effectstest.RequireFullNoCaveats(t, oracle, name)
	}
}

func TestOldB744KeeperOfTheDeadNeedsAnOpponentTwoCreatureCardsBehind(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	keeper := effectstest.PushCatalogPermanent(g, me.ID, "Keeper of the Dead", "Creature — Human Wizard", oldB744KeeperDead, false)
	bear := effectstest.B27Push(g, opp.ID, "Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	zombie := effectstest.B27Push(g, opp.ID, "Zombie", "Creature — Zombie", "", "{1}{B}", 2, 2, "B")
	thirdsBear := effectstest.B27Push(g, third.ID, "Third Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	for i := 0; i < 3; i++ {
		effectstest.PushCatalogGraveyardCard(me, "Mine", "Creature — Bear", "", 1, 1)
	}
	effectstest.PushCatalogGraveyardCard(me, "Mine Spell", "Sorcery", "", 0, 0)
	effectstest.PushCatalogGraveyardCard(opp, "Theirs", "Creature — Bear", "", 1, 1)
	targets := func(creature uuid.UUID) game.ActivateAbilityParams {
		return game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{
			{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: creature}}}
	}
	effectstest.DcMana(t, g, me.ID, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, targets(zombie)); err == nil {
		t.Fatal("a black creature was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, targets(thirdsBear)); err == nil {
		t.Fatal("a creature another player controls was accepted")
	}
	// 3 creature cards against 1 is a difference of two, which is enough.
	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, targets(bear)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the nonblack creature was not destroyed")
	}
	if !g.Battlefield.Contains(zombie) || !g.Battlefield.Contains(thirdsBear) {
		t.Error("another creature was destroyed")
	}
}

func TestOldB744KeeperOfTheDeadRefusesAnOpponentOnlyOneBehind(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	keeper := effectstest.PushCatalogPermanent(g, me.ID, "Keeper of the Dead", "Creature — Human Wizard", oldB744KeeperDead, false)
	bear := effectstest.B27Push(g, opp.ID, "Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	for i := 0; i < 2; i++ {
		effectstest.PushCatalogGraveyardCard(me, "Mine", "Creature — Bear", "", 1, 1)
	}
	effectstest.PushCatalogGraveyardCard(opp, "Theirs", "Creature — Bear", "", 1, 1)
	effectstest.DcMana(t, g, me.ID, "{B}")
	err := g.ActivateCatalogAbility(me.ID, keeper, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: bear}}})
	if err == nil {
		t.Fatal("an opponent with only one fewer creature card was accepted")
	}
}

func TestOldB744WallOfNetsExilesWhatItBlockedAndReturnsItWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	bystander := effectstest.B12Creature(g, me.ID, "Bystander", "Creature — Bear", 2, 2)
	wall := effectstest.B12Push(g, opp.ID, "Wall of Nets", "Creature — Wall", oldB744WallNets, 0, 7)
	effectstest.AttackIntoBlocks(t, g, opp.ID, attacker)
	if err := g.DeclareBlocker(wall, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInCombatDeclaration(t, g)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Fatal("the blocked creature was not exiled at end of combat")
	}
	if !g.Battlefield.Contains(bystander) {
		t.Error("a creature the Wall did not block was exiled")
	}
	if z := g.FindCardZoneForEffect(attacker); z == nil || z.Kind != game.ZoneExile {
		t.Fatalf("the blocked creature is in %+v, want exile", z)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(wall); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	// It returns as a new object (CR 400.7), under its owner's control.
	if effectstest.CountOnBattlefield(g, "Attacker", me.ID) != 1 {
		t.Fatalf("the exiled creature did not return to its owner's battlefield when the Wall left")
	}
	if effectstest.CountOnBattlefield(g, "Attacker", opp.ID) != 0 {
		t.Error("it returned under the Wall's controller")
	}
}

func TestOldB744WallOfNetsExilesNothingItDidNotBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.B12Push(g, opp.ID, "Wall of Nets", "Creature — Wall", oldB744WallNets, 0, 7)
	effectstest.AttackIntoBlocks(t, g, opp.ID, attacker)
	effectstest.LockInCombatDeclaration(t, g)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(attacker) {
		t.Error("an unblocked attacker was exiled")
	}
}

func TestOldB744PlaguebearerKillsANonblackCreatureOfManaValueX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bearer := effectstest.PushCatalogPermanent(g, me.ID, "Plaguebearer", "Creature — Zombie", oldB744Plague, false)
	two := effectstest.B27Push(g, opp.ID, "Two", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	three := effectstest.B27Push(g, opp.ID, "Three", "Creature — Bear", "", "{2}{G}", 3, 3, "G")
	black := effectstest.B27Push(g, opp.ID, "Black Two", "Creature — Zombie", "", "{1}{B}", 2, 2, "B")
	effectstest.DcMana(t, g, me.ID, "{C}{C}{C}{C}{B}")
	try := func(target uuid.UUID, x int) error {
		return g.ActivateCatalogAbility(me.ID, bearer, 0, game.ActivateAbilityParams{Strict: true, XValue: x, Targets: effectstest.CardRefs(target)})
	}
	if err := try(three, 2); err == nil {
		t.Fatal("a mana value 3 creature was accepted at X=2")
	}
	if err := try(black, 2); err == nil {
		t.Fatal("a black creature was accepted")
	}
	if err := try(two, 2); err != nil {
		t.Fatalf("activate at X=2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(two) {
		t.Error("the mana value 2 creature survived")
	}
	if !g.Battlefield.Contains(three) || !g.Battlefield.Contains(black) {
		t.Error("another creature was destroyed")
	}
	// X=2 cost {4}{B}: the pool is empty now.
	if err := try(three, 3); err == nil {
		t.Error("activated with no mana")
	}
}

func TestOldB744DauthiWarlordCountsShadowCreaturesOnTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warlord := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dauthi Warlord", TypeLine: "Creature — Dauthi Soldier", OracleID: oldB744Warlord,
		Keywords: []string{"shadow"}, Power: 0, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	if p, tough := effectstest.PtOf(t, g, warlord); p != 1 || tough != 1 {
		t.Errorf("alone it is %d/%d, want 1/1 (it counts itself)", p, tough)
	}
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Shade", TypeLine: "Creature — Shade", Keywords: []string{"shadow"},
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "A Shadow Land", TypeLine: "Land", Keywords: []string{"shadow"}, Owner: opp.ID, Controller: opp.ID,
	})
	if p, tough := effectstest.PtOf(t, g, warlord); p != 2 || tough != 1 {
		t.Errorf("with one other shadow creature it is %d/%d, want 2/1", p, tough)
	}
}

func TestOldB744SoltariVisionaryDestroysAnEnchantmentOfTheDamagedPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	vis := effectstest.B12Push(g, me.ID, "Soltari Visionary", "Creature — Soltari Cleric", oldB744Visionary, 2, 2)
	theirs1 := effectstest.AcPermanent(g, opp.ID, "Their Aura", "Enchantment", 0, 0)
	theirs2 := effectstest.AcPermanent(g, opp.ID, "Their Other Aura", "Enchantment", 0, 0)
	elsewhere := effectstest.AcPermanent(g, other.ID, "Third Aura", "Enchantment", 0, 0)
	mine := effectstest.AcPermanent(g, me.ID, "My Aura", "Enchantment", 0, 0)
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(vis, opp.ID, 2); err != nil {
			t.Fatal(err)
		}
	})
	if p := effectstest.LatestPickTarget(g, me.ID); p == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.PickCard(t, g, me.ID, theirs1)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs1) {
		t.Error("the chosen enchantment survived")
	}
	for name, id := range map[string]uuid.UUID{"their other": theirs2, "third player's": elsewhere, "mine": mine} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s enchantment was destroyed", name)
		}
	}
}

func TestOldB744SoltariVisionaryHasNoTargetAgainstAPlayerWithoutEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vis := effectstest.B12Push(g, me.ID, "Soltari Visionary", "Creature — Soltari Cleric", oldB744Visionary, 2, 2)
	mine := effectstest.AcPermanent(g, me.ID, "My Aura", "Enchantment", 0, 0)
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(vis, opp.ID, 2); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatalf("a target was asked for: %+v", p)
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my own enchantment was destroyed")
	}
}

func TestOldB744DauthiJackalDestroysABlockingCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	jackal := effectstest.B12Push(g, me.ID, "Dauthi Jackal", "Creature — Dauthi Jackal", oldB744Jackal, 2, 1)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Wall", 0, 4)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 2)
	effectstest.AttackIntoBlocks(t, g, opp.ID, attacker)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInCombatDeclaration(t, g)
	effectstest.DcMana(t, g, me.ID, "{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, jackal, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(idle)}); err == nil {
		t.Fatal("a creature that is not blocking was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, jackal, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(blocker)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(jackal) {
		t.Error("the Jackal was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(blocker) {
		t.Error("the blocking creature survived")
	}
	if !g.Battlefield.Contains(idle) {
		t.Error("another creature was destroyed")
	}
}

func TestOldB744LimitedResourcesKeepsFiveLandsAndBansLandPlayAtTen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	var mine, theirs []uuid.UUID
	for i := 0; i < 7; i++ {
		mine = append(mine, effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest"))
	}
	for i := 0; i < 3; i++ {
		theirs = append(theirs, effectstest.SeedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island"))
	}
	var thirds []uuid.UUID
	for i := 0; i < 6; i++ {
		thirds = append(thirds, effectstest.SeedLandOnBattlefield(g, third.ID, "Swamp", "Basic Land — Swamp"))
	}
	effectstest.CastCatalogSpell(t, g, "Limited Resources", "Enchantment", oldB744Limited, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.ChooseCardsChoiceFor(g, opp.ID); p != nil {
		t.Fatal("a player with only three lands was asked")
	}
	if err := g.ResolveChooseCards(effectstest.ChooseCardsChoiceFor(g, me.ID).ID, me.ID, mine[:4]); err == nil {
		t.Fatal("keeping four lands was accepted")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, mine[2:]...)
	effectstest.AnswerChooseCards(t, g, third.ID, thirds[:5]...)
	effectstest.PassPriorityAroundTable(t, g)
	count := func(ids []uuid.UUID) int {
		n := 0
		for _, id := range ids {
			if g.Battlefield.Contains(id) {
				n++
			}
		}
		return n
	}
	if count(mine) != 5 || count(theirs) != 3 || count(thirds) != 5 {
		t.Errorf("lands left: %d / %d / %d, want 5 / 3 / 5", count(mine), count(theirs), count(thirds))
	}
	// 13 lands on the battlefield: nobody can play a land.
	land := effectstest.LpHandLand(me, "Forest")
	effectstest.LpAssertRefused(t, g, me, land, effectstest.LpPlay(t, g, me, land), "Players can't play lands as long as ten or more lands are on the battlefield.")
	if effectstest.LpGate(g, opp, "Forest", game.ZoneHand) == nil {
		t.Error("an opponent may play a land with ten or more on the battlefield")
	}
	// Drop under ten and the ban lifts.
	g.WithWriteLock(func() {
		for _, id := range append(append([]uuid.UUID{}, mine[2:5]...), thirds[:2]...) {
			_ = g.DestroyPermanentForEffect(id)
		}
	})
	if err := effectstest.LpGate(g, opp, "Forest", game.ZoneHand); err != nil {
		t.Errorf("with fewer than ten lands the ban still holds: %v", err)
	}
}

func TestOldB744SkyshroudWarBeastCountsTheChosenPlayersNonbasicLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, chosen, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.SeedLandOnBattlefield(g, chosen.ID, "Forest", "Basic Land — Forest")
	effectstest.SeedLandOnBattlefield(g, chosen.ID, "Volcanic Island", "Land — Island Mountain")
	effectstest.SeedLandOnBattlefield(g, chosen.ID, "Strip Mine", "Land")
	effectstest.SeedLandOnBattlefield(g, other.ID, "Tundra", "Land — Plains Island")
	effectstest.SeedLandOnBattlefield(g, me.ID, "Mine", "Land")
	beast := effectstest.CastCatalogSpell(t, g, "Skyshroud War Beast", "Creature — Beast", oldB744WarBeast, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChosenPlayer(t, g, me.ID, chosen.ID)
	if p, tough := effectstest.PtOf(t, g, beast); p != 2 || tough != 2 {
		t.Errorf("it is %d/%d, want 2/2 (the chosen player's two nonbasic lands)", p, tough)
	}
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Another", TypeLine: "Land", Owner: chosen.ID, Controller: chosen.ID})
	if p, tough := effectstest.PtOf(t, g, beast); p != 3 || tough != 3 {
		t.Errorf("it is %d/%d, want 3/3 after another nonbasic land", p, tough)
	}
}

func TestOldB744CrashingBoarsMakeTheDefenderChooseABlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	boars := effectstest.B12Push(g, me.ID, "Crashing Boars", "Creature — Boar", oldB744Boars, 4, 4)
	a := effectstest.B12Creature(g, opp.ID, "Blocker A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Blocker B", "Creature — Bear", 2, 2)
	tapped := effectstest.B12Creature(g, opp.ID, "Tapped", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, tapped).Tapped = true })
	effectstest.AdvanceToStepInTurn(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(boars, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, opp.ID)
	if prompt == nil {
		effectstest.PassPriorityAroundTable(t, g)
		prompt = effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, opp.ID)
	}
	if prompt == nil {
		t.Fatal("the defending player was not asked to choose a creature")
	}
	if err := g.ResolveOwnPermanents(prompt.ID, opp.ID, []uuid.UUID{tapped}); err == nil {
		t.Fatal("a tapped creature was accepted")
	}
	if err := g.ResolveOwnPermanents(prompt.ID, opp.ID, []uuid.UUID{a}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToStepInTurn(t, g, game.StepDeclareBlockers)
	// The chosen creature must block; declaring no block is refused.
	if err := g.FinishBlocks(opp.ID); err == nil {
		t.Fatal("the chosen creature was allowed not to block the Boars")
	}
	if err := g.DeclareBlocker(b, boars); err != nil {
		t.Fatalf("another creature blocking: %v", err)
	}
	if err := g.FinishBlocks(opp.ID); err == nil {
		t.Fatal("a different creature blocking satisfied the requirement")
	}
	if err := g.DeclareBlocker(a, boars); err != nil {
		t.Fatalf("the chosen creature blocking: %v", err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
}

func TestOldB744JackalopeHerdReturnsToHandWhenYouCastASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	herd := effectstest.B12Push(g, me.ID, "Jackalope Herd", "Creature — Rabbit Beast", oldB744Herd, 4, 5)
	// An opponent's spell does nothing.
	theirs := effectstest.HandCardOf(g, opp, "Shock", "Instant", "{R}", "")
	if err := g.CastSpell(opp.ID, theirs, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(herd) {
		t.Fatal("an opponent's spell bounced the Herd")
	}
	spell := effectstest.HandCardOf(g, me, "Shock", "Instant", "{R}", "")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(herd) || !me.Hand.Contains(herd) {
		t.Error("the Herd did not return to its owner's hand")
	}
}
