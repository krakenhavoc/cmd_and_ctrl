package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b186_test.go — Guilds of Ravnica leftovers (S82 older sets, batch 186).

const (
	s82Old186PeltOracle       = "2348f3d4-02cc-4611-a7e0-b77604d5f2e8"
	s82Old186EtrataOracle     = "2ee87667-ec9d-4ced-b7e6-df3fd18b0ed6"
	s82Old186MonumentOracle   = "3984478b-debb-4a65-ab5b-4bc134e03d82"
	s82Old186ColossusOracle   = "6970f285-dc2b-437b-b099-ca1d3ef106ab"
	s82Old186GargoyleOracle   = "a175de44-3a7e-4165-8f9a-8173c0df13d6"
	s82Old186NecroliskOracle  = "b3418d92-81de-4575-b8da-563be6b31b91"
	s82Old186ChallengerOracle = "bc6b40b1-0d17-4dd2-bec9-f1653611dba3"
	s82Old186TrollOracle      = "c7c35dd9-8c3b-4be5-b310-61db5e73ed6c"
	s82Old186ResearchOracle   = "edac6c23-cff5-4ee2-9227-756b1dc6ab0c"
)

func TestS82Old186CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82Old186PeltOracle:       "Pelt Collector",
		s82Old186EtrataOracle:     "Etrata, the Silencer",
		s82Old186MonumentOracle:   "Rampaging Monument",
		s82Old186ColossusOracle:   "Worldsoul Colossus",
		s82Old186GargoyleOracle:   "Gatekeeper Gargoyle",
		s82Old186NecroliskOracle:  "Undercity Necrolisk",
		s82Old186ChallengerOracle: "Boros Challenger",
		s82Old186TrollOracle:      "Charnel Troll",
		s82Old186ResearchOracle:   "Firemind's Research",
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

func s82Old186CastCreature(t *testing.T, g *game.Game, me *game.Player, name string, power, toughness int) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Bear", Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func s82Old186Counters(t *testing.T, g *game.Game, id uuid.UUID, kind string) int {
	t.Helper()
	n := 0
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			n = c.Counters[kind]
		}
	})
	return n
}

func TestS82Old186PeltCollectorGrowsOnBiggerCreaturesAndGetsTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pelt := effectstest.B21Push(g, me.ID, "Pelt Collector", "Creature — Elf Warrior", s82Old186PeltOracle, 1, 1)
	effectstest.AdvanceToMain(t, g)
	s82Old186CastCreature(t, g, me, "Small", 1, 1)
	if got := s82Old186Counters(t, g, pelt, game.CounterPlusOne); got != 0 {
		t.Fatalf("an equal-power creature entered: %d counters, want 0", got)
	}
	big := s82Old186CastCreature(t, g, me, "Big", 3, 3)
	if got := s82Old186Counters(t, g, pelt, game.CounterPlusOne); got != 1 {
		t.Fatalf("a bigger creature entered: %d counters, want 1", got)
	}
	effectstest.B18Kill(t, g, big)
	if got := s82Old186Counters(t, g, pelt, game.CounterPlusOne); got != 2 {
		t.Fatalf("a power-3 creature died while Pelt had power 2: %d counters, want 2", got)
	}
	if effectstest.HasEffectiveKeyword(t, g, pelt, "trample") {
		t.Error("trample with two counters")
	}
	huge := effectstest.B16Creature(g, me.ID, "Huge", "Creature — Ogre", 6, 6, "G")
	effectstest.B18Kill(t, g, huge)
	if got := s82Old186Counters(t, g, pelt, game.CounterPlusOne); got != 3 {
		t.Fatalf("%d counters, want 3", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, pelt, "trample") {
		t.Error("no trample with three +1/+1 counters")
	}
}

