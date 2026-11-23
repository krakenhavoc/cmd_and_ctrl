package testss82m

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_neo_b07_cards_test.go — S82 NEO batch 7 (#5493): the channel cards.
// Every test plays the real catalog card through the engine.

const (
	neoB07Crab      = "cd727e02-5064-4b89-9398-e3cc4463ba56"
	neoB07Sniper    = "b72fd535-0051-4f59-9d99-75895f751006"
	neoB07Favor     = "08af6fdc-4068-4652-b1fc-cda5309c7b71"
	neoB07Sphere    = "53bfffae-b99e-44bd-8954-6cf236f93f26"
	neoB07Boar      = "90266ce6-4360-4de4-9b06-adff7d6c378b"
	neoB07Ronin     = "5e891927-b6cb-40ae-8270-0e72820fd7f5"
	neoB07Archer    = "657f6318-aa09-46f1-ab69-896b17a48950"
	neoB07Earth     = "43abfce5-b1dd-4784-bb91-d57b2c28f6a6"
	neoB07Prototype = "7346c266-19d8-4432-acf5-f00aff074ad5"
	neoB07Samurai   = "b767bd0a-a8fe-49d7-a3e7-4f9ac6a3280f"
	neoB07Trespass  = "d5cdfa86-a234-4ddd-bfa0-4a3bc51eebce"
	neoB07Preserver = "a8ce90be-f0ab-4f8d-896a-cde8aaf24579"
	neoB07Turtle    = "871e4c4e-8105-4cb9-b1cd-53c5f8e64494"
)

func TestNeoB07EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		neoB07Crab: "Mirrorshell Crab", neoB07Sniper: "Twinshot Sniper", neoB07Favor: "Favor of Jukai",
		neoB07Sphere: "Mnemonic Sphere", neoB07Boar: "Ironhoof Boar", neoB07Ronin: "Reinforced Ronin",
		neoB07Archer: "Bamboo Grove Archer", neoB07Earth: "Roaring Earth", neoB07Prototype: "Moonsnare Prototype",
		neoB07Samurai: "Sunblade Samurai", neoB07Trespass: "Saiba Trespassers",
		neoB07Preserver: "Jukai Preserver", neoB07Turtle: "Colossal Skyturtle",
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

func neoB07Creature(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: owner, Controller: owner,
	})
}

// --- Mirrorshell Crab -----------------------------------------------------------------

func TestNeoB07CrabHasWard(t *testing.T) {
	spec, _ := effects.Lookup(neoB07Crab)
	if len(spec.Triggered) != 1 || len(spec.Activated) != 1 {
		t.Fatalf("triggered %d, activated %d; want ward and one channel", len(spec.Triggered), len(spec.Activated))
	}
}

func TestNeoB07CrabCountersASpellUnlessItsControllerPays(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		spell := effectstest.CastCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
		id := effectstest.PushCatalogHandCard(me, "Mirrorshell Crab", "Artifact Creature — Crab", neoB07Crab)
		if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{U}"); err != nil {
			t.Fatal(err)
		}
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(spell)}); err != nil {
			t.Fatalf("channel: %v", err)
		}
		if !me.Graveyard.Contains(id) {
			t.Error("the discard is a cost: the Crab must be in the graveyard")
		}
		effectstest.PassUntilTaxed(t, g, me.ID)
		if pay {
			effectstest.FillPool(me, 3)
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		if g.Stack.Contains(spell) != pay {
			t.Errorf("paid=%v: the spell is still on the stack = %v", pay, g.Stack.Contains(spell))
		}
	}
}

func TestNeoB07CrabCountersAnAbilityUnlessItsControllerPays(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
		src := neoB07Creature(g, opp.ID, "Their Bear")
		fired := false
		ability := effectstest.TaAnnounce(t, g, opp.ID, src, "Their ability", nil, &fired)
		if _, err := effectstest.NeoB07Channel(t, g, "Mirrorshell Crab", "Artifact Creature — Crab", neoB07Crab, "{C}{C}{U}", 0,
			game.ActivateAbilityParams{Targets: effectstest.CardTarget(ability)}); err != nil {
			t.Fatalf("channel: %v", err)
		}
		effectstest.PassUntilTaxed(t, g, opp.ID)
		if pay {
			effectstest.FillPool(opp, 3)
		}
		effectstest.AnswerPayUnless(t, g, opp.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if fired != pay {
			t.Errorf("paid=%v: the ability resolved = %v", pay, fired)
		}
		_ = me
	}
}

