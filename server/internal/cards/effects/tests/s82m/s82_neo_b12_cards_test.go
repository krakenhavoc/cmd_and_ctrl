package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_neo_b12_cards_test.go — Neon Dynasty batch 12 (#5493): the Auras.

const (
	neoAkkiWarPaint        = "e01ab6d8-8419-436e-a8fb-10aa46ca3d05"
	neoGiftOfWrath         = "9ba31dd0-d78e-4639-a550-bd72cac5898a"
	neoAwakenedAwareness   = "6629e5f3-ac59-48f8-b09c-3b98f59b9922"
	neoClawingTorment      = "0ac5c149-f6af-46b6-974e-14689e9a583d"
	neoShortCircuit        = "5ea0e3a5-9c38-4da4-98a6-41c14be512d9"
	neoHistoriansWisdom    = "7c2c460d-8059-41ff-b2d7-374aa498f011"
	neoTamiyosCompleation  = "71d8b7f5-a5da-4528-815e-51c65626b681"
	neoTwistedEmbrace      = "2e29efee-1eed-4bef-8574-0f6ff99d3cda"
	neoGraftedGrowth       = "c4525b1c-5c3d-494e-b660-ba5a5a28f621"
	neoUnquenchableFury    = "1c69a41e-122a-42c1-98d4-56035d68cfc8"
	neoCracklingEmergence  = "c984787c-f883-42c9-ae5d-9277c84dba01"
	neoHarmoniousEmergence = "ad8fd4e9-f6ae-4f24-b806-4509c8f4ee47"
)

func TestS82NeoB12EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		neoAkkiWarPaint: "Akki War Paint", neoGiftOfWrath: "Gift of Wrath",
		neoAwakenedAwareness: "Awakened Awareness", neoClawingTorment: "Clawing Torment",
		neoShortCircuit: "Short Circuit", neoHistoriansWisdom: "Historian's Wisdom",
		neoTamiyosCompleation: "Tamiyo's Compleation", neoTwistedEmbrace: "Twisted Embrace",
		neoGraftedGrowth: "Grafted Growth", neoUnquenchableFury: "Unquenchable Fury",
		neoCracklingEmergence: "Crackling Emergence", neoHarmoniousEmergence: "Harmonious Emergence",
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

func neoB12Artifact(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: owner, Controller: owner})
}

func neoB12Land(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: owner, Controller: owner})
}

func neoB12Flier(g *game.Game, owner uuid.UUID, p, tough int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Flier", TypeLine: "Creature — Bird",
		Power: p, Toughness: tough, Keywords: []string{"flying"}, Owner: owner, Controller: owner})
}

func neoB12CastStrict(t *testing.T, g *game.Game, name, oracle string, target uuid.UUID) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: effectstest.AuraTypeLine, OracleID: oracle, Owner: me.ID, Controller: me.ID})
	})
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(target)})
}

func TestS82NeoB12AkkiWarPaintPumpsACreatureAndWaitsOnANoncreatureArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	rock := neoB12Artifact(g, me.ID)
	effectstest.Cmdr2B19Attach(g, me.ID, host, "Akki War Paint", neoAkkiWarPaint)
	effectstest.Cmdr2B19Attach(g, me.ID, rock, "Akki War Paint", neoAkkiWarPaint)
	if p, tt := effectstest.PtOf(t, g, host); p != 4 || tt != 3 {
		t.Errorf("enchanted creature is %d/%d, want 4/3", p, tt)
	}
	if c, _ := effectstest.BattlefieldCard(g, rock); c.Power != 0 || c.Toughness != 0 {
		t.Errorf("noncreature artifact got a pump: %d/%d", c.Power, c.Toughness)
	}
}

