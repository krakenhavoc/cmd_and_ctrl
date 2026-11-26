package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b467_cards_test.go — Mirrodin Besieged, S82 OLD batch 467 (#5576).
// Every test plays the real catalog entry through the engine.

const (
	oldB467SteelSabotage = "0e71fc56-0332-4e5e-8658-9257cc0d8a1d"
	oldB467Fuel          = "667cffdd-8951-4223-8895-fa46e9f14b96"
	oldB467PistusStrike  = "138f9daa-19d4-4290-94a7-09e00dbcebb0"
	oldB467DivineOffer   = "231f8edb-4ea1-44be-8794-b76a31462dfc"
	oldB467Rally         = "4145c75d-d052-40d2-8ef4-0a08071582b8"
	oldB467Hero          = "f18799c4-ec54-4435-be51-03ea25809289"
	oldB467Keepers       = "6f2f4c31-74dc-4a52-96ad-94c3d51398d1"
	oldB467Cryptoplasm   = "e7bef318-2720-45bc-b663-aafd724bc511"
	oldB467BurnImpure    = "9503c801-eb6f-4f34-9062-c56105509b7b"
	oldB467SerumRaker    = "7bb7b937-a7ad-4f5c-b173-df71556c48e9"
	oldB467Pool          = "cb53bd72-dee8-46fa-b0f9-1eee3a99379a"
	oldB467Mine          = "bbb50cf5-cd07-4ae8-bfa8-d747e44fb1ef"
	oldB467RotWolf       = "7409906d-7111-48fa-bfea-8b95e0fdb7dc"
)

func TestOldB467EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB467SteelSabotage: "Steel Sabotage",
		oldB467Fuel:          "Fuel for the Cause",
		oldB467PistusStrike:  "Pistus Strike",
		oldB467DivineOffer:   "Divine Offering",
		oldB467Rally:         "Rally the Forces",
		oldB467Hero:          "Hero of Oxid Ridge",
		oldB467Keepers:       "Melira's Keepers",
		oldB467Cryptoplasm:   "Cryptoplasm",
		oldB467BurnImpure:    "Burn the Impure",
		oldB467SerumRaker:    "Serum Raker",
		oldB467Pool:          "Knowledge Pool",
		oldB467Mine:          "Magnetic Mine",
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

func TestOldB467SteelSabotageCountersAnArtifactSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.AvB1Hand(me, game.Card{Name: "Mind Stone", TypeLine: "Artifact", ManaCost: "{2}"})
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the artifact: %v", err)
	}
	sab := effectstest.AvB1Hand(me, game.Card{Name: "Steel Sabotage", TypeLine: "Instant", OracleID: oldB467SteelSabotage})
	if err := g.CastSpell(me.ID, sab, game.CastSpellParams{Modes: []int{0}, Targets: effectstest.CardRefs(rock)}); err != nil {
		t.Fatalf("Steel Sabotage's counter mode: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !me.Graveyard.Contains(rock) {
		t.Error("the artifact spell is countered into its owner's graveyard")
	}
}

func TestOldB467SteelSabotageBouncesAnArtifactButNotACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	refused := effectstest.AvB1Hand(me, game.Card{Name: "Steel Sabotage", TypeLine: "Instant", OracleID: oldB467SteelSabotage})
	if err := g.CastSpell(me.ID, refused, game.CastSpellParams{Modes: []int{1}, Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("the return mode took a nonartifact creature")
	}
	sab := effectstest.AvB1Hand(me, game.Card{Name: "Steel Sabotage", TypeLine: "Instant", OracleID: oldB467SteelSabotage})
	if err := g.CastSpell(me.ID, sab, game.CastSpellParams{Modes: []int{1}, Targets: effectstest.CardRefs(rock)}); err != nil {
		t.Fatalf("Steel Sabotage's return mode: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Hand.Contains(rock) {
		t.Error("the artifact did not return to its owner's hand")
	}
}

func TestOldB467FuelForTheCauseCountersThenProliferates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	counted := effectstest.B12Creature(g, me.ID, "Counted", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, counted).Counters = map[string]int{"+1/+1": 1} })
	rock := effectstest.AvB1Hand(me, game.Card{Name: "Mind Stone", TypeLine: "Artifact", ManaCost: "{2}"})
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the artifact: %v", err)
	}
	fuel := effectstest.AvB1Hand(me, game.Card{Name: "Fuel for the Cause", TypeLine: "Instant", OracleID: oldB467Fuel})
	if err := g.CastSpell(me.ID, fuel, game.CastSpellParams{Targets: effectstest.CardRefs(rock)}); err != nil {
		t.Fatalf("cast Fuel for the Cause: %v", err)
	}
	effectstest.SettleAnsweringProliferate(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the spell was not countered")
	}
	if got := effectstest.BattlefieldCardCopy(t, g, counted).Counters["+1/+1"]; got != 2 {
		t.Errorf("+1/+1 counters %d, want 2 after the proliferate", got)
	}
}