func TestNeoB07CrabChannelIsRefusedShortOfManaOrWithNoTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	spell := effectstest.CastCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	if _, err := effectstest.NeoB07Channel(t, g, "Mirrorshell Crab", "Artifact Creature — Crab", neoB07Crab, "{C}{U}", 0,
		game.ActivateAbilityParams{Targets: effectstest.CardTarget(spell)}); err == nil {
		t.Error("channel went through with {C}{U} for {2}{U}")
	}
	g2 := effectstest.NewCatalogGame(t)
	if _, err := effectstest.NeoB07Channel(t, g2, "Mirrorshell Crab", "Artifact Creature — Crab", neoB07Crab, "{C}{C}{U}", 0,
		game.ActivateAbilityParams{}); err == nil {
		t.Error("channel went through with nothing to target")
	}
}

// --- Twinshot Sniper ------------------------------------------------------------------

func TestNeoB07SniperEntersAndDealsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	life := opp.Life
	id := effectstest.CastCatalogSpell(t, g, "Twinshot Sniper", "Artifact Creature — Goblin Archer", neoB07Sniper, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d", opp.Life, life-2)
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "reach") {
		t.Error("no reach")
	}
}

func TestNeoB07SniperChannelDealsTwoToAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	life := opp.Life
	id, err := effectstest.NeoB07Channel(t, g, "Twinshot Sniper", "Artifact Creature — Goblin Archer", neoB07Sniper, "{C}{R}", 0,
		game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if err != nil {
		t.Fatalf("channel: %v", err)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the Sniper is not in the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d", opp.Life, life-2)
	}
}

func TestNeoB07SniperChannelKillsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := neoB07Creature(g, opp.ID, "Their Bear")
	if _, err := effectstest.NeoB07Channel(t, g, "Twinshot Sniper", "Artifact Creature — Goblin Archer", neoB07Sniper, "{C}{R}", 0,
		game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the 2/2 survived 2 damage")
	}
}

// --- Favor of Jukai -------------------------------------------------------------------

func TestNeoB07FavorChannelPumpsUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := neoB07Creature(g, me.ID, "Bear")
	if _, err := effectstest.NeoB07Channel(t, g, "Favor of Jukai", "Enchantment — Aura", neoB07Favor, "{C}{G}", 0,
		game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, bear); p != 5 || th != 5 {
		t.Errorf("the bear is %d/%d, want 5/5", p, th)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "reach") {
		t.Error("no reach")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, bear); p != 2 || th != 2 {
		t.Errorf("next turn the bear is %d/%d, want 2/2", p, th)
	}
}

func TestNeoB07FavorOnAnArtifactOnlyWorksWhileItIsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID,
	})
	bear := neoB07Creature(g, me.ID, "Bear")
	effectstest.CastCatalogSpell(t, g, "Favor of Jukai", "Enchantment — Aura", neoB07Favor, effectstest.CardTarget(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if v := effectstest.S58p6View(t, g, rock); v.IsCreature() || effectstest.HasEffectiveKeyword(t, g, rock, "reach") {
		t.Error("a non-creature artifact got the bonus")
	}
	// Enchanting a creature gives it +3/+3 and reach.
	effectstest.CastCatalogSpell(t, g, "Favor of Jukai", "Enchantment — Aura", neoB07Favor, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, bear); p != 5 || th != 5 {
		t.Errorf("the enchanted bear is %d/%d, want 5/5", p, th)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "reach") {
		t.Error("the enchanted bear has no reach")
	}
}

// --- Mnemonic Sphere ------------------------------------------------------------------

func TestNeoB07SphereSacrificeDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sphere := effectstest.PushCatalogPermanent(g, me.ID, "Mnemonic Sphere", "Artifact", neoB07Sphere, false)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{U}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, sphere, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(sphere) {
		t.Error("the Sphere was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d → %d, want two cards", hand, got)
	}
}