func TestS82NeoB12EnchantClausesRefuseTheWrongPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := neoB12Land(g, me.ID)
	for name, oracle := range map[string]string{"Akki War Paint": neoAkkiWarPaint, "Short Circuit": neoShortCircuit, "Tamiyo's Compleation": neoTamiyosCompleation} {
		if err := neoB12CastStrict(t, g, name, oracle, land); err == nil {
			t.Errorf("%s may enchant a land", name)
		}
	}
	// Twisted Embrace enchants only what you control; Emergences only your lands.
	oppCreature := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	if err := neoB12CastStrict(t, g, "Twisted Embrace", neoTwistedEmbrace, oppCreature); err == nil {
		t.Error("Twisted Embrace enchanted an opponent's creature")
	}
	oppLand := neoB12Land(g, g.Seats[1].ID)
	if err := neoB12CastStrict(t, g, "Crackling Emergence", neoCracklingEmergence, oppLand); err == nil {
		t.Error("Crackling Emergence enchanted an opponent's land")
	}
	creature := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	if err := neoB12CastStrict(t, g, "Grafted Growth", neoGraftedGrowth, creature); err == nil {
		t.Error("Grafted Growth enchanted a creature")
	}
}

func TestS82NeoB12ShortCircuitShrinksAndGroundsACreatureButNotAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := neoB12Flier(g, me.ID, 4, 4)
	rock := neoB12Artifact(g, me.ID)
	effectstest.Cmdr2B19Attach(g, me.ID, host, "Short Circuit", neoShortCircuit)
	effectstest.Cmdr2B19Attach(g, me.ID, rock, "Short Circuit", neoShortCircuit)
	if p, tt := effectstest.PtOf(t, g, host); p != 1 || tt != 4 {
		t.Errorf("enchanted creature is %d/%d, want 1/4", p, tt)
	}
	if effectstest.HasEffectiveKeyword(t, g, host, "flying") {
		t.Error("the enchanted creature still has flying")
	}
	if c, _ := effectstest.BattlefieldCard(g, rock); c.Power != 0 {
		t.Errorf("noncreature artifact got -3/-0: power %d", c.Power)
	}
}

func TestS82NeoB12ShortCircuitHasFlash(t *testing.T) {
	spec, _ := effects.Lookup(neoShortCircuit)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("printed keywords %v, want flash", spec.PrintedKeywords)
	}
	spec, _ = effects.Lookup(neoTamiyosCompleation)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("Tamiyo's Compleation printed keywords %v, want flash", spec.PrintedKeywords)
	}
}

func TestS82NeoB12GiftOfWrathPumpsAndLeavesASpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	aura := effectstest.Cmdr2B19Attach(g, me.ID, host, "Gift of Wrath", neoGiftOfWrath)
	if p, tt := effectstest.PtOf(t, g, host); p != 4 || tt != 4 {
		t.Errorf("enchanted creature is %d/%d, want 4/4", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, host, "menace") {
		t.Error("the enchanted creature lacks menace")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.NamedOnBattlefield(g, "Spirit", me.ID); n != 1 {
		t.Fatalf("%d Spirit tokens after the Aura left, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" {
			if c.Power != 2 || c.Toughness != 2 || !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "menace") || len(c.Colors) != 1 || c.Colors[0] != "R" {
				t.Errorf("Spirit token is %d/%d colors %v, want a 2/2 red creature with menace", c.Power, c.Toughness, c.Colors)
			}
		}
	}
	if p, tt := effectstest.PtOf(t, g, host); p != 2 || tt != 2 {
		t.Errorf("host after the Aura left is %d/%d, want 2/2", p, tt)
	}
}

func TestS82NeoB12GiftOfWrathTokenWhenTheHostDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	effectstest.Cmdr2B19Attach(g, me.ID, host, "Gift of Wrath", neoGiftOfWrath)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(host) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.NamedOnBattlefield(g, "Spirit", me.ID); n != 1 {
		t.Errorf("%d Spirit tokens after the enchanted creature died, want 1", n)
	}
}

