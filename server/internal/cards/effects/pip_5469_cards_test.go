package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pip_5469_cards_test.go — Fallout, batch 1 (#5469, S82): fifteen PIP
// cards built from seams the engine already has. Each test plays the
// real catalog entry.

const (
	pipb1AlmostPerfectOracle          = "617eae63-b5f8-4014-8269-7c66f4388ac7"
	pipb1IntelligenceBobbleheadOracle = "16e6bd4b-fbf1-4bbe-9271-a74f59320e0d"
	pipb1StrengthBobbleheadOracle     = "bd0e22d1-4789-406d-9293-3ade30365b32"
	pipb1CharismaBobbleheadOracle     = "77897962-7d67-4fa8-858f-2f35a322d7c6"
	pipb1SuperMutantScavengerOracle   = "2972b9c8-d9b3-4bb3-b9c5-9aa70fb1ac71"
	pipb1PaladinDanseOracle           = "ac830812-2f33-4f9e-ba3a-ea665989011c"
	pipb1AtomizeOracle                = "4b457bc2-2d8c-42cb-9a63-bb4c459d14e3"
	pipb1RadstormOracle               = "d4cde3d5-eb13-4ad5-8178-08780dc85d39"
	pipb1BreakDownOracle              = "73094bde-5466-41a4-88f9-a98e67f5ed2e"
	pipb1PowderGangerOracle           = "b34507d3-55ba-4629-9581-7983d0025e89"
	pipb1WastelandRaiderOracle        = "29439669-c9be-46f3-a62f-16de51fefe88"
	pipb1SecuritronSquadronOracle     = "96312b75-3110-4506-9f95-0c6c17d9da11"
	pipb1JunktownOracle               = "c470a802-931f-4b63-92fd-ef9cf2e796dd"
	pipb1BighornerRancherOracle       = "826e6a6f-3b66-4225-87ee-04e345e047b3"
	pipb1Vault75Oracle                = "e81fa87f-80ce-4db6-940e-d30309bc70e6"
)

func TestPipb1CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		pipb1AlmostPerfectOracle:          "Almost Perfect",
		pipb1IntelligenceBobbleheadOracle: "Intelligence Bobblehead",
		pipb1StrengthBobbleheadOracle:     "Strength Bobblehead",
		pipb1CharismaBobbleheadOracle:     "Charisma Bobblehead",
		pipb1SuperMutantScavengerOracle:   "Super Mutant Scavenger",
		pipb1PaladinDanseOracle:           "Paladin Danse, Steel Maverick",
		pipb1AtomizeOracle:                "Atomize",
		pipb1RadstormOracle:               "Radstorm",
		pipb1BreakDownOracle:              "Break Down",
		pipb1PowderGangerOracle:           "Powder Ganger",
		pipb1WastelandRaiderOracle:        "Wasteland Raider",
		pipb1SecuritronSquadronOracle:     "Securitron Squadron",
		pipb1JunktownOracle:               "Junktown",
		pipb1BighornerRancherOracle:       "Bighorner Rancher",
		pipb1Vault75Oracle:                "Vault 75: Middle School",
	} {
		def, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Name != name || def.Completeness != CompletenessFull || len(def.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, def.Name, def.Completeness, def.Caveats)
		}
	}
}

// pipb1Bobblehead puts a catalog Bobblehead onto the battlefield.
func pipb1Bobblehead(g *game.Game, owner uuid.UUID, name, oracle string) uuid.UUID {
	return pushCatalogPermanent(g, owner, name, "Artifact — Bobblehead", oracle, false)
}

// pipb1Activate activates a catalog permanent's own ability with the
// mana gate on.
func pipb1Activate(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, index int, mana string, params game.ActivateAbilityParams) {
	t.Helper()
	if mana != "" {
		floatMana(t, g, p, mana)
	}
	params.Strict = true
	if err := g.ActivateCatalogAbility(p.ID, id, index, params); err != nil {
		t.Fatalf("activate: %v", err)
	}
}

// --- Almost Perfect -----------------------------------------------------------

