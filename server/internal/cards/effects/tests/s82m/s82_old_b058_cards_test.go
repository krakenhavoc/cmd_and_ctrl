package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 set OLD (older sets), batch 058, Commander 2021 (#5504). Each test
// plays the real catalog entry through the engine.

const (
	oldB058AuthorOracle   = "8bad4e43-4375-4e7f-897e-d28be72308d7"
	oldB058RoutOracle     = "d69aa624-23d1-406a-b7e0-9140a6f57929"
	oldB058ArashiOracle   = "9277b680-a146-4d4f-9649-39cddafdb585"
	oldB058TeysaOracle    = "98b283dc-96a6-45a3-8a3f-11458497f358"
	oldB058MuseOracle     = "0b65dc8d-a8bf-4fa1-a44e-6a3782d18c0f"
	oldB058CyclopsOracle  = "19a11c33-269a-45d3-86c7-827965053f26"
	oldB058GiftOracle     = "43a86d36-9653-4498-a533-8b30b919cdbf"
	oldB058SwarmOracle    = "4a7a5ea9-b633-459f-bcd4-40de53afce0e"
	oldB058EncoreOracle   = "921e7f6b-bf49-46f1-9fdd-b0e5fde5a5a1"
	oldB058BracersOracle  = "bd56225e-8c80-4328-9ca0-0fff4969573d"
	oldB058ReinterpretOrc = "e197104f-5382-4cb2-8afe-5ca0594629c2"
	oldB058BoltOracle     = "4457ed35-7c10-48c8-9776-456485fdf070"
)

func TestOldB058EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB058AuthorOracle:   "Author of Shadows",
		oldB058RoutOracle:     "Rout",
		oldB058ArashiOracle:   "Arashi, the Sky Asunder",
		oldB058TeysaOracle:    "Teysa, Envoy of Ghosts",
		oldB058MuseOracle:     "Muse Vortex",
		oldB058CyclopsOracle:  "Erratic Cyclops",
		oldB058GiftOracle:     "Gift of Paradise",
		oldB058SwarmOracle:    "Swarm Intelligence",
		oldB058EncoreOracle:   "Fiery Encore",
		oldB058BracersOracle:  "Battlemage's Bracers",
		oldB058ReinterpretOrc: "Reinterpret",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// handCard puts a card with the given cost in p's hand.
func oldB058Hand(p *game.Player, name, typeLine, cost, oracle string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	return id
}

// --- Erratic Cyclops ------------------------------------------------------

func TestOldB058CyclopsGetsPlusXForAnInstantOrSorcerysManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cyclops := effectstest.B12Push(g, me.ID, "Erratic Cyclops", "Creature — Cyclops Shaman", oldB058CyclopsOracle, 0, 8)
	effectstest.AdvanceToMain(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, cyclops, "trample") {
		t.Error("no trample")
	}
	effectstest.CastFromHandForTest(t, g, me, "Three Drop", "Sorcery", "{2}{R}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, cyclops); p != 3 || th != 8 {
		t.Errorf("Cyclops is %d/%d after a mana value 3 sorcery, want 3/8", p, th)
	}
	effectstest.CastFromHandForTest(t, g, me, "Two Drop", "Instant", "{1}{R}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, cyclops); p != 5 {
		t.Errorf("Cyclops power %d after a second spell of mana value 2, want 5", p)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, cyclops); p != 0 {
		t.Errorf("power %d next turn, want the boost gone", p)
	}
}

func TestOldB058CyclopsIgnoresCreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cyclops := effectstest.B12Push(g, me.ID, "Erratic Cyclops", "Creature — Cyclops Shaman", oldB058CyclopsOracle, 0, 8)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastFromHandForTest(t, g, me, "Bear", "Creature — Bear", "{2}{G}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, cyclops); p != 0 {
		t.Errorf("a creature spell gave +%d/+0", p)
	}
}

// --- Swarm Intelligence ---------------------------------------------------

func oldB058Bolt(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) {
	t.Helper()
	id := oldB058Hand(me, "Lightning Bolt", "Instant", "{R}", oldB058BoltOracle)
	effectstest.B06AddMana(me, "R")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: target}}}); err != nil {
		t.Fatalf("cast Bolt: %v", err)
	}
}

