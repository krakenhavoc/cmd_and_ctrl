package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_neo_b02_cards_test.go — Kamigawa: Neon Dynasty, S82 batch 2 (#5493):
// simple instants and sorceries. Each test casts the real catalog entry and
// asserts what its text says.

const (
	neoB02Boon      = "44e1946d-01b5-45db-abdc-2f6d6480cf60"
	neoB02Lucky     = "cbb39e5c-e31a-44e9-be0c-2003c8bce599"
	neoB02Winds     = "721c3255-953c-49cd-a5dd-852c3d09eae4"
	neoB02Light     = "83917116-d2d8-4701-a79f-af17486706ae"
	neoB02Wanderer  = "4497d0b9-cf4b-4c9d-bf7e-8ed85aa6b5b8"
	neoB02Dead      = "c1157d79-38ea-4a88-b346-7b12614ff78a"
	neoB02Regent    = "54402930-84ea-4c3e-9880-ca631e47485f"
	neoB02Explosive = "57544b37-fd5d-4da5-bbb6-7b4a09501dd1"
	neoB02Season    = "410ce48e-d907-4bc5-ad90-2067161a0e04"
	neoB02Incision  = "a959a91a-3e1d-4bd4-9e14-5ee977540146"
	neoB02Repel     = "aa108554-f1a0-45b6-aa15-597d42be7a8a"
	neoB02Debt      = "c6c1e576-7197-4067-8cd1-aa535188e1b8"
	neoB02Pursuit   = "23b7fd18-3129-448f-881f-947ba988f98d"
)

func TestS82NeoB02EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		neoB02Boon: "Boon of Boseiju", neoB02Lucky: "Lucky Offering", neoB02Winds: "Invoke the Winds",
		neoB02Light: "Light the Way", neoB02Wanderer: "Wanderer's Intervention",
		neoB02Dead: "You Are Already Dead", neoB02Regent: "Regent's Authority",
		neoB02Explosive: "Explosive Entry", neoB02Season: "Season of Renewal",
		neoB02Incision: "Planar Incision", neoB02Repel: "Repel the Vile",
		neoB02Debt: "Debt to the Kami", neoB02Pursuit: "Kaito's Pursuit",
	}
	for oracle, name := range want {
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

func neoB02Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	return g, me, opp
}

// neoB02Cast puts the card into the caster's hand and casts it. Strict
// refuses what the table would.
func neoB02Cast(g *game.Game, me *game.Player, name, typeLine, oracle string, modes []int, strict bool, targets ...game.TargetRef) error {
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: me.ID, Controller: me.ID})
	return g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets, Modes: modes, Strict: strict})
}

func neoB02Card(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

func neoB02Perm(g *game.Game, owner uuid.UUID, name, typeLine, cost string, p, tough int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		ManaCost: cost, Power: p, Toughness: tough, Owner: owner, Controller: owner})
}

func neoB02Counters(g *game.Game, id uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		if c, ok := effectstest.BattlefieldCard(g, id); ok {
			n = c.Counters["+1/+1"]
		}
	})
	return n
}

func TestS82NeoB02BoonOfBoseijuPumpsByTheGreatestManaValueAndUntaps(t *testing.T) {
	g, me, _ := neoB02Table(t)
	bear := neoB02Perm(g, me.ID, "Bear", "Creature — Bear", "{1}{G}", 2, 2)
	neoB02Perm(g, me.ID, "Big Rock", "Artifact", "{5}", 0, 0)
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, bear).Tapped = true })
	if err := neoB02Cast(g, me, "Boon of Boseiju", "Instant", neoB02Boon, nil, true, neoB02Card(bear)); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, bear); p != 7 || tt != 7 {
		t.Errorf("Bear is %d/%d, want 7/7 (+5/+5)", p, tt)
	}
	if c, _ := effectstest.BattlefieldCard(g, bear); c.Tapped {
		t.Error("Bear stayed tapped")
	}
}

func TestS82NeoB02BoonOfBoseijuRefusesANonCreature(t *testing.T) {
	g, me, _ := neoB02Table(t)
	rock := neoB02Perm(g, me.ID, "Rock", "Artifact", "{1}", 0, 0)
	if err := neoB02Cast(g, me, "Boon of Boseiju", "Instant", neoB02Boon, nil, true, neoB02Card(rock)); err == nil {
		t.Error("a noncreature target was accepted")
	}
}

