package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// coin_flip_wins_test.go — "whenever you win a coin flip" and the
// Vehicle tokens with crew (#3767): Setzer, Wandering Gambler (deck
// request #3655), Chance Encounter, Tavern Scoundrel and Mu Yanling,
// Wind Rider.

const (
	setzerOracle          = "acdd5082-7b3c-4fd3-941b-d19dd72698ec"
	chanceEncounterOracle = "8bd33b5c-0777-41ef-9e17-fb30ce69ec25"
	tavernScoundrelOracle = "68c4f86f-ff12-44a6-8f18-f2ee9753b2cd"
	muYanlingOracle       = "aaafcd69-6295-4d79-8e67-98df0c9bae88"
)

// emitCoinFlip fakes one coin's EventFlipCoin, which is all a "whenever
// you win a coin flip" trigger watches, and runs the priority boundary
// that puts a trigger it caused on the stack, as a real flip's
// resolution does.
func emitCoinFlip(g *game.Game, flipper uuid.UUID, won bool) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventFlipCoin, Actor: flipper, Label: "heads", Call: "heads", Won: won})
	})
	g.RunStateChecksForTest()
}

// battlefieldTokenNamed returns the first token with this name on the
// battlefield.
func battlefieldTokenNamed(t *testing.T, g *game.Game, name string) game.Card {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if IsToken(c) && c.Name == name {
			return c
		}
	}
	t.Fatalf("no %s token on the battlefield", name)
	return game.Card{}
}