func TestOldB058SwarmIntelligenceCopiesAnInstant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, me.ID, "Swarm Intelligence", oldB058SwarmOracle, "Enchantment")
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	oldB058Bolt(t, g, me, opp.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 6 {
		t.Errorf("the opponent lost %d life, want 6 (the spell and its copy)", got)
	}
}

func TestOldB058SwarmIntelligenceMayDeclineAndSkipsCreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, me.ID, "Swarm Intelligence", oldB058SwarmOracle, "Enchantment")
	effectstest.AdvanceToMain(t, g)
	bear := oldB058Hand(me, "Bear", "Creature — Bear", "{1}", "")
	effectstest.B06AddMana(me, "G")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if effectstest.OwedMays(g, me.ID) != 0 {
		t.Error("offered to copy a creature spell")
	}
	effectstest.PassPriorityAroundTable(t, g)
	life := opp.Life
	oldB058Bolt(t, g, me, opp.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 3 {
		t.Errorf("the opponent lost %d life, want 3 (no copy)", got)
	}
}

// --- Gift of Paradise -----------------------------------------------------

func TestOldB058GiftOfParadiseGainsLifeAndGrantsTwoManaOfOneColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.B12Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Gift of Paradise", effectstest.AuraTypeLine, oldB058GiftOracle, effectstest.CardRefs(land))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
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

func TestOldB058GiftOfParadiseRefusesACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	creature := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	id := oldB058Hand(me, "Gift of Paradise", effectstest.AuraTypeLine, "", oldB058GiftOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(creature)}); err == nil {
		t.Error("Gift of Paradise enchanted a creature")
	}
}

// --- Rout -----------------------------------------------------------------

func TestOldB058RoutDestroysEveryCreatureEvenARegeneratingOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 3)
	land := effectstest.B12Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.RegenerateForEffect(theirs) })
	id := oldB058Hand(me, "Rout", "Sorcery", "", oldB058RoutOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("a creature survived Rout")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("a land was destroyed")
	}
}

func TestOldB058RoutHasFlashOnlyForTwoMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	id := oldB058Hand(me, "Rout", "Sorcery", "{3}{W}{W}", oldB058RoutOracle)
	fund := func(colors ...string) {
		g.WithWriteLock(func() {
			me.ManaPool = nil
			for _, c := range colors {
				me.ManaPool.AddMana(game.ManaToken{Color: c})
			}
		})
	}
	fund("W", "W", "C", "C", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast Rout at instant speed for its printed cost")
	}
	fund("W", "W", "C", "C", "C", "C", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("cast Rout at instant speed paying {2} more: %v", err)
	}
	if n := len(me.ManaPool); n != 0 {
		t.Errorf("%d mana left in the pool, want all seven spent", n)
	}
}

// --- Arashi, the Sky Asunder ---------------------------------------------

func oldB058Flyer(g *game.Game, owner uuid.UUID, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird", Power: 1, Toughness: toughness,
		Keywords: []string{"flying"}, Owner: owner, Controller: owner,
	})
}

func TestOldB058ArashiTapsToDealXDamageToAFlyingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arashi := effectstest.B12Push(g, me.ID, "Arashi, the Sky Asunder", "Legendary Creature — Spirit", oldB058ArashiOracle, 5, 5)
	bird := oldB058Flyer(g, opp.ID, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.ApaMana(me, "G", "G", "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, arashi, 0, game.ActivateAbilityParams{Strict: true, XValue: 3, Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("targeted a creature without flying")
	}
	effectstest.ApaActivate(t, g, me, arashi, 0, game.ActivateAbilityParams{XValue: 3, Targets: effectstest.CardRefs(bird)})
	if g.Battlefield.Contains(bird) {
		t.Error("the 3-toughness flyer survived X = 3")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the ground creature died")
	}
	if !effectstest.AlliTapped(t, g, arashi) {
		t.Error("Arashi did not tap")
	}
}

