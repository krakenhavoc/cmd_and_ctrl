package effects

import (
	"testing"

	"github.com/google/uuid"

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
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func neoB12Artifact(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: owner, Controller: owner})
}

func neoB12Land(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: owner, Controller: owner})
}

func neoB12Flier(g *game.Game, owner uuid.UUID, p, tough int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Flier", TypeLine: "Creature — Bird",
		Power: p, Toughness: tough, Keywords: []string{"flying"}, Owner: owner, Controller: owner})
}

func neoB12CastStrict(t *testing.T, g *game.Game, name, oracle string, target uuid.UUID) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: auraTypeLine, OracleID: oracle, Owner: me.ID, Controller: me.ID})
	})
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: cardRefs(target)})
}

func TestS82NeoB12AkkiWarPaintPumpsACreatureAndWaitsOnANoncreatureArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := b12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	rock := neoB12Artifact(g, me.ID)
	cmdr2B19Attach(g, me.ID, host, "Akki War Paint", neoAkkiWarPaint)
	cmdr2B19Attach(g, me.ID, rock, "Akki War Paint", neoAkkiWarPaint)
	if p, tt := ptOf(t, g, host); p != 4 || tt != 3 {
		t.Errorf("enchanted creature is %d/%d, want 4/3", p, tt)
	}
	if c, _ := battlefieldCard(g, rock); c.Power != 0 || c.Toughness != 0 {
		t.Errorf("noncreature artifact got a pump: %d/%d", c.Power, c.Toughness)
	}
}

func TestS82NeoB12EnchantClausesRefuseTheWrongPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := neoB12Land(g, me.ID)
	for name, oracle := range map[string]string{"Akki War Paint": neoAkkiWarPaint, "Short Circuit": neoShortCircuit, "Tamiyo's Compleation": neoTamiyosCompleation} {
		if err := neoB12CastStrict(t, g, name, oracle, land); err == nil {
			t.Errorf("%s may enchant a land", name)
		}
	}
	// Twisted Embrace enchants only what you control; Emergences only your lands.
	oppCreature := b12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	if err := neoB12CastStrict(t, g, "Twisted Embrace", neoTwistedEmbrace, oppCreature); err == nil {
		t.Error("Twisted Embrace enchanted an opponent's creature")
	}
	oppLand := neoB12Land(g, g.Seats[1].ID)
	if err := neoB12CastStrict(t, g, "Crackling Emergence", neoCracklingEmergence, oppLand); err == nil {
		t.Error("Crackling Emergence enchanted an opponent's land")
	}
	creature := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	if err := neoB12CastStrict(t, g, "Grafted Growth", neoGraftedGrowth, creature); err == nil {
		t.Error("Grafted Growth enchanted a creature")
	}
}

func TestS82NeoB12ShortCircuitShrinksAndGroundsACreatureButNotAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := neoB12Flier(g, me.ID, 4, 4)
	rock := neoB12Artifact(g, me.ID)
	cmdr2B19Attach(g, me.ID, host, "Short Circuit", neoShortCircuit)
	cmdr2B19Attach(g, me.ID, rock, "Short Circuit", neoShortCircuit)
	if p, tt := ptOf(t, g, host); p != 1 || tt != 4 {
		t.Errorf("enchanted creature is %d/%d, want 1/4", p, tt)
	}
	if hasEffectiveKeyword(t, g, host, "flying") {
		t.Error("the enchanted creature still has flying")
	}
	if c, _ := battlefieldCard(g, rock); c.Power != 0 {
		t.Errorf("noncreature artifact got -3/-0: power %d", c.Power)
	}
}

func TestS82NeoB12ShortCircuitHasFlash(t *testing.T) {
	spec, _ := Lookup(neoShortCircuit)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("printed keywords %v, want flash", spec.PrintedKeywords)
	}
	spec, _ = Lookup(neoTamiyosCompleation)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("Tamiyo's Compleation printed keywords %v, want flash", spec.PrintedKeywords)
	}
}