// TestWinACoinFlipTriggersOnlyOnTheFlippersOwnWins pins the matcher:
// your won flip triggers, a lost flip, an opponent's won flip and a
// face-only flip that nobody wins do not.
func TestWinACoinFlipTriggersOnlyOnTheFlippersOwnWins(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Setzer, Wandering Gambler", setzerOracle, "Legendary Creature — Human Rogue Pilot")

	emitCoinFlip(g, me.ID, false)
	emitCoinFlip(g, opp.ID, true)
	g.WithWriteLock(func() {
		if _, err := g.FlipCoinsForEffect(game.RandomDraw{Player: me.ID}, 3); err != nil {
			t.Fatalf("FlipCoinsForEffect: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Treasure"); n != 0 {
		t.Fatalf("a lost, an opponent's or a face-only flip made %d Treasures", n)
	}

	emitCoinFlip(g, me.ID, true)
	emitCoinFlip(g, me.ID, true)
	passPriorityAroundTable(t, g)
	if n, tapped := countTokensNamed(g, me.ID, "Treasure"); n != 4 || tapped != 4 {
		t.Fatalf("two won flips: %d Treasures, %d tapped; want 4 and 4", n, tapped)
	}
}

// TestSetzerMakesTheBlackjackAVehicleThatCrewsAndFlies: the ETB token
// is a legendary 3/3 colorless Vehicle artifact with flying that is not
// a creature until crewed, and crew 2 makes it one.
func TestSetzerMakesTheBlackjackAVehicleThatCrewsAndFlies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Setzer, Wandering Gambler", "Legendary Creature — Human Rogue Pilot", setzerOracle, nil)
	passPriorityAroundTable(t, g)

	bj := battlefieldTokenNamed(t, g, "The Blackjack")
	if bj.Controller != me.ID || !bj.IsLegendary() || !bj.IsArtifact() || !bj.HasSubtype("Vehicle") {
		t.Fatalf("The Blackjack is %q under %s", bj.TypeLine, bj.Controller)
	}
	if bj.IsCreature() {
		t.Fatal("an uncrewed Vehicle token is a creature")
	}
	if len(bj.Colors) != 0 {
		t.Errorf("The Blackjack has colours %v, want colorless", bj.Colors)
	}
	if !effectiveAbilitiesContain(t, g, bj.InstanceID, "flying") {
		t.Error("The Blackjack has no flying")
	}

	crewer := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Two-Power Crewer", TypeLine: "Creature — Human",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	if err := g.ActivateCatalogAbility(me.ID, bj.InstanceID, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{crewer}}); err != nil {
		t.Fatalf("crew 2: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, _ := g.LookupCardForEffect(bj.InstanceID)
	if !c.IsCreature() || effectivePower(t, g, bj.InstanceID) != 3 || effectiveToughness(t, g, bj.InstanceID) != 3 {
		t.Fatalf("crewed Blackjack: creature=%v %d/%d, want a 3/3 creature", c.IsCreature(),
			effectivePower(t, g, bj.InstanceID), effectiveToughness(t, g, bj.InstanceID))
	}
}

// TestSetzerFlipsForEachVehicleThatConnects: a Vehicle you control
// dealing combat damage to a player asks for a coin call; a won call
// makes two tapped Treasures and a lost one makes none. A non-Vehicle
// creature connecting does not flip.
func TestSetzerFlipsForEachVehicleThatConnects(t *testing.T) {
	sawWin, sawLoss := false, false
	for i := 0; i < 64 && !(sawWin && sawLoss); i++ {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		pushPermanentForTest(g, me.ID, "Setzer, Wandering Gambler", setzerOracle, "Legendary Creature — Human Rogue Pilot")
		bear := pushPermanentForTest(g, me.ID, "Grizzly Bears", "", "Creature — Bear")
		copter := pushPermanentForTest(g, me.ID, "Crewed Copter", "", "Artifact Creature — Vehicle")

		dealCombatDamageToPlayer(g, bear, opp.ID, 2)
		passPriorityAroundTable(t, g)
		if latestChoiceOfKind(g, game.PendingChoiceCoinCall) != nil {
			t.Fatal("a non-Vehicle creature's combat damage asked for a coin call")
		}

		dealCombatDamageToPlayer(g, copter, opp.ID, 3)
		passPriorityAroundTable(t, g)
		flip := latestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if flip == nil || flip.Chooser != me.ID {
			t.Fatalf("the Vehicle's damage asked for no coin call from Setzer's controller: %+v", flip)
		}
		if err := g.ResolveCoinCall(flip.ID, me.ID, "heads"); err != nil {
			t.Fatalf("ResolveCoinCall: %v", err)
		}
		won := false
		for _, ev := range randomEvents(g, game.EventFlipCoin) {
			won = ev.Won
		}
		passPriorityAroundTable(t, g)
		n, tapped := countTokensNamed(g, me.ID, "Treasure")
		if won {
			sawWin = true
			if n != 2 || tapped != 2 {
				t.Fatalf("won flip: %d Treasures, %d tapped; want 2 tapped", n, tapped)
			}
		} else {
			sawLoss = true
			if n != 0 {
				t.Fatalf("lost flip made %d Treasures", n)
			}
		}
	}
	if !sawWin || !sawLoss {
		t.Fatalf("64 flips never both won and lost (win %v, loss %v)", sawWin, sawLoss)
	}
}

// TestChanceEncounterWinsAtTenLuckCounters: each won flip is one luck
// counter, nine is not enough at upkeep, and ten wins the game.
func TestChanceEncounterWinsAtTenLuckCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ce := pushPermanentForTest(g, me.ID, "Chance Encounter", chanceEncounterOracle, "Enchantment")

	for i := 0; i < 9; i++ {
		emitCoinFlip(g, me.ID, true)
		passPriorityAroundTable(t, g)
	}
	emitCoinFlip(g, me.ID, false)
	passPriorityAroundTable(t, g)
	c, _ := g.LookupCardForEffect(ce)
	if got := c.Counters["luck"]; got != 9 {
		t.Fatalf("nine wins and a loss left %d luck counters, want 9", got)
	}

	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	if triggerOnStack(g, ce) != nil || g.State != game.StateActive {
		t.Fatalf("nine luck counters triggered the win (state %v, outcome %+v)", g.State, g.Outcome)
	}

	emitCoinFlip(g, me.ID, true)
	passPriorityAroundTable(t, g)
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if g.Outcome == nil || g.Outcome.Winner != me.ID {
		t.Fatalf("ten luck counters at upkeep: outcome %+v, want a win for seat 0", g.Outcome)
	}
}

// TestTavernScoundrelFlipsForTreasure: the activated ability sacrifices
// another permanent and asks for a call; a win makes two untapped
// Treasures.
func TestTavernScoundrelFlipsForTreasure(t *testing.T) {
	sawWin, sawLoss := false, false
	for i := 0; i < 64 && !(sawWin && sawLoss); i++ {
		g := newCatalogGame(t)
		me := g.Seats[0]
		ts := pushCatalogPermanent(g, me.ID, "Tavern Scoundrel", "Creature — Human Rogue", tavernScoundrelOracle, false)
		fodder := pushPermanentForTest(g, me.ID, "Fodder", "", "Artifact")
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})

		if err := g.ActivateCatalogAbility(me.ID, ts, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		if g.Battlefield.Contains(fodder) {
			t.Fatal("the other permanent was not sacrificed")
		}
		passPriorityAroundTable(t, g)
		flip := latestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if flip == nil || flip.Chooser != me.ID {
			t.Fatalf("Tavern Scoundrel asked for no coin call: %+v", flip)
		}
		if err := g.ResolveCoinCall(flip.ID, me.ID, "tails"); err != nil {
			t.Fatalf("ResolveCoinCall: %v", err)
		}
		won := false
		for _, ev := range randomEvents(g, game.EventFlipCoin) {
			won = ev.Won
		}
		passPriorityAroundTable(t, g)
		n, tapped := countTokensNamed(g, me.ID, "Treasure")
		if won {
			sawWin = true
			if n != 2 || tapped != 0 {
				t.Fatalf("won flip: %d Treasures, %d tapped; want 2 untapped", n, tapped)
			}
		} else {
			sawLoss = true
			if n != 0 {
				t.Fatalf("lost flip made %d Treasures", n)
			}
		}
	}
	if !sawWin || !sawLoss {
		t.Fatalf("64 flips never both won and lost (win %v, loss %v)", sawWin, sawLoss)
	}
}

// TestMuYanlingMakesAVehicleThatFliesAndDrawsOnFlyingDamage: the token is
// a 3/2 Vehicle with crew 1 that has flying from Mu Yanling, and flying
// creatures connecting draw once per player while a creature without
// flying draws nothing.
func TestMuYanlingMakesAVehicleThatFliesAndDrawsOnFlyingDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpell(t, g, "Mu Yanling, Wind Rider", "Legendary Creature — Human Wizard Pilot", muYanlingOracle, nil)
	passPriorityAroundTable(t, g)

	v := battlefieldTokenNamed(t, g, "Vehicle")
	if v.IsCreature() || !v.HasSubtype("Vehicle") || v.Power != 3 || v.Toughness != 2 {
		t.Fatalf("the Vehicle token is %q %d/%d", v.TypeLine, v.Power, v.Toughness)
	}
	if !effectiveAbilitiesContain(t, g, v.InstanceID, "flying") {
		t.Fatal("Mu Yanling did not give the Vehicle flying")
	}
	crewer := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "One-Power Crewer", TypeLine: "Creature — Human",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	if err := g.ActivateCatalogAbility(me.ID, v.InstanceID, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{crewer}}); err != nil {
		t.Fatalf("crew 1: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, _ := g.LookupCardForEffect(v.InstanceID); !c.IsCreature() {
		t.Fatal("crew 1 did not make the Vehicle a creature")
	}

	hand := me.Hand.Size()
	dealCombatDamageToPlayer(g, crewer, opp.ID, 1)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Fatalf("a creature without flying drew: hand %d -> %d", hand, got)
	}
	dealCombatDamageToPlayer(g, v.InstanceID, opp.ID, 3)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("the flying Vehicle connecting: hand %d -> %d, want +1", hand, got)
	}
}
