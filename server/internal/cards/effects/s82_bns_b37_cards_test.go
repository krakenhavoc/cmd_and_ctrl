package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s82BnsRecklessScholarOracle   = "604aded2-64c2-4474-b5b7-8c83e0e3fe76"
	s82BnsReliefCaptainOracle     = "97e088a3-65eb-42f5-85b8-6aeb770a3f56"
	s82BnsRuthlessKnaveOracle     = "00225c40-27ff-410f-a591-5c788c6a2bd6"
	s82BnsSandsteppeOutcastOracle = "fccd17b1-ca0f-4649-a002-5cbf9666b510"
	s82BnsSatyrHopliteOracle      = "ee7dc820-10ee-429a-a1ef-ecd6a11f5608"
	s82BnsScholarCombustionOracle = "ebae61bf-6e8c-4cc9-8071-b63e781f6530"
	s82BnsScionDarknessOracle     = "bd940cab-d23a-44bf-ab96-8aa1dcd3f0ff"
	s82BnsScionSwarmOracle        = "ccaeabe1-42fa-4f29-bf80-d9b73498eec6"
	s82BnsScythecatCubOracle      = "929117cd-caa6-402b-9d7b-1f684d75695c"
	s82BnsSerpentAssassinOracle   = "3bc7d3a7-ddb4-4eaa-882e-404d8f2926fb"
	s82BnsServantScaleOracle      = "7d404420-e3de-47a8-a159-2f634a73f2c9"
)

func TestS82BnsB37EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82BnsRecklessScholarOracle:   "Reckless Scholar",
		s82BnsReliefCaptainOracle:     "Relief Captain",
		s82BnsRuthlessKnaveOracle:     "Ruthless Knave",
		s82BnsSandsteppeOutcastOracle: "Sandsteppe Outcast",
		s82BnsSatyrHopliteOracle:      "Satyr Hoplite",
		s82BnsScholarCombustionOracle: "Scholar of Combustion",
		s82BnsScionDarknessOracle:     "Scion of Darkness",
		s82BnsScionSwarmOracle:        "Scion of the Swarm",
		s82BnsScythecatCubOracle:      "Scythecat Cub",
		s82BnsSerpentAssassinOracle:   "Serpent Assassin",
		s82BnsServantScaleOracle:      "Servant of the Scale",
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

func TestS82BnsB37RecklessScholarLootsTheTargetPlayer(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, _ := blbB1Seats(g)
	scholar := pushCatalogPermanent(g, me.ID, "Reckless Scholar", "Creature — Human Wizard", s82BnsRecklessScholarOracle, true)
	if err := g.ActivateCatalogAbility(me.ID, scholar, 0, game.ActivateAbilityParams{Targets: acPlayerRef(opp.ID)}); err == nil {
		t.Error("a summoning-sick Scholar activated its {T} ability")
	}
	scholar = pushCatalogPermanent(g, me.ID, "Reckless Scholar", "Creature — Human Wizard", s82BnsRecklessScholarOracle, false)
	before := opp.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, scholar, 0, game.ActivateAbilityParams{Targets: acPlayerRef(opp.ID)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Hand.Size() < before {
		t.Errorf("hand %d → %d: the draw must come before the discard", before, opp.Hand.Size())
	}
}

func TestS82BnsB37ReliefCaptainSupportsOtherCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	captain := castAndResolveCreature(t, g, "Relief Captain", "Creature — Kor Knight Ally", s82BnsReliefCaptainOracle)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no support target prompt")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, append(tdCard(a), tdCard(b)...)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if plusOnes(g, a) != 1 || plusOnes(g, b) != 1 {
		t.Errorf("counters a=%d b=%d, want 1 each", plusOnes(g, a), plusOnes(g, b))
	}
	if plusOnes(g, captain) != 0 {
		t.Error("support put a counter on the Captain itself")
	}
}

func TestS82BnsB37RuthlessKnaveMakesTreasureAndDrawsOffThree(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	knave := pushCatalogPermanent(g, me.ID, "Ruthless Knave", "Creature — Orc Pirate", s82BnsRuthlessKnaveOracle, false)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, knave, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err == nil {
		t.Fatal("activated without paying {2}{B}")
	}
	floatForTest(g, me, "CCB")
	if err := g.ActivateCatalogAbility(me.ID, knave, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed")
	}
	if n := tokensNamedControlledBy(g, "Treasure", me.ID); n != 2 {
		t.Fatalf("%d Treasures, want 2", n)
	}
	var treasures []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Treasure" && c.Controller == me.ID {
			treasures = append(treasures, c.InstanceID)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, knave, 1, game.ActivateAbilityParams{SacrificeIDs: treasures, Strict: true}); err == nil {
		t.Fatal("two Treasures paid for \"sacrifice three\"")
	}
	extra := b12Push(g, me.ID, "Treasure", "Token Artifact — Treasure", "", 0, 0)
	treasures = append(treasures, extra)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, knave, 1, game.ActivateAbilityParams{SacrificeIDs: treasures, Strict: true}); err != nil {
		t.Fatalf("sacrifice three: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, me.Hand.Size())
	}
}