func TestS82Old186EtrataExilesWithAHitCounterAndShufflesAway(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	etrata := effectstest.B21Push(g, me.ID, "Etrata, the Silencer", "Legendary Creature — Vampire Assassin", s82Old186EtrataOracle, 3, 5)
	bear := effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	effectstest.DealCombatDamageToPlayer(g, etrata, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickCard(t, g, me.ID, bear)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !g.Exile.Contains(bear) {
		t.Fatal("the target was not exiled")
	}
	if got := s82Old186Counters(t, g, bear, "hit"); got != 1 {
		t.Errorf("hit counters on the exiled card: %d, want 1", got)
	}
	if g.Battlefield.Contains(etrata) {
		t.Error("Etrata stayed on the battlefield; her owner shuffles her into their library")
	}
	if !me.Library.Contains(etrata) {
		t.Error("Etrata is not in her owner's library")
	}
}

func TestS82Old186EtrataMakesTheThirdHitCardLoseTheGame(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	etrata := effectstest.B21Push(g, me.ID, "Etrata, the Silencer", "Legendary Creature — Vampire Assassin", s82Old186EtrataOracle, 3, 5)
	for i := 0; i < 2; i++ {
		g.Exile.PushTop(game.Card{InstanceID: uuid.New(), Name: "Hit", TypeLine: "Creature — Bear", Counters: map[string]int{"hit": 1}, Owner: opp.ID, Controller: opp.ID})
	}
	effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	effectstest.DealCombatDamageToPlayer(g, etrata, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		var bear uuid.UUID
		g.ReadSnapshot(func() {
			for _, c := range g.Battlefield.Cards {
				if c.Controller == opp.ID && c.IsCreature() {
					bear = c.InstanceID
				}
			}
		})
		effectstest.PickCard(t, g, me.ID, bear)
		effectstest.PassPriorityAroundTable(t, g)
	}
	var out bool
	g.ReadSnapshot(func() { out = opp.Eliminated })
	if !out {
		t.Error("the player who owns three hit-counter cards in exile did not lose")
	}
}

func TestS82Old186EtrataCantBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	etrata := effectstest.B21Push(g, me.ID, "Etrata, the Silencer", "Legendary Creature — Vampire Assassin", s82Old186EtrataOracle, 3, 5)
	wall := effectstest.B16Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 8, "W")
	effectstest.BrAttack(t, g, etrata)
	effectstest.BrRefusal(t, g.DeclareBlocker(wall, etrata), game.BlockReasonCantBeBlocked)
}

func TestS82Old186RampagingMonumentEntersWithThreeAndGrowsOnMulticolor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Rampaging Monument", TypeLine: "Artifact Creature — Cleric", OracleID: s82Old186MonumentOracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := s82Old186Counters(t, g, id, game.CounterPlusOne); got != 3 {
		t.Fatalf("entered with %d counters, want 3", got)
	}
	if got := effectstest.EffectivePowerWithCounters(t, g, id); got != 3 {
		t.Errorf("power %d, want 3", got)
	}
	mono := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: mono, Name: "Mono", TypeLine: "Sorcery", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, mono, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast mono: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := s82Old186Counters(t, g, id, game.CounterPlusOne); got != 3 {
		t.Fatalf("a mono-coloured spell added a counter: %d", got)
	}
	multi := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: multi, Name: "Multi", TypeLine: "Sorcery", Colors: []string{"R", "W"}, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, multi, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast multi: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := s82Old186Counters(t, g, id, game.CounterPlusOne); got != 4 {
		t.Errorf("a multicolored spell: %d counters, want 4", got)
	}
}

func TestS82Old186WorldsoulColossusEntersWithXCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Worldsoul Colossus", TypeLine: "Creature — Elemental", ManaCost: "{X}{G}{W}", OracleID: s82Old186ColossusOracle, Owner: me.ID, Controller: me.ID})
	})
	effectstest.FloatMana(t, g, me, "{G}{W}{C}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 3}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := s82Old186Counters(t, g, id, game.CounterPlusOne); got != 3 {
		t.Errorf("entered with %d counters, want X = 3", got)
	}
	if got := effectstest.EffectivePowerWithCounters(t, g, id); got != 3 {
		t.Errorf("power %d, want 3", got)
	}
	spec, _ := effects.Lookup(s82Old186ColossusOracle)
	if spec.TapCost == nil {
		t.Error("Convoke is missing")
	}
}

func TestS82Old186GatekeeperGargoyleEntersWithACounterPerGate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Permanent(g, me.ID, "Gate A", "Land — Gate")
	effectstest.B12Permanent(g, me.ID, "Gate B", "Land — Gate")
	effectstest.B12Permanent(g, g.Seats[1].ID, "Their Gate", "Land — Gate")
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Gatekeeper Gargoyle", TypeLine: "Artifact Creature — Gargoyle", OracleID: s82Old186GargoyleOracle, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := s82Old186Counters(t, g, id, game.CounterPlusOne); got != 2 {
		t.Errorf("entered with %d counters, want 2 (the Gates you control)", got)
	}
	if got := effectstest.EffectivePowerWithCounters(t, g, id); got != 5 {
		t.Errorf("power %d, want 5", got)
	}
}

