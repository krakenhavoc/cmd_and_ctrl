package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// repeating_delayed_cards_test.go — #2169 / CR 603.7b: the cards that set
// up a trigger for the rest of the turn.

const (
	greatTrainHeistOracle   = "afe2f7a4-9440-4d93-801f-a18b627efb21"
	bubblingMuckOracle      = "3c7f2b29-9f42-41ab-a2d4-7a450fb0242d"
	highTideOracle          = "dc671205-f2fa-454f-9957-921a6069ad53"
	glimpseOfNatureOracle   = "5fae52ed-1b84-406e-951d-b0d329dc48a7"
	benefactorsDraughtOracl = "85113e73-26f7-4938-b7a7-607e704af0ca"
)

func rdCountNamed(g *game.Game, name string, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == name && c.Controller == controller {
			n++
		}
	}
	return n
}

func emitCombatHit(g *game.Game, source, target uuid.UUID) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: source, Target: target, Amount: 2, Combat: true})
	})
}

func TestGreatTrainHeistTreasureForEveryCreatureThatConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)

	castModal(t, g, "Great Train Heist", "Instant", greatTrainHeistOracle,
		[]int{2}, []game.TargetRef{modeRef(game.TargetPlayer, opp.ID, 0, 0)})
	passPriorityAroundTable(t, g)

	emitCombatHit(g, a, opp.ID)
	passPriorityAroundTable(t, g)
	emitCombatHit(g, b, opp.ID)
	passPriorityAroundTable(t, g)
	if n := rdCountNamed(g, "Treasure", me.ID); n != 2 {
		t.Fatalf("%d Treasures after two creatures connected, want 2", n)
	}
	var tapped int
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Treasure" && c.Tapped {
			tapped++
		}
	}
	if tapped != 2 {
		t.Errorf("%d of the Treasures are tapped, want both", tapped)
	}

	// Another player is not "that player".
	emitCombatHit(g, a, third.ID)
	passPriorityAroundTable(t, g)
	// A creature an opponent controls is not "a creature you control".
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	emitCombatHit(g, theirs, opp.ID)
	passPriorityAroundTable(t, g)
	if n := rdCountNamed(g, "Treasure", me.ID); n != 2 {
		t.Errorf("%d Treasures after non-matching hits, want still 2", n)
	}

	// Gone with the turn.
	advanceToUpkeepOf(t, g, 1)
	emitCombatHit(g, a, opp.ID)
	passPriorityAroundTable(t, g)
	if n := rdCountNamed(g, "Treasure", me.ID); n != 2 {
		t.Errorf("%d Treasures after the turn ended, want still 2", n)
	}
}

func TestGreatTrainHeistPumpsAndGrantsFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	castModal(t, g, "Great Train Heist", "Instant", greatTrainHeistOracle, []int{1}, nil)
	passPriorityAroundTable(t, g)

	if got := effectivePower(t, g, mine); got != 3 {
		t.Errorf("my creature's power = %d, want 3", got)
	}
	if got := effectivePower(t, g, theirs); got != 2 {
		t.Errorf("their creature's power = %d, want 2", got)
	}
	has := false
	for _, a := range effectiveAbilities(t, g, mine) {
		has = has || a == "first strike"
	}
	if !has {
		t.Error("my creature did not gain first strike")
	}
}

func TestGreatTrainHeistUntapsAllYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tapped := func(owner uuid.UUID, name string) uuid.UUID {
		return pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear",
			Power: 2, Toughness: 2, Tapped: true, Owner: owner, Controller: owner})
	}
	mine := tapped(me.ID, "Bear A")
	theirs := tapped(opp.ID, "Their Bear")

	castModal(t, g, "Great Train Heist", "Instant", greatTrainHeistOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)

	var mineTapped, theirsTapped bool
	for _, c := range g.BattlefieldCardsForEffect() {
		switch c.InstanceID {
		case mine:
			mineTapped = c.Tapped
		case theirs:
			theirsTapped = c.Tapped
		}
	}
	if mineTapped {
		t.Error("my creature is still tapped")
	}
	if !theirsTapped {
		t.Error("their creature was untapped; only yours do")
	}
}

