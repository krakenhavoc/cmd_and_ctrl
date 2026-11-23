package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b184_test.go — GAP batch 184 (#5948, Starter Commander Decks):
// every card cast through the engine.

const (
	gap184UnconsciousOracle = "38abf769-e4ef-4b47-9c79-c8cda0b90f6c"
	gap184HarvestOracle     = "ec6d52f9-1c8f-48da-951c-be8d7813c7be"
	gap184ProwessOracle     = "e1a776d2-fb3a-4215-a14e-4b6e9bda95df"
	gap184SunsOracle        = "526cd570-046e-402a-986c-f6ece68eec47"
	gap184HerdOracle        = "30aa362b-7ab1-486e-9802-1171e0dc2416"
)

func TestS82Gap184CardsAreRegisteredFull(t *testing.T) {
	for _, id := range []string{gap184UnconsciousOracle, gap184HarvestOracle, gap184ProwessOracle, gap184SunsOracle, gap184HerdOracle} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS82Gap184CollectiveUnconsciousDrawsPerCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedBear(g, me.ID)
	effectstest.SeedBear(g, me.ID)
	effectstest.SeedBear(g, opp.ID)
	effectstest.PushTypedCard(g, me.ID, "Rock", "Artifact", "{1}")
	before := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Collective Unconscious", "Sorcery", gap184UnconsciousOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// CastCatalogSpell adds the spell to the hand and casting removes it.
	if got := me.Hand.Size() - before; got != 2 {
		t.Errorf("drew %d cards, want 2 (two creatures of mine)", got)
	}
}

func TestS82Gap184HarvestSeasonFetchesOneBasicPerTappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tapped1 := effectstest.B12Creature(g, me.ID, "Tapped One", "Creature — Bear", 2, 2)
	tapped2 := effectstest.B12Creature(g, me.ID, "Tapped Two", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Untapped", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, tapped1).Tapped = true
		effectstest.FindBattlefieldCardForTest(g, tapped2).Tapped = true
	})
	var forests []uuid.UUID
	for i := 0; i < 3; i++ {
		forests = append(forests, effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest"))
	}
	effectstest.StapleLibraryCard(me, "Stomping Ground", "Land — Mountain Forest")

	effectstest.CastCatalogSpell(t, g, "Harvest Season", "Sorcery", gap184HarvestOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Stomping Ground") != uuid.Nil {
		t.Error("a nonbasic land was offered")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, forests); err == nil {
		t.Error("took three basics with two tapped creatures")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, forests[0], forests[1])
	for _, id := range forests[:2] {
		got, ok := effectstest.BattlefieldCard(g, id)
		if !ok || !got.Tapped {
			t.Errorf("fetched Forest on battlefield=%v tapped=%v, want both", ok, got.Tapped)
		}
	}
	if g.Battlefield.Contains(forests[2]) {
		t.Error("a third Forest was fetched")
	}
}

func TestS82Gap184HarvestSeasonWithNoTappedCreatureFindsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Creature(g, me.ID, "Untapped", "Creature — Bear", 2, 2)
	forest := effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Harvest Season", "Sorcery", gap184HarvestOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SearchChoiceFor(g, me.ID); c != nil {
		effectstest.AnswerSearchByID(t, g, me.ID)
	}
	if g.Battlefield.Contains(forest) {
		t.Error("X = 0 put a land onto the battlefield")
	}
}

func TestS82Gap184SwelteringSunsDamagesEveryCreatureAndCycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := effectstest.PushVanillaCreature(g, opp.ID, "Small", 2, 3)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
	big := effectstest.PushVanillaCreature(g, opp.ID, "Big", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Sweltering Suns", "Sorcery", gap184SunsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) || g.Battlefield.Contains(mine) {
		t.Error("a creature with toughness 3 or less survived 3 damage")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("a 4/4 died to 3 damage")
	}
	spec, _ := effects.Lookup(gap184SunsOracle)
	if len(spec.Activated) != 1 {
		t.Fatalf("%d activated abilities, want cycling", len(spec.Activated))
	}
}

func TestS82Gap184SwelteringSunsCyclesFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Sweltering Suns", TypeLine: "Sorcery", ManaCost: "{1}{R}{R}",
			OracleID: gap184SunsOracle, Owner: me.ID, Controller: me.ID})
	})
	for i := 0; i < 3; i++ {
		effectstest.B06AddMana(me, "C")
	}
	before := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand %d -> %d after cycling, want unchanged (discard one, draw one)", before, got)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the cycled card is not in the graveyard")
	}
}

func TestS82Gap184StormHerdMakesAPegasusPerLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Life = 7
	effectstest.CastCatalogSpell(t, g, "Storm Herd", "Sorcery", gap184HerdOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Pegasus", me.ID); n != 7 {
		t.Fatalf("%d Pegasus tokens at 7 life, want 7", n)
	}
	for _, c := range effectstest.TokensNamed(g, "Pegasus") {
		if c.Power != 1 || c.Toughness != 1 || !game.HasKeyword(&c, "flying") {
			t.Errorf("Pegasus is %d/%d with keywords %v, want a 1/1 flier", c.Power, c.Toughness, c.Keywords)
		}
	}
}

func TestS82Gap184HuntersProwessPumpsGrantsTrampleAndDrawsThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.SeedBear(g, me.ID)
	other := effectstest.SeedBear(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Hunter's Prowess", "Sorcery", gap184ProwessOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the target is %d/%d, want 5/5", p, tough)
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, bear), "trample") {
		t.Error("the target has no trample")
	}
	if p := effectstest.EffectivePower(t, g, other); p != 2 {
		t.Errorf("the other Bear has power %d, want 2", p)
	}
	hand := me.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 5)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 5 {
		t.Errorf("drew %d cards for 5 combat damage, want 5", got)
	}
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%4)
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("the pump outlived the turn: power %d", p)
	}
	hand = me.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Error("the granted draw outlived the turn")
	}
}

func TestS82Gap184HuntersProwessRefusesAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	if err := effectstest.CastCatalogSpellErr(t, g, "Hunter's Prowess", "Sorcery", gap184ProwessOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}); err == nil {
		t.Fatal("Hunter's Prowess was cast at a player")
	}
}