func TestS82BnsB37SandsteppeOutcastModes(t *testing.T) {
	t.Run("counter", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		outcast := castRingCreature(t, g, "Sandsteppe Outcast", "Creature — Human Warrior", s82BnsSandsteppeOutcastOracle, 2, 1)
		passPriorityAroundTable(t, g)
		pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if pick == nil {
			t.Fatal("no mode prompt")
		}
		if err := g.ResolveModePick(pick.ID, me.ID, []int{0}); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
		if plusOnes(g, outcast) != 1 {
			t.Errorf("%d counters, want 1", plusOnes(g, outcast))
		}
		if n := tokensNamedControlledBy(g, "Spirit", me.ID); n != 0 {
			t.Errorf("%d Spirits, want 0", n)
		}
	})
	t.Run("spirit", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		outcast := castRingCreature(t, g, "Sandsteppe Outcast", "Creature — Human Warrior", s82BnsSandsteppeOutcastOracle, 2, 1)
		passPriorityAroundTable(t, g)
		pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if pick == nil {
			t.Fatal("no mode prompt")
		}
		if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
		if plusOnes(g, outcast) != 0 {
			t.Errorf("%d counters, want 0", plusOnes(g, outcast))
		}
		if n := tokensNamedControlledBy(g, "Spirit", me.ID); n != 1 {
			t.Fatalf("%d Spirits, want 1", n)
		}
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Spirit" && IsToken(c) && (c.Power != 1 || c.Toughness != 1) {
				t.Errorf("Spirit is %d/%d", c.Power, c.Toughness)
			}
		}
	})
}

func TestS82BnsB37SatyrHopliteHeroicOnlyOnItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hoplite := pushCatalogPermanent(g, me.ID, "Satyr Hoplite", "Creature — Satyr Soldier", s82BnsSatyrHopliteOracle, false)
	other := pr7Creature(g, me.ID, "Other", 2, "W")
	castCatalogSpell(t, g, "Indestructible Aura", "Instant", pr7aIndestructibleAura, []game.TargetRef{pr7aCard(other)})
	passPriorityAroundTable(t, g)
	if plusOnes(g, hoplite) != 0 {
		t.Fatal("heroic triggered for a spell targeting another creature")
	}
	castCatalogSpell(t, g, "Djeru's Resolve", "Instant", pr7aDjerusResolve, []game.TargetRef{pr7aCard(hoplite)})
	passPriorityAroundTable(t, g)
	if plusOnes(g, hoplite) != 1 {
		t.Errorf("%d counters, want 1", plusOnes(g, hoplite))
	}
}

func TestS82BnsB37ScholarOfCombustionExilesAnInstantForACast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	bolt := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bolt, Name: "Shock", TypeLine: "Instant", ManaCost: "{R}", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	castAndResolveCreature(t, g, "Scholar of Combustion", "Creature — Human Wizard", s82BnsScholarCombustionOracle)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
			t.Fatal("a creature card was a legal target")
		}
		pickCard(t, g, me.ID, bolt)
	} else {
		t.Fatal("no target prompt")
	}
	passPriorityAroundTable(t, g)
	if !inExile(g, bolt) {
		t.Fatal("the instant was not exiled")
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the creature card left the graveyard")
	}
	floatForTest(g, me, "R")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
}

func TestS82BnsB37ScionOfDarknessReanimatesFromTheDamagedPlayersGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scion := b43Catalog(g, me.ID, "Scion of Darkness", "Creature — Avatar", s82BnsScionDarknessOracle, 6, 6)
	victim := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: victim, Name: "Dead Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	mine := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: mine, Name: "My Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	dealCombatDamageToPlayer(g, scion, opp.ID, 6)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, victim)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Fatal("the creature card did not come back")
	}
	if got := controllerOf(t, g, victim); got != me.ID {
		t.Errorf("it came back under %s, want your control", got)
	}
	if g.Battlefield.Contains(mine) {
		t.Error("your own graveyard card came back")
	}
	if spec, _ := Lookup(s82BnsScionDarknessOracle); len(spec.Activated) != 1 || !spec.Activated[0].Cycling {
		t.Error("Scion of Darkness has no cycling ability")
	}
}

