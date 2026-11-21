package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s82g57DogUmbra        = "e475c96f-6cf8-4139-92fd-f68ecaae4893"
	s82g57Insight         = "81a398ee-e534-43b7-88fe-45155d6aba3a"
	s82g57FlareMalice     = "4ce156f8-6b0b-41ce-9e8f-b2735c757406"
	s82g57WitherBloom     = "cf1e1d8d-c02e-4b66-b1db-49ec48fa4b03"
	s82g57PyreticRebirth  = "a38ace5b-34f5-4105-b6ee-dedb86037729"
	s82g57Abstruse        = "61a23402-3244-4b5e-9134-98e1f82eda65"
	s82g57SignatureSlam   = "5ae23752-47a6-4031-8398-b8f6c77cfaad"
	s82g57PlanarGenesis   = "b368121b-9984-4a01-8d35-d1e4a8a87539"
	s82g57FlareDuplicate  = "f586ecb8-310c-4f40-a766-6a7ca46a254d"
	s82g57SiegeSmash      = "4e07d30a-d29a-49bc-ab49-41ea4116a18c"
	s82g57FlareDenial     = "437cfd16-94bc-43ba-afd8-e9e9d028a4ca"
	s82g57InstantTypeLine = "Instant"
)

func TestS82GapB57EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82g57DogUmbra:       "Dog Umbra",
		s82g57Insight:        "Eviscerator's Insight",
		s82g57FlareMalice:    "Flare of Malice",
		s82g57WitherBloom:    "Wither and Bloom",
		s82g57PyreticRebirth: "Pyretic Rebirth",
		s82g57Abstruse:       "Abstruse Appropriation",
		s82g57SignatureSlam:  "Signature Slam",
		s82g57PlanarGenesis:  "Planar Genesis",
		s82g57FlareDuplicate: "Flare of Duplication",
		s82g57SiegeSmash:     "Siege Smash",
		s82g57FlareDenial:    "Flare of Denial",
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

// --- Dog Umbra -----------------------------------------------------

func TestS82GapB57DogUmbraHasUmbraArmorOnYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	aura := effectstest.PushAuraOnLand(t, g, me.ID, "Dog Umbra", s82g57DogUmbra, bear)
	if got := effectstest.RestrictionsOf(t, g, bear); got&game.CantAttackOrBlock != 0 {
		t.Errorf("my own creature is restricted: %v", got.Names())
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatal(err)
		}
	})
	if !g.Battlefield.Contains(bear) {
		t.Fatal("umbra armor should save my creature from a destroy")
	}
	if !me.Graveyard.Contains(aura) {
		t.Error("Dog Umbra should be destroyed instead")
	}
}

func TestS82GapB57DogUmbraStopsAnotherPlayersCreatureAndHasNoArmorThere(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	aura := effectstest.PushAuraOnLand(t, g, me.ID, "Dog Umbra", s82g57DogUmbra, ogre)
	effectstest.AssertRestrictions(t, g, ogre, game.CantAttackOrBlock)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(ogre); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(ogre) {
		t.Fatal("no umbra armor on another player's creature: it is destroyed")
	}
	if !me.Graveyard.Contains(aura) {
		t.Error("the Aura goes to its owner's graveyard with the dead creature")
	}
}

func TestS82GapB57DogUmbraFlipsWhenControlChanges(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.PushAuraOnLand(t, g, me.ID, "Dog Umbra", s82g57DogUmbra, bear)
	g.WithWriteLock(func() { g.GainControlForEffect(uuid.Nil, bear, opp.ID, game.IndefiniteDuration(), "theft") })
	effectstest.AssertRestrictions(t, g, bear, game.CantAttackOrBlock)
	g.WithWriteLock(func() { g.GainControlForEffect(uuid.Nil, bear, me.ID, game.IndefiniteDuration(), "back") })
	if got := effectstest.RestrictionsOf(t, g, bear); got&game.CantAttackOrBlock != 0 {
		t.Errorf("control returned to the Aura's controller but the creature is still restricted: %v", got.Names())
	}
}