func TestOldB058ArashiChannelDamagesEveryFlyerButNotGroundCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mineFlyer := oldB058Flyer(g, me.ID, 2)
	theirFlyer := oldB058Flyer(g, opp.ID, 2)
	bigFlyer := oldB058Flyer(g, opp.ID, 5)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogHandCard(me, "Arashi, the Sky Asunder", "Legendary Creature — Spirit", oldB058ArashiOracle)
	effectstest.FloatMana(t, g, me, "{C}{C}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{Strict: true, XValue: 2}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mineFlyer) || g.Battlefield.Contains(theirFlyer) {
		t.Error("a 2-toughness flyer survived X = 2")
	}
	if !g.Battlefield.Contains(bigFlyer) || !g.Battlefield.Contains(bear) {
		t.Error("a creature that should survive died")
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the discarded Arashi is not in the graveyard")
	}
}

// --- Teysa, Envoy of Ghosts ----------------------------------------------

func TestOldB058TeysaDestroysACreatureThatHitsHerControllerAndMakesASpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	teysa := effectstest.B12Push(g, opp.ID, "Teysa, Envoy of Ghosts", "Legendary Creature — Human Advisor", oldB058TeysaOracle, 4, 4)
	for _, kw := range []string{"vigilance", "protection from creatures"} {
		if !effectstest.HasEffectiveKeyword(t, g, teysa, kw) {
			t.Errorf("Teysa has no %s", kw)
		}
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	effectstest.AttackInto(t, g, map[uuid.UUID]uuid.UUID{bear: opp.ID})
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature that dealt combat damage was not destroyed")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("a creature that dealt no damage was destroyed")
	}
	if n := effectstest.B27CountNamed(g, opp.ID, "Spirit"); n != 1 {
		t.Errorf("%d Spirit tokens, want 1", n)
	}
	if opp.Life != 40-2 {
		t.Errorf("life %d, want 38", opp.Life)
	}
}

func TestOldB058TeysaDoesNotTriggerOnDamageToAnotherPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.B12Push(g, opp.ID, "Teysa, Envoy of Ghosts", "Legendary Creature — Human Advisor", oldB058TeysaOracle, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AttackInto(t, g, map[uuid.UUID]uuid.UUID{bear: third.ID})
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a creature that hit another player was destroyed")
	}
}

// --- Battlemage's Bracers -------------------------------------------------

func TestOldB058BracersGiveHasteAndCopyAnAbilityForOneMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bracers := effectstest.B12Push(g, me.ID, "Battlemage's Bracers", "Artifact — Equipment", oldB058BracersOracle, 0, 0)
	arashi := effectstest.B12Push(g, me.ID, "Arashi, the Sky Asunder", "Legendary Creature — Spirit", oldB058ArashiOracle, 5, 5)
	bird := oldB058Flyer(g, opp.ID, 3)
	bird2 := oldB058Flyer(g, opp.ID, 3)
	effectstest.AdvanceToMain(t, g)
	if effectstest.HasEffectiveKeyword(t, g, arashi, "haste") {
		t.Fatal("Arashi has haste before it is equipped")
	}
	effectstest.AttachForTest(t, g, bracers, arashi)
	if !effectstest.HasEffectiveKeyword(t, g, arashi, "haste") {
		t.Error("the equipped creature has no haste")
	}
	effectstest.ApaMana(me, "G", "G", "G", "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, arashi, 0, game.ActivateAbilityParams{Strict: true, XValue: 3, Targets: effectstest.CardRefs(bird)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pay := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless)
	if pay == nil {
		t.Fatal("the Bracers did not offer the {1}")
	}
	effectstest.ApaMana(me, "R")
	if err := g.ResolvePayUnless(pay.ID, me.ID, true); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bird2}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bird) || g.Battlefield.Contains(bird2) {
		t.Error("the ability and its copy should have killed both flyers")
	}
}

func TestOldB058BracersCopyNeedsThePayment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bracers := effectstest.B12Push(g, me.ID, "Battlemage's Bracers", "Artifact — Equipment", oldB058BracersOracle, 0, 0)
	arashi := effectstest.B12Push(g, me.ID, "Arashi, the Sky Asunder", "Legendary Creature — Spirit", oldB058ArashiOracle, 5, 5)
	bird := oldB058Flyer(g, opp.ID, 3)
	bird2 := oldB058Flyer(g, opp.ID, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.AttachForTest(t, g, bracers, arashi)
	effectstest.ApaMana(me, "G", "G", "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, arashi, 0, game.ActivateAbilityParams{Strict: true, XValue: 3, Targets: effectstest.CardRefs(bird)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pay := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless)
	if pay == nil {
		t.Fatal("the Bracers did not offer the {1}")
	}
	if err := g.ResolvePayUnless(pay.ID, me.ID, false); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bird) {
		t.Error("the original ability did not resolve")
	}
	if !g.Battlefield.Contains(bird2) {
		t.Error("a copy was made with no mana to pay for it")
	}
}