func TestS82NeoB12HistoriansWisdomDrawsOnlyForTheGreatestPower(t *testing.T) {
	for _, tc := range []struct {
		name      string
		hostPower int
		wantDraw  int
	}{
		{"host is greatest after the pump", 3, 1},
		{"host ties the other creature after the pump", 4, 1},
		{"host is not greatest", 1, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.SeedLibrary(me, "A", "B", "C")
			host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", tc.hostPower, 2)
			effectstest.B12Creature(g, g.Seats[1].ID, "Big", "Creature — Giant", 4, 4)
			effectstest.AdvanceToMain(t, g)
			before := me.Hand.Size()
			effectstest.CastCatalogSpell(t, g, "Historian's Wisdom", effectstest.AuraTypeLine, neoHistoriansWisdom, effectstest.CardRefs(host))
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Hand.Size() - before; got != tc.wantDraw {
				t.Errorf("hand grew by %d, want %d", got, tc.wantDraw)
			}
			if p, tt := effectstest.PtOf(t, g, host); p != tc.hostPower+2 || tt != 3 {
				t.Errorf("host is %d/%d, want %d/3", p, tt, tc.hostPower+2)
			}
		})
	}
}

func TestS82NeoB12HistoriansWisdomOnAnArtifactDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "A", "B", "C")
	rock := neoB12Artifact(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	before := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Historian's Wisdom", effectstest.AuraTypeLine, neoHistoriansWisdom, effectstest.CardRefs(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 0 {
		t.Errorf("hand grew by %d for a noncreature artifact", got)
	}
}

func TestS82NeoB12AwakenedAwarenessPutsXCountersAndSetsBasePT(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 5, 5)
	effectstest.ToMainForCost(t, g)
	effectstest.FloatForTest(g, me, "UUUUUU")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Awakened Awareness", TypeLine: effectstest.AuraTypeLine, ManaCost: "{X}{U}{U}",
			OracleID: neoAwakenedAwareness, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 3, Targets: effectstest.CardRefs(host)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, host)
	if got := c.Counters[game.CounterPlusOne]; got != 3 {
		t.Fatalf("%d +1/+1 counters, want 3", got)
	}
	if p, tt := effectstest.PtOf(t, g, host); p != 4 || tt != 4 {
		t.Errorf("enchanted creature is %d/%d, want base 1/1 plus three counters (4/4)", p, tt)
	}
}

func TestS82NeoB12AwakenedAwarenessLeavesANoncreatureArtifactAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := neoB12Artifact(g, me.ID)
	effectstest.Cmdr2B19Attach(g, me.ID, rock, "Awakened Awareness", neoAwakenedAwareness)
	if c, _ := effectstest.BattlefieldCard(g, rock); c.Power != 0 || c.Toughness != 0 {
		t.Errorf("noncreature artifact has a base P/T: %d/%d", c.Power, c.Toughness)
	}
}

func TestS82NeoB12ClawingTormentShrinksBarsBlockingAndDrainsItsControllerEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	a, b := g.Seats[0], g.Seats[1]
	hostA := effectstest.B12Creature(g, a.ID, "Host A", "Creature — Bear", 3, 3)
	hostB := neoB12Artifact(g, b.ID)
	effectstest.Cmdr2B19Attach(g, g.Seats[2].ID, hostA, "Clawing Torment", neoClawingTorment)
	effectstest.Cmdr2B19Attach(g, g.Seats[2].ID, hostB, "Clawing Torment", neoClawingTorment)
	if p, tt := effectstest.PtOf(t, g, hostA); p != 2 || tt != 2 {
		t.Errorf("enchanted creature is %d/%d, want 2/2", p, tt)
	}
	var cantBlockA, cantBlockB bool
	g.ReadSnapshot(func() {
		ca, _ := g.LookupCardForEffect(hostA)
		cb, _ := g.LookupCardForEffect(hostB)
		cantBlockA = ca.Effective().Restrictions&game.CantBlock != 0
		cantBlockB = cb.Effective().Restrictions&game.CantBlock != 0
	})
	if !cantBlockA || cantBlockB {
		t.Errorf("can't block: creature %v (want true), artifact %v (want false)", cantBlockA, cantBlockB)
	}
	lifeA, lifeB, lifeC := effectstest.LifeOf(g, a.ID), effectstest.LifeOf(g, b.ID), effectstest.LifeOf(g, g.Seats[2].ID)
	effectstest.AdvanceTo(t, g, game.StepUpkeep)
	effectstest.PassPriorityAroundTable(t, g)
	active := g.Seats[g.Turn.ActiveSeat].ID
	for _, tc := range []struct {
		name string
		id   uuid.UUID
		was  int
	}{{"A", a.ID, lifeA}, {"B", b.ID, lifeB}} {
		want := tc.was
		if tc.id == active {
			want--
		}
		if got := effectstest.LifeOf(g, tc.id); got != want {
			t.Errorf("seat %s life %d, want %d on its own upkeep only", tc.name, got, want)
		}
	}
	if got := effectstest.LifeOf(g, g.Seats[2].ID); got != lifeC {
		t.Errorf("the Aura's controller lost life: %d → %d", lifeC, got)
	}
}