func TestS82NeoB02LuckyOfferingDestroysASmallArtifactAndGainsThree(t *testing.T) {
	g, me, opp := neoB02Table(t)
	small := neoB02Perm(g, opp.ID, "Trinket", "Artifact", "{3}", 0, 0)
	big := neoB02Perm(g, opp.ID, "Engine", "Artifact", "{4}", 0, 0)
	if err := neoB02Cast(g, me, "Lucky Offering", "Sorcery", neoB02Lucky, nil, true, neoB02Card(big)); err == nil {
		t.Fatal("a mana value 4 artifact was accepted")
	}
	life := me.Life
	if err := neoB02Cast(g, me, "Lucky Offering", "Sorcery", neoB02Lucky, nil, true, neoB02Card(small)); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, small); ok {
		t.Error("the artifact survived")
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestS82NeoB02InvokeTheWindsStealsAndUntapsForGood(t *testing.T) {
	g, me, opp := neoB02Table(t)
	bear := neoB02Perm(g, opp.ID, "Their Bear", "Creature — Bear", "{1}{G}", 2, 2)
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, bear).Tapped = true })
	if err := neoB02Cast(g, me, "Invoke the Winds", "Sorcery", neoB02Winds, nil, true, neoB02Card(bear)); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, bear)
	if !ok || c.Controller != me.ID || c.Tapped {
		t.Fatalf("not an untapped permanent of mine: %+v", c)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, bear); c.Controller != me.ID {
		t.Error("control reverted at end of turn")
	}
}

func TestS82NeoB02InvokeTheWindsTakesAnArtifactAndRefusesALand(t *testing.T) {
	g, me, opp := neoB02Table(t)
	rock := neoB02Perm(g, opp.ID, "Rock", "Artifact", "{2}", 0, 0)
	land := neoB02Perm(g, opp.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	if err := neoB02Cast(g, me, "Invoke the Winds", "Sorcery", neoB02Winds, nil, true, neoB02Card(land)); err == nil {
		t.Error("a land was accepted")
	}
	if err := neoB02Cast(g, me, "Invoke the Winds", "Sorcery", neoB02Winds, nil, true, neoB02Card(rock)); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, rock); c.Controller != me.ID {
		t.Error("the artifact was not taken")
	}
}

func TestS82NeoB02LightTheWayCountersAndUntapsACreatureOrVehicle(t *testing.T) {
	g, me, _ := neoB02Table(t)
	bear := neoB02Perm(g, me.ID, "Bear", "Creature — Bear", "{1}{G}", 2, 2)
	car := neoB02Perm(g, me.ID, "Car", "Artifact — Vehicle", "{2}", 3, 3)
	land := neoB02Perm(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, car).Tapped = true })
	if err := neoB02Cast(g, me, "Light the Way", "Instant", neoB02Light, []int{0}, true, neoB02Card(land)); err == nil {
		t.Error("a land was accepted for the counter mode")
	}
	for _, id := range []uuid.UUID{bear, car} {
		if err := neoB02Cast(g, me, "Light the Way", "Instant", neoB02Light, []int{0}, true, neoB02Card(id)); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if neoB02Counters(g, id) != 1 {
			t.Errorf("%s has no +1/+1 counter", id)
		}
	}
	if c, _ := effectstest.BattlefieldCard(g, car); c.Tapped {
		t.Error("the Vehicle stayed tapped")
	}
}

func TestS82NeoB02LightTheWayReturnsYourPermanentToHand(t *testing.T) {
	g, me, opp := neoB02Table(t)
	mine := neoB02Perm(g, me.ID, "Bear", "Creature — Bear", "{1}{G}", 2, 2)
	theirs := neoB02Perm(g, opp.ID, "Their Bear", "Creature — Bear", "{1}{G}", 2, 2)
	if err := neoB02Cast(g, me, "Light the Way", "Instant", neoB02Light, []int{1}, true, neoB02Card(theirs)); err == nil {
		t.Error("an opponent's permanent was accepted")
	}
	if err := neoB02Cast(g, me, "Light the Way", "Instant", neoB02Light, []int{1}, true, neoB02Card(mine)); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, mine); ok || !me.Hand.Contains(mine) {
		t.Error("the permanent did not return to its owner's hand")
	}
}