func TestPipb1AlmostPerfectSetsBase9By10AndGrantsIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Almost Perfect", "Enchantment — Aura", pipb1AlmostPerfectOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, bear); p != 9 || tt != 10 {
		t.Fatalf("enchanted Bear is %d/%d, want 9/10", p, tt)
	}
	assertKeywords(t, g, bear, "indestructible")
	if err := g.AddCounter(bear, game.CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}
	if p, tt := ptOf(t, g, bear); p != 10 || tt != 11 {
		t.Errorf("with a +1/+1 counter: %d/%d, want 10/11 (the counter applies after the base)", p, tt)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	if !g.Battlefield.Contains(bear) {
		t.Error("the enchanted creature was destroyed while indestructible")
	}
}

// --- Bobbleheads ---------------------------------------------------------------

func TestPipb1IntelligenceBobbleheadDrawsOnePerBobblehead(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bob := pipb1Bobblehead(g, me.ID, "Intelligence Bobblehead", pipb1IntelligenceBobbleheadOracle)
	pipb1Bobblehead(g, me.ID, "Charisma Bobblehead", pipb1CharismaBobbleheadOracle)
	b12Permanent(g, me.ID, "Mind Stone", "Artifact")
	pipb1Bobblehead(g, g.Seats[1].ID, "Strength Bobblehead", pipb1StrengthBobbleheadOracle)
	hand := me.Hand.Size()
	pipb1Activate(t, g, me, bob, 0, "{C}{C}{C}{C}{C}", game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards with two Bobbleheads (and a non-Bobblehead artifact, and an opponent's), want 2", got)
	}
}

func TestPipb1StrengthBobbleheadPutsXCountersAsASorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bob := pipb1Bobblehead(g, me.ID, "Strength Bobblehead", pipb1StrengthBobbleheadOracle)
	pipb1Bobblehead(g, me.ID, "Intelligence Bobblehead", pipb1IntelligenceBobbleheadOracle)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	// The draw step is not a main phase: activate only as a sorcery.
	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, bob, 0, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err == nil {
		t.Fatal("Strength Bobblehead was activated outside a main phase")
	}

	advanceToMain(t, g)
	pipb1Activate(t, g, me, bob, 0, "{C}{C}{C}", game.ActivateAbilityParams{Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bear, game.CounterPlusOne); n != 2 {
		t.Errorf("the target has %d +1/+1 counters with two Bobbleheads, want 2", n)
	}
}

func TestPipb1CharismaBobbleheadMakesASoldierPerBobblehead(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bob := pipb1Bobblehead(g, me.ID, "Charisma Bobblehead", pipb1CharismaBobbleheadOracle)
	pipb1Bobblehead(g, me.ID, "Strength Bobblehead", pipb1StrengthBobbleheadOracle)
	pipb1Bobblehead(g, me.ID, "Intelligence Bobblehead", pipb1IntelligenceBobbleheadOracle)
	pipb1Activate(t, g, me, bob, 0, "{C}{C}{C}{C}", game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if n := b28TokensNamed(g, me.ID, "Soldier"); n != 3 {
		t.Errorf("%d Soldier tokens with three Bobbleheads, want 3", n)
	}
}

// --- Super Mutant Scavenger ---------------------------------------------------

func TestPipb1SuperMutantScavengerReturnsAnAuraOrEquipmentOnEnterAndDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := pushGraveyardCardTyped(me, "Old Aura", "Enchantment — Aura")
	gear := pushGraveyardCardTyped(me, "Old Gear", "Artifact — Equipment")
	body := pushGraveyardCardTyped(me, "Old Body", "Creature — Bear")
	theirs := pushGraveyardCardTyped(opp, "Their Aura", "Enchantment — Aura")

	scav := enterFromHand(t, g, me.ID, "Super Mutant Scavenger", "Creature — Mutant Warrior", pipb1SuperMutantScavengerOracle)
	assertKeywords(t, g, scav, "trample")
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt as the Scavenger entered")
	}
	if !hasID(pick.PickTargetCards, aura) || !hasID(pick.PickTargetCards, gear) {
		t.Errorf("an Aura or Equipment card in my graveyard is not offered: %v", pick.PickTargetCards)
	}
	if hasID(pick.PickTargetCards, body) || hasID(pick.PickTargetCards, theirs) {
		t.Errorf("a creature card or an opponent's Aura is offered: %v", pick.PickTargetCards)
	}
	pickCard(t, g, me.ID, aura)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(aura) {
		t.Fatal("the Aura was not returned as the Scavenger entered")
	}

	killCreature(t, g, me.ID, scav)
	if latestPickTarget(g, me.ID) == nil {
		t.Fatal("no target prompt as the Scavenger died")
	}
	pickCard(t, g, me.ID, gear)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(gear) {
		t.Error("the Equipment was not returned as the Scavenger died")
	}
	if me.Hand.Contains(body) {
		t.Error("the creature card was returned")
	}
}

