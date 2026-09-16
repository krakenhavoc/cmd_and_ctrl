package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// proliferate_test.go — S21's Proliferate primitive (CR 701.34) and
// the five catalog cards that were gated on it.

const (
	steadyProgressOracle = "d2145ad3-fe7e-459b-b155-1b3ed089e936"
	karnsBastionOracle   = "9fb8cd81-403a-4988-8f1c-b8eccf8abd9c"
	contagionClaspOracle = "43f2d81e-aa01-4fa9-9046-6a27a05dbd2d"
	fluxChannelerOracle  = "83874e60-291b-47a7-ba9f-69437fa7e3c7"
	evolutionSageOracle  = "b45fdeab-00cc-4422-af9f-66f30a880a7c"
	inexorableTideOracle = "40ec0a47-badf-4074-b0a8-749bb7c17b95"
)

// pushCounterCreature seeds a 4/4 so a -1/-1 counter (or two) does
// not immediately kill it — pushCatalogPermanent's 1/1 dies to the
// CR 704.5f SBA the moment the first counter lands, which would make
// every assertion below read "card not found".
func pushCounterCreature(g *game.Game, owner uuid.UUID, name string, counter string, n int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Bear",
		Power: 4, Toughness: 4, Owner: owner, Controller: owner,
	})
	if n > 0 {
		_ = g.AddCounter(id, counter, n)
	}
	return id
}

func counterCount(g *game.Game, id uuid.UUID, kind string) int {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Counters[kind]
		}
	}
	return -1
}

// The auto-pick is the whole sandbox simplification, so it gets the
// first and longest test: everything of mine that a counter helps,
// everything of theirs that a counter hurts, and nothing else.
func TestProliferatePickIsBeneficialAndAsymmetric(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	myGrowing := pushCounterCreature(g, me.ID, "Mine Growing", game.CounterPlusOne, 1)
	myShrinking := pushCounterCreature(g, me.ID, "Mine Shrinking", game.CounterMinusOne, 1)
	theirShrinking := pushCounterCreature(g, opp.ID, "Theirs Shrinking", game.CounterMinusOne, 1)
	theirGrowing := pushCounterCreature(g, opp.ID, "Theirs Growing", game.CounterPlusOne, 1)
	bare := pushCounterCreature(g, me.ID, "Bare", "", 0)

	if err := g.AddPlayerCounter(me.ID, game.CounterExperience, 1); err != nil {
		t.Fatalf("AddPlayerCounter: %v", err)
	}
	if err := g.AddPlayerCounter(opp.ID, game.CounterPoison, 2); err != nil {
		t.Fatalf("AddPlayerCounter: %v", err)
	}
	if err := g.AddPlayerCounter(opp.ID, game.CounterExperience, 1); err != nil {
		t.Fatalf("AddPlayerCounter: %v", err)
	}

	cards, players := BeneficialProliferateChoice(g, me.ID)
	if len(cards) != 2 {
		t.Fatalf("chose %d permanents, want 2 (mine growing + theirs shrinking): %v", len(cards), cards)
	}
	chosen := map[uuid.UUID]bool{}
	for _, id := range cards {
		chosen[id] = true
	}
	if !chosen[myGrowing] || !chosen[theirShrinking] {
		t.Errorf("wrong permanents chosen: %v", cards)
	}
	if chosen[myShrinking] || chosen[theirGrowing] || chosen[bare] {
		t.Errorf("a counter that helps the wrong player (or no counter at all) must not be chosen: %v", cards)
	}
	// Me: experience only, all helpful. Opp: poison AND experience —
	// proliferate would hand them both, so they are not a legal
	// beneficial choice.
	if len(players) != 1 || players[0] != me.ID {
		t.Errorf("players chosen = %v, want just the proliferating player", players)
	}
}

func TestSteadyProgressProliferatesAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushCounterCreature(g, me.ID, "Mine", game.CounterPlusOne, 1)
	theirs := pushCounterCreature(g, opp.ID, "Theirs", game.CounterMinusOne, 1)
	if err := g.AddPlayerCounter(opp.ID, game.CounterPoison, 3); err != nil {
		t.Fatalf("AddPlayerCounter: %v", err)
	}

	before := len(me.Hand.Cards)
	castCatalogSpell(t, g, "Steady Progress", "Instant", steadyProgressOracle, nil)
	passPriorityAroundTable(t, g)

	if got := counterCount(g, mine, game.CounterPlusOne); got != 2 {
		t.Errorf("my +1/+1 counters = %d, want 2", got)
	}
	if got := counterCount(g, theirs, game.CounterMinusOne); got != 2 {
		t.Errorf("their -1/-1 counters = %d, want 2", got)
	}
	if opp.Counters[game.CounterPoison] != 4 {
		t.Errorf("opponent poison = %d, want 4", opp.Counters[game.CounterPoison])
	}
	if opp.Poison != 4 {
		t.Errorf("legacy Player.Poison = %d, want it mirrored to 4", opp.Poison)
	}
	// The spell was seeded into hand and cast (net zero), then the
	// card draw added one.
	if got := len(me.Hand.Cards); got != before+1 {
		t.Errorf("hand = %d, want %d (cast the spell, drew a card)", got, before+1)
	}
}

// Nothing to choose is a legal proliferate, not an error: no counters
// anywhere on the board means the spell still resolves and still
// draws.
func TestSteadyProgressWithNoCountersIsANoOp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := len(me.Hand.Cards)
	castCatalogSpell(t, g, "Steady Progress", "Instant", steadyProgressOracle, nil)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != before+1 {
		t.Errorf("hand = %d, want %d (cast the spell, drew a card)", got, before+1)
	}
}

func TestKarnsBastionProliferatesForFourAndATap(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bastion := pushCatalogPermanent(g, me.ID, "Karn's Bastion", "Land", karnsBastionOracle, false)
	mine := pushCounterCreature(g, me.ID, "Mine", game.CounterPlusOne, 2)

	if err := g.ActivateCatalogAbility(me.ID, bastion, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := counterCount(g, mine, game.CounterPlusOne); got != 2 {
		t.Errorf("proliferate must wait for resolution, counters = %d", got)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, mine, game.CounterPlusOne); got != 3 {
		t.Errorf("+1/+1 counters = %d, want 3", got)
	}
	// The mana ability is a separate list and must still be there —
	// a land whose only ability was the proliferate would be
	// unplayable in a real deck.
	if len(game.ManaAbilitiesForCard(game.Card{OracleID: karnsBastionOracle})) != 1 {
		t.Errorf("Karn's Bastion should keep its {T}: Add {C}")
	}
}

func TestContagionClaspShrinksThenProliferates(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushCounterCreature(g, opp.ID, "Victim", "", 0)

	clasp := castCatalogSpell(t, g, "Contagion Clasp", "Artifact", contagionClaspOracle, nil)
	// The ETB trigger targets, so it queues a pick before it can go
	// on the stack.
	for i := 0; i < 6 && latestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatalf("Contagion Clasp's ETB should ask for a target creature")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: victim}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, victim, game.CounterMinusOne); got != 1 {
		t.Fatalf("victim -1/-1 counters = %d, want 1", got)
	}

	// And the activated half grows the counter it just placed.
	if err := g.ActivateCatalogAbility(me.ID, clasp, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, victim, game.CounterMinusOne); got != 2 {
		t.Errorf("victim -1/-1 counters = %d, want 2", got)
	}
}

func TestFluxChannelerProliferatesOnNoncreatureSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Flux Channeler", "Creature — Human Wizard", fluxChannelerOracle, false)
	mine := pushCounterCreature(g, me.ID, "Mine", game.CounterPlusOne, 1)

	castCatalogSpell(t, g, "Some Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, mine, game.CounterPlusOne); got != 2 {
		t.Errorf("a noncreature spell should proliferate: counters = %d, want 2", got)
	}

	castCatalogSpell(t, g, "Some Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, mine, game.CounterPlusOne); got != 2 {
		t.Errorf("a CREATURE spell must not proliferate: counters = %d, want 2", got)
	}
}

func TestInexorableTideProliferatesOnEverySpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Inexorable Tide", "Enchantment", inexorableTideOracle, false)
	mine := pushCounterCreature(g, me.ID, "Mine", game.CounterPlusOne, 1)

	castCatalogSpell(t, g, "Some Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, mine, game.CounterPlusOne); got != 2 {
		t.Errorf("the Tide sees creature spells too: counters = %d, want 2", got)
	}
}

func TestEvolutionSageProliferatesOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Evolution Sage", "Creature — Elf Druid", evolutionSageOracle, false)
	mine := pushCounterCreature(g, me.ID, "Mine", game.CounterPlusOne, 1)

	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if got := counterCount(g, mine, game.CounterPlusOne); got != 2 {
		t.Errorf("landfall should proliferate: counters = %d, want 2", got)
	}
}