func TestNeoB07SphereChannelDrawsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	id, err := effectstest.NeoB07Channel(t, g, "Mnemonic Sphere", "Artifact", neoB07Sphere, "{U}", 1, game.ActivateAbilityParams{})
	if err != nil {
		t.Fatalf("channel: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) {
		t.Error("the Sphere is not in the graveyard")
	}
	// +1 for the pushed card, -1 for the discard, +1 for the draw.
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want one more card than before", hand, got)
	}
	// Channel works from the hand only: a battlefield Sphere has no such offer.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	sphere := effectstest.PushCatalogPermanent(g2, me2.ID, "Mnemonic Sphere", "Artifact", neoB07Sphere, false)
	effectstest.AdvanceToMain(t, g2)
	if err := g2.AddManaForEffect(me2.ID, uuid.Nil, "{U}"); err != nil {
		t.Fatal(err)
	}
	if err := g2.ActivateCatalogAbility(me2.ID, sphere, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("channel was activated from the battlefield")
	}
}

// --- Ironhoof Boar --------------------------------------------------------------------

func TestNeoB07BoarChannelGivesPlusThreePlusOneAndTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := neoB07Creature(g, me.ID, "Bear")
	if _, err := effectstest.NeoB07Channel(t, g, "Ironhoof Boar", "Artifact Creature — Boar", neoB07Boar, "{C}{R}", 0,
		game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, bear); p != 5 || th != 3 {
		t.Errorf("the bear is %d/%d, want 5/3", p, th)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("no trample")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, bear); p != 2 || th != 2 {
		t.Errorf("next turn the bear is %d/%d, want 2/2", p, th)
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("trample outlived the turn")
	}
}

func TestNeoB07BoarHasTrampleAndHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	boar := effectstest.PushCatalogPermanent(g, g.Seats[0].ID, "Ironhoof Boar", "Artifact Creature — Boar", neoB07Boar, false)
	for _, kw := range []string{"trample", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, boar, kw) {
			t.Errorf("no %s", kw)
		}
	}
}

// --- Reinforced Ronin -----------------------------------------------------------------

func TestNeoB07RoninReturnsToHandAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ronin := effectstest.PushCatalogPermanent(g, me.ID, "Reinforced Ronin", "Artifact Creature — Human Samurai", neoB07Ronin, false)
	if !effectstest.HasEffectiveKeyword(t, g, ronin, "haste") {
		t.Error("no haste")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ronin) || !me.Hand.Contains(ronin) {
		t.Error("the Ronin was not returned to its owner's hand at the end step")
	}
}

func TestNeoB07RoninChannelDrawsACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	if _, err := effectstest.NeoB07Channel(t, g, "Reinforced Ronin", "Artifact Creature — Human Samurai", neoB07Ronin, "{C}{R}", 0,
		game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want +1 (pushed, discarded, drew)", hand, got)
	}
}

// --- Bamboo Grove Archer --------------------------------------------------------------

func TestNeoB07ArcherChannelDestroysOnlyAFlyer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird", Power: 1, Toughness: 1,
		Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	bear := neoB07Creature(g, opp.ID, "Their Bear")
	if _, err := effectstest.NeoB07Channel(t, g, "Bamboo Grove Archer", "Enchantment Creature — Snake Archer", neoB07Archer, "{C}{C}{C}{C}{G}", 0,
		game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Fatal("the channel targeted a creature without flying")
	}
	g2 := effectstest.NewCatalogGame(t)
	opp2 := g2.Seats[1]
	flyer2 := effectstest.PushBattlefieldCardWithTimestamp(g2, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird", Power: 1, Toughness: 1,
		Keywords: []string{"flying"}, Owner: opp2.ID, Controller: opp2.ID,
	})
	if _, err := effectstest.NeoB07Channel(t, g2, "Bamboo Grove Archer", "Enchantment Creature — Snake Archer", neoB07Archer, "{C}{C}{C}{C}{G}", 0,
		game.ActivateAbilityParams{Targets: effectstest.CardTarget(flyer2)}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(flyer2) {
		t.Error("the flyer survived")
	}
	_ = flyer
}