// --- Paladin Danse, Steel Maverick ---------------------------------------------

func TestPipb1PaladinDanseExilesItselfToProtectArtifactsAndHumans(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	danse := b12Push(g, me.ID, "Paladin Danse, Steel Maverick", "Legendary Artifact Creature — Synth Knight", pipb1PaladinDanseOracle, 3, 3)
	assertKeywords(t, g, danse, "vigilance", "lifelink")
	human := b12Creature(g, me.ID, "My Human", "Creature — Human Soldier", 2, 2)
	robot := b12Creature(g, me.ID, "My Robot", "Artifact Creature — Robot", 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	theirHuman := b12Creature(g, opp.ID, "Their Human", "Creature — Human Soldier", 2, 2)

	pipb1Activate(t, g, me, danse, 0, "", game.ActivateAbilityParams{})
	if g.Battlefield.Contains(danse) || !g.Exile.Contains(danse) {
		t.Fatal("Paladin Danse was not exiled as the cost")
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{human, robot} {
		if !effectiveAbilitiesContain(t, g, id, "indestructible") {
			t.Errorf("%s (an artifact or Human creature of mine) did not gain indestructible", id)
		}
	}
	for _, id := range []uuid.UUID{bear, rock, theirHuman} {
		if effectiveAbilitiesContain(t, g, id, "indestructible") {
			t.Errorf("%s gained indestructible: a non-Human non-artifact, a noncreature artifact, or an opponent's", id)
		}
	}
}

// --- Atomize --------------------------------------------------------------------

func TestPipb1AtomizeDestroysANonlandPermanentThenProliferates(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	relic := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	land := b12Permanent(g, opp.ID, "Their Land", "Land")
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := g.AddCounter(mine, game.CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}
	if err := castCatalogSpellErr(t, g, "Atomize", "Instant", pipb1AtomizeOracle, cardRefs(land)); err == nil {
		t.Fatal("Atomize was cast at a land")
	}
	castCatalogSpell(t, g, "Atomize", "Instant", pipb1AtomizeOracle, cardRefs(relic))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Error("the target was not destroyed")
	}
	if proliferatePrompt(g) == nil {
		t.Fatal("no proliferate prompt after the destroy")
	}
	answerProliferate(t, g, mine)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, mine, game.CounterPlusOne); n != 2 {
		t.Errorf("my creature has %d +1/+1 counters after the proliferate, want 2", n)
	}
}

// --- Radstorm ---------------------------------------------------------------------

func TestPipb1RadstormProliferatesOncePerStormCopy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := g.AddCounter(mine, game.CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}
	castCatalogSpell(t, g, "Earlier Spell", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Radstorm", "Instant", pipb1RadstormOracle, nil)
	settleAnsweringProliferate(t, g)
	// One spell before it: one copy and the original, two proliferates.
	if n := countersOn(g, mine, game.CounterPlusOne); n != 3 {
		t.Errorf("my creature has %d +1/+1 counters, want 3 (1 + two proliferates)", n)
	}
}

// --- Break Down -------------------------------------------------------------------

func TestPipb1BreakDownDestroysAnArtifactOrEnchantmentAndMakesJunk(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := b12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Break Down", "Instant", pipb1BreakDownOracle, cardRefs(bear)); err == nil {
		t.Fatal("Break Down was cast at a creature")
	}
	castCatalogSpell(t, g, "Break Down", "Instant", pipb1BreakDownOracle, cardRefs(aura))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the enchantment was not destroyed")
	}
	if n := b28TokensNamed(g, me.ID, "Junk"); n != 1 {
		t.Errorf("%d Junk tokens, want 1", n)
	}
}