func TestS82NeoB12GiftOfWrathPumpsAndLeavesASpirit(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	host := b12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	aura := cmdr2B19Attach(g, me.ID, host, "Gift of Wrath", neoGiftOfWrath)
	if p, tt := ptOf(t, g, host); p != 4 || tt != 4 {
		t.Errorf("enchanted creature is %d/%d, want 4/4", p, tt)
	}
	if !hasEffectiveKeyword(t, g, host, "menace") {
		t.Error("the enchanted creature lacks menace")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	passPriorityAroundTable(t, g)
	if n := namedOnBattlefield(g, "Spirit", me.ID); n != 1 {
		t.Fatalf("%d Spirit tokens after the Aura left, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" {
			if c.Power != 2 || c.Toughness != 2 || !hasEffectiveKeyword(t, g, c.InstanceID, "menace") || len(c.Colors) != 1 || c.Colors[0] != "R" {
				t.Errorf("Spirit token is %d/%d colors %v, want a 2/2 red creature with menace", c.Power, c.Toughness, c.Colors)
			}
		}
	}
	if p, tt := ptOf(t, g, host); p != 2 || tt != 2 {
		t.Errorf("host after the Aura left is %d/%d, want 2/2", p, tt)
	}
}

func TestS82NeoB12GiftOfWrathTokenWhenTheHostDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	host := b12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	cmdr2B19Attach(g, me.ID, host, "Gift of Wrath", neoGiftOfWrath)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(host) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if n := namedOnBattlefield(g, "Spirit", me.ID); n != 1 {
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
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			seedLibrary(me, "A", "B", "C")
			host := b12Creature(g, me.ID, "Host", "Creature — Bear", tc.hostPower, 2)
			b12Creature(g, g.Seats[1].ID, "Big", "Creature — Giant", 4, 4)
			advanceToMain(t, g)
			before := me.Hand.Size()
			castCatalogSpell(t, g, "Historian's Wisdom", auraTypeLine, neoHistoriansWisdom, cardRefs(host))
			passPriorityAroundTable(t, g)
			if got := me.Hand.Size() - before; got != tc.wantDraw {
				t.Errorf("hand grew by %d, want %d", got, tc.wantDraw)
			}
			if p, tt := ptOf(t, g, host); p != tc.hostPower+2 || tt != 3 {
				t.Errorf("host is %d/%d, want %d/3", p, tt, tc.hostPower+2)
			}
		})
	}
}

func TestS82NeoB12HistoriansWisdomOnAnArtifactDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B", "C")
	rock := neoB12Artifact(g, me.ID)
	advanceToMain(t, g)
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Historian's Wisdom", auraTypeLine, neoHistoriansWisdom, cardRefs(rock))
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 0 {
		t.Errorf("hand grew by %d for a noncreature artifact", got)
	}
}

func TestS82NeoB12AwakenedAwarenessPutsXCountersAndSetsBasePT(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	host := b12Creature(g, me.ID, "Host", "Creature — Bear", 5, 5)
	toMainForCost(t, g)
	floatForTest(g, me, "UUUUUU")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Awakened Awareness", TypeLine: auraTypeLine, ManaCost: "{X}{U}{U}",
			OracleID: neoAwakenedAwareness, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 3, Targets: cardRefs(host)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, host)
	if got := c.Counters[game.CounterPlusOne]; got != 3 {
		t.Fatalf("%d +1/+1 counters, want 3", got)
	}
	if p, tt := ptOf(t, g, host); p != 4 || tt != 4 {
		t.Errorf("enchanted creature is %d/%d, want base 1/1 plus three counters (4/4)", p, tt)
	}
}

func TestS82NeoB12AwakenedAwarenessLeavesANoncreatureArtifactAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := neoB12Artifact(g, me.ID)
	cmdr2B19Attach(g, me.ID, rock, "Awakened Awareness", neoAwakenedAwareness)
	if c, _ := battlefieldCard(g, rock); c.Power != 0 || c.Toughness != 0 {
		t.Errorf("noncreature artifact has a base P/T: %d/%d", c.Power, c.Toughness)
	}
}

func TestS82NeoB12ClawingTormentShrinksBarsBlockingAndDrainsItsControllerEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	a, b := g.Seats[0], g.Seats[1]
	hostA := b12Creature(g, a.ID, "Host A", "Creature — Bear", 3, 3)
	hostB := neoB12Artifact(g, b.ID)
	cmdr2B19Attach(g, g.Seats[2].ID, hostA, "Clawing Torment", neoClawingTorment)
	cmdr2B19Attach(g, g.Seats[2].ID, hostB, "Clawing Torment", neoClawingTorment)
	if p, tt := ptOf(t, g, hostA); p != 2 || tt != 2 {
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
	lifeA, lifeB, lifeC := lifeOf(g, a.ID), lifeOf(g, b.ID), lifeOf(g, g.Seats[2].ID)
	advanceTo(t, g, game.StepUpkeep)
	passPriorityAroundTable(t, g)
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
		if got := lifeOf(g, tc.id); got != want {
			t.Errorf("seat %s life %d, want %d on its own upkeep only", tc.name, got, want)
		}
	}
	if got := lifeOf(g, g.Seats[2].ID); got != lifeC {
		t.Errorf("the Aura's controller lost life: %d → %d", lifeC, got)
	}
}

func TestS82NeoB12TamiyosCompleationTapsAndSilencesACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	host := neoB12Flier(g, g.Seats[1].ID, 3, 3)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Tamiyo's Compleation", auraTypeLine, neoTamiyosCompleation, cardRefs(host))
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, host)
	if !c.Tapped {
		t.Error("the enchanted creature is not tapped")
	}
	if hasEffectiveKeyword(t, g, host, "flying") {
		t.Error("the enchanted creature kept flying")
	}
	_ = me
}

func TestS82NeoB12TamiyosCompleationUnattachesAnEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bearer := b12Creature(g, me.ID, "Bearer", "Creature — Bear", 2, 2)
	sword := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Sword", TypeLine: "Artifact — Equipment",
		Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: bearer}})
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Tamiyo's Compleation", auraTypeLine, neoTamiyosCompleation, cardRefs(sword))
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, sword)
	if !c.Tapped {
		t.Error("the Equipment is not tapped")
	}
	if c.AttachedTo.Kind == game.TargetCard {
		t.Error("the Equipment is still attached to the creature")
	}
}