func TestS82NeoB02WanderersInterventionHitsOnlyAttackersAndBlockers(t *testing.T) {
	g, me, opp := neoB02Table(t)
	atk := neoB02Perm(g, me.ID, "Attacker", "Creature — Bear", "{2}", 2, 4)
	idle := neoB02Perm(g, me.ID, "Idle", "Creature — Bear", "{2}", 2, 4)
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, atk).SummonedThisTurn = false })
	effectstest.AttackWith(t, g, opp.ID, atk)
	if err := neoB02Cast(g, me, "Wanderer's Intervention", "Instant", neoB02Wanderer, nil, true, neoB02Card(idle)); err == nil {
		t.Error("a creature that is neither attacking nor blocking was accepted")
	}
	if err := neoB02Cast(g, me, "Wanderer's Intervention", "Instant", neoB02Wanderer, nil, true, neoB02Card(atk)); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, atk); ok {
		t.Error("4 damage did not kill the 2/4 attacker")
	}
}

func TestS82NeoB02YouAreAlreadyDeadNeedsDamageAndDraws(t *testing.T) {
	g, me, opp := neoB02Table(t)
	hurt := neoB02Perm(g, opp.ID, "Hurt Bear", "Creature — Bear", "{2}", 2, 5)
	fresh := neoB02Perm(g, opp.ID, "Fresh Bear", "Creature — Bear", "{2}", 2, 5)
	effectstest.DdCombatDamage(g, me.ID, uuid.New(), hurt, 1)
	if err := neoB02Cast(g, me, "You Are Already Dead", "Instant", neoB02Dead, nil, true, neoB02Card(fresh)); err == nil {
		t.Error("an undamaged creature was accepted")
	}
	hand := me.Hand.Size()
	if err := neoB02Cast(g, me, "You Are Already Dead", "Instant", neoB02Dead, nil, true, neoB02Card(hurt)); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, hurt); ok {
		t.Error("the damaged creature survived")
	}
	// The spell enters and leaves the hand when cast; the draw is +1.
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, expected one card drawn", hand, me.Hand.Size())
	}
}

func TestS82NeoB02RegentsAuthorityPumpsOrCountersLegendsAndEnchantmentCreatures(t *testing.T) {
	g, me, _ := neoB02Table(t)
	plain := neoB02Perm(g, me.ID, "Bear", "Creature — Bear", "{2}", 2, 2)
	legend := neoB02Perm(g, me.ID, "Legend", "Legendary Creature — Human", "{2}", 2, 2)
	ench := neoB02Perm(g, me.ID, "Spirit", "Enchantment Creature — Spirit", "{2}", 2, 2)
	for id, wantCounter := range map[uuid.UUID]int{plain: 0, legend: 1, ench: 1} {
		if err := neoB02Cast(g, me, "Regent's Authority", "Instant", neoB02Regent, nil, true, neoB02Card(id)); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if n := neoB02Counters(g, id); n != wantCounter {
			t.Errorf("%s: %d counters, want %d", id, n, wantCounter)
		}
		if p, tt := effectstest.PtOf(t, g, id); p != 4 || tt != 4 {
			t.Errorf("%s is %d/%d, want 4/4", id, p, tt)
		}
	}
	// Until end of turn: the +1/+1 wears off, the counter stays.
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	if p, _ := effectstest.PtOf(t, g, legend); p != 3 {
		t.Errorf("the legend is %d/x after cleanup, want 3 (counter stays)", p)
	}
	if p, _ := effectstest.PtOf(t, g, plain); p != 2 {
		t.Errorf("the plain creature is %d/x after cleanup, want 2", p)
	}
}