func TestS82NeoB12TamiyosCompleationTapsAndSilencesACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	host := neoB12Flier(g, g.Seats[1].ID, 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Tamiyo's Compleation", effectstest.AuraTypeLine, neoTamiyosCompleation, effectstest.CardRefs(host))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, host)
	if !c.Tapped {
		t.Error("the enchanted creature is not tapped")
	}
	if effectstest.HasEffectiveKeyword(t, g, host, "flying") {
		t.Error("the enchanted creature kept flying")
	}
	_ = me
}

func TestS82NeoB12TamiyosCompleationUnattachesAnEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bearer := effectstest.B12Creature(g, me.ID, "Bearer", "Creature — Bear", 2, 2)
	sword := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Sword", TypeLine: "Artifact — Equipment",
		Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: bearer}})
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Tamiyo's Compleation", effectstest.AuraTypeLine, neoTamiyosCompleation, effectstest.CardRefs(sword))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, sword)
	if !c.Tapped {
		t.Error("the Equipment is not tapped")
	}
	if c.AttachedTo.Kind == game.TargetCard {
		t.Error("the Equipment is still attached to the creature")
	}
}

func TestS82NeoB12TamiyosCompleationDoesNotUntap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	effectstest.Cmdr2B19Attach(g, g.Seats[1].ID, host, "Tamiyo's Compleation", neoTamiyosCompleation)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(host) })
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if c, _ := effectstest.BattlefieldCard(g, host); !c.Tapped {
		t.Error("the enchanted creature untapped during its controller's untap step")
	}
}

func TestS82NeoB12TwistedEmbraceDestroysAnOpposingCreatureAndPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	bystander := effectstest.B12Creature(g, me.ID, "Bystander", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Twisted Embrace", effectstest.AuraTypeLine, neoTwistedEmbrace, effectstest.CardRefs(mine))
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	for _, id := range pick.PickTargetCards {
		if id == bystander || id == mine {
			t.Errorf("my own creature %s is offered as a target", id)
		}
	}
	effectstest.PickTriggerTarget(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the opponent's creature was not destroyed")
	}
	if !g.Battlefield.Contains(bystander) {
		t.Error("my own creature was destroyed")
	}
	if p, tt := effectstest.PtOf(t, g, mine); p != 3 || tt != 3 {
		t.Errorf("enchanted creature is %d/%d, want 3/3", p, tt)
	}
}

func TestS82NeoB12GraftedGrowthCountersACreatureAndTapsTheLandForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := neoB12Land(g, me.ID)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Grafted Growth", effectstest.AuraTypeLine, neoGraftedGrowth, effectstest.CardRefs(land))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, bear)
	if got := c.Counters[game.CounterPlusOne]; got != 1 {
		t.Errorf("%d +1/+1 counters on the creature, want 1", got)
	}
	var produced string
	g.ReadSnapshot(func() {
		card, _ := g.LookupCardForEffect(land)
		abs, origins := game.ManaAbilitiesWithOrigins(card)
		for i, o := range origins {
			if o.Granted() {
				produced = abs[i].Produced
			}
		}
	})
	if produced != effects.OneColorOfAmount(2) {
		t.Errorf("granted mana ability produces %q, want two mana of any one color", produced)
	}
}