func TestOldB467PistusStrikeDestroysAFlyerAndPoisonsItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	flyer := effectstest.PushCatalogPermanent(g, opp.ID, "Serum Raker", "Creature — Phyrexian Drake", oldB467SerumRaker, false)
	ground := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Pistus Strike", "Instant", oldB467PistusStrike, effectstest.CardRefs(ground)); err == nil {
		t.Fatal("Pistus Strike took a creature without flying")
	}
	effectstest.CastCatalogSpell(t, g, "Pistus Strike", "Instant", oldB467PistusStrike, effectstest.CardRefs(flyer))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(flyer) {
		t.Error("the flyer survived")
	}
	if got := effectstest.PoisonOn(opp); got != 1 {
		t.Errorf("poison counters %d, want 1", got)
	}
}

func TestOldB467DivineOfferingDestroysAnArtifactAndGainsItsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big Rock", TypeLine: "Artifact", ManaCost: "{4}", Owner: opp.ID, Controller: opp.ID})
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Divine Offering", "Instant", oldB467DivineOffer, effectstest.CardRefs(bear)); err == nil {
		t.Fatal("Divine Offering took a nonartifact")
	}
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Divine Offering", "Instant", oldB467DivineOffer, effectstest.CardRefs(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	if me.Life != life+4 {
		t.Errorf("life %d → %d, want +4", life, me.Life)
	}
}

func TestOldB467RallyTheForcesBoostsOnlyAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a)
	rally := effectstest.AvB1Hand(me, game.Card{Name: "Rally the Forces", TypeLine: "Instant", OracleID: oldB467Rally})
	if err := g.CastSpell(me.ID, rally, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Rally the Forces: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, a); p != 3 {
		t.Errorf("the attacker's power is %d, want 3", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, a, "first strike") {
		t.Error("the attacker has no first strike")
	}
	if p, _ := effectstest.EffectivePTOf(t, g, idle); p != 2 {
		t.Errorf("a non-attacker's power is %d, want 2", p)
	}
	if effectstest.HasEffectiveKeyword(t, g, idle, "first strike") {
		t.Error("a non-attacker gained first strike")
	}
}

func TestOldB467HeroOfOxidRidgeBattleCriesAndBarsSmallBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hero := effectstest.B12Push(g, me.ID, "Hero of Oxid Ridge", "Creature — Human Knight", oldB467Hero, 4, 2)
	buddy := effectstest.B12Creature(g, me.ID, "Buddy", "Creature — Bear", 2, 2)
	tiny := effectstest.B12Creature(g, opp.ID, "Tiny", "Creature — Insect", 1, 1)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, hero, buddy)
	effectstest.SettleOrderingTriggers(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, buddy); p != 3 {
		t.Errorf("the other attacker's power is %d, want 3 (battle cry)", p)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, hero); p != 4 {
		t.Errorf("the Hero's own power is %d, want 4", p)
	}
	oldB465CanBlockB467(t, g, buddy, tiny, false)
	oldB465CanBlockB467(t, g, buddy, big, true)
}

func oldB465CanBlockB467(t *testing.T, g *game.Game, attacker, blocker uuid.UUID, want bool) {
	t.Helper()
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		a, _ := g.LookupCardForEffect(attacker)
		b, _ := g.LookupCardForEffect(blocker)
		if r := g.BlockPairRefusalLocked(&a, &b); r.Legal() != want {
			t.Errorf("blocker %s may block = %v, want %v (%q)", b.Name, r.Legal(), want, r.Reason)
		}
	})
}

func TestOldB467MelirasKeepersCantHaveCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	keepers := effectstest.B12Push(g, me.ID, "Melira's Keepers", "Creature — Human Warrior", oldB467Keepers, 4, 4)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	effectstest.EffectivePTOf(t, g, keepers) // settle the layers
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(keepers, game.CounterPlusOne, 2)
		_ = g.AddCounterForEffect(other, game.CounterPlusOne, 2)
	})
	if got := effectstest.BattlefieldCardCopy(t, g, keepers).Counters["+1/+1"]; got != 0 {
		t.Errorf("the Keepers has %d +1/+1 counters, want none", got)
	}
	if got := effectstest.BattlefieldCardCopy(t, g, other).Counters["+1/+1"]; got != 2 {
		t.Errorf("another creature has %d counters, want 2", got)
	}
}

func TestOldB467BurnTheImpureHitsTheControllerOnlyOfAnInfectCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	wolf := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rot Wolf", TypeLine: "Creature — Phyrexian Wolf",
		OracleID: oldB467RotWolf, Power: 2, Toughness: 2, Keywords: []string{"infect"}, Owner: opp.ID, Controller: opp.ID})
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 5, 5)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Burn the Impure", "Instant", oldB467BurnImpure, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Fatalf("a creature without infect cost its controller %d life", life-opp.Life)
	}
	effectstest.CastCatalogSpell(t, g, "Burn the Impure", "Instant", oldB467BurnImpure, effectstest.CardRefs(wolf))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wolf) {
		t.Error("the 2/2 Wolf survived 3 damage")
	}
	if opp.Life != life-3 {
		t.Errorf("life %d → %d, want 3 damage to the infect creature's controller", life, opp.Life)
	}
}

