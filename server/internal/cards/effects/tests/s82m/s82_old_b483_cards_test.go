package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b483_cards_test.go — Rise of the Eldrazi, batch 483 (S82, #5581).
// Each test plays the real catalog entry through the engine.

const (
	oldB483BroodwardenOracle = "14537413-3628-4152-b23f-78b0e9c416ac"
	oldB483BramblesnapOracle = "312ce96b-6abe-4f72-a205-9b27c885a5fe"
	oldB483InvokerOracle     = "41b49a8e-09c5-4fd1-95f9-e5a51ce24f69"
	oldB483NullOracle        = "8bc17c50-f4c9-46dd-9294-3036ed46de91"
	oldB483ChannelersOracle  = "d3daca53-d40e-4537-a16a-f1e97333bc87"
	oldB483LightmineOracle   = "20490b29-4f08-48a2-bf54-2ba118e746ac"
	oldB483DemonOracle       = "90514aa5-84b6-4be7-b6d0-b05529c97140"
	oldB483LastKissOracle    = "bc59bf58-373d-4562-9a67-0721fcf2f098"
	oldB483BoulderfootOracle = "c16367eb-1c0a-4197-8e2f-2d74e4381024"
	oldB483LeafArrowOracle   = "d3d67bf0-2e9e-463f-8df8-990ab906fdec"
)

func TestOldB483EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB483BroodwardenOracle: "Broodwarden", oldB483BramblesnapOracle: "Bramblesnap",
		oldB483InvokerOracle: "Lavafume Invoker", oldB483NullOracle: "Escaped Null",
		oldB483ChannelersOracle: "Mul Daya Channelers", oldB483LightmineOracle: "Lightmine Field",
		oldB483DemonOracle: "Hellcarver Demon", oldB483LastKissOracle: "Last Kiss",
		oldB483BoulderfootOracle: "Akoum Boulderfoot", oldB483LeafArrowOracle: "Leaf Arrow",
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

func TestOldB483BroodwardenPumpsOnlyYourEldraziSpawn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warden := effectstest.B12Push(g, me.ID, "Broodwarden", "Creature — Eldrazi Drone", oldB483BroodwardenOracle, 4, 4)
	spawn := effectstest.B12Push(g, me.ID, "Eldrazi Spawn", "Creature — Eldrazi Spawn", "", 0, 1)
	drone := effectstest.B12Push(g, me.ID, "Other Drone", "Creature — Eldrazi Drone", "", 1, 1)
	theirs := effectstest.B12Push(g, opp.ID, "Eldrazi Spawn", "Creature — Eldrazi Spawn", "", 0, 1)
	for _, tc := range []struct {
		name         string
		id           uuid.UUID
		power, tough int
	}{{"your Spawn", spawn, 2, 2}, {"the Broodwarden", warden, 4, 4}, {"another Eldrazi", drone, 1, 1}, {"their Spawn", theirs, 0, 1}} {
		if p, tough := effectstest.PtOf(t, g, tc.id); p != tc.power || tough != tc.tough {
			t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tough, tc.power, tc.tough)
		}
	}
}

func TestOldB483BramblesnapGrowsForEachCreatureTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	snap := effectstest.B12Push(g, me.ID, "Bramblesnap", "Creature — Elemental", oldB483BramblesnapOracle, 1, 1)
	if !effectstest.HasEffectiveKeyword(t, g, snap, "trample") {
		t.Error("no trample")
	}
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, snap, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without naming a creature to tap")
	}
	effectstest.ApaActivate(t, g, me, snap, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, snap); p != 2 || tough != 2 {
		t.Errorf("Bramblesnap is %d/%d, want 2/2", p, tough)
	}
	if !effectstest.TappedForTest(t, g, bear) {
		t.Error("the Bear was not tapped")
	}
	// It may tap itself for the cost.
	effectstest.ApaActivate(t, g, me, snap, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{snap}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, snap); p != 3 || tough != 3 {
		t.Errorf("Bramblesnap is %d/%d after tapping itself, want 3/3", p, tough)
	}
}