// --- Powder Ganger ----------------------------------------------------------------

func TestPipb1PowderGangerDestroysUpToOneArtifactAsItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	enchantment := b12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	enterFromHand(t, g, me.ID, "Powder Ganger", "Creature — Human Rogue", pipb1PowderGangerOracle)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt as Powder Ganger entered")
	}
	if hasID(pick.PickTargetCards, enchantment) {
		t.Error("an enchantment is offered as an artifact")
	}
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the artifact was not destroyed")
	}
}

// --- Wasteland Raider ---------------------------------------------------------------

func TestPipb1WastelandRaiderMakesEachPlayerSacrificeACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := seedCreature(g, "Bear", opp.ID)
	raider := enterFromHand(t, g, me.ID, "Wasteland Raider", "Creature — Human Mercenary", pipb1WastelandRaiderOracle)
	if triggerOnStack(g, raider) == nil {
		t.Fatal("no enters trigger on the stack")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the opponent did not sacrifice their only creature")
	}
	if g.Battlefield.Contains(raider) {
		t.Error("each player includes the Raider's controller: it was its only creature")
	}
}

// --- Securitron Squadron ------------------------------------------------------------

func TestPipb1SecuritronSquadronCountersYourCreatureTokens(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	squadron := b12Push(g, me.ID, "Securitron Squadron", "Artifact Creature — Robot", pipb1SecuritronSquadronOracle, 2, 2)
	assertKeywords(t, g, squadron, "vigilance")
	var soldier, junk, rat uuid.UUID
	var err error
	g.WithWriteLock(func() {
		var ids []uuid.UUID
		if ids, err = g.CreateTokensForEffect(me.ID, TokenCard("1/1 white Soldier"), 1, game.TokenEntryOptions{}); err != nil || len(ids) != 1 {
			return
		}
		soldier = ids[0]
		if ids, err = g.CreateTokensForEffect(me.ID, JunkToken(), 1, game.TokenEntryOptions{}); err != nil || len(ids) != 1 {
			return
		}
		junk = ids[0]
		if ids, err = g.CreateTokensForEffect(opp.ID, TokenCard("1/1 black Rat"), 1, game.TokenEntryOptions{}); err != nil || len(ids) != 1 {
			return
		}
		rat = ids[0]
	})
	if err != nil || soldier == uuid.Nil || junk == uuid.Nil || rat == uuid.Nil {
		t.Fatalf("create tokens: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, soldier, game.CounterPlusOne); n != 1 {
		t.Errorf("my Soldier token has %d +1/+1 counters, want 1", n)
	}
	if n := countersOn(g, junk, game.CounterPlusOne); n != 0 {
		t.Errorf("a noncreature token got %d counters", n)
	}
	if n := countersOn(g, rat, game.CounterPlusOne); n != 0 {
		t.Errorf("an opponent's creature token got %d counters", n)
	}
	if n := countersOn(g, squadron, game.CounterPlusOne); n != 0 {
		t.Errorf("the nontoken Squadron got %d counters", n)
	}
}

func TestPipb1SecuritronSquadronSquadCopiesCountThemselves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	id := tmtB11Hand(g, me, "Securitron Squadron", "Artifact Creature — Robot", pipb1SecuritronSquadronOracle, "{1}{W}", 2, 2)
	floatMana(t, g, me, "{W}{C}{C}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("cast with squad paid once: %v", err)
	}
	// The card's trigger and the copy's own are both mine: I order them
	// (CR 603.3b).
	settleOrderingTriggers(t, g)
	ids := tmtB1Named(g, "Securitron Squadron")
	if len(ids) != 2 {
		t.Fatalf("%d Securitron Squadrons, want the card and one token copy", len(ids))
	}
	tokens := 0
	for _, sid := range ids {
		want := 0
		if IsToken(battlefieldCardCopy(t, g, sid)) {
			tokens++
			// The copy is a creature token entering under my control: the
			// card sees it, and so does the copy itself.
			want = 2
		}
		if n := countersOn(g, sid, game.CounterPlusOne); n != want {
			t.Errorf("Squadron %s has %d +1/+1 counters, want %d", sid, n, want)
		}
	}
	if tokens != 1 {
		t.Errorf("%d token Squadrons, want 1", tokens)
	}
}