func TestS82BnsB37ScionOfDarknessDeclinedLeavesTheGraveyardAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scion := b43Catalog(g, me.ID, "Scion of Darkness", "Creature — Avatar", s82BnsScionDarknessOracle, 6, 6)
	victim := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: victim, Name: "Dead Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	dealCombatDamageToPlayer(g, scion, opp.ID, 6)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, victim)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("a declined reanimation happened")
	}
}

func TestS82BnsB37ScionOfTheSwarmGrowsOncePerLifeGainEvent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	scion := pushCatalogPermanent(g, me.ID, "Scion of the Swarm", "Creature — Vampire Cleric", s82BnsScionSwarmOracle, false)
	for i := 0; i < 2; i++ {
		g.WithWriteLock(func() {
			if err := g.ChangePlayerLifeForEffect(scion, me.ID, 3); err != nil {
				t.Fatal(err)
			}
		})
		g.RunStateChecksForTest()
		passPriorityAroundTable(t, g)
	}
	if plusOnes(g, scion) != 2 {
		t.Errorf("%d counters after two gains, want 2 (one per event)", plusOnes(g, scion))
	}
}

func TestS82BnsB37ScythecatCubCountersThenDoubles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cub := pushCatalogPermanent(g, me.ID, "Scythecat Cub", "Creature — Cat", s82BnsScythecatCubOracle, false)
	ally := b12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	counters := func() int { return plusOnes(g, ally) }
	g.WithWriteLock(func() {
		findBattlefieldCardForTest(g, ally).Counters = map[string]int{game.CounterPlusOne: 2}
	})
	playLandFromHand(t, g, "Forest", "")
	answerPickTargetIfOpen(t, g, ally)
	passPriorityAroundTable(t, g)
	if counters() != 3 {
		t.Fatalf("after the first landfall %d counters, want 3", counters())
	}
	playLandFromHand(t, g, "Forest", "")
	answerPickTargetIfOpen(t, g, ally)
	passPriorityAroundTable(t, g)
	if counters() != 6 {
		t.Fatalf("after the second landfall %d counters, want 6 (doubled)", counters())
	}
	playLandFromHand(t, g, "Forest", "")
	answerPickTargetIfOpen(t, g, ally)
	passPriorityAroundTable(t, g)
	if counters() != 7 {
		t.Errorf("after the third landfall %d counters, want 7", counters())
	}
	_ = cub
}

func TestS82BnsB37SerpentAssassinMayDestroyANonblackCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	black := usgB3Colored(g, opp.ID, "Black Beast", "B")
	green := usgB3Colored(g, opp.ID, "Green Beast", "G")
	castAndResolveCreature(t, g, "Serpent Assassin", "Creature — Snake Assassin", s82BnsSerpentAssassinOracle)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: black}); err == nil {
		t.Fatal("the trigger targeted a black creature")
	}
	pickTriggerTarget(t, g, me.ID, green)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green) || !g.Battlefield.Contains(black) {
		t.Errorf("green on battlefield %v, black %v", g.Battlefield.Contains(green), g.Battlefield.Contains(black))
	}
}

func TestS82BnsB37SerpentAssassinDeclinedDestroysNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	green := usgB3Colored(g, opp.ID, "Green Beast", "G")
	castAndResolveCreature(t, g, "Serpent Assassin", "Creature — Snake Assassin", s82BnsSerpentAssassinOracle)
	pickTriggerTarget(t, g, me.ID, green)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(green) {
		t.Error("a declined destroy still destroyed")
	}
}

func TestS82BnsB37ServantOfTheScaleEntersWithACounterAndPassesItsCountersOn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ally := b12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	servant := castAndResolveCreature(t, g, "Servant of the Scale", "Creature — Human Soldier", s82BnsServantScaleOracle)
	if plusOnes(g, servant) != 1 {
		t.Fatalf("Servant has %d counters, want 1", plusOnes(g, servant))
	}
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, servant).Counters = map[string]int{game.CounterPlusOne: 3} })
	if err := g.DestroyPermanentForEffect(servant); err != nil {
		t.Fatal(err)
	}
	g.RunStateChecksForTest()
	pickTriggerTarget(t, g, me.ID, ally)
	passPriorityAroundTable(t, g)
	if plusOnes(g, ally) != 3 {
		t.Errorf("ally has %d counters, want 3", plusOnes(g, ally))
	}
}
