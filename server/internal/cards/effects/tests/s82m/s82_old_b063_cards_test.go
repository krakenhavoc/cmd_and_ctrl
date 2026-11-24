package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b063_cards_test.go — Kaldheim, batch 063 (foretell): each test plays
// the real catalog entry through the engine.

const (
	oldB063Jarl      = "294c1fed-9473-4fbe-b10e-71a7a432b98d"
	oldB063Shepherd  = "2c5aa155-646c-47f6-ba5c-6e1b41fd4bb2"
	oldB063Starnheim = "495ebe60-f0e5-4f0f-9d99-b2c64c95dbf4"
	oldB063Charger   = "603e518c-8f12-432c-ae03-790db5be77bf"
	oldB063SkullRaid = "7d22c7ed-2f45-42af-b999-b921f4f4db25"
	oldB063Struggle  = "a3bdb026-e09d-443f-ba02-fe0d5412fc97"
	oldB063DualStrk  = "a498ca5f-7743-47b7-a57a-471efa1a99e8"
	oldB063Tide      = "b0bd6135-4257-4f71-bdfe-71fb3f7a36a5"
	oldB063Epiphany  = "bc842838-1993-4480-b8fe-f9d9f2dee7a4"
)

func TestOldB063EveryCardIsRegisteredFullWithItsForetellCost(t *testing.T) {
	for oracle, tc := range map[string]struct{ name, cost string }{
		oldB063Jarl:      {"Jarl of the Forsaken", "{1}{B}"},
		oldB063Shepherd:  {"Shepherd of the Cosmos", "{3}{W}"},
		oldB063Starnheim: {"Starnheim Unleashed", "{X}{X}{W}"},
		oldB063Charger:   {"Cosmos Charger", "{2}{U}"},
		oldB063SkullRaid: {"Skull Raid", "{1}{B}"},
		oldB063Struggle:  {"Struggle for Skemfar", "{G}"},
		oldB063DualStrk:  {"Dual Strike", "{R}"},
		oldB063Tide:      {"Return Upon the Tide", "{3}{B}"},
		oldB063Epiphany:  {"Alrund's Epiphany", "{4}{U}{U}"},
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != tc.name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v registered=%v, want Full with no caveats", tc.name, spec, ok)
			continue
		}
		found := false
		for _, sa := range spec.SpecialActions {
			if sa.Kind == game.SpecialActionForetell && sa.CastCost == tc.cost {
				found = true
			}
		}
		if !found {
			t.Errorf("%s declares no foretell %s", tc.name, tc.cost)
		}
	}
}

// oldB063Foretell foretells a hand card on the active seat's turn, then moves
// to a later turn (the card may not be cast the turn it was foretold).
func oldB063Foretell(t *testing.T, g *game.Game, me *game.Player, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	card := effectstest.HandCardForTest(me, name, typeLine, oracle)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.PerformSpecialAction(me.ID, card, game.SpecialActionForetell, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("foretell %s: %v", name, err)
	}
	g.WithWriteLock(func() { g.Turn.Seq++ })
	return card
}

func TestOldB063JarlDestroysOnlyADamagedOpposingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hurt := effectstest.B12Creature(g, opp.ID, "Hurt Ogre", "Creature — Ogre", 5, 5)
	whole := effectstest.B12Creature(g, opp.ID, "Whole Ogre", "Creature — Ogre", 5, 5)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(uuid.Nil, hurt, 1)
		_ = g.DealDamageToCreatureForEffect(uuid.Nil, mine, 1)
	})
	effectstest.CastRingCreature(t, g, "Jarl of the Forsaken", "Creature — Zombie Cleric", oldB063Jarl, 3, 2)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	for _, bad := range []uuid.UUID{whole, mine} {
		if err := g.ResolvePickTarget(p.ID, p.Chooser, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Errorf("target %v is not a damaged creature an opponent controls", bad)
		}
	}
	effectstest.AnswerPickTarget(t, g, hurt)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hurt) {
		t.Error("the damaged creature survived")
	}
	if !g.Battlefield.Contains(whole) {
		t.Error("the undamaged creature died")
	}
}