func TestNeoB07ArcherHasDefenderAndReach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	archer := effectstest.PushCatalogPermanent(g, g.Seats[0].ID, "Bamboo Grove Archer", "Enchantment Creature — Snake Archer", neoB07Archer, false)
	for _, kw := range []string{"defender", "reach"} {
		if !effectstest.HasEffectiveKeyword(t, g, archer, kw) {
			t.Errorf("no %s", kw)
		}
	}
}

// --- Roaring Earth --------------------------------------------------------------------

func TestNeoB07EarthLandfallPutsACounterOnACreatureOrVehicle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Roaring Earth", "Enchantment", neoB07Earth, false)
	bear := neoB07Creature(g, me.ID, "Bear")
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the bear has %d +1/+1 counters, want 1", got)
	}
}

func TestNeoB07EarthChannelAnimatesALandWithXCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	if _, err := effectstest.NeoB07Channel(t, g, "Roaring Earth", "Enchantment", neoB07Earth, "{C}{C}{C}{G}{G}", 0,
		game.ActivateAbilityParams{Targets: effectstest.CardTarget(land), XValue: 3}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	v := effectstest.S58p6View(t, g, land)
	if !v.IsCreature() || !v.IsLand() || !v.HasSubtype("Spirit") {
		t.Fatalf("the land is %v, want a Spirit creature land", v.Effective().Types)
	}
	if !slices.Contains(v.Effective().Colors, "G") {
		t.Errorf("colours %v, want green", v.Effective().Colors)
	}
	if p, th := neoB07PT(t, g, land); p != 3 || th != 3 {
		t.Errorf("the land is %d/%d, want 3/3 (0/0 with three counters)", p, th)
	}
	if !effectstest.HasEffectiveKeyword(t, g, land, "haste") {
		t.Error("no haste")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, _ := neoB07PT(t, g, land); p != 3 {
		t.Errorf("a turn later it is %d/_, want it to stay a creature (no duration)", p)
	}
}

func TestNeoB07EarthChannelWithXZeroStillAnimatesAndDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	if _, err := effectstest.NeoB07Channel(t, g, "Roaring Earth", "Enchantment", neoB07Earth, "{G}{G}", 0,
		game.ActivateAbilityParams{Targets: effectstest.CardTarget(land), XValue: 0}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(land) {
		t.Error("a 0/0 land creature stayed on the battlefield")
	}
}

func TestNeoB07EarthChannelRefusesAnOpponentsLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := effectstest.SeedLandOnBattlefield(g, g.Seats[1].ID, "Forest", "Basic Land — Forest")
	if _, err := effectstest.NeoB07Channel(t, g, "Roaring Earth", "Enchantment", neoB07Earth, "{C}{G}{G}", 0,
		game.ActivateAbilityParams{Targets: effectstest.CardTarget(land), XValue: 1}); err == nil {
		t.Error("channel targeted an opponent's land")
	}
}

// --- Moonsnare Prototype --------------------------------------------------------------

func TestNeoB07PrototypeTapsAnotherPermanentForColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	proto := effectstest.PushCatalogPermanent(g, me.ID, "Moonsnare Prototype", "Artifact", neoB07Prototype, false)
	// With nothing else to tap, the ability can't be paid (the {T} taps the
	// Prototype itself, so the other permanent is another one).
	if err := g.ActivateManaAbility(me.ID, proto, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("activated with no other untapped artifact or creature")
	}
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Fresh Bear", "Creature — Bear", "", true)
	if err := g.ActivateManaAbility(me.ID, proto, 0, game.ManaAbilityParams{TapIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !effectstest.BattlefieldTapped(g, proto) || !effectstest.BattlefieldTapped(g, bear) {
		t.Error("both the Prototype and the other permanent must be tapped")
	}
	if got := effectstest.ColorlessFloating(g, me); got != 1 {
		t.Errorf("{C} floating %d, want 1", got)
	}
}

func TestNeoB07PrototypeChannelLetsTheOwnerChooseTopOrBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	size := opp.Library.Size()
	if _, err := effectstest.NeoB07Channel(t, g, "Moonsnare Prototype", "Artifact", neoB07Prototype, "{C}{C}{C}{C}{U}", 0,
		game.ActivateAbilityParams{Targets: effectstest.CardTarget(theirs)}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The prompt belongs to the OWNER of the permanent, not to the channeler.
	var asked *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil {
			asked = c
		}
	}
	if asked == nil || asked.Chooser != opp.ID || asked.Chooser == me.ID {
		t.Fatalf("pending choice %+v, want one addressed to the permanent's owner", asked)
	}
	_ = size
}

func TestNeoB07PrototypeChannelRefusesALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := effectstest.SeedLandOnBattlefield(g, g.Seats[1].ID, "Forest", "Basic Land — Forest")
	if _, err := effectstest.NeoB07Channel(t, g, "Moonsnare Prototype", "Artifact", neoB07Prototype, "{C}{C}{C}{C}{U}", 0,
		game.ActivateAbilityParams{Targets: effectstest.CardTarget(land)}); err == nil {
		t.Error("channel targeted a land")
	}
}

// --- Sunblade Samurai -----------------------------------------------------------------

func TestNeoB07SamuraiChannelFetchesAPlainsAndGainsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	plains := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	life := me.Life
	if _, err := effectstest.NeoB07Channel(t, g, "Sunblade Samurai", "Enchantment Creature — Human Samurai", neoB07Samurai, "{C}{C}", 0,
		game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, plains)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(plains) {
		t.Error("the Plains did not reach the hand")
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestNeoB07SamuraiChannelGainsLifeEvenWithNoPlains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	if _, err := effectstest.NeoB07Channel(t, g, "Sunblade Samurai", "Enchantment Creature — Human Samurai", neoB07Samurai, "{C}{C}", 0,
		game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

// --- Saiba Trespassers ----------------------------------------------------------------

func TestNeoB07TrespassersChannelTapsAndFreezesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := neoB07Creature(g, opp.ID, "A")
	b := neoB07Creature(g, opp.ID, "B")
	c := neoB07Creature(g, opp.ID, "C")
	if _, err := effectstest.NeoB07Channel(t, g, "Saiba Trespassers", "Artifact Creature — Moonfolk Rogue", neoB07Trespass, "{C}{C}{C}{U}", 0,
		game.ActivateAbilityParams{Targets: effectstest.CardTargets(a, b, c)}); err == nil {
		t.Fatal("three targets were accepted for up to two")
	}
	g2 := effectstest.NewCatalogGame(t)
	opp2 := g2.Seats[1]
	a2 := neoB07Creature(g2, opp2.ID, "A")
	b2 := neoB07Creature(g2, opp2.ID, "B")
	c2 := neoB07Creature(g2, opp2.ID, "C")
	if _, err := effectstest.NeoB07Channel(t, g2, "Saiba Trespassers", "Artifact Creature — Moonfolk Rogue", neoB07Trespass, "{C}{C}{C}{U}", 0,
		game.ActivateAbilityParams{Targets: effectstest.CardTargets(a2, b2)}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if !effectstest.BattlefieldTapped(g2, a2) || !effectstest.BattlefieldTapped(g2, b2) || effectstest.BattlefieldTapped(g2, c2) {
		t.Error("exactly the two targets must be tapped")
	}
	if neoB07Holds(t, g2, a2) != 1 || neoB07Holds(t, g2, b2) != 1 || neoB07Holds(t, g2, c2) != 0 {
		t.Errorf("don't-untap holds: %d %d %d, want 1 1 0", neoB07Holds(t, g2, a2), neoB07Holds(t, g2, b2), neoB07Holds(t, g2, c2))
	}
	_, _, _ = a, b, c
}

func TestNeoB07TrespassersChannelRefusesYourOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	mine := neoB07Creature(g, g.Seats[g.Turn.ActiveSeat].ID, "Mine")
	if _, err := effectstest.NeoB07Channel(t, g, "Saiba Trespassers", "Artifact Creature — Moonfolk Rogue", neoB07Trespass, "{C}{C}{C}{U}", 0,
		game.ActivateAbilityParams{Targets: effectstest.CardTarget(mine)}); err == nil {
		t.Error("channel targeted a creature its controller controls")
	}
}

// --- Jukai Preserver ------------------------------------------------------------------

func TestNeoB07PreserverEntersWithACounterOnTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := neoB07Creature(g, me.ID, "Bear")
	effectstest.CastCatalogSpell(t, g, "Jukai Preserver", "Enchantment Creature — Human Druid", neoB07Preserver, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the bear has %d counters, want 1", got)
	}
}