// --- Reinterpret ----------------------------------------------------------

func TestOldB058ReinterpretCountersAndOffersACheaperSpellFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	cheap := oldB058Hand(me, "Cheap Creature", "Creature — Bear", "{1}{G}", "")
	equal := oldB058Hand(me, "Equal Sorcery", "Sorcery", "{2}{R}", "")
	dear := oldB058Hand(me, "Dear Creature", "Creature — Bear", "{3}{G}", "")
	land := oldB058Hand(me, "Forest", "Basic Land — Forest", "", "")
	effectstest.AdvanceToMain(t, g)
	// An opponent's three-mana spell is on the stack.
	theirs := uuid.New()
	g.WithWriteLock(func() {
		opp.Hand.PushTop(game.Card{InstanceID: theirs, Name: "Their Sorcery", TypeLine: "Instant", ManaCost: "{2}{B}", Owner: opp.ID, Controller: opp.ID})
	})
	if err := g.CastSpell(opp.ID, theirs, game.CastSpellParams{}); err != nil {
		t.Fatalf("their cast: %v", err)
	}
	id := oldB058Hand(me, "Reinterpret", "Instant", "", oldB058ReinterpretOrc)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(theirs)}); err != nil {
		t.Fatalf("cast Reinterpret: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(theirs) {
		t.Error("the targeted spell was not countered")
	}
	offer := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, cheap, equal)
	for _, c := range offer.DiscardOptions {
		if c == dear || c == land {
			t.Errorf("a spell with a greater mana value, or a land, is offered: %v", c)
		}
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, equal, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("casting it free: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.ManaPool.Total() != 0 {
		t.Error("the cast was not free")
	}
	if !me.Graveyard.Contains(equal) {
		t.Error("the cast sorcery did not resolve")
	}
}

func TestOldB058ReinterpretOffersNothingWhenTheTargetIsGone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cheap := oldB058Hand(me, "Cheap Creature", "Creature — Bear", "{1}", "")
	effectstest.AdvanceToMain(t, g)
	theirs := uuid.New()
	g.WithWriteLock(func() {
		opp.Hand.PushTop(game.Card{InstanceID: theirs, Name: "Their Instant", TypeLine: "Instant", ManaCost: "{2}", Owner: opp.ID, Controller: opp.ID})
	})
	if err := g.CastSpell(opp.ID, theirs, game.CastSpellParams{}); err != nil {
		t.Fatalf("their cast: %v", err)
	}
	id := oldB058Hand(me, "Reinterpret", "Instant", "", oldB058ReinterpretOrc)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(theirs)}); err != nil {
		t.Fatalf("cast Reinterpret: %v", err)
	}
	g.WithWriteLock(func() { _ = g.CounterKeyedForEffect(theirs, game.KeyedTail{}) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil || !me.Hand.Contains(cheap) {
		t.Error("an offer was made though the target was already gone")
	}
}

// --- Fiery Encore ---------------------------------------------------------

func TestOldB058FieryEncoreRummagesAndDamagesForTheDiscardedCardsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	victim := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Ogre", TypeLine: "Creature — Ogre",
		Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 4)
	three := oldB058Hand(me, "Three Drop", "Sorcery", "{1}{R}{R}", "")
	effectstest.AdvanceToMain(t, g)
	id := oldB058Hand(me, "Fiery Encore", "Sorcery", "", oldB058EncoreOracle)
	g.WithWriteLock(func() {
		me.Hand.Cards = me.Hand.Cards[:0]
		me.Hand.PushTop(game.Card{InstanceID: three, Name: "Three Drop", TypeLine: "Sorcery", ManaCost: "{1}{R}{R}", Owner: me.ID, Controller: me.ID})
	})
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Fiery Encore", TypeLine: "Sorcery", OracleID: oldB058EncoreOracle, Owner: me.ID, Controller: me.ID})
	})
	handBefore := me.Hand.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, three)
	if !me.Graveyard.Contains(three) {
		t.Fatal("the discarded card is not in the graveyard")
	}
	if me.Hand.Size() != handBefore-1 {
		t.Errorf("hand %d, want %d (the spell left, one discarded, one drawn)", me.Hand.Size(), handBefore-1)
	}
	p := effectstest.PassUntilPickTarget(t, g, me.ID)
	if p == nil {
		t.Fatal("no target was asked for the reflexive damage")
	}
	for _, c := range p.PickTargetCards {
		if c == id {
			t.Error("the spell itself is a candidate")
		}
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: victim}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the 3-toughness creature survived 3 damage")
	}
	if !g.Battlefield.Contains(bystander) {
		t.Error("a creature that was not the target died")
	}
}

