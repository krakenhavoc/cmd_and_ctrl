package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

func TestS82BnsB37RecklessScholarLootsTheTargetPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	scholar := effectstest.PushCatalogPermanent(g, me.ID, "Reckless Scholar", "Creature — Human Wizard", s82BnsRecklessScholarOracle, true)
	if err := g.ActivateCatalogAbility(me.ID, scholar, 0, game.ActivateAbilityParams{Targets: effectstest.AcPlayerRef(opp.ID)}); err == nil {
		t.Error("a summoning-sick Scholar activated its {T} ability")
	}
	scholar = effectstest.PushCatalogPermanent(g, me.ID, "Reckless Scholar", "Creature — Human Wizard", s82BnsRecklessScholarOracle, false)
	before := opp.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, scholar, 0, game.ActivateAbilityParams{Targets: effectstest.AcPlayerRef(opp.ID)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() < before {
		t.Errorf("hand %d → %d: the draw must come before the discard", before, opp.Hand.Size())
	}
}

func TestS82BnsB37ReliefCaptainSupportsOtherCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	captain := effectstest.CastAndResolveCreature(t, g, "Relief Captain", "Creature — Kor Knight Ally", s82BnsReliefCaptainOracle)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no support target prompt")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, append(effectstest.TdCard(a), effectstest.TdCard(b)...)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, a) != 1 || effectstest.PlusOnes(g, b) != 1 {
		t.Errorf("counters a=%d b=%d, want 1 each", effectstest.PlusOnes(g, a), effectstest.PlusOnes(g, b))
	}
	if effectstest.PlusOnes(g, captain) != 0 {
		t.Error("support put a counter on the Captain itself")
	}
}

func TestS82BnsB37RuthlessKnaveMakesTreasureAndDrawsOffThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	knave := effectstest.PushCatalogPermanent(g, me.ID, "Ruthless Knave", "Creature — Orc Pirate", s82BnsRuthlessKnaveOracle, false)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, knave, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err == nil {
		t.Fatal("activated without paying {2}{B}")
	}
	effectstest.FloatForTest(g, me, "CCB")
	if err := g.ActivateCatalogAbility(me.ID, knave, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed")
	}
	if n := effectstest.TokensNamedControlledBy(g, "Treasure", me.ID); n != 2 {
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
	extra := effectstest.B12Push(g, me.ID, "Treasure", "Token Artifact — Treasure", "", 0, 0)
	treasures = append(treasures, extra)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, knave, 1, game.ActivateAbilityParams{SacrificeIDs: treasures, Strict: true}); err != nil {
		t.Fatalf("sacrifice three: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, me.Hand.Size())
	}
}

func TestS82BnsB37SandsteppeOutcastModes(t *testing.T) {
	t.Run("counter", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		outcast := effectstest.CastRingCreature(t, g, "Sandsteppe Outcast", "Creature — Human Warrior", s82BnsSandsteppeOutcastOracle, 2, 1)
		effectstest.PassPriorityAroundTable(t, g)
		pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if pick == nil {
			t.Fatal("no mode prompt")
		}
		if err := g.ResolveModePick(pick.ID, me.ID, []int{0}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.PlusOnes(g, outcast) != 1 {
			t.Errorf("%d counters, want 1", effectstest.PlusOnes(g, outcast))
		}
		if n := effectstest.TokensNamedControlledBy(g, "Spirit", me.ID); n != 0 {
			t.Errorf("%d Spirits, want 0", n)
		}
	})
	t.Run("spirit", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		outcast := effectstest.CastRingCreature(t, g, "Sandsteppe Outcast", "Creature — Human Warrior", s82BnsSandsteppeOutcastOracle, 2, 1)
		effectstest.PassPriorityAroundTable(t, g)
		pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if pick == nil {
			t.Fatal("no mode prompt")
		}
		if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.PlusOnes(g, outcast) != 0 {
			t.Errorf("%d counters, want 0", effectstest.PlusOnes(g, outcast))
		}
		if n := effectstest.TokensNamedControlledBy(g, "Spirit", me.ID); n != 1 {
			t.Fatalf("%d Spirits, want 1", n)
		}
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Spirit" && effects.IsToken(c) && (c.Power != 1 || c.Toughness != 1) {
				t.Errorf("Spirit is %d/%d", c.Power, c.Toughness)
			}
		}
	})
}

func TestS82BnsB37SatyrHopliteHeroicOnlyOnItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hoplite := effectstest.PushCatalogPermanent(g, me.ID, "Satyr Hoplite", "Creature — Satyr Soldier", s82BnsSatyrHopliteOracle, false)
	other := effectstest.Pr7Creature(g, me.ID, "Other", 2, "W")
	effectstest.CastCatalogSpell(t, g, "Indestructible Aura", "Instant", effectstest.Pr7aIndestructibleAura, []game.TargetRef{effectstest.Pr7aCard(other)})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, hoplite) != 0 {
		t.Fatal("heroic triggered for a spell targeting another creature")
	}
	effectstest.CastCatalogSpell(t, g, "Djeru's Resolve", "Instant", effectstest.Pr7aDjerusResolve, []game.TargetRef{effectstest.Pr7aCard(hoplite)})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, hoplite) != 1 {
		t.Errorf("%d counters, want 1", effectstest.PlusOnes(g, hoplite))
	}
}

func TestS82BnsB37ScholarOfCombustionExilesAnInstantForACast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	bolt := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bolt, Name: "Shock", TypeLine: "Instant", ManaCost: "{R}", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	effectstest.CastAndResolveCreature(t, g, "Scholar of Combustion", "Creature — Human Wizard", s82BnsScholarCombustionOracle)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
			t.Fatal("a creature card was a legal target")
		}
		effectstest.PickCard(t, g, me.ID, bolt)
	} else {
		t.Fatal("no target prompt")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, bolt) {
		t.Fatal("the instant was not exiled")
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the creature card left the graveyard")
	}
	effectstest.FloatForTest(g, me, "R")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
}

