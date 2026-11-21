package testsa

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const aangAirbendingMasterOracle = "f3176779-74f7-4136-a701-438feeead7a0"

// "Airbend ANOTHER target creature", mandatory: Aang is not in the
// legal set, the chosen creature is airbent, and its owner holds the
// {2} grant.
func TestAangAirbendingMasterAirbendsAnotherCreatureNotHimself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Their Blocker")

	aang := effectstest.CastAndResolveCreature(t, g, "Aang, Airbending Master",
		"Legendary Creature — Human Avatar Ally", aangAirbendingMasterOracle)
	if p := effectstest.LatestConfirmFor(g, me.ID); p != nil {
		t.Fatalf("the enters trigger is mandatory; it should not ask \"you may\"")
	}
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatalf("no target prompt for the enters trigger")
	}
	if effectstest.HasID(pick.PickTargetCards, aang) {
		t.Errorf("Aang is offered as his own target; the clause says ANOTHER")
	}
	if !effectstest.HasID(pick.PickTargetCards, victim) {
		t.Fatalf("the opponent's creature is not a legal target: %v", pick.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AssertAirbent(t, g, victim, opp.ID)
	if !g.Battlefield.Contains(aang) {
		t.Errorf("Aang left the battlefield")
	}
	// An opponent's creature leaving is not "a creature you control".
	if got := me.Counters[game.CounterExperience]; got != 0 {
		t.Errorf("experience = %d after airbending an opponent's creature, want 0", got)
	}
}

// With no other creature the trigger has no legal target and is
// removed (CR 603.3d) — no prompt, and Aang stays.
func TestAangAirbendingMasterAloneTriggersNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]

	aang := effectstest.CastAndResolveCreature(t, g, "Aang, Airbending Master",
		"Legendary Creature — Human Avatar Ally", aangAirbendingMasterOracle)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatalf("a target prompt opened with no other creature on the battlefield")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aang) {
		t.Errorf("Aang airbent himself")
	}
}

// Airbending your OWN creature is a departure without dying: one
// experience counter.
func TestAangAirbendingMasterAirbendingYourOwnCreatureGivesExperience(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "My Bear")

	effectstest.CastAndResolveCreature(t, g, "Aang, Airbending Master",
		"Legendary Creature — Human Avatar Ally", aangAirbendingMasterOracle)
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AssertAirbent(t, g, mine, me.ID)
	if got := me.Counters[game.CounterExperience]; got != 1 {
		t.Errorf("experience = %d, want 1", got)
	}
}

// "One or more … leave without dying": a mass bounce of three
// creatures (Aang among them — he counts himself) is ONE counter; a
// later single bounce is another; a death is none.
func TestAangAirbendingMasterExperienceIsOncePerBatchAndNotOnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	aang := effectstest.B12Push(g, me.ID, "Aang, Airbending Master", "Legendary Creature — Human Avatar Ally", aangAirbendingMasterOracle, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	doomed := effectstest.B12Creature(g, me.ID, "Doomed", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)

	// A death does not count (CR 700.4).
	effectstest.CastCatalogSpell(t, g, "Doom Blade", "Instant", effectstest.DoomBladeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: doomed}})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(doomed) {
		t.Fatalf("Doom Blade did not kill the creature")
	}
	if got := me.Counters[game.CounterExperience]; got != 0 {
		t.Fatalf("experience = %d after a creature died, want 0", got)
	}

	// One bounce: one counter.
	effectstest.CastCatalogSpell(t, g, "Unsummon", "Instant", effectstest.B829UnsummonOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Counters[game.CounterExperience]; got != 1 {
		t.Fatalf("experience = %d after one bounce, want 1", got)
	}

	// Put two creatures back and bounce the whole board, Aang included:
	// one batch, one counter — and Aang leaving counts for himself.
	effectstest.B12Creature(g, me.ID, "Second", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Third", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Evacuation", "Instant", effectstest.B829EvacuationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(aang) {
		t.Fatalf("Evacuation did not bounce Aang")
	}
	if got := me.Counters[game.CounterExperience]; got != 2 {
		t.Errorf("experience = %d after a mass bounce, want 2 (one per batch)", got)
	}
}

// Aang leaving alone still gives a counter: "one or more creatures you
// control" includes him (CR 603.10a look-back).
func TestAangAirbendingMasterCountsHimselfLeaving(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	aang := effectstest.B12Push(g, me.ID, "Aang, Airbending Master", "Legendary Creature — Human Avatar Ally", aangAirbendingMasterOracle, 4, 4)
	effectstest.AdvanceToMain(t, g)

	effectstest.CastCatalogSpell(t, g, "Unsummon", "Instant", effectstest.B829UnsummonOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: aang}})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(aang) {
		t.Fatalf("Unsummon did not bounce Aang")
	}
	if got := me.Counters[game.CounterExperience]; got != 1 {
		t.Errorf("experience = %d after Aang himself left, want 1", got)
	}
}

// The upkeep makes one 1/1 white Ally per experience counter, on its
// controller's upkeep only.
func TestAangAirbendingMasterUpkeepMakesAnAllyPerExperienceCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	effectstest.B12Push(g, me.ID, "Aang, Airbending Master", "Legendary Creature — Human Avatar Ally", aangAirbendingMasterOracle, 4, 4)
	if err := g.AddPlayerCounterForEffect(me.ID, game.CounterExperience, 3); err != nil {
		t.Fatalf("seed experience: %v", err)
	}

	countAllies := func() int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.IsToken() && c.Name == "Ally" {
				if c.Power != 1 || c.Toughness != 1 {
					t.Errorf("Ally token is %d/%d, want 1/1", c.Power, c.Toughness)
				}
				n++
			}
		}
		return n
	}

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := countAllies(); got != 3 {
		t.Fatalf("Allies after the upkeep = %d, want 3", got)
	}

	// Another player's upkeep makes nothing.
	effectstest.AdvanceToUpkeepOf(t, g, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := countAllies(); got != 3 {
		t.Errorf("Allies after an opponent's upkeep = %d, want still 3", got)
	}
}