func TestOldB058FieryEncoreDiscardingALandDealsNoDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 1)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	land := oldB058Hand(me, "Forest", "Basic Land — Forest", "", "")
	id := oldB058Hand(me, "Fiery Encore", "Sorcery", "", oldB058EncoreOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, land)
	if p := effectstest.PassUntilPickTarget(t, g, me.ID); p != nil {
		t.Error("a land discard created the damage trigger")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Error("the creature took damage from a land discard")
	}
}

func TestOldB058FieryEncoreStormCopiesEachRummage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastFromHandForTest(t, g, me, "Earlier Spell", "Instant", "", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	a := oldB058Hand(me, "Land A", "Basic Land — Forest", "", "")
	b := oldB058Hand(me, "Land B", "Basic Land — Forest", "", "")
	id := oldB058Hand(me, "Fiery Encore", "Sorcery", "", oldB058EncoreOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// One prior spell: the copy and the original each ask for a discard.
	effectstest.AnswerDiscard(t, g, me.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, b)
	if !me.Graveyard.Contains(a) || !me.Graveyard.Contains(b) {
		t.Error("the copy and the original did not each discard")
	}
}

// --- Muse Vortex ----------------------------------------------------------

func oldB058MuseTable(t *testing.T) (*game.Game, *game.Player, []game.Card) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cards := []game.Card{
		effectstest.HobB10Card(me, "Cheap Sorcery", "Sorcery", "{1}{R}"),
		effectstest.HobB10Card(me, "Big Instant", "Instant", "{4}{R}"),
		effectstest.HobB10Card(me, "A Creature", "Creature — Bear", "{1}{G}"),
		effectstest.HobB10Card(me, "A Forest", "Basic Land — Forest", ""),
		effectstest.HobB10Card(me, "Fifth Card", "Sorcery", "{R}"),
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.HobB10Stack(me, cards...)
	return g, me, cards
}

func TestOldB058MuseVortexCastsOneThenSortsTheRest(t *testing.T) {
	g, me, c := oldB058MuseTable(t)
	cheap, big, bear, forest, fifth := c[0].InstanceID, c[1].InstanceID, c[2].InstanceID, c[3].InstanceID, c[4].InstanceID
	effectstest.CastFromHandForTest(t, g, me, "Muse Vortex", "Sorcery", "{X}{U}{U}", oldB058MuseOracle, game.CastSpellParams{XValue: 4})
	effectstest.PassPriorityAroundTable(t, g)
	offer := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, cheap)
	for _, id := range offer.DiscardOptions {
		if id == big || id == bear || id == forest || id == fifth {
			t.Errorf("a card that cannot be cast this way is offered: %v", id)
		}
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, cheap, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("casting it free: %v", err)
	}
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		effectstest.DeclineResolutionCast(t, g, me.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(cheap) {
		t.Error("the cast sorcery did not resolve")
	}
	if !me.Hand.Contains(big) {
		t.Error("the instant that was not cast did not go to the hand")
	}
	if me.Hand.Contains(bear) || me.Hand.Contains(forest) {
		t.Error("a creature or land went to the hand")
	}
	n := me.Library.Size()
	if !me.Library.Contains(bear) || !me.Library.Contains(forest) {
		t.Fatal("the creature and the land are not on the library")
	}
	if me.Library.Cards[n-1].InstanceID != fifth {
		t.Error("the fifth card is no longer on top")
	}
	bottom := map[uuid.UUID]bool{me.Library.Cards[0].InstanceID: true, me.Library.Cards[1].InstanceID: true}
	if !bottom[bear] || !bottom[forest] {
		t.Error("the creature and land are not the bottom two cards")
	}
	if g.Exile.Contains(bear) || g.Exile.Contains(big) {
		t.Error("an exiled card was left in exile")
	}
}

func TestOldB058MuseVortexDecliningPutsInstantsAndSorceriesInHand(t *testing.T) {
	g, me, c := oldB058MuseTable(t)
	cheap, big, bear := c[0].InstanceID, c[1].InstanceID, c[2].InstanceID
	effectstest.CastFromHandForTest(t, g, me, "Muse Vortex", "Sorcery", "{X}{U}{U}", oldB058MuseOracle, game.CastSpellParams{XValue: 3})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, cheap)
	effectstest.DeclineResolutionCast(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(cheap) || !me.Hand.Contains(big) {
		t.Error("the instant and sorcery cards are not in the hand")
	}
	if !me.Library.Contains(bear) || g.Exile.Contains(bear) {
		t.Error("the creature is not on the library")
	}
}

func TestOldB058MuseVortexWithNothingCastableStillSorts(t *testing.T) {
	g, me, c := oldB058MuseTable(t)
	big, bear := c[1].InstanceID, c[2].InstanceID
	g.WithWriteLock(func() {
		effectstest.HobB10Stack(me, c[1], c[2], c[3])
	})
	effectstest.CastFromHandForTest(t, g, me, "Muse Vortex", "Sorcery", "{X}{U}{U}", oldB058MuseOracle, game.CastSpellParams{XValue: 3})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		t.Fatal("an offer was made though nothing could be cast")
	}
	if !me.Hand.Contains(big) || !me.Library.Contains(bear) {
		t.Error("the exiled cards were not sorted")
	}
}