func TestS82NeoB12GraftedGrowthCountersAVehicle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := neoB12Land(g, me.ID)
	van := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Van", TypeLine: "Artifact — Vehicle",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Grafted Growth", effectstest.AuraTypeLine, neoGraftedGrowth, effectstest.CardRefs(land))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, van)
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, van)
	if got := c.Counters[game.CounterPlusOne]; got != 1 {
		t.Errorf("%d +1/+1 counters on the Vehicle, want 1", got)
	}
}

func TestS82NeoB12UnquenchableFuryDealsDamageEqualToTheirHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	effectstest.Cmdr2B19Attach(g, me.ID, host, "Unquenchable Fury", neoUnquenchableFury)
	g.WithWriteLock(func() {
		for i := 0; i < 3; i++ {
			opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
		}
	})
	want := opp.Hand.Size()
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.DeclareAttack(t, g, opp.ID, host)
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - effectstest.LifeOf(g, opp.ID); got != want {
		t.Errorf("the defending player lost %d life, want %d (their hand)", got, want)
	}
}

func TestS82NeoB12UnquenchableFuryReturnsToHandWhenItGoesToYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	aura := effectstest.Cmdr2B19Attach(g, me.ID, host, "Unquenchable Fury", neoUnquenchableFury)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(host) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(aura) {
		t.Error("the Aura did not return to its owner's hand")
	}
}

func TestS82NeoB12EmergenceAnimatesTheLandAndSavesItFromDestruction(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, color string
		p, tt               int
		keywords            []string
	}{
		{"Crackling Emergence", neoCracklingEmergence, "R", 3, 3, []string{"haste"}},
		{"Harmonious Emergence", neoHarmoniousEmergence, "G", 4, 5, []string{"vigilance", "haste"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			land := neoB12Land(g, me.ID)
			aura := effectstest.Cmdr2B19Attach(g, me.ID, land, tc.name, tc.oracle)
			var eff game.Characteristic
			g.ReadSnapshot(func() {
				c, _ := g.LookupCardForEffect(land)
				eff = c.Effective()
			})
			if !effects.TypeListHasEffects(eff.Types, "Creature") || !effects.TypeListHasEffects(eff.Types, "Land") || !effects.TypeListHasEffects(eff.Subtypes, "Spirit") {
				t.Errorf("types %v subtypes %v, want a Spirit creature that is still a land", eff.Types, eff.Subtypes)
			}
			if len(eff.Colors) != 1 || eff.Colors[0] != tc.color {
				t.Errorf("colors %v, want %s", eff.Colors, tc.color)
			}
			if p, tt := effectstest.PtOf(t, g, land); p != tc.p || tt != tc.tt {
				t.Errorf("land is %d/%d, want %d/%d", p, tt, tc.p, tc.tt)
			}
			for _, kw := range tc.keywords {
				if !effectstest.HasEffectiveKeyword(t, g, land, kw) {
					t.Errorf("land lacks %s", kw)
				}
			}
			// A destruction is replaced: the Aura is sacrificed and the land survives, indestructible.
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(land) })
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(land) {
				t.Fatal("the land was destroyed")
			}
			if g.Battlefield.Contains(aura) {
				t.Error("the Aura was not sacrificed")
			}
			if !effectstest.HasEffectiveKeyword(t, g, land, "indestructible") {
				t.Error("the land did not gain indestructible")
			}
			// It stays safe for the rest of the turn.
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(land) })
			if !g.Battlefield.Contains(land) {
				t.Error("an indestructible land was destroyed")
			}
		})
	}
}

func TestS82NeoB12EmergenceDoesNotReplaceASacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := neoB12Land(g, me.ID)
	aura := effectstest.Cmdr2B19Attach(g, me.ID, land, "Crackling Emergence", neoCracklingEmergence)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(land) })
	if g.Battlefield.Contains(land) {
		t.Error("a sacrificed land stayed on the battlefield")
	}
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(aura) {
		t.Error("the Aura stayed after its land was sacrificed")
	}
}