func TestS82NeoB02ExplosiveEntryTakesEitherHalfOrBoth(t *testing.T) {
	g, me, opp := neoB02Table(t)
	rock := neoB02Perm(g, opp.ID, "Rock", "Artifact", "{2}", 0, 0)
	bear := neoB02Perm(g, me.ID, "Bear", "Creature — Bear", "{2}", 2, 2)
	// Both clauses.
	if err := neoB02Cast(g, me, "Explosive Entry", "Sorcery", neoB02Explosive, nil, true,
		neoB02Card(rock), neoB02Card(bear)); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, rock); ok {
		t.Error("the artifact survived")
	}
	if neoB02Counters(g, bear) != 1 {
		t.Error("no counter on the creature")
	}
	// Only the creature.
	if err := neoB02Cast(g, me, "Explosive Entry", "Sorcery", neoB02Explosive, nil, true, neoB02Card(bear)); err != nil {
		t.Fatalf("a cast naming only the creature was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if neoB02Counters(g, bear) != 2 {
		t.Error("the creature-only cast did not add a counter")
	}
	// No targets at all.
	if err := neoB02Cast(g, me, "Explosive Entry", "Sorcery", neoB02Explosive, nil, true); err != nil {
		t.Fatalf("a cast with no targets was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestS82NeoB02SeasonOfRenewalReturnsEitherOrBoth(t *testing.T) {
	g, me, _ := neoB02Table(t)
	creature := uuid.New()
	ench := uuid.New()
	land := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: creature, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: ench, Name: "Old Aura", TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	if err := neoB02Cast(g, me, "Season of Renewal", "Instant", neoB02Season, []int{0}, true, neoB02Card(land)); err == nil {
		t.Error("a land card was accepted for the creature mode")
	}
	if err := neoB02Cast(g, me, "Season of Renewal", "Instant", neoB02Season, []int{0, 1}, true,
		neoB02Card(creature), neoB02Card(ench)); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(creature) || !me.Hand.Contains(ench) {
		t.Error("both cards did not return to hand")
	}
}

func TestS82NeoB02SeasonOfRenewalOneModeAlone(t *testing.T) {
	g, me, _ := neoB02Table(t)
	creature := uuid.New()
	ench := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: creature, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: ench, Name: "Old Aura", TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID})
	if err := neoB02Cast(g, me, "Season of Renewal", "Instant", neoB02Season, []int{1}, true, neoB02Card(ench)); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ench) || me.Hand.Contains(creature) {
		t.Error("only the enchantment should have returned")
	}
}

func TestS82NeoB02PlanarIncisionReturnsItWithACounterUnderItsOwner(t *testing.T) {
	g, me, opp := neoB02Table(t)
	bear := neoB02Perm(g, opp.ID, "Their Bear", "Creature — Bear", "{2}", 2, 2)
	rock := neoB02Perm(g, opp.ID, "Rock", "Artifact", "{2}", 0, 0)
	land := neoB02Perm(g, opp.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	if err := neoB02Cast(g, me, "Planar Incision", "Instant", neoB02Incision, nil, true, neoB02Card(land)); err == nil {
		t.Error("a land was accepted")
	}
	if err := neoB02Cast(g, me, "Planar Incision", "Instant", neoB02Incision, nil, true, neoB02Card(bear)); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var back *game.Card
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			c := g.Battlefield.Cards[i]
			if c.Name == "Their Bear" {
				back = &c
			}
		}
	})
	if back == nil || back.Controller != opp.ID || back.Counters["+1/+1"] != 1 {
		t.Fatalf("the creature did not return under its owner with a counter: %+v", back)
	}
	if back.InstanceID == bear {
		t.Error("the creature came back as the same object (CR 400.7)")
	}
	if err := neoB02Cast(g, me, "Planar Incision", "Instant", neoB02Incision, nil, true, neoB02Card(rock)); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestS82NeoB02RepelTheVileExilesABigCreatureOrAnEnchantment(t *testing.T) {
	g, me, opp := neoB02Table(t)
	small := neoB02Perm(g, opp.ID, "Small", "Creature — Bear", "{2}", 3, 3)
	big := neoB02Perm(g, opp.ID, "Big", "Creature — Beast", "{5}", 4, 4)
	aura := neoB02Perm(g, opp.ID, "Aura", "Enchantment", "{2}", 0, 0)
	if err := neoB02Cast(g, me, "Repel the Vile", "Instant", neoB02Repel, []int{0}, true, neoB02Card(small)); err == nil {
		t.Error("a power 3 creature was accepted")
	}
	if err := neoB02Cast(g, me, "Repel the Vile", "Instant", neoB02Repel, []int{0}, true, neoB02Card(big)); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, big); ok || !zoneHasExiled(g, big) {
		t.Error("the big creature was not exiled")
	}
	if err := neoB02Cast(g, me, "Repel the Vile", "Instant", neoB02Repel, []int{1}, true, neoB02Card(aura)); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, aura); ok {
		t.Error("the enchantment was not exiled")
	}
}