// --- Eviscerator's Insight -----------------------------------------

func s82g57Library(me *game.Player, n int) {
	for i := 0; i < n; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
}

func TestS82GapB57EvisceratorsInsightSacrificesDrawsAndFlashesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	s82g57Library(me, 6)
	victim := effectstest.SeedCreature(g, "Doomed", me.ID)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	land := effectstest.SeedLand(g, me.ID, "Forest", "Basic Land — Forest", "")

	if _, err := effectstest.TryCastWithSacrifice(g, "Eviscerator's Insight", s82g57InstantTypeLine, s82g57Insight, nil); err == nil {
		t.Fatal("cast without the sacrifice")
	}
	if _, err := effectstest.TryCastWithSacrifice(g, "Eviscerator's Insight", s82g57InstantTypeLine, s82g57Insight, []uuid.UUID{land}); err == nil {
		t.Fatal("a land paid \"sacrifice an artifact or creature\"")
	}
	before := me.Hand.Size()
	id := effectstest.CastWithSacrifice(t, g, "Eviscerator's Insight", s82g57InstantTypeLine, s82g57Insight, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature was not sacrificed")
	}
	if got := me.Hand.Size() - before; got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
	if !me.Graveyard.Contains(id) {
		t.Fatal("Insight is not in the graveyard")
	}
	// Flashback pays the additional cost again, here with an artifact.
	before = me.Hand.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", SacrificeIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed to the flashback cast")
	}
	if got := me.Hand.Size() - before; got != 2 {
		t.Errorf("flashback drew %d cards, want 2", got)
	}
	if !g.Exile.Contains(id) {
		t.Error("the flashed-back Insight is not exiled")
	}
}

// --- Flare of Malice -----------------------------------------------

func TestS82GapB57FlareOfMaliceSacrificesBlackCreatureAndEdictsTheGreatest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	flare := effectstest.HandCardAtMain(t, g, "Flare of Malice", s82g57InstantTypeLine, s82g57FlareMalice)
	black := effectstest.SeedColoredCreature(g, me.ID, "Zombie", []string{"B"}, 0)
	green := effectstest.SeedColoredCreature(g, me.ID, "Elf", []string{"G"}, 0)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Zombie", TypeLine: "Token Creature — Zombie",
		Colors: []string{"B"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	big := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big", TypeLine: "Creature — Giant",
		ManaCost: "{4}{R}", Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID})
	small := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Small", TypeLine: "Creature — Rat",
		ManaCost: "{B}", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})

	for name, id := range map[string]uuid.UUID{"a green creature": green, "a black token": token} {
		if err := effectstest.CastForAltCost(g, flare, "", "sacrifice", []uuid.UUID{id}, nil); err == nil {
			t.Fatalf("%s paid \"sacrifice a nontoken black creature\"", name)
		}
	}
	if err := effectstest.CastForAltCost(g, flare, "", "sacrifice", []uuid.UUID{black}, nil); err != nil {
		t.Fatalf("Flare for a nontoken black creature: %v", err)
	}
	if g.Battlefield.Contains(black) {
		t.Error("the black creature was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the opponent's greatest-mana-value creature survived")
	}
	if !g.Battlefield.Contains(small) {
		t.Error("the opponent's smaller creature was sacrificed too")
	}
}

// --- Flare of Denial -----------------------------------------------

func TestS82GapB57FlareOfDenialCountersASpellForABlueCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	blue := effectstest.SeedColoredCreature(g, me.ID, "Merfolk", []string{"U"}, 0)
	red := effectstest.SeedColoredCreature(g, me.ID, "Goblin", []string{"R"}, 0)
	oppLife := g.Seats[2].Life
	spell := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[2].ID}})
	flare := effectstest.HandCardAtMain(t, g, "Flare of Denial", s82g57InstantTypeLine, s82g57FlareDenial)
	at := []game.TargetRef{{Kind: game.TargetCard, ID: spell}}
	if err := effectstest.CastForAltCost(g, flare, "", "sacrifice", []uuid.UUID{red}, at); err == nil {
		t.Fatal("a red creature paid \"sacrifice a nontoken blue creature\"")
	}
	if err := effectstest.CastForAltCost(g, flare, "", "sacrifice", []uuid.UUID{blue}, at); err != nil {
		t.Fatalf("Flare of Denial: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(blue) {
		t.Error("the blue creature was not sacrificed")
	}
	if g.Seats[2].Life != oppLife {
		t.Errorf("the countered Bolt still dealt damage: %d -> %d", oppLife, g.Seats[2].Life)
	}
	if !opp.Graveyard.Contains(spell) {
		t.Error("the countered Bolt is not in its owner's graveyard")
	}
}