func TestS82Old186UndercityNecroliskSacrificesForACounterAndMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	necro := effectstest.B21Push(g, me.ID, "Undercity Necrolisk", "Creature — Zombie Lizard", s82Old186NecroliskOracle, 3, 3)
	fodder := effectstest.B16Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1, "G")
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, necro, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{necro}}); err == nil {
		t.Fatal("the Necrolisk sacrificed itself; the cost says another creature")
	}
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.B16Activate(t, g, me.ID, necro, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still around")
	}
	if got := s82Old186Counters(t, g, necro, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters, want 1", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, necro, "menace") {
		t.Error("no menace until end of turn")
	}
}

func TestS82Old186BorosChallengerMentorsAndPumpsItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	chal := effectstest.B21Push(g, me.ID, "Boros Challenger", "Creature — Human Soldier", s82Old186ChallengerOracle, 2, 3)
	small := effectstest.Ally(g, "Small", 1)
	prompt := effectstest.MentorAttack(t, g, chal, small)
	if err := g.ResolvePickTarget(prompt.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: small}); err != nil {
		t.Fatalf("pick mentor target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := s82Old186Counters(t, g, small, game.CounterPlusOne); got != 1 {
		t.Errorf("the mentored creature has %d counters, want 1", got)
	}
	effectstest.FloatMana(t, g, me, "{R}{W}{C}{C}")
	effectstest.B16Activate(t, g, me.ID, chal, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, chal); got != 3 {
		t.Errorf("power %d, want 3 after the pump", got)
	}
}

func TestS82Old186CharnelTrollEatsACreatureCardOrIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	troll := effectstest.B21Push(g, me.ID, "Charnel Troll", "Creature — Troll", s82Old186TrollOracle, 4, 4)
	dead := effectstest.PushGraveyardPermanent(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	effectstest.PushGraveyardPermanent(me, "Dead Land", "Land", "")
	effectstest.AdvanceToStepOf(t, g, 0, game.StepUpkeep)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("not asked which creature card to exile")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != dead || pick.ChooseMin != 1 {
		t.Fatalf("offered %v (min %d), want only the creature card and no way to decline", pick.ChooseCards, pick.ChooseMin)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{dead}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(dead) {
		t.Error("the creature card was not exiled")
	}
	if got := s82Old186Counters(t, g, troll, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters, want 1", got)
	}
	// A second upkeep with no creature card left sacrifices it.
	for i := 0; i < len(g.Seats); i++ {
		effectstest.EndTurn(t, g)
	}
	effectstest.AdvanceToStepOf(t, g, 0, game.StepUpkeep)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(troll) {
		t.Error("the Troll survived an upkeep with no creature card in the graveyard")
	}
}

func TestS82Old186CharnelTrollDiscardsACreatureCardForACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	troll := effectstest.B21Push(g, me.ID, "Charnel Troll", "Creature — Troll", s82Old186TrollOracle, 4, 4)
	card := effectstest.AvB1Hand(me, game.Card{Name: "Hand Bear", TypeLine: "Creature — Bear"})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{B}{G}")
	effectstest.B16Activate(t, g, me.ID, troll, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{card}})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(card) {
		t.Error("the creature card was not discarded")
	}
	if got := s82Old186Counters(t, g, troll, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters, want 1", got)
	}
}

func TestS82Old186FiremindsResearchCountsCastsAndSpendsCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	res := effectstest.B21Push(g, me.ID, "Firemind's Research", "Enchantment", s82Old186ResearchOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 5; i++ {
		effectstest.CastCatalogSpell(t, g, "Spell", "Sorcery", "", nil)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := s82Old186Counters(t, g, res, game.CounterCharge); got != 5 {
		t.Fatalf("%d charge counters after five sorceries, want 5", got)
	}
	creature := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: creature, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := s82Old186Counters(t, g, res, game.CounterCharge); got != 5 {
		t.Errorf("a creature spell added a counter: %d", got)
	}
	top := effectstest.SeedLibrary(me, "Top Card")[0]
	effectstest.FloatMana(t, g, me, "{C}{U}")
	effectstest.B16Activate(t, g, me.ID, res, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(top) {
		t.Error("the first ability draws a card")
	}
	if got := s82Old186Counters(t, g, res, game.CounterCharge); got != 3 {
		t.Errorf("%d charge counters after removing two, want 3", got)
	}
	effectstest.FloatMana(t, g, me, "{C}{R}")
	if err := g.ActivateCatalogAbility(me.ID, res, 1, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Fatal("removed five counters from a Research with three")
	}
	for i := 0; i < 2; i++ {
		effectstest.CastCatalogSpell(t, g, "Spell", "Sorcery", "", nil)
		effectstest.PassPriorityAroundTable(t, g)
	}
	life := opp.Life
	effectstest.B16Activate(t, g, me.ID, res, 1, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-5 {
		t.Errorf("life %d → %d, want -5", life, opp.Life)
	}
}