func TestNeoB07PreserverChannelPutsCountersOnUpToTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := neoB07Creature(g, me.ID, "A")
	b := neoB07Creature(g, me.ID, "B")
	c := neoB07Creature(g, me.ID, "C")
	if _, err := effectstest.NeoB07Channel(t, g, "Jukai Preserver", "Enchantment Creature — Human Druid", neoB07Preserver, "{C}{C}{G}", 0,
		game.ActivateAbilityParams{Targets: effectstest.CardTargets(a, b, c)}); err == nil {
		t.Fatal("three targets were accepted for up to two")
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	a2 := neoB07Creature(g2, me2.ID, "A")
	b2 := neoB07Creature(g2, me2.ID, "B")
	c2 := neoB07Creature(g2, me2.ID, "C")
	if _, err := effectstest.NeoB07Channel(t, g2, "Jukai Preserver", "Enchantment Creature — Human Druid", neoB07Preserver, "{C}{C}{G}", 0,
		game.ActivateAbilityParams{Targets: effectstest.CardTargets(a2, b2)}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if effectstest.CountersOn(g2, a2, game.CounterPlusOne) != 1 || effectstest.CountersOn(g2, b2, game.CounterPlusOne) != 1 || effectstest.CountersOn(g2, c2, game.CounterPlusOne) != 0 {
		t.Error("exactly the two targets must get a counter")
	}
	_, _, _ = a, b, c
}

// --- Colossal Skyturtle ---------------------------------------------------------------

func TestNeoB07SkyturtleHasTwoChannelAbilitiesAndWard(t *testing.T) {
	spec, _ := effects.Lookup(neoB07Turtle)
	if len(spec.Activated) != 2 || len(spec.Triggered) != 1 {
		t.Fatalf("activated %d, triggered %d; want two channels and ward", len(spec.Activated), len(spec.Triggered))
	}
	g := effectstest.NewCatalogGame(t)
	turtle := effectstest.PushCatalogPermanent(g, g.Seats[0].ID, "Colossal Skyturtle", "Enchantment Creature — Turtle", neoB07Turtle, false)
	if !effectstest.HasEffectiveKeyword(t, g, turtle, "flying") {
		t.Error("no flying")
	}
}

func TestNeoB07SkyturtleReturnsACardFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	card := game.Card{InstanceID: uuid.New(), Name: "Old Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID}
	me.Graveyard.PushTop(card)
	id, err := effectstest.NeoB07Channel(t, g, "Colossal Skyturtle", "Enchantment Creature — Turtle", neoB07Turtle, "{C}{C}{G}", 0,
		game.ActivateAbilityParams{Targets: effectstest.CardTarget(card.InstanceID)})
	if err != nil {
		t.Fatalf("channel: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(card.InstanceID) {
		t.Error("the card did not return to hand")
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the discarded Skyturtle should stay in the graveyard")
	}
}

func TestNeoB07SkyturtleCannotReturnItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogHandCard(me, "Colossal Skyturtle", "Enchantment Creature — Turtle", neoB07Turtle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{G}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(id)}); err == nil {
		t.Error("the Skyturtle targeted itself in the hand")
	}
}

func TestNeoB07SkyturtleBouncesACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := neoB07Creature(g, opp.ID, "Their Bear")
	if _, err := effectstest.NeoB07Channel(t, g, "Colossal Skyturtle", "Enchantment Creature — Turtle", neoB07Turtle, "{C}{U}", 1,
		game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("the creature was not returned to its owner's hand")
	}
}

// neoB07Holds counts the "doesn't untap next untap step" records on a permanent.
func neoB07Holds(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("battlefield card %s is missing", id)
	}
	return len(c.NextUntapSkips)
}

// neoB07PT is the power and toughness after a layer recompute.
func neoB07PT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	p, th, _, _ := effectstest.EarthbentBody(t, g, id)
	return p, th
}