// --- Flare of Duplication ------------------------------------------

func TestS82GapB57FlareOfDuplicationCopiesAnInstant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	red := effectstest.SeedColoredCreature(g, me.ID, "Goblin", []string{"R"}, 0)
	blue := effectstest.SeedColoredCreature(g, me.ID, "Merfolk", []string{"U"}, 0)
	victim := g.Seats[2]
	spell := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	flare := effectstest.HandCardAtMain(t, g, "Flare of Duplication", s82g57InstantTypeLine, s82g57FlareDuplicate)
	at := []game.TargetRef{{Kind: game.TargetCard, ID: spell}}
	if err := effectstest.CastForAltCost(g, flare, "", "sacrifice", []uuid.UUID{blue}, at); err == nil {
		t.Fatal("a blue creature paid \"sacrifice a nontoken red creature\"")
	}
	other := g.Seats[3]
	before, otherBefore := victim.Life, other.Life
	if err := effectstest.CastForAltCost(g, flare, "", "sacrifice", []uuid.UUID{red}, at); err != nil {
		t.Fatalf("Flare of Duplication: %v", err)
	}
	for i := 0; i < 8 && effectstest.CopyPromptFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	p := effectstest.CopyPromptFor(g, me.ID)
	if p == nil {
		t.Fatal("the copy opened no new-targets prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: other.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(red) {
		t.Error("the red creature was not sacrificed")
	}
	if victim.Life != before-3 || other.Life != otherBefore-3 {
		t.Errorf("the Bolt and its re-aimed copy: victim %d -> %d, other %d -> %d, want 3 each", before, victim.Life, otherBefore, other.Life)
	}
}

// --- Siege Smash ---------------------------------------------------

func TestS82GapB57SiegeSmashBothModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spec, _ := effects.Lookup(s82g57SiegeSmash)
	split := false
	for _, k := range spec.PrintedKeywords {
		split = split || k == game.KeywordSplitSecond
	}
	if !split {
		t.Error("Siege Smash does not print split second")
	}
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.D1658Cast(t, g, "Siege Smash", s82g57InstantTypeLine, s82g57SiegeSmash,
		game.CastSpellParams{Modes: []int{0}, Targets: effectstest.CardRefs(rock)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived the first mode")
	}
	effectstest.D1658Cast(t, g, "Siege Smash", s82g57InstantTypeLine, s82g57SiegeSmash,
		game.CastSpellParams{Modes: []int{1}, Targets: effectstest.CardRefs(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 5 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 5/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the Bear did not gain trample")
	}
}

func TestS82GapB57SiegeSmashRefusesACreatureForTheArtifactMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.D1658CastErr(t, g, "Siege Smash", s82g57InstantTypeLine, s82g57SiegeSmash,
		game.CastSpellParams{Modes: []int{0}, Targets: effectstest.CardRefs(bear), Strict: true}); err == nil {
		t.Error("a non-artifact was accepted for \"destroy target artifact\"")
	}
}

// --- Signature Slam ------------------------------------------------

func TestS82GapB57SignatureSlamCountsEveryModifiedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	effectstest.SeedColoredCreature(g, me.ID, "Counter Bear", nil, 1) // modified already, power 3
	enchanted := effectstest.B12Creature(g, me.ID, "Aura Bear", "Creature — Bear", 4, 4)
	effectstest.PushAuraOnLand(t, g, me.ID, "Pacifism-like", "", enchanted)
	theirs := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 20)
	effectstest.SeedColoredCreature(g, opp.ID, "Their Counter Bear", nil, 1) // modified, but not mine: deals no damage
	effectstest.D1658Cast(t, g, "Signature Slam", s82g57InstantTypeLine, s82g57SignatureSlam,
		game.CastSpellParams{Targets: effectstest.CardRefs(plain, theirs)})
	effectstest.PassPriorityAroundTable(t, g)
	// plain gains a counter (3 power), counter bear 3, aura bear 4 = 10.
	if got := effectstest.DamageMarkedOn(g, theirs); got != 10 {
		t.Errorf("the Wall took %d, want 10 (3 + 3 + 4)", got)
	}
}