func TestS82BnsB37ScionOfDarknessReanimatesFromTheDamagedPlayersGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scion := effectstest.B43Catalog(g, me.ID, "Scion of Darkness", "Creature — Avatar", s82BnsScionDarknessOracle, 6, 6)
	victim := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: victim, Name: "Dead Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	mine := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: mine, Name: "My Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.DealCombatDamageToPlayer(g, scion, opp.ID, 6)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Fatal("the creature card did not come back")
	}
	if got := effectstest.ControllerOf(t, g, victim); got != me.ID {
		t.Errorf("it came back under %s, want your control", got)
	}
	if g.Battlefield.Contains(mine) {
		t.Error("your own graveyard card came back")
	}
	if spec, _ := effects.Lookup(s82BnsScionDarknessOracle); len(spec.Activated) != 1 || !spec.Activated[0].Cycling {
		t.Error("Scion of Darkness has no cycling ability")
	}
}

func TestS82BnsB37ScionOfDarknessDeclinedLeavesTheGraveyardAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scion := effectstest.B43Catalog(g, me.ID, "Scion of Darkness", "Creature — Avatar", s82BnsScionDarknessOracle, 6, 6)
	victim := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: victim, Name: "Dead Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	effectstest.DealCombatDamageToPlayer(g, scion, opp.ID, 6)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("a declined reanimation happened")
	}
}

func TestS82BnsB37ScionOfTheSwarmGrowsOncePerLifeGainEvent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	scion := effectstest.PushCatalogPermanent(g, me.ID, "Scion of the Swarm", "Creature — Vampire Cleric", s82BnsScionSwarmOracle, false)
	for i := 0; i < 2; i++ {
		g.WithWriteLock(func() {
			if err := g.ChangePlayerLifeForEffect(scion, me.ID, 3); err != nil {
				t.Fatal(err)
			}
		})
		g.RunStateChecksForTest()
		effectstest.PassPriorityAroundTable(t, g)
	}
	if effectstest.PlusOnes(g, scion) != 2 {
		t.Errorf("%d counters after two gains, want 2 (one per event)", effectstest.PlusOnes(g, scion))
	}
}

func TestS82BnsB37ScythecatCubCountersThenDoubles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cub := effectstest.PushCatalogPermanent(g, me.ID, "Scythecat Cub", "Creature — Cat", s82BnsScythecatCubOracle, false)
	ally := effectstest.B12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	counters := func() int { return effectstest.PlusOnes(g, ally) }
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, ally).Counters = map[string]int{game.CounterPlusOne: 2}
	})
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.AnswerPickTargetIfOpen(t, g, ally)
	effectstest.PassPriorityAroundTable(t, g)
	if counters() != 3 {
		t.Fatalf("after the first landfall %d counters, want 3", counters())
	}
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.AnswerPickTargetIfOpen(t, g, ally)
	effectstest.PassPriorityAroundTable(t, g)
	if counters() != 6 {
		t.Fatalf("after the second landfall %d counters, want 6 (doubled)", counters())
	}
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.AnswerPickTargetIfOpen(t, g, ally)
	effectstest.PassPriorityAroundTable(t, g)
	if counters() != 7 {
		t.Errorf("after the third landfall %d counters, want 7", counters())
	}
	_ = cub
}

func TestS82BnsB37SerpentAssassinMayDestroyANonblackCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	black := effectstest.UsgB3Colored(g, opp.ID, "Black Beast", "B")
	green := effectstest.UsgB3Colored(g, opp.ID, "Green Beast", "G")
	effectstest.CastAndResolveCreature(t, g, "Serpent Assassin", "Creature — Snake Assassin", s82BnsSerpentAssassinOracle)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: black}); err == nil {
		t.Fatal("the trigger targeted a black creature")
	}
	effectstest.PickTriggerTarget(t, g, me.ID, green)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green) || !g.Battlefield.Contains(black) {
		t.Errorf("green on battlefield %v, black %v", g.Battlefield.Contains(green), g.Battlefield.Contains(black))
	}
}

func TestS82BnsB37SerpentAssassinDeclinedDestroysNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	green := effectstest.UsgB3Colored(g, opp.ID, "Green Beast", "G")
	effectstest.CastAndResolveCreature(t, g, "Serpent Assassin", "Creature — Snake Assassin", s82BnsSerpentAssassinOracle)
	effectstest.PickTriggerTarget(t, g, me.ID, green)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(green) {
		t.Error("a declined destroy still destroyed")
	}
}

func TestS82BnsB37ServantOfTheScaleEntersWithACounterAndPassesItsCountersOn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ally := effectstest.B12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	servant := effectstest.CastAndResolveCreature(t, g, "Servant of the Scale", "Creature — Human Soldier", s82BnsServantScaleOracle)
	if effectstest.PlusOnes(g, servant) != 1 {
		t.Fatalf("Servant has %d counters, want 1", effectstest.PlusOnes(g, servant))
	}
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, servant).Counters = map[string]int{game.CounterPlusOne: 3}
	})
	if err := g.DestroyPermanentForEffect(servant); err != nil {
		t.Fatal(err)
	}
	g.RunStateChecksForTest()
	effectstest.PickTriggerTarget(t, g, me.ID, ally)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, ally) != 3 {
		t.Errorf("ally has %d counters, want 3", effectstest.PlusOnes(g, ally))
	}
}
