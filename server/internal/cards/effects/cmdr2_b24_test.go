package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	cmdr2FerraforOracle        = "a46f9136-b904-403e-b263-e99c9a2fe2ec"
	cmdr2CommunalBrewingOracle = "99deac41-584b-43f2-804e-c56ae0fef047"
	cmdr2LlanowarRebornOracle  = "92acb789-0e42-465c-ac16-40fefec48805"
	cmdr2AetherbornOracle      = "ff6427f6-d2b1-4e33-984d-c06bd15d7c2c"
	cmdr2GenesisHydraOracle    = "4345c359-b9d6-43c8-997f-e00746e6e318"
	cmdr2BossChauffeurOracle   = "5a6176cc-98bf-43ac-815a-1f19c4e9329c"
	cmdr2XavierSalOracle       = "50a9c647-47ba-4732-852a-f0ab2b4549c9"
	cmdr2DevilishValetOracle   = "e1e0dc37-20c9-4236-ae01-e0891ce83afb"
	cmdr2RoseRoomOracle        = "d73f02e0-53fa-4437-ab3a-89a9a1db27c5"
	cmdr2JackedRabbitOracle    = "5a343a5f-3f9b-43c3-8532-8391de7a5443"
)

func cmdr2PowerOf(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return avB1Card(t, g, id).CurrentPower()
}

// cmdr2Unsick lets a creature that was just cast attack and tap.
func cmdr2Unsick(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		if c := findBattlefieldCardForTest(g, id); c != nil {
			c.SummonedThisTurn = false
		}
	})
}

func cmdr2Named(g *game.Game, owner uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == owner {
			n++
		}
	}
	return n
}

func TestCmdr2DevilishValetDoublesPowerPerCreatureEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	valet := b12Push(g, me.ID, "Devilish Valet", "Creature — Devil Warrior", cmdr2DevilishValetOracle, 1, 3)
	// An opponent's creature does not trigger it (played first: the
	// opponent's main phase is another turn, which would end the doubling).
	b13PlayAs(t, g, 1, "Their Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if got := cmdr2PowerOf(t, g, valet); got != 1 {
		t.Fatalf("an opponent's creature doubled it: %d", got)
	}
	aangAdvanceToMain(t, g, 0)
	cmdr2B23Cast(t, g, "Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	if got := cmdr2PowerOf(t, g, valet); got != 2 {
		t.Errorf("power after one creature entered = %d, want 2", got)
	}
	cmdr2B23Cast(t, g, "Bear Two", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	if got := cmdr2PowerOf(t, g, valet); got != 4 {
		t.Errorf("power after a second creature = %d, want 4 (doubled again)", got)
	}
}

func TestCmdr2RoseRoomTreasurerMakesTwoTreasuresThenOffersX(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Rose Room Treasurer", "Creature — Ogre Warrior", cmdr2RoseRoomOracle, 4, 3)
	for i := 1; i <= 2; i++ {
		cmdr2B23Cast(t, g, "Bear", "Creature — Bear", "", 2, 2)
		passPriorityAroundTable(t, g)
		if got := cmdr2Named(g, me.ID, "Treasure"); got != i {
			t.Fatalf("after creature %d: %d Treasures, want %d", i, got, i)
		}
	}
	// The third resolution asks for X instead of a Treasure.
	life := opp.Life
	fillPool(me, 3)
	cmdr2B23Cast(t, g, "Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	if got := cmdr2Named(g, me.ID, "Treasure"); got != 2 {
		t.Errorf("third resolution made a Treasure: %d", got)
	}
	answerNumber(t, g, me.ID, 3)
	answerPayUnless(t, g, me.ID, true)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d -> %d, want 3 damage", life, opp.Life)
	}
}

func TestCmdr2JackedRabbitRavenousAndTokensEqualToPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand := me.Hand.Size()
	rabbit := castXSpell(t, g, "Jacked Rabbit", "Creature — Rabbit Warrior", cmdr2JackedRabbitOracle, "{X}{1}{W}", 5, nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, rabbit, game.CounterPlusOne); got != 5 {
		t.Fatalf("entered with %d counters, want 5", got)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want a card drawn at X = 5 (cast adds one, casting removes it, draw adds one)", hand, got)
	}
	cmdr2Unsick(g, rabbit)
	declareAttack(t, g, opp.ID, rabbit)
	passPriorityAroundTable(t, g)
	// The test card has no printed power, so its power is its five counters.
	want := cmdr2PowerOf(t, g, rabbit)
	if got := cmdr2Named(g, me.ID, "Rabbit"); want != 5 || got != want {
		t.Errorf("%d Rabbit tokens, want %d (the attacker's power, 5)", got, want)
	}
}