func TestS82GapB57SignatureSlamRefusesTheWrongSides(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	for name, targets := range map[string][]game.TargetRef{
		"two of mine":       effectstest.CardRefs(mine, other),
		"theirs then mine":  effectstest.CardRefs(theirs, mine),
		"theirs and theirs": effectstest.CardRefs(theirs, theirs),
	} {
		if err := effectstest.D1658CastErr(t, g, "Signature Slam", s82g57InstantTypeLine, s82g57SignatureSlam,
			game.CastSpellParams{Targets: targets, Strict: true}); err == nil {
			t.Errorf("%s was accepted", name)
		}
	}
}

// --- Pyretic Rebirth -----------------------------------------------

func TestS82GapB57PyreticRebirthReturnsAndDealsItsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Fatty", TypeLine: "Creature — Giant", ManaCost: "{3}{R}",
		Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 20)
	effectstest.D1658Cast(t, g, "Pyretic Rebirth", s82g57InstantTypeLine, s82g57PyreticRebirth,
		game.CastSpellParams{Targets: effectstest.CardRefs(dead, wall)})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Error("the creature card did not return to hand")
	}
	if got := effectstest.DamageMarkedOn(g, wall); got != 4 {
		t.Errorf("the Wall took %d, want 4 (the card's mana value)", got)
	}
}

func TestS82GapB57PyreticRebirthWithNoRecipientOnlyReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Rock", TypeLine: "Artifact", ManaCost: "{2}", Owner: me.ID, Controller: me.ID})
	effectstest.D1658Cast(t, g, "Pyretic Rebirth", s82g57InstantTypeLine, s82g57PyreticRebirth,
		game.CastSpellParams{Targets: effectstest.CardRefs(dead)})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Error("the artifact card did not return to hand")
	}
}

func TestS82GapB57PyreticRebirthRefusesALandCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	if err := effectstest.D1658CastErr(t, g, "Pyretic Rebirth", s82g57InstantTypeLine, s82g57PyreticRebirth,
		game.CastSpellParams{Targets: effectstest.CardRefs(land), Strict: true}); err == nil {
		t.Error("a land card was accepted for \"artifact or creature card\"")
	}
}

// --- Wither and Bloom ----------------------------------------------

func TestS82GapB57WitherAndBloomShrinksThenPutsACounterFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	id := effectstest.D1658Cast(t, g, "Wither and Bloom", s82g57InstantTypeLine, s82g57WitherBloom,
		game.CastSpellParams{Targets: effectstest.CardRefs(ogre)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Fatal("a 3/3 survived -3/-3")
	}
	if !me.Graveyard.Contains(id) {
		t.Fatal("Wither and Bloom is not in the graveyard")
	}
	// Their creature is not a legal target for the counter.
	theirs := effectstest.B12Creature(g, opp.ID, "Rat", "Creature — Rat", 1, 1)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{B}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs), Strict: true}); err == nil {
		t.Fatal("the counter ability targeted a creature I do not control")
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(mine), Strict: true}); err != nil {
		t.Fatalf("graveyard ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Error("the card was not exiled as a cost")
	}
	if n := effectstest.CardByID(g, mine).Counters[game.CounterPlusOne]; n != 1 {
		t.Errorf("the Bear has %d +1/+1 counters, want 1", n)
	}
}

