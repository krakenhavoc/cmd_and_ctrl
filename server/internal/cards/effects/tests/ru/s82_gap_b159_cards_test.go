package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b159_cards_test.go — GAP batch 159 (#5948): five Jumpstart 2022
// creatures that were not in the catalog. Each test plays the real catalog
// entry through the engine and asserts what its text says.

const (
	gapB159VaultRobberOracle = "68c90a32-d6ee-4595-8517-5f4cb0e4062d"
	gapB159AspirantOracle    = "448c84f9-712b-4970-8719-2ab9c2a59d20"
	gapB159PhoenixOracle     = "5297f646-8235-47da-a290-a6883b374977"
	gapB159YidaroOracle      = "6c0e22f2-f0f3-43e6-87c5-c543032112d8"
	gapB159CallerOracle      = "c66dc0d3-b98c-40a2-be35-e455b4658828"
)

func TestGapB159EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB159VaultRobberOracle: "Vault Robber",
		gapB159AspirantOracle:    "War-Name Aspirant",
		gapB159PhoenixOracle:     "Warcry Phoenix",
		gapB159YidaroOracle:      "Yidaro, Wandering Monster",
		gapB159CallerOracle:      "Avalanche Caller",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v registered=%v, want Full with no caveats", name, spec, ok)
		}
	}
}

func TestGapB159VaultRobberExilesACreatureCardForATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	robber := effectstest.PushCatalogPermanent(g, me.ID, "Vault Robber", "Creature — Dwarf Rogue", gapB159VaultRobberOracle, false)
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, robber, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no creature card in the graveyard")
	}
	spell := effectstest.PushToGraveyard(g, me, game.Card{Name: "Shock", TypeLine: "Instant"})
	if err := g.ActivateCatalogAbility(me.ID, robber, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{spell}}); err == nil {
		t.Fatal("exiled a noncreature card to pay the cost")
	}
	gy := effectstest.PushCatalogGraveyardCard(me, "Dead Bear", "Creature — Bear", "", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, robber, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{gy}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.ZoneOfCard(g, gy); z != game.ZoneExile {
		t.Errorf("the creature card is in zone %v, want exile", z)
	}
	treasures := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.IsToken() && c.Name == "Treasure" && c.Controller == me.ID {
				treasures++
			}
		}
	})
	if treasures != 1 {
		t.Errorf("%d Treasure tokens, want 1", treasures)
	}
	if c, _ := effectstest.BattlefieldCard(g, robber); !c.Tapped {
		t.Error("the Robber did not tap")
	}
}

func TestGapB159WarNameAspirantRaidAndEvasion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	first := effectstest.AvB16CastCreature(t, g, "War-Name Aspirant", "Creature — Human Warrior", gapB159AspirantOracle, 2, 1)
	if n := effectstest.CountersOn(g, first, game.CounterPlusOne); n != 0 {
		t.Fatalf("%d counters without an attack this turn, want 0", n)
	}
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	second := effectstest.AvB16CastCreature(t, g, "War-Name Aspirant", "Creature — Human Warrior", gapB159AspirantOracle, 2, 1)
	if n := effectstest.CountersOn(g, second, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after attacking this turn, want 1", n)
	}
}

func TestGapB159WarNameAspirantCantBeBlockedByPowerOneOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aspirant := effectstest.B12Push(g, me.ID, "War-Name Aspirant", "Creature — Human Warrior", gapB159AspirantOracle, 2, 1)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 1, 1)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, aspirant)
	if err := g.DeclareBlocker(small, aspirant); err == nil {
		t.Error("a creature with power 1 blocked the Aspirant")
	}
	if err := g.DeclareBlocker(big, aspirant); err != nil {
		t.Errorf("a creature with power 2 can't block the Aspirant: %v", err)
	}
}