// --- Junktown -------------------------------------------------------------------------

func TestPipb1JunktownSacrificesForThreeJunk(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	town := pushCatalogPermanent(g, me.ID, "Junktown", "Land", pipb1JunktownOracle, false)
	pipb1Activate(t, g, me, town, 0, "{R}{C}{C}{C}{C}", game.ActivateAbilityParams{})
	if g.Battlefield.Contains(town) {
		t.Fatal("Junktown was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if n := b28TokensNamed(g, me.ID, "Junk"); n != 3 {
		t.Errorf("%d Junk tokens, want 3", n)
	}
}

// --- Bighorner Rancher ------------------------------------------------------------------

func TestPipb1BighornerRancherTapsForGreatestPowerAndSacrificesForGreatestToughness(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rancher := b12Push(g, me.ID, "Bighorner Rancher", "Creature — Human Ranger", pipb1BighornerRancherOracle, 2, 5)
	assertKeywords(t, g, rancher, "vigilance")
	b12Creature(g, me.ID, "My Brute", "Creature — Ogre", 4, 3)
	b12Creature(g, me.ID, "My Wall", "Creature — Wall", 1, 4)
	b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 7, 7)

	if err := g.ActivateManaAbility(me.ID, rancher, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolCount(me, "G"); got != 4 {
		t.Errorf("added %d {G}, want 4 (the greatest power among my creatures, not the opponent's 7)", got)
	}

	life := me.Life
	pipb1Activate(t, g, me, rancher, 0, "", game.ActivateAbilityParams{})
	if g.Battlefield.Contains(rancher) {
		t.Fatal("the Rancher was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if got := me.Life - life; got != 4 {
		t.Errorf("gained %d life, want 4 (the greatest toughness among the other creatures, not the Rancher's 5)", got)
	}
}

// --- Vault 75: Middle School --------------------------------------------------------------

func TestPipb1Vault75ExilesTheBigThenCountersYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myBig := b12Creature(g, me.ID, "My Big", "Creature — Beast", 4, 4)
	mySmall := b12Creature(g, me.ID, "My Small", "Creature — Bear", 2, 2)
	theirBig := b12Creature(g, opp.ID, "Their Big", "Creature — Beast", 5, 5)
	theirSmall := b12Creature(g, opp.ID, "Their Small", "Creature — Bear", 3, 3)

	saga := castCatalogSpell(t, g, "Vault 75: Middle School", "Enchantment — Saga", pipb1Vault75Oracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{myBig, theirBig} {
		if g.Battlefield.Contains(id) || !g.Exile.Contains(id) {
			t.Errorf("a creature with power 4 or greater (%s) was not exiled by chapter I", id)
		}
	}
	for _, id := range []uuid.UUID{mySmall, theirSmall} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("a creature with power less than 4 (%s) was exiled", id)
		}
	}

	g.WithWriteLock(func() { _ = g.AddCounterForEffect(saga, game.CounterLore, 1) })
	passPriorityAroundTable(t, g)
	if n := countersOn(g, mySmall, game.CounterPlusOne); n != 1 {
		t.Errorf("chapter II: my creature has %d +1/+1 counters, want 1", n)
	}
	if n := countersOn(g, theirSmall, game.CounterPlusOne); n != 0 {
		t.Errorf("chapter II put %d counters on an opponent's creature", n)
	}

	g.WithWriteLock(func() { _ = g.AddCounterForEffect(saga, game.CounterLore, 1) })
	passPriorityAroundTable(t, g)
	if n := countersOn(g, mySmall, game.CounterPlusOne); n != 2 {
		t.Errorf("chapter III: my creature has %d +1/+1 counters, want 2", n)
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}