func TestBubblingMuckAndHighTideAddForTheTurn(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, land, color string
	}{
		{"Bubbling Muck", "Sorcery", bubblingMuckOracle, "Swamp", "B"},
		{"High Tide", "Instant", highTideOracle, "Island", "U"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			mine := b12Permanent(g, me.ID, tc.land, "Basic Land — "+tc.land)
			theirs := b12Permanent(g, opp.ID, tc.land, "Basic Land — "+tc.land)
			plains := b12Permanent(g, me.ID, "Plains", "Basic Land — Plains")

			castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			passPriorityAroundTable(t, g)

			if err := g.ActivateManaAbility(me.ID, mine, 0, game.ManaAbilityParams{}); err != nil {
				t.Fatal(err)
			}
			if got := poolCount(me, tc.color); got != 2 {
				t.Errorf("my %s mana = %d, want 2", tc.color, got)
			}
			if err := g.ActivateManaAbility(opp.ID, theirs, 0, game.ManaAbilityParams{}); err != nil {
				t.Fatal(err)
			}
			if got := poolCount(opp, tc.color); got != 2 {
				t.Errorf("their %s mana = %d, want 2: it is every player's", tc.color, got)
			}
			if err := g.ActivateManaAbility(me.ID, plains, 0, game.ManaAbilityParams{}); err != nil {
				t.Fatal(err)
			}
			if got := poolCount(me, tc.color); got != 2 {
				t.Errorf("a Plains changed my %s mana to %d", tc.color, got)
			}
			if len(g.PendingTriggers) != 0 || len(g.StackMeta) != 0 {
				t.Error("the triggered mana ability used the stack")
			}

			advanceToUpkeepOf(t, g, 1)
			again := b12Permanent(g, opp.ID, tc.land, "Basic Land — "+tc.land)
			if err := g.ActivateManaAbility(opp.ID, again, 0, game.ManaAbilityParams{}); err != nil {
				t.Fatal(err)
			}
			if got := poolCount(opp, tc.color); got != 1 {
				t.Errorf("next turn's %s mana = %d, want 1: it ended at cleanup", tc.color, got)
			}
		})
	}
}

func TestGlimpseOfNatureDrawsForEveryCreatureSpellThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")

	castCatalogSpell(t, g, "Glimpse of Nature", "Sorcery", glimpseOfNatureOracle, nil)
	passPriorityAroundTable(t, g)
	before := handSize(me)

	for i := 0; i < 2; i++ {
		g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventCast, Actor: me.ID, CardID: bear}) })
		passPriorityAroundTable(t, g)
	}
	if got := handSize(me) - before; got != 2 {
		t.Fatalf("drew %d cards for two creature spells, want 2", got)
	}
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventCast, Actor: me.ID, CardID: rock}) })
	passPriorityAroundTable(t, g)
	if got := handSize(me) - before; got != 2 {
		t.Errorf("a noncreature spell drew a card (%d total)", got)
	}
}

func TestBenefactorsDraughtUntapsAllAndDrawsPerOpposingBlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Tapped: true, Owner: opp.ID, Controller: opp.ID})
	before := handSize(me)

	castCatalogSpell(t, g, "Benefactor's Draught", "Instant", benefactorsDraughtOracl, nil)
	passPriorityAroundTable(t, g)

	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == theirs && c.Tapped {
			t.Error("an opponent's creature was not untapped: it is all creatures")
		}
	}
	if got := handSize(me) - before; got != 1 {
		t.Fatalf("drew %d from the spell itself, want 1", got)
	}
	// Two opposing blockers draw twice; my own blocker does not.
	block := func(actor, blocker uuid.UUID) {
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: actor, CardID: blocker, Target: mine, Amount: 1})
		})
		passPriorityAroundTable(t, g)
	}
	block(opp.ID, theirs)
	block(opp.ID, uuid.New())
	block(me.ID, mine)
	if got := handSize(me) - before; got != 3 {
		t.Errorf("after two opposing blocks hand grew by %d, want 3 in all", got)
	}
}