func TestCmdr2JackedRabbitDoesNotDrawBelowFive(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	rabbit := castXSpell(t, g, "Jacked Rabbit", "Creature — Rabbit Warrior", cmdr2JackedRabbitOracle, "{X}{1}{W}", 4, nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, rabbit, game.CounterPlusOne); got != 4 {
		t.Fatalf("entered with %d counters, want 4", got)
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d -> %d, want no draw at X = 4", hand, got)
	}
}

func TestCmdr2BossChauffeurCountersTokensAndDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	b12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	boss := cmdr2B23Cast(t, g, "Boss's Chauffeur", "Creature — Elf Citizen", cmdr2BossChauffeurOracle, 0, 0)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, boss, game.CounterPlusOne); got != 3 {
		t.Fatalf("entered with %d counters, want 3 (one plus two other creatures you control)", got)
	}
	cmdr2B23Cast(t, g, "Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, boss, game.CounterPlusOne); got != 4 {
		t.Errorf("alliance: %d counters, want 4", got)
	}
	b13PlayAs(t, g, 1, "Their Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if got := counterCount(g, boss, game.CounterPlusOne); got != 4 {
		t.Errorf("an opponent's creature grew it: %d", got)
	}
	b15Destroy(t, g, boss)
	if got := cmdr2Named(g, me.ID, "Citizen"); got != 4 {
		t.Errorf("%d Citizens on death, want 4 (one per counter)", got)
	}
}

func TestCmdr2BossChauffeurWithNoCountersMakesNoCitizens(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	boss := b12Push(g, me.ID, "Boss's Chauffeur", "Creature — Elf Citizen", cmdr2BossChauffeurOracle, 2, 2)
	b15Destroy(t, g, boss)
	if got := cmdr2Named(g, me.ID, "Citizen"); got != 0 {
		t.Errorf("%d Citizens, want none without counters", got)
	}
}

func TestCmdr2AetherbornMarauderMovesChosenCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	b28AddCounter(t, g, a, 3)
	b28AddCounter(t, g, b, 2)
	b28AddCounter(t, g, theirs, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(b, "charge", 1) })
	marauder := cmdr2B23Cast(t, g, "Aetherborn Marauder", "Creature — Aetherborn Rogue", cmdr2AetherbornOracle, 2, 2)
	passPriorityAroundTable(t, g)
	answerNumber(t, g, me.ID, 2) // from A
	answerNumber(t, g, me.ID, 0) // from B
	passPriorityAroundTable(t, g)
	if got := counterCount(g, marauder, game.CounterPlusOne); got != 2 {
		t.Errorf("Marauder counters = %d, want 2", got)
	}
	if got := counterCount(g, a, game.CounterPlusOne); got != 1 {
		t.Errorf("A counters = %d, want 1", got)
	}
	if got := counterCount(g, b, game.CounterPlusOne); got != 2 {
		t.Errorf("B counters = %d, want 2 (zero chosen)", got)
	}
	if got := counterCount(g, theirs, game.CounterPlusOne); got != 2 {
		t.Errorf("an opponent's permanent lost counters: %d", got)
	}
	if got := counterCount(g, b, "charge"); got != 1 {
		t.Errorf("a non-+1/+1 counter moved: %d", got)
	}
}

func TestCmdr2AetherbornMarauderWithNothingToMoveDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	marauder := cmdr2B23Cast(t, g, "Aetherborn Marauder", "Creature — Aetherborn Rogue", cmdr2AetherbornOracle, 2, 2)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, marauder, game.CounterPlusOne); got != 0 {
		t.Errorf("counters = %d", got)
	}
	if payAmountPrompt(g, g.Seats[0].ID) != nil {
		t.Error("asked about counters with none to move")
	}
}