func zoneHasExiled(g *game.Game, id uuid.UUID) bool {
	z := g.FindCardZoneForEffect(id)
	return z != nil && z.Kind == game.ZoneExile
}

func TestS82NeoB02DebtToTheKamiLetsTheOpponentChooseWhatToExile(t *testing.T) {
	g, me, opp := neoB02Table(t)
	a := neoB02Perm(g, opp.ID, "Bear A", "Creature — Bear", "{2}", 2, 2)
	b := neoB02Perm(g, opp.ID, "Bear B", "Creature — Bear", "{2}", 2, 2)
	ench := neoB02Perm(g, opp.ID, "Aura", "Enchantment", "{2}", 0, 0)
	if err := neoB02Cast(g, me, "Debt to the Kami", "Instant", neoB02Debt, []int{0}, true,
		game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, b)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, b); ok || !zoneHasExiled(g, b) {
		t.Error("the chosen creature was not exiled")
	}
	if _, ok := effectstest.BattlefieldCard(g, a); !ok {
		t.Error("the other creature was exiled too")
	}
	if _, ok := effectstest.BattlefieldCard(g, ench); !ok {
		t.Error("the creature mode exiled an enchantment")
	}
	if err := neoB02Cast(g, me, "Debt to the Kami", "Instant", neoB02Debt, []int{1}, true,
		game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(chooseCardsChoiceForSafe(g, opp.ID)) > 0 {
		effectstest.AnswerChooseCards(t, g, opp.ID, ench)
	}
	if _, ok := effectstest.BattlefieldCard(g, ench); ok {
		t.Error("the enchantment was not exiled")
	}
}

func chooseCardsChoiceForSafe(g *game.Game, who uuid.UUID) []uuid.UUID {
	if c := effectstest.ChooseCardsChoiceFor(g, who); c != nil {
		return c.ChooseCards
	}
	return nil
}

func TestS82NeoB02DebtToTheKamiRefusesYourOwnPlayerAsTarget(t *testing.T) {
	g, me, _ := neoB02Table(t)
	if err := neoB02Cast(g, me, "Debt to the Kami", "Instant", neoB02Debt, []int{0}, true,
		game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
		t.Error("the caster was accepted as the target opponent")
	}
}

func TestS82NeoB02KaitosPursuitDiscardsTwoAndGivesNinjasAndRoguesMenace(t *testing.T) {
	g, me, opp := neoB02Table(t)
	ninja := neoB02Perm(g, me.ID, "Ninja", "Creature — Human Ninja", "{2}", 2, 2)
	rogue := neoB02Perm(g, me.ID, "Rogue", "Creature — Human Rogue", "{2}", 2, 2)
	bear := neoB02Perm(g, me.ID, "Bear", "Creature — Bear", "{2}", 2, 2)
	theirs := neoB02Perm(g, opp.ID, "Their Ninja", "Creature — Human Ninja", "{2}", 2, 2)
	effectstest.PushHandCard(g, opp)
	effectstest.PushHandCard(g, opp)
	effectstest.PushHandCard(g, opp)
	hand := opp.Hand.Size()
	if err := neoB02Cast(g, me, "Kaito's Pursuit", "Sorcery", neoB02Pursuit, nil, true,
		game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, opp.ID, opp.Hand.Cards[0].InstanceID, opp.Hand.Cards[1].InstanceID)
	if got := hand - opp.Hand.Size(); got != 2 {
		t.Errorf("discarded %d, want 2", got)
	}
	for id, want := range map[uuid.UUID]bool{ninja: true, rogue: true, bear: false, theirs: false} {
		if got := effectstest.HasEffectiveKeyword(t, g, id, "menace"); got != want {
			t.Errorf("%s menace = %v, want %v", id, got, want)
		}
	}
}