func TestS82NeoB12TamiyosCompleationDoesNotUntap(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := b12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	cmdr2B19Attach(g, g.Seats[1].ID, host, "Tamiyo's Compleation", neoTamiyosCompleation)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(host) })
	advanceToUpkeepOf(t, g, 0)
	if c, _ := battlefieldCard(g, host); !c.Tapped {
		t.Error("the enchanted creature untapped during its controller's untap step")
	}
}

func TestS82NeoB12TwistedEmbraceDestroysAnOpposingCreatureAndPumps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	bystander := b12Creature(g, me.ID, "Bystander", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Twisted Embrace", auraTypeLine, neoTwistedEmbrace, cardRefs(mine))
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	for _, id := range pick.PickTargetCards {
		if id == bystander || id == mine {
			t.Errorf("my own creature %s is offered as a target", id)
		}
	}
	pickTriggerTarget(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the opponent's creature was not destroyed")
	}
	if !g.Battlefield.Contains(bystander) {
		t.Error("my own creature was destroyed")
	}
	if p, tt := ptOf(t, g, mine); p != 3 || tt != 3 {
		t.Errorf("enchanted creature is %d/%d, want 3/3", p, tt)
	}
}

func TestS82NeoB12GraftedGrowthCountersACreatureAndTapsTheLandForTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := neoB12Land(g, me.ID)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Grafted Growth", auraTypeLine, neoGraftedGrowth, cardRefs(land))
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, bear)
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
	if produced != OneColorOfAmount(2) {
		t.Errorf("granted mana ability produces %q, want two mana of any one color", produced)
	}
}

func TestS82NeoB12GraftedGrowthCountersAVehicle(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := neoB12Land(g, me.ID)
	van := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Van", TypeLine: "Artifact — Vehicle",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Grafted Growth", auraTypeLine, neoGraftedGrowth, cardRefs(land))
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, van)
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, van)
	if got := c.Counters[game.CounterPlusOne]; got != 1 {
		t.Errorf("%d +1/+1 counters on the Vehicle, want 1", got)
	}
}

func TestS82NeoB12UnquenchableFuryDealsDamageEqualToTheirHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	host := b12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	cmdr2B19Attach(g, me.ID, host, "Unquenchable Fury", neoUnquenchableFury)
	g.WithWriteLock(func() {
		for i := 0; i < 3; i++ {
			opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
		}
	})
	want := opp.Hand.Size()
	life := lifeOf(g, opp.ID)
	declareAttack(t, g, opp.ID, host)
	passPriorityAroundTable(t, g)
	if got := life - lifeOf(g, opp.ID); got != want {
		t.Errorf("the defending player lost %d life, want %d (their hand)", got, want)
	}
}

func TestS82NeoB12UnquenchableFuryReturnsToHandWhenItGoesToYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	host := b12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	aura := cmdr2B19Attach(g, me.ID, host, "Unquenchable Fury", neoUnquenchableFury)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(host) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
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
			g := newCatalogGame(t)
			me := g.Seats[0]
			land := neoB12Land(g, me.ID)
			aura := cmdr2B19Attach(g, me.ID, land, tc.name, tc.oracle)
			var eff game.Characteristic
			g.ReadSnapshot(func() {
				c, _ := g.LookupCardForEffect(land)
				eff = c.Effective()
			})
			if !typeListHasEffects(eff.Types, "Creature") || !typeListHasEffects(eff.Types, "Land") || !typeListHasEffects(eff.Subtypes, "Spirit") {
				t.Errorf("types %v subtypes %v, want a Spirit creature that is still a land", eff.Types, eff.Subtypes)
			}
			if len(eff.Colors) != 1 || eff.Colors[0] != tc.color {
				t.Errorf("colors %v, want %s", eff.Colors, tc.color)
			}
			if p, tt := ptOf(t, g, land); p != tc.p || tt != tc.tt {
				t.Errorf("land is %d/%d, want %d/%d", p, tt, tc.p, tc.tt)
			}
			for _, kw := range tc.keywords {
				if !hasEffectiveKeyword(t, g, land, kw) {
					t.Errorf("land lacks %s", kw)
				}
			}
			// A destruction is replaced: the Aura is sacrificed and the land survives, indestructible.
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(land) })
			passPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(land) {
				t.Fatal("the land was destroyed")
			}
			if g.Battlefield.Contains(aura) {
				t.Error("the Aura was not sacrificed")
			}
			if !hasEffectiveKeyword(t, g, land, "indestructible") {
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := neoB12Land(g, me.ID)
	aura := cmdr2B19Attach(g, me.ID, land, "Crackling Emergence", neoCracklingEmergence)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(land) })
	if g.Battlefield.Contains(land) {
		t.Error("a sacrificed land stayed on the battlefield")
	}
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(aura) {
		t.Error("the Aura stayed after its land was sacrificed")
	}
}
