package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tidus_a_test.go — the Tidus, Yuna's Guardian deck request (#3246),
// slice A: the cards existing seams carry.

const (
	tdDestroyEvilOracle       = "b75e667a-1ceb-445a-acf4-ae1fd2dae9e7"
	tdSungrassPrairieOracle   = "0a28dff0-2bd6-4105-b73a-b6c4735833fd"
	tdGratefulApparitionOracl = "f1b3e0fc-1dc3-4eba-bd54-557f036f5ea9"
	tdLuluOracle              = "969fa9b9-161f-461e-8878-9ebcff5a2669"
	tdPathOfDiscoveryOracle   = "52714c3c-6821-44dd-88ba-813554df9914"
	tdBroodmothOracle         = "28c7c816-07e7-42fb-923c-bf149ba28b38"
	tdRejuvenatorOracle       = "b5f5d6e3-e72e-40a5-a5c7-410617b6f933"
	tdSummonersSendingOracle  = "46b01055-4839-4b95-aa53-b954e9669e03"
	tdShelindaOracle          = "b06c9a4b-f309-45b3-ae5b-0ce3f067a5cb"
	tdOakaOracle              = "ed0fc294-de9b-4c64-b97d-e3a41fd2f46a"
	tdGenerousPatronOracle    = "fb084141-cb80-479e-b5d4-114c1f31eb6c"
	tdTogetherForeverOracle   = "36b19ec0-d581-4213-bfae-1d7808a2f60d"
	tdForgeOfHeroesOracle     = "77807103-bcd5-479f-bedd-f5d97aa6d3d2"
	tdMaesterSeymourOracle    = "e7e41303-69a8-4617-b418-142cf41edb9b"
)

func TestTidusASliceIsRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		tdDestroyEvilOracle:       "Destroy Evil",
		tdSungrassPrairieOracle:   "Sungrass Prairie",
		tdGratefulApparitionOracl: "Grateful Apparition",
		tdLuluOracle:              "Lulu, Stern Guardian",
		tdPathOfDiscoveryOracle:   "Path of Discovery",
		tdBroodmothOracle:         "Luminous Broodmoth",
		tdRejuvenatorOracle:       "Rampant Rejuvenator",
		tdSummonersSendingOracle:  "Summoner's Sending",
		tdShelindaOracle:          "Shelinda, Yevon Acolyte",
		tdOakaOracle:              "O'aka, Traveling Merchant",
		tdGenerousPatronOracle:    "Generous Patron",
		tdTogetherForeverOracle:   "Together Forever",
		tdForgeOfHeroesOracle:     "Forge of Heroes",
		tdMaesterSeymourOracle:    "Maester Seymour",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered as %q", oracle, spec.Name)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: Completeness = %v, Caveats = %v", name, spec.Completeness, spec.Caveats)
		}
	}
}

// tdCreature is a creature with the given power and toughness, able to
// attack and tap.
func tdCreature(g *game.Game, owner uuid.UUID, name string, power, toughness int) uuid.UUID {
	return b12Push(g, owner, name, "Creature — Test", "", power, toughness)
}

// tdCastCreature casts a vanilla creature from the active seat's hand
// in a main phase and returns its id.
func tdCastCreature(t *testing.T, g *game.Game, name string, power, toughness int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Test",
		Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

func tdCard(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func TestTidusADestroyEvilKillsABigCreatureOrAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	small := tdCreature(g, opp.ID, "Small", 2, 2)
	big := tdCreature(g, opp.ID, "Big", 1, 4)
	ench := b12Push(g, opp.ID, "Their Enchantment", "Enchantment", "", 0, 0)

	if _, err := castModalErr(t, g, "Destroy Evil", "Instant", tdDestroyEvilOracle, []int{0}, tdCard(small)); err == nil {
		t.Error("a creature with toughness 2 was a legal target")
	}
	castModal(t, g, "Destroy Evil", "Instant", tdDestroyEvilOracle, []int{0}, tdCard(big))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the toughness-4 creature survived")
	}
	castModal(t, g, "Destroy Evil", "Instant", tdDestroyEvilOracle, []int{1}, tdCard(ench))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment survived")
	}
}

// castModalErr is castModal returning the announce error.
func castModalErr(t *testing.T, g *game.Game, name, typeLine, oracle string, modes []int, targets []game.TargetRef) (uuid.UUID, error) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return id, g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes, Targets: targets})
}

func TestTidusASungrassPrairieFiltersOneIntoGreenWhite(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := pushCatalogPermanent(g, me.ID, "Sungrass Prairie", "Land", tdSungrassPrairieOracle, false)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if mcPoolCount(me, "G") != 1 || mcPoolCount(me, "W") != 1 || mcPoolCount(me, "C") != 0 {
		t.Errorf("pool = %+v, want {G}{W} for the {1}", me.ManaPool)
	}
}

func TestTidusAGratefulApparitionProliferatesOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	app := b12Push(g, me.ID, "Grateful Apparition", "Creature — Spirit", tdGratefulApparitionOracl, 1, 1)
	bear := tdCreature(g, me.ID, "Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	attackWith(t, g, opp.ID, app)
	passPriorityAroundTable(t, g)
	answerProliferate(t, g, bear)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("the Bear has %d +1/+1 counters after the proliferate, want 2", got)
	}
}

func TestTidusALuluStunsACreatureAttackingYouAndProliferates(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lulu := b12Push(g, me.ID, "Lulu, Stern Guardian", "Legendary Creature — Human Wizard", tdLuluOracle, 2, 3)
	raider := tdCreature(g, opp.ID, "Raider", 3, 3)

	advanceToDeclareAttackersOf(t, g, 1)
	if err := g.DeclareAttacker(raider, me.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, raider)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, raider, game.CounterStun); got != 1 {
		t.Fatalf("the attacker has %d stun counters, want 1", got)
	}

	fillPoolColored(me, "U", 4)
	if err := g.ActivateCatalogAbility(me.ID, lulu, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Lulu: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerProliferate(t, g, raider)
	if got := countersOn(g, raider, game.CounterStun); got != 2 {
		t.Errorf("after the proliferate the attacker has %d stun counters, want 2", got)
	}
}

func TestTidusAPathOfDiscoveryMakesAnEnteringCreatureExplore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Path of Discovery", "Enchantment", tdPathOfDiscoveryOracle, false)
	land := putOnTopForTest(me, "Forest", "Basic Land — Forest")
	tdCastCreature(t, g, "Bear", 2, 2)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) {
		t.Error("the revealed land did not go to hand")
	}

	top := putOnTopForTest(me, "Some Sorcery", "Sorcery")
	bear := tdCastCreature(t, g, "Second Bear", 2, 2)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the explorer has %d +1/+1 counters, want 1", got)
	}
	answerExploreQuestion(t, g, me.ID, true)
	if !me.Graveyard.Contains(top) {
		t.Error("the revealed nonland was not put into the graveyard")
	}
}

func TestTidusALuminousBroodmothReturnsACreatureWithoutFlyingWithAFlyingCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Luminous Broodmoth", "Creature — Insect", tdBroodmothOracle, 3, 4)
	bear := tdCreature(g, me.ID, "Bear", 2, 2)
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the creature did not return")
	}
	if got := countersOn(g, bear, game.CounterFlying); got != 1 {
		t.Errorf("the returned creature has %d flying counters, want 1", got)
	}
	// It has flying now, so a second death does not bring it back.
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("a creature with flying came back")
	}
}

func TestTidusARampantRejuvenatorFetchesBasicsEqualToItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var forests []uuid.UUID
	for i := 0; i < 3; i++ {
		forests = append(forests, putOnTopForTest(me, "Forest", "Basic Land — Forest"))
	}
	hydra := castCatalogSpell(t, g, "Rampant Rejuvenator", "Creature — Plant Hydra", tdRejuvenatorOracle, nil)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, hydra, game.CounterPlusOne); got != 2 {
		t.Fatalf("it entered with %d +1/+1 counters, want 2", got)
	}
	lands := func() int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.Name == "Forest" {
				n++
			}
		}
		return n
	}
	before := lands()
	killCreature(t, g, me.ID, hydra)
	passPriorityAroundTable(t, g)
	p := latestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, me.ID)
	if p == nil {
		t.Fatal("no search prompt")
	}
	if err := g.ResolveSearchLibrary(p.ID, me.ID, forests[:3]); err == nil {
		t.Error("three lands were accepted for a power of 2")
	}
	if err := g.ResolveSearchLibrary(p.ID, me.ID, forests[:2]); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if got := lands() - before; got != 2 {
		t.Errorf("%d basic lands entered, want 2", got)
	}
}

func TestTidusASummonersSendingExilesForASpiritThatGrowsOnABigCard(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	pushCatalogPermanent(g, me.ID, "Summoner's Sending", "Enchantment", tdSummonersSendingOracle, false)
	big := pushGraveyardCreature(g, g.Seats[1].ID, "Big Dead", "{3}{G}")

	advanceToEndStepOf(t, g, seat)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, big)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(big) {
		t.Fatal("the creature card was not exiled")
	}
	var spirit *game.Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == me.ID && c.Name == "Spirit" {
			spirit = c
		}
	}
	if spirit == nil {
		t.Fatal("no Spirit token")
	}
	if spirit.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("the Spirit has %d +1/+1 counters, want 1 for a mana value 4 card", spirit.Counters[game.CounterPlusOne])
	}
}

func TestTidusAShelindaCountersTheSmallerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shelinda := b12Push(g, me.ID, "Shelinda, Yevon Acolyte", "Legendary Creature — Human Cleric", tdShelindaOracle, 2, 2)

	small := tdCastCreature(t, g, "Small", 1, 1)
	passPriorityAroundTable(t, g)
	if countersOn(g, small, game.CounterPlusOne) != 1 || countersOn(g, shelinda, game.CounterPlusOne) != 0 {
		t.Errorf("a 1/1 entering: it has %d, Shelinda %d; want 1 and 0",
			countersOn(g, small, game.CounterPlusOne), countersOn(g, shelinda, game.CounterPlusOne))
	}
	big := tdCastCreature(t, g, "Big", 5, 5)
	passPriorityAroundTable(t, g)
	if countersOn(g, big, game.CounterPlusOne) != 0 || countersOn(g, shelinda, game.CounterPlusOne) != 1 {
		t.Errorf("a 5/5 entering: it has %d, Shelinda %d; want 0 and 1",
			countersOn(g, big, game.CounterPlusOne), countersOn(g, shelinda, game.CounterPlusOne))
	}
}

func TestTidusAOakaRemovesACounterToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	oaka := b12Push(g, me.ID, "O'aka, Traveling Merchant", "Legendary Creature — Human Citizen", tdOakaOracle, 1, 2)
	bear := tdCreature(g, me.ID, "Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, oaka, 0, game.ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{bear}, CounterKind: game.CounterPlusOne,
	}); err != nil {
		t.Fatalf("activate O'aka: %v", err)
	}
	if got := countersOn(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("the counter was not removed as a cost (%d left)", got)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
}

func TestTidusAGenerousPatronSupportsAndDrawsPerOpposingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	a := tdCreature(g, opp.ID, "Their A", 2, 2)
	b := tdCreature(g, opp.ID, "Their B", 2, 2)
	castCatalogSpell(t, g, "Generous Patron", "Creature — Elf Advisor", tdGenerousPatronOracle, nil)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no support target prompt")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, append(tdCard(a), tdCard(b)...)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if countersOn(g, a, game.CounterPlusOne) != 1 || countersOn(g, b, game.CounterPlusOne) != 1 {
		t.Fatal("support did not put a counter on each target")
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("drew %d, want 2 (one per opposing creature)", me.Hand.Size()-hand)
	}
}

func TestTidusATogetherForeverReturnsTheMarkedCreatureToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	tf := pushCatalogPermanent(g, me.ID, "Together Forever", "Enchantment", tdTogetherForeverOracle, false)
	bear := tdCreature(g, me.ID, "Bear", 2, 2)
	plain := tdCreature(g, me.ID, "Plain", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })

	fillPoolColored(me, "W", 1)
	if err := g.ActivateCatalogAbility(me.ID, tf, 0, game.ActivateAbilityParams{Targets: tdCard(plain)}); err == nil {
		t.Error("a creature with no counter was a legal target")
	}
	fillPoolColored(me, "W", 1)
	if err := g.ActivateCatalogAbility(me.ID, tf, 0, game.ActivateAbilityParams{Targets: tdCard(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the creature did not return to its owner's hand")
	}
}

func TestTidusAForgeOfHeroesCountersACommanderThatEnteredThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	forge := pushCatalogPermanent(g, me.ID, "Forge of Heroes", "Land", tdForgeOfHeroesOracle, false)
	cmdr := tdCreature(g, me.ID, "My Commander", 3, 3)
	old := tdCreature(g, me.ID, "Old Commander", 3, 3)
	for i := range g.Battlefield.Cards {
		if id := g.Battlefield.Cards[i].InstanceID; id == cmdr || id == old {
			g.Battlefield.Cards[i].IsCommander = true
		}
	}
	if g.TurnTally.Entered == nil {
		g.TurnTally.Entered = map[uuid.UUID]int{}
	}
	g.TurnTally.Entered[cmdr] = 1
	delete(g.TurnTally.Entered, old)

	if err := g.ActivateCatalogAbility(me.ID, forge, 0, game.ActivateAbilityParams{Targets: tdCard(old)}); err == nil {
		t.Error("a commander that did not enter this turn was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, forge, 0, game.ActivateAbilityParams{Targets: tdCard(cmdr)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, cmdr, game.CounterPlusOne); got != 1 {
		t.Errorf("the commander has %d +1/+1 counters, want 1", got)
	}
}

func TestTidusAMaesterSeymourCountersByPowerAndMonstrosityCountsCounters(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	seymour := b12Push(g, me.ID, "Maester Seymour", "Legendary Creature — Human Elf Cleric", tdMaesterSeymourOracle, 1, 3)
	bear := tdCreature(g, me.ID, "Bear", 2, 2)

	advanceTo(t, g, game.StepBeginCombat)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Fatalf("the Bear has %d counters, want 1 (Seymour's power)", got)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterStun, 1) })

	fillPoolColored(me, "G", 5)
	if err := g.ActivateCatalogAbility(me.ID, seymour, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate monstrosity: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, seymour, game.CounterPlusOne); got != 2 {
		t.Errorf("Seymour has %d +1/+1 counters, want 2 (one +1/+1 and one stun among creatures)", got)
	}
}