// --- Author of Shadows ----------------------------------------------------

func oldB058Graveyards(g *game.Game) (me, opp, third *game.Player, spell, land, creature, mine uuid.UUID) {
	me, opp, third = g.Seats[0], g.Seats[1], g.Seats[2]
	spell, land, creature = uuid.New(), uuid.New(), uuid.New()
	mine = uuid.New()
	g.WithWriteLock(func() {
		opp.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Their Sorcery", TypeLine: "Sorcery", ManaCost: "{2}{B}{B}", Owner: opp.ID, Controller: opp.ID})
		opp.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Their Swamp", TypeLine: "Basic Land — Swamp", Owner: opp.ID, Controller: opp.ID})
		third.Graveyard.PushTop(game.Card{InstanceID: creature, Name: "Their Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: third.ID, Controller: third.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: mine, Name: "My Card", TypeLine: "Sorcery", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	})
	return
}

func TestOldB058AuthorOfShadowsExilesOpponentsGraveyardsAndLetsYouCastOneAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, third, spell, land, creature, mine := oldB058Graveyards(g)
	_ = opp
	_ = third
	effectstest.CastFromHandForTest(t, g, me, "Author of Shadows", "Creature — Shade Warlock", "{4}{B}", oldB058AuthorOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if len(opp.Graveyard.Cards) != 0 || len(third.Graveyard.Cards) != 0 {
		t.Error("an opponent's graveyard was not exiled")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("your own graveyard was exiled")
	}
	for _, id := range []uuid.UUID{spell, land, creature} {
		if !g.Exile.Contains(id) {
			t.Errorf("card %s is not in exile", id)
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, spell)
	effectstest.PassPriorityAroundTable(t, g)
	// Red mana pays for the {B} part of the chosen sorcery: any color.
	effectstest.B06AddMana(me, "R", "R", "R", "R")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, FromZone: string(game.ZoneExile)}); err != nil {
		t.Fatalf("casting the chosen card with off-color mana: %v", err)
	}
	// The cards not chosen carry no permission.
	effectstest.B06AddMana(me, "G", "G")
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{Strict: true, FromZone: string(game.ZoneExile)}); err == nil {
		t.Error("cast a card that was not chosen")
	}
}