func TestCmdr2GenesisHydraEntersWithXCountersAndPutsAPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cheap := seedLibraryTop(me, "Cheap Bear", "Creature — Bear", "{1}{G}")
	pricey := seedLibraryTop(me, "Big Rock", "Artifact", "{6}")
	land := seedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	hydra := castXSpell(t, g, "Genesis Hydra", "Creature — Plant Hydra", cmdr2GenesisHydraOracle, "{X}{G}{G}", 3, nil)
	passPriorityAroundTable(t, g)
	// The cast trigger resolved first: pick the cheap permanent.
	pick := latestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no put-onto-the-battlefield prompt")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	if !offered[cheap] || offered[pricey] || offered[land] || len(offered) != 1 {
		t.Fatalf("offered %v, want only the cheap nonland permanent", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{cheap}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, hydra, game.CounterPlusOne); got != 3 {
		t.Errorf("Hydra counters = %d, want 3", got)
	}
	if z := b12ZoneOf(g, cheap); z != game.ZoneBattlefield {
		t.Errorf("Cheap Bear in %v, want the battlefield", z)
	}
	if z := b12ZoneOf(g, pricey); z != game.ZoneLibrary {
		t.Errorf("Big Rock (mana value 6) in %v, want the library", z)
	}
	if z := b12ZoneOf(g, land); z != game.ZoneLibrary {
		t.Errorf("a land in %v, want the library", z)
	}
}

func TestCmdr2GenesisHydraAtXZeroRevealsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	top := seedLibraryTop(me, "Cheap Bear", "Creature — Bear", "{1}{G}")
	castXSpell(t, g, "Genesis Hydra", "Creature — Plant Hydra", cmdr2GenesisHydraOracle, "{X}{G}{G}", 0, nil)
	passPriorityAroundTable(t, g)
	if z := b12ZoneOf(g, top); z != game.ZoneLibrary {
		t.Errorf("top card moved to %v at X = 0", z)
	}
}

func TestCmdr2CommunalBrewingDrawsForOpponentsAndCountersCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand := opp.Hand.Size()
	brew := castCatalogSpell(t, g, "Communal Brewing", "Enchantment", cmdr2CommunalBrewingOracle, nil)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetPlayer, ID: g.Seats[2].ID},
	}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != hand+1 {
		t.Errorf("opponent hand %d -> %d, want +1", hand, got)
	}
	if got := counterCount(g, brew, "ingredient"); got != 3 {
		t.Errorf("ingredient counters = %d, want 3 (one, plus one per card drawn by two opponents)", got)
	}
	bear := cmdr2B23Cast(t, g, "Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, game.CounterPlusOne); got != 3 {
		t.Errorf("a creature cast: %d +1/+1 counters, want 3", got)
	}
}

func TestCmdr2CommunalBrewingWithNoTargetsHasOneCounterAndNoncreatureGetsNothing(t *testing.T) {
	g := newCatalogGame(t)
	brew := b12Push(g, g.Seats[0].ID, "Communal Brewing", "Enchantment", cmdr2CommunalBrewingOracle, 0, 0)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(brew, "ingredient", 2) })
	rock := castCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, rock, game.CounterPlusOne); got != 0 {
		t.Errorf("a noncreature spell got %d counters", got)
	}
}

func TestCmdr2XavierSalPopulatesByRemovingACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	xavier := b12Push(g, me.ID, "Xavier Sal, Infested Captain", "Legendary Creature — Human Fungus Pirate", cmdr2XavierSalOracle, 3, 3)
	token := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Token Creature — Goblin",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	b28AddCounter(t, g, token, 1)
	advanceToMain(t, g)
	// Only "another" permanent: Xavier's own counters cannot pay.
	b28AddCounter(t, g, xavier, 1)
	p7Activate(t, g, me, xavier, 0, game.ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{token}, CounterCounts: []int{1}, CounterKinds: []string{"+1/+1"}})
	if got := cmdr2Named(g, me.ID, "Goblin"); got != 2 {
		t.Errorf("%d Goblins, want 2 after populate", got)
	}
	if got := counterCount(g, token, "+1/+1"); got != 0 {
		t.Errorf("counter not removed: %d", got)
	}
}

func TestCmdr2XavierSalRefusesItsOwnCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	xavier := b12Push(g, me.ID, "Xavier Sal, Infested Captain", "Legendary Creature — Human Fungus Pirate", cmdr2XavierSalOracle, 3, 3)
	b28AddCounter(t, g, xavier, 1)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, xavier, 0, game.ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{xavier}, CounterCounts: []int{1}, CounterKinds: []string{"+1/+1"}}); err == nil {
		t.Error("paid with its own counter")
	}
}

func TestCmdr2XavierSalSacrificesAnotherCreatureToProliferate(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	xavier := b12Push(g, me.ID, "Xavier Sal, Infested Captain", "Legendary Creature — Human Fungus Pirate", cmdr2XavierSalOracle, 3, 3)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	grown := b12Creature(g, me.ID, "Grown", "Creature — Bear", 2, 2)
	b28AddCounter(t, g, grown, 1)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, xavier, 1, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{xavier}}); err == nil {
		t.Error("sacrificed itself")
	}
	p7Activate(t, g, me, xavier, 1, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	settleAnsweringProliferate(t, g)
	if z := b12ZoneOf(g, fodder); z != game.ZoneGraveyard {
		t.Errorf("fodder in %v, want graveyard", z)
	}
	if got := counterCount(g, grown, game.CounterPlusOne); got != 2 {
		t.Errorf("proliferate: %d counters, want 2", got)
	}
}

func TestCmdr2FerraforMakesSaprolingsForCountersAndDoubles(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	b28AddCounter(t, g, theirs, 3)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(theirs, "-1/-1", 1) })
	noncreature := b12Permanent(g, opp.ID, "Rock", "Artifact")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(noncreature, "charge", 5) })
	ferrafor := cmdr2B23Cast(t, g, "Ferrafor, Young Yew", "Legendary Creature — Treefolk Druid", cmdr2FerraforOracle, 4, 7)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	// Three +1/+1 and one -1/-1 cancel by CR 704.5q to two counters; a
	// noncreature's five charge counters do not count.
	if got := cmdr2Named(g, me.ID, "Saproling"); got != 2 {
		t.Errorf("%d Saprolings, want 2 (the counters among their creatures)", got)
	}
	// {T}: double the counters on target creature.
	b28AddCounter(t, g, ferrafor, 2)
	cmdr2Unsick(g, ferrafor)
	advanceToMain(t, g)
	p7Activate(t, g, me, ferrafor, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{cardRef(ferrafor)}})
	if got := counterCount(g, ferrafor, game.CounterPlusOne); got != 4 {
		t.Errorf("doubled counters = %d, want 4", got)
	}
}

func TestCmdr2LlanowarRebornEntersTappedWithACounterAndGrafts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := playLandFromHand(t, g, "Llanowar Reborn", cmdr2LlanowarRebornOracle)
	passPriorityAroundTable(t, g)
	if !b31Tapped(t, g, land) {
		t.Error("the land did not enter tapped")
	}
	if got := counterCount(g, land, game.CounterPlusOne); got != 1 {
		t.Fatalf("land counters = %d, want 1 (graft 1)", got)
	}
	// Another creature enters (an opponent's too): you may move the counter.
	bear := cmdr2B23Cast(t, g, "Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("creature counters = %d, want 1 after the move", got)
	}
	if got := counterCount(g, land, game.CounterPlusOne); got != 0 {
		t.Errorf("land counters = %d, want 0 after the move", got)
	}
	// With no counter left the ability no longer triggers.
	cmdr2B23Cast(t, g, "Bear Two", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	if len(g.PendingTriggers) != 0 {
		t.Error("graft triggered with no counter on the land")
	}
}

func TestCmdr2LlanowarRebornMayDeclineAndWorksOnOpponentCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := playLandFromHand(t, g, "Llanowar Reborn", cmdr2LlanowarRebornOracle)
	passPriorityAroundTable(t, g)
	cmdr2B23Cast(t, g, "Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, land, game.CounterPlusOne); got != 1 {
		t.Errorf("declined, yet the counter moved: land has %d", got)
	}
	theirs := b13PlayAs(t, g, 1, "Their Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, theirs, game.CounterPlusOne); got != 1 {
		t.Errorf("an opponent's creature entering: %d counters, want 1", got)
	}
}