// gapB159Attack seeds `attackers` creatures for the active seat, a Warcry
// Phoenix in its graveyard, and declares the attack.
func gapB159Attack(t *testing.T, g *game.Game, attackers int) (me *game.Player, phoenix uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	phoenix = effectstest.PushCatalogGraveyardCard(me, "Warcry Phoenix", "Creature — Phoenix", gapB159PhoenixOracle, 3, 3)
	var ids []uuid.UUID
	for i := 0; i < attackers; i++ {
		ids = append(ids, effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 1))
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range ids {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	return me, phoenix
}

func TestGapB159WarcryPhoenixReturnsTappedAndAttackingForThreeMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, phoenix := gapB159Attack(t, g, 3)
	effectstest.FloatForTest(g, me, "CCR")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionPick && c.Chooser == me.ID {
			// "What does it attack?": the first option is an opponent.
			if err := g.ResolveOptionPick(c.ID, me.ID, 0); err != nil {
				t.Fatalf("ResolveOptionPick: %v", err)
			}
			break
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, phoenix)
	if !ok {
		t.Fatal("the Phoenix did not return to the battlefield")
	}
	if !c.Tapped {
		t.Error("it should enter tapped")
	}
	if c.AttackingTarget == uuid.Nil {
		t.Error("it should enter attacking")
	}
}

func TestGapB159WarcryPhoenixStaysIfYouDeclineOrAttackWithTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, phoenix := gapB159Attack(t, g, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		effectstest.AnswerPayUnless(t, g, me.ID, false)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, phoenix); ok {
		t.Error("the Phoenix returned without the payment")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, phoenix2 := gapB159Attack(t, g2, 2)
	effectstest.PassPriorityAroundTable(t, g2)
	if effectstest.HasPayUnlessFor(g2, me2.ID) {
		t.Error("the Phoenix asked for a payment after an attack with two creatures")
	}
	if z := effectstest.ZoneOfCard(g2, phoenix2); z != game.ZoneGraveyard {
		t.Errorf("the Phoenix is in zone %v, want the graveyard", z)
	}
}

func TestGapB159YidaroShufflesInUntilTheFourthCycle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	for i := 1; i <= 4; i++ {
		id, me := effectstest.CycleFromHand(t, g, "Yidaro, Wandering Monster", "Legendary Creature — Dinosaur Turtle", gapB159YidaroOracle, "{C}{R}")
		effectstest.PassPriorityAroundTable(t, g)
		got := effectstest.ZoneOfCard(g, id)
		// Shuffled in, the card can be the very card the cycling then draws,
		// so the library and the hand are both a shuffle-in's result.
		shuffledIn := got == game.ZoneLibrary || got == game.ZoneHand
		if i < 4 && !shuffledIn {
			t.Fatalf("cycle %d: the card is in zone %v, want it shuffled into the library", i, got)
		}
		if i == 4 && got != game.ZoneBattlefield {
			t.Fatalf("cycle %d: the card is in zone %v, want the battlefield", i, got)
		}
		if me.Graveyard.Contains(id) {
			t.Fatalf("cycle %d: the card stayed in the graveyard", i)
		}
	}
}

func TestGapB159AvalancheCallerAnimatesASnowLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	caller := effectstest.PushCatalogPermanent(g, me.ID, "Avalanche Caller", "Snow Creature — Human Wizard", gapB159CallerOracle, false)
	snow := effectstest.PushCatalogPermanent(g, me.ID, "Snow-Covered Island", "Basic Snow Land — Island", "", false)
	plain := effectstest.PushCatalogPermanent(g, me.ID, "Island", "Basic Land — Island", "", false)
	ref := func(id uuid.UUID) []game.TargetRef { return []game.TargetRef{{Kind: game.TargetCard, ID: id}} }
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, caller, 0, game.ActivateAbilityParams{Strict: true, Targets: ref(plain)}); err == nil {
		t.Fatal("a nonsnow land was a legal target")
	}
	effectstest.B16Activate(t, g, me.ID, caller, 0, game.ActivateAbilityParams{Strict: true, Targets: ref(snow)})
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, snow)
	if !ok {
		t.Fatal("the land left the battlefield")
	}
	if !c.IsCreature() || !c.IsLand() {
		t.Errorf("creature=%v land=%v, want a land creature", c.IsCreature(), c.IsLand())
	}
	if p := effectstest.EffectivePower(t, g, snow); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
	if !c.HasSubtype("Elemental") {
		t.Error("the land is not an Elemental")
	}
	for _, kw := range []string{"hexproof", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, snow, kw) {
			t.Errorf("the land has no %s", kw)
		}
	}
	if pc, _ := effectstest.BattlefieldCard(g, plain); pc.IsCreature() {
		t.Error("the other land became a creature")
	}
}