func TestOldB063ShepherdReturnsAPermanentCardWithManaValueTwoOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cheap := uuid.New()
	dear := uuid.New()
	spell := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: cheap, Name: "Cheap Relic", TypeLine: "Artifact", ManaCost: "{2}", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: dear, Name: "Dear Beast", TypeLine: "Creature — Beast", ManaCost: "{2}{G}", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Cheap Bolt", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	effectstest.CastRingCreature(t, g, "Shepherd of the Cosmos", "Creature — Angel Warrior", oldB063Shepherd, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	for _, bad := range []uuid.UUID{dear, spell} {
		if err := g.ResolvePickTarget(p.ID, p.Chooser, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Errorf("target %v should be illegal", bad)
		}
	}
	effectstest.AnswerPickTarget(t, g, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Error("the artifact card did not return to the battlefield")
	}
}

func TestOldB063StruggleForSkemfarPutsACounterThenFights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Imp", "Creature — Imp", 2, 3)
	effectstest.ToMainForCost(t, g)
	effectstest.B06AddMana(me, "G", "G", "G", "G")
	id := effectstest.HandCardFull(me, "Struggle for Skemfar", "Sorcery", "{3}{G}", oldB063Struggle, []string{"G"})
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, mine, "+1/+1"); n != 1 {
		t.Errorf("my creature has %d counters, want 1", n)
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the 2/3 survived 3 damage from the fight")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my 3/3 died to 2 damage")
	}
}

func TestOldB063StruggleForSkemfarWithNoSecondTargetStillGivesTheCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.ToMainForCost(t, g)
	effectstest.B06AddMana(me, "G", "G", "G", "G")
	id := effectstest.HandCardFull(me, "Struggle for Skemfar", "Sorcery", "{3}{G}", oldB063Struggle, []string{"G"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, mine, "+1/+1"); n != 1 {
		t.Errorf("my creature has %d counters, want 1", n)
	}
}

func TestOldB063ReturnUponTheTideMakesElfWarriorsOnlyForAnElf(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		tokens         int
	}{{"Elf Druid", "Creature — Elf Druid", 2}, {"Bear", "Creature — Bear", 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			dead := effectstest.PushGraveyardCardTyped(me, tc.name, tc.typeLine)
			effectstest.CastCatalogSpell(t, g, "Return Upon the Tide", "Sorcery", oldB063Tide, effectstest.CardTarget(dead))
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(dead) {
				t.Fatal("the creature card did not return")
			}
			if n, _ := effectstest.CountTokensNamed(g, me.ID, "Elf Warrior"); n != tc.tokens {
				t.Errorf("%d Elf Warrior tokens, want %d", n, tc.tokens)
			}
		})
	}
}

func TestOldB063AlrundsEpiphanyMakesBirdsTakesAnExtraTurnAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	spell := effectstest.CastCatalogSpell(t, g, "Alrund's Epiphany", "Sorcery", oldB063Epiphany, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, g.Seats[me].ID, "Bird"); n != 2 {
		t.Errorf("%d Birds, want 2", n)
	}
	if g.Seats[me].Graveyard.Contains(spell) || !g.Exile.Contains(spell) {
		t.Error("Alrund's Epiphany should exile itself")
	}
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, me, true, "after Alrund's Epiphany")
}