// --- Abstruse Appropriation ----------------------------------------

func TestS82GapB57AbstruseAppropriationExilesAndLetsColorlessPayColoredCosts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Elf", TypeLine: "Creature — Elf",
		ManaCost: "{1}{G}", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	land := effectstest.SeedLand(g, opp.ID, "Forest", "Basic Land — Forest", "")
	if err := effectstest.D1658CastErr(t, g, "Abstruse Appropriation", s82g57InstantTypeLine, s82g57Abstruse,
		game.CastSpellParams{Targets: effectstest.CardRefs(land), Strict: true}); err == nil {
		t.Fatal("a land was accepted for \"nonland permanent\"")
	}
	effectstest.D1658Cast(t, g, "Abstruse Appropriation", s82g57InstantTypeLine, s82g57Abstruse,
		game.CastSpellParams{Targets: effectstest.CardRefs(elf)})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(elf) {
		t.Fatal("the permanent was not exiled")
	}
	// Red mana cannot stand in for {G}: only colorless can.
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}{R}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, elf, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Fatal("red mana paid {G}")
	}
	me.ManaPool.EmptyPool()
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, elf, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("casting the exiled card with colorless mana: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(elf) {
		t.Error("the stolen Elf did not enter the battlefield")
	}
}

// --- Planar Genesis ------------------------------------------------

func TestS82GapB57PlanarGenesisLandTappedOrCardToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	lib := effectstest.DftMPLibraryTop(me, effectstest.C2b15Bear(), effectstest.C2b15Forest(), effectstest.C2b15Spell(), effectstest.C2b15Bear(), effectstest.C2b15Bear())
	effectstest.D1658Cast(t, g, "Planar Genesis", s82g57InstantTypeLine, s82g57PlanarGenesis, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if ch == nil || ch.ChooseMin != 0 || ch.ChooseMax != 1 {
		t.Fatalf("want an optional pick of one land, got %+v", ch)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, lib[1])
	if !g.Battlefield.Contains(lib[1]) || !effectstest.C2b16IsTapped(g, lib[1]) {
		t.Error("the land is not on the battlefield tapped")
	}
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a land was put, so no card goes to hand")
	}
	for _, i := range []int{0, 2, 3} {
		if !me.Library.Contains(lib[i]) {
			t.Errorf("card %d is not back in the library", i)
		}
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top != lib[4] {
		t.Error("the fifth card is not on top: the looked-at cards went to the bottom")
	}
}

func TestS82GapB57PlanarGenesisDecliningTheLandTakesACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	lib := effectstest.DftMPLibraryTop(me, effectstest.C2b15Bear(), effectstest.C2b15Forest(), effectstest.C2b15Spell(), effectstest.C2b15Bear())
	effectstest.D1658Cast(t, g, "Planar Genesis", s82g57InstantTypeLine, s82g57PlanarGenesis, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID) // decline the land
	ch := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if ch == nil || ch.ChooseMin != 1 || ch.ChooseMax != 1 {
		t.Fatalf("want a mandatory pick of one card, got %+v", ch)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, lib[2])
	if !me.Hand.Contains(lib[2]) {
		t.Error("the chosen card is not in hand")
	}
	if g.Battlefield.Contains(lib[1]) {
		t.Error("the declined land entered the battlefield")
	}
	for _, i := range []int{0, 1, 3} {
		if !me.Library.Contains(lib[i]) {
			t.Errorf("card %d is not back in the library", i)
		}
	}
}

func TestS82GapB57PlanarGenesisWithNoLandTakesACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	lib := effectstest.DftMPLibraryTop(me, effectstest.C2b15Bear(), effectstest.C2b15Spell(), effectstest.C2b15Bear(), effectstest.C2b15Bear())
	effectstest.D1658Cast(t, g, "Planar Genesis", s82g57InstantTypeLine, s82g57PlanarGenesis, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, lib[0])
	if !me.Hand.Contains(lib[0]) {
		t.Error("the chosen card is not in hand")
	}
}