func TestOldB467SerumRakerMakesEachPlayerDiscardWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raker := effectstest.PushCatalogPermanent(g, me.ID, "Serum Raker", "Creature — Phyrexian Drake", oldB467SerumRaker, false)
	effectstest.AdvanceToMain(t, g)
	for _, p := range []*game.Player{me, opp} {
		for i := 0; i < 3; i++ {
			effectstest.HandSpell(p, "Card", "Sorcery", "{9}")
		}
	}
	mine, theirs := me.Hand.Size(), opp.Hand.Size()
	effectstest.P1Destroy(t, g, raker)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range []*game.Player{me, opp} {
		if effectstest.DiscardChoiceFor(g, p.ID) == nil {
			t.Fatalf("player %v was not asked to discard", p.ID)
		}
		effectstest.DiscardFromHand(t, g, p.ID)
	}
	g.SettleResolution()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != mine-1 || opp.Hand.Size() != theirs-1 {
		t.Errorf("hands %d → %d and %d → %d, want one card each", mine, me.Hand.Size(), theirs, opp.Hand.Size())
	}
}

func TestOldB467CryptoplasmCopiesAnotherCreatureAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plasm := effectstest.B12Push(g, me.ID, "Cryptoplasm", "Creature — Shapeshifter", oldB467Cryptoplasm, 2, 2)
	effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	effectstest.AdvanceToStepOf(t, g, 0, game.StepUpkeep)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the upkeep trigger")
	}
	if effectstest.HasID(pick.PickTargetCards, plasm) {
		t.Error("the Cryptoplasm offered itself as 'another' target")
	}
	var ogre uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Ogre" {
			ogre = c.InstanceID
		}
	}
	effectstest.PickCard(t, g, me.ID, ogre)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.BattlefieldCardCopy(t, g, plasm)
	if c.Name != "Ogre" {
		t.Fatalf("the Cryptoplasm is %q, want a copy of the Ogre", c.Name)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, plasm); p != 4 || tough != 4 {
		t.Errorf("the copy is %d/%d, want 4/4", p, tough)
	}
}

func TestOldB467MagneticMineDamagesTheControllerOfAnotherArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushCatalogPermanent(g, me.ID, "Magnetic Mine", "Artifact", oldB467Mine, false)
	theirs := effectstest.B12Permanent(g, opp.ID, "Stone", "Artifact")
	mineRock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	myLife, oppLife := me.Life, opp.Life
	effectstest.P1Destroy(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 || me.Life != myLife {
		t.Fatalf("an opponent's artifact: their life %d → %d, mine %d → %d, want them -2", oppLife, opp.Life, myLife, me.Life)
	}
	effectstest.P1Destroy(t, g, mineRock)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife-2 {
		t.Fatalf("my own artifact: life %d → %d, want -2", myLife, me.Life)
	}
	effectstest.P1Destroy(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 {
		t.Error("a nonartifact creature dying hurt its controller")
	}
	effectstest.P1Destroy(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife-2 {
		t.Error("the Mine's own destruction triggered it ('another artifact')")
	}
}

func TestOldB467KnowledgePoolImprintsThenExilesCastSpellsAndOffersAFreeCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	before := g.Exile.Size()
	effectstest.CastCatalogSpell(t, g, "Knowledge Pool", "Artifact", oldB467Pool, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := g.Exile.Size() - before; got != 12 {
		t.Fatalf("%d cards exiled by the imprint, want twelve (three from each of four libraries)", got)
	}
	var pool uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Knowledge Pool" {
			pool = c.InstanceID
		}
	}
	linked := 0
	for _, c := range g.Exile.Cards {
		if c.ExiledWith.ID == pool {
			linked++
		}
	}
	if linked != 12 {
		t.Fatalf("%d cards are linked to the Pool, want 12", linked)
	}
	// Seed a free-castable card among the linked ones.
	var target uuid.UUID
	g.WithWriteLock(func() {
		for i := range g.Exile.Cards {
			if g.Exile.Cards[i].ExiledWith.ID == pool {
				g.Exile.Cards[i].Name, g.Exile.Cards[i].TypeLine = "Linked Rock", "Artifact"
				g.Exile.Cards[i].ManaCost = "{3}"
				target = g.Exile.Cards[i].InstanceID
				break
			}
		}
	})
	spell := effectstest.AvB1Hand(me, game.Card{Name: "Cast Rock", TypeLine: "Artifact", ManaCost: "{2}"})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast from hand: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(spell) {
		t.Fatal("the spell cast from hand was not exiled")
	}
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, target)
	c := effectstest.ResolutionCastOfferFor(g, me.ID)
	for _, id := range c.DiscardOptions {
		if id == spell {
			t.Error("the spell just exiled was offered; only OTHER exiled cards may be cast")
		}
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, target, game.CastSpellParams{}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(target) {
		t.Error("the free-cast artifact did not resolve onto the battlefield")
	}
	if got := effectstest.BattlefieldNamed(g, me.ID, "Cast Rock"); got != 0 {
		t.Error("the exiled spell resolved anyway")
	}
}