func TestOldB063SkullRaidDrawsTheDifferenceWhenTheyHaveFewerCards(t *testing.T) {
	for _, tc := range []struct {
		hand  int
		draws int
	}{{3, 0}, {1, 1}, {0, 2}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		opp.Hand.Cards = nil
		var hand []uuid.UUID
		for i := 0; i < tc.hand; i++ {
			hand = append(hand, effectstest.PushGraveyardCardTyped(opp, "x", "Instant"))
		}
		// PushGraveyardCardTyped puts them in the graveyard; move them to the hand.
		opp.Graveyard.Cards = nil
		for _, id := range hand {
			opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Card", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
		}
		effectstest.ToMainForCost(t, g)
		effectstest.B06AddMana(me, "B", "B", "B", "B")
		id := effectstest.HandCardFull(me, "Skull Raid", "Sorcery", "{3}{B}", oldB063SkullRaid, []string{"B"})
		mine := me.Hand.Size()
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if pick := effectstest.LatestChooseCardsFor(g, opp.ID); pick != nil {
			n := min(2, len(pick.ChooseCards))
			if err := g.ResolveChooseCards(pick.ID, opp.ID, pick.ChooseCards[:n]); err != nil {
				t.Fatalf("discard: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
		}
		if opp.Hand.Size() != max(0, tc.hand-2) {
			t.Errorf("hand %d: the opponent has %d cards left", tc.hand, opp.Hand.Size())
		}
		if got := me.Hand.Size(); got != mine-1+tc.draws {
			t.Errorf("hand %d: I drew to %d, want %d (%d draws)", tc.hand, got, mine-1+tc.draws, tc.draws)
		}
	}
}

func TestOldB063StarnheimUnleashedMakesOneAngelOrXWhenForetold(t *testing.T) {
	t.Run("from hand", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.CastCatalogSpell(t, g, "Starnheim Unleashed", "Sorcery", oldB063Starnheim, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if n, _ := effectstest.CountTokensNamed(g, me.ID, "Angel Warrior"); n != 1 {
			t.Errorf("%d Angel Warrior tokens, want 1", n)
		}
	})
	t.Run("foretold", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		card := oldB063Foretell(t, g, me, "Starnheim Unleashed", "Sorcery", oldB063Starnheim)
		me.ManaPool.AddMana(game.ManaToken{Color: "W"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
		if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: game.AltCostKeyForetell, XValue: 2}); err != nil {
			t.Fatalf("cast the foretold spell for X=2: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if n, _ := effectstest.CountTokensNamed(g, me.ID, "Angel Warrior"); n != 2 {
			t.Errorf("%d Angel Warrior tokens, want X = 2", n)
		}
	})
}

func TestOldB063CosmosChargerCutsForetellCostAndOpensOtherTurns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Cosmos Charger", "Creature — Horse Spirit", oldB063Charger, false)
	// During an opponent's turn the foretell action is open to its controller.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	card := effectstest.HandCardForTest(me, "Struggle for Skemfar", "Sorcery", oldB063Struggle)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.PerformSpecialAction(me.ID, card, game.SpecialActionForetell, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("foretell for {1} on another player's turn: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left in the pool, want the {1} spent", me.ManaPool.Total())
	}
	// Without the Charger it stays closed on another player's turn.
	other := effectstest.HandCardForTest(opp, "Struggle for Skemfar", "Sorcery", oldB063Struggle)
	_ = other
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.AdvanceToUpkeepOf(t, g2, 1)
	c2 := effectstest.HandCardForTest(me2, "Struggle for Skemfar", "Sorcery", oldB063Struggle)
	me2.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g2.PerformSpecialAction(me2.ID, c2, game.SpecialActionForetell, game.SpecialActionParams{Strict: true}); err == nil {
		t.Error("foretold during another player's turn with no Charger")
	}
}

func TestOldB063DualStrikeCopiesOnlyTheNextCheapInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	effectstest.B06AddMana(me, "R", "R", "R", "R", "R", "R", "R", "R", "R", "R", "R", "R")
	strike := effectstest.HandCardFull(me, "Dual Strike", "Instant", "{R}{R}", oldB063DualStrk, []string{"R"})
	if err := g.CastSpell(me.ID, strike, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast Dual Strike: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	// A mana value 5 instant is not copied.
	big := effectstest.HandCardFull(me, "Big Spell", "Instant", "{4}{R}", "", []string{"R"})
	if err := g.CastSpell(me.ID, big, game.CastSpellParams{Strict: true, Targets: target}); err != nil {
		t.Fatalf("cast the big spell: %v", err)
	}
	if n := len(g.StackMeta); n != 1 {
		t.Fatalf("%d items on the stack after a mana value 5 spell, want just the spell", n)
	}
	// A mana value 1 instant is, once; the next one is not.
	bolt := effectstest.HandCardFull(me, "Test Bolt", "Instant", "{R}", "", []string{"R"})
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Strict: true, Targets: target}); err != nil {
		t.Fatalf("cast the bolt: %v", err)
	}
	if n := len(g.StackMeta); n != 3 {
		t.Fatalf("%d items on the stack after the cheap spell, want the big spell, the bolt and the bolt's copy", n)
	}
	bolt2 := effectstest.HandCardFull(me, "Test Bolt Two", "Instant", "{R}", "", []string{"R"})
	if err := g.CastSpell(me.ID, bolt2, game.CastSpellParams{Strict: true, Targets: target}); err != nil {
		t.Fatalf("cast the second bolt: %v", err)
	}
	if n := len(g.StackMeta); n != 4 {
		t.Errorf("%d items on the stack, want 4: the trigger fires only for the next spell", n)
	}
}