func TestOldB483LavafumeInvokerPumpsYourCreaturesForEight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	inv := effectstest.B12Push(g, me.ID, "Lavafume Invoker", "Creature — Goblin Shaman", oldB483InvokerOracle, 2, 2)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.FloatForTest(g, me, "CCCCCCC")
	if err := g.ActivateCatalogAbility(me.ID, inv, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with only seven mana")
	}
	effectstest.FloatForTest(g, me, "C")
	effectstest.ApaActivate(t, g, me, inv, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, tc := range []struct {
		name  string
		id    uuid.UUID
		power int
	}{{"the Invoker", inv, 5}, {"your Bear", bear, 5}, {"their Bear", theirs, 2}} {
		if p, tough := effectstest.PtOf(t, g, tc.id); p != tc.power || tough != 2 {
			t.Errorf("%s is %d/%d, want %d/2", tc.name, p, tough, tc.power)
		}
	}
}

func TestOldB483EscapedNullGetsPlusFiveWhenBlockingOrBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	null := effectstest.B12Push(g, me.ID, "Escaped Null", "Creature — Zombie Berserker", oldB483NullOracle, 1, 2)
	if !effectstest.HasEffectiveKeyword(t, g, null, "lifelink") {
		t.Error("no lifelink")
	}
	blocker := effectstest.B12Push(g, opp.ID, "Wall", "Creature — Wall", "", 0, 9)
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, null, blocker)
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, null); p != 6 || tough != 2 {
		t.Errorf("Escaped Null is %d/%d after being blocked, want 6/2", p, tough)
	}
}

func TestOldB483EscapedNullGetsPlusFiveWhenItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	null := effectstest.B12Push(g, opp.ID, "Escaped Null", "Creature — Zombie Berserker", oldB483NullOracle, 1, 2)
	attacker := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, attacker, null)
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, null); p != 6 || tough != 2 {
		t.Errorf("Escaped Null is %d/%d after blocking, want 6/2", p, tough)
	}
}

func TestOldB483MulDayaChannelersReadsTheTopCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ch := effectstest.B12Push(g, me.ID, "Mul Daya Channelers", "Creature — Elf Druid Shaman", oldB483ChannelersOracle, 2, 2)
	push := func(name, typeLine string) {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Owner: me.ID, Controller: me.ID})
		g.WithWriteLock(func() {})
	}
	push("Spell", "Sorcery")
	if p, tough := effectstest.PtOf(t, g, ch); p != 2 || tough != 2 {
		t.Errorf("with a sorcery on top it is %d/%d, want 2/2", p, tough)
	}
	if err := g.ActivateManaAbility(me.ID, ch, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err == nil {
		t.Error("made mana with a sorcery on top")
	}
	push("Bear", "Creature — Bear")
	if p, tough := effectstest.PtOf(t, g, ch); p != 5 || tough != 5 {
		t.Errorf("with a creature on top it is %d/%d, want 5/5", p, tough)
	}
	if err := g.ActivateManaAbility(me.ID, ch, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err == nil {
		t.Error("made mana with a creature on top")
	}
	push("Forest", "Basic Land — Forest")
	if p, tough := effectstest.PtOf(t, g, ch); p != 2 || tough != 2 {
		t.Errorf("with a land on top it is %d/%d, want 2/2", p, tough)
	}
	if err := g.ActivateManaAbility(me.ID, ch, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("with a land on top: %v", err)
	}
	if n := me.ManaPool.CountColor("U"); n != 2 || me.ManaPool.Total() != 2 {
		t.Errorf("pool holds %d {U} of %d, want 2 of 2", n, me.ManaPool.Total())
	}
}

func TestOldB483LightmineFieldDamagesEachAttackerByTheNumberAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, opp.ID, "Lightmine Field", "Enchantment", oldB483LightmineOracle, 0, 0)
	a := effectstest.B12Push(g, me.ID, "Bear A", "Creature — Bear", "", 2, 2)
	b := effectstest.B12Push(g, me.ID, "Bear B", "Creature — Bear", "", 2, 2)
	big := effectstest.B12Push(g, me.ID, "Ogre", "Creature — Ogre", "", 4, 4)
	home := effectstest.B12Push(g, me.ID, "Stay Home", "Creature — Bear", "", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a, b, big)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	for _, id := range []uuid.UUID{a, b} {
		if oldB483OnBattlefield(g, id) {
			t.Errorf("a 2/2 attacker survived three damage")
		}
	}
	if !oldB483OnBattlefield(g, big) {
		t.Error("the 4/4 attacker died to three damage")
	}
	if !oldB483OnBattlefield(g, home) {
		t.Error("a creature that did not attack was hurt")
	}
}

func TestOldB483LastKissDamagesAndGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	victim := effectstest.B12Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	life := me.Life
	effectstest.CastFromHandForTest(t, g, me, "Last Kiss", "Instant", "{2}{B}", oldB483LastKissOracle,
		game.CastSpellParams{Targets: effectstest.CardTarget(victim)})
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if oldB483OnBattlefield(g, victim) {
		t.Error("the 2/2 survived two damage")
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestOldB483LeafArrowOnlyTargetsCreaturesWithFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bird := effectstest.B30Flyer(g, opp.ID, "Bird", "Creature — Bird", 2, 3)
	ground := effectstest.B12Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.FloatForTest(g, me, "G")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Leaf Arrow", TypeLine: "Instant", ManaCost: "{G}", OracleID: oldB483LeafArrowOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(ground)}); err == nil {
		t.Fatal("targeted a creature without flying")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(bird)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if oldB483OnBattlefield(g, bird) {
		t.Error("the 2/3 flier survived three damage")
	}
}

func TestOldB483AkoumBoulderfootPingsAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	life := opp.Life
	effectstest.PlsB1Cast(t, g, "Akoum Boulderfoot", "Creature — Giant Warrior", oldB483BoulderfootOracle, "", 4, 5)
	pick := effectstest.PassUntilPickTarget(t, g, me.ID)
	if pick == nil {
		t.Fatal("no target was asked for")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, effectstest.PlayerTarget(opp.ID)); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d → %d, want 1 damage", life, opp.Life)
	}
}

// oldB483OnBattlefield is whether the card is on the battlefield.
func oldB483OnBattlefield(g *game.Game, id uuid.UUID) bool {
	return effectstest.BattlefieldCardFor(g, id) != nil
}

func TestOldB483HellcarverDemonSacrificesDiscardsExilesAndCastsForFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	demon := effectstest.B12Push(g, me.ID, "Hellcarver Demon", "Creature — Demon", oldB483DemonOracle, 6, 6)
	if !effectstest.HasEffectiveKeyword(t, g, demon, "flying") {
		t.Error("no flying")
	}
	other1 := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	other2 := effectstest.PushCatalogPermanent(g, me.ID, "Plains", "Basic Land — Plains", "", false)
	theirs := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	for me.Hand.Size() > 0 {
		_, _ = me.Hand.PopTop()
	}
	hand := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: hand, Name: "Hand Card", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	for me.Library.Size() > 0 {
		_, _ = me.Library.PopTop()
	}
	var exiled []uuid.UUID
	untouched := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: untouched, Name: "Seventh", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	for i := 0; i < 5; i++ {
		id := uuid.New()
		exiled = append(exiled, id)
		me.Library.PushTop(game.Card{InstanceID: id, Name: "Spell", TypeLine: "Sorcery", ManaCost: "{2}{R}", Owner: me.ID, Controller: me.ID})
	}
	land := uuid.New()
	exiled = append(exiled, land)
	me.Library.PushTop(game.Card{InstanceID: land, Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	effectstest.DealCombatDamageToPlayer(g, demon, opp.ID, 6)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{other1, other2} {
		if oldB483OnBattlefield(g, id) {
			t.Errorf("another permanent of yours was not sacrificed")
		}
	}
	if !oldB483OnBattlefield(g, demon) {
		t.Error("the Demon itself was sacrificed")
	}
	if !oldB483OnBattlefield(g, theirs) {
		t.Error("an opponent's permanent was sacrificed")
	}
	if me.Hand.Contains(hand) {
		t.Error("the hand was not discarded")
	}
	if me.Library.Size() != 1 || !me.Library.Contains(untouched) {
		t.Errorf("library %d cards, want only the seventh", me.Library.Size())
	}
	offer := effectstest.ResolutionCastOfferFor(g, me.ID)
	if offer == nil {
		t.Fatal("no free-cast offer for the exiled spells")
	}
	for _, id := range exiled[:5] {
		found := false
		for _, o := range offer.DiscardOptions {
			found = found || o == id
		}
		if !found {
			t.Errorf("an exiled spell is not offered")
		}
	}
	cast := exiled[0]
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, cast, game.CastSpellParams{}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		effectstest.DeclineResolutionCast(t, g, me.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range exiled[1:] {
		if z := effectstest.ZoneOf(g, id); z != game.ZoneExile {
			t.Errorf("an uncast card is in %q, want exile", z)
		}
	}
}
