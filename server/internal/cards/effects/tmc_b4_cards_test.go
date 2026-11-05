package effects

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmc_b4_cards_test.go — TMNT Commander batch 4 (#5011). Each test plays
// the real catalog entry through the engine and asserts what its text
// says.

const (
	tmcB4RodneyOracle    = "cd8761ef-b57b-456b-aa12-374ac825087e"
	tmcB4BigAppleOracle  = "dd01ef1f-f6be-498f-82e0-dc04833e685f"
	tmcB4HideoutOracle   = "b639e1fe-d099-4cab-a0d0-a1b33c7f31dd"
	tmcB4BiogenicOracle  = "d2d84ba5-0d20-4ce3-8f1f-93cd9ef94f5f"
	tmcB4SteelbaneOracle = "e38e3723-05f5-4a51-8364-1cda19f9cc49"
)

func TestTmcB4EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tmcB4RodneyOracle:    "Roadkill Rodney",
		tmcB4BigAppleOracle:  "Big Apple, 3 a.m.",
		tmcB4HideoutOracle:   "Hidden Hideout",
		tmcB4BiogenicOracle:  "Biogenic Ooze",
		tmcB4SteelbaneOracle: "Steelbane Hydra",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Roadkill Rodney ---------------------------------------------------------

// tmcB4CastRodney casts Roadkill Rodney strictly from my hand with `squads`
// squad payments, floating exactly what that costs.
func tmcB4CastRodney(t *testing.T, g *game.Game, squads int) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[0]
	toMainForCost(t, g)
	id := tmtB11Hand(g, me, "Roadkill Rodney", "Artifact Creature — Robot", tmcB4RodneyOracle, "{2}", 2, 1)
	mana := "{C}{C}"
	var paid []int
	for i := 0; i < squads; i++ {
		mana += "{C}{C}{C}"
		paid = append(paid, 0)
	}
	floatMana(t, g, me, mana)
	return id, g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: paid})
}

func TestTmcB4RodneySquadTwiceMakesTwoCopies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rodney, err := tmcB4CastRodney(t, g, 2)
	if err != nil {
		t.Fatalf("cast with squad paid twice: %v", err)
	}
	passPriorityAroundTable(t, g)
	ids := tmtB1Named(g, "Roadkill Rodney")
	if len(ids) != 3 {
		t.Fatalf("%d Roadkill Rodneys, want the card and two token copies", len(ids))
	}
	tokens := 0
	for _, id := range ids {
		c := tmtB1Card(t, g, id)
		if c.Controller != me.ID {
			t.Errorf("a Rodney is controlled by %v", c.Controller)
		}
		if c.IsToken() {
			tokens++
		}
		assertKeywords(t, g, id, "deathtouch")
	}
	if tokens != 2 || tmtB1Card(t, g, rodney).IsToken() {
		t.Errorf("%d tokens, want 2 copies beside the cast card", tokens)
	}
	// The copies were never cast, so they paid no squad cost and copy
	// nothing themselves.
	passPriorityAroundTable(t, g)
	if n := len(tmtB1Named(g, "Roadkill Rodney")); n != 3 {
		t.Errorf("%d Rodneys after the table settled, want 3", n)
	}
}

func TestTmcB4RodneyUnpaidSquadMakesNoCopies(t *testing.T) {
	g := newCatalogGame(t)
	if _, err := tmcB4CastRodney(t, g, 0); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := len(tmtB1Named(g, "Roadkill Rodney")); n != 1 {
		t.Errorf("%d Rodneys, want 1", n)
	}
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered {
			t.Errorf("a trigger %q is on the stack with no squad paid", item.Label)
		}
	}
}

func TestTmcB4RodneySquadNeedsItsManaUnderTheStrictGate(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	id := tmtB11Hand(g, me, "Roadkill Rodney", "Artifact Creature — Robot", tmcB4RodneyOracle, "{2}", 2, 1)
	floatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err == nil {
		t.Fatal("a squad cast was accepted with four mana for a {5} total")
	}
}

// The ruling: a Rodney that leaves before its squad trigger resolves still
// makes its copies, from last-known information.
func TestTmcB4RodneyCopiesEvenIfItLeftInResponse(t *testing.T) {
	g := newCatalogGame(t)
	rodney, err := tmcB4CastRodney(t, g, 1)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	passUntilSpellResolves(t, g)
	b27Kill(g, rodney)
	passPriorityAroundTable(t, g)
	ids := tmtB1Named(g, "Roadkill Rodney")
	if len(ids) != 1 || !tmtB1Card(t, g, ids[0]).IsToken() {
		t.Fatalf("battlefield Rodneys %v, want exactly one token copy", ids)
	}
	if c := tmtB1Card(t, g, ids[0]); c.Power != 2 || c.Toughness != 1 {
		t.Errorf("the copy is %d/%d, want 2/1", c.Power, c.Toughness)
	}
}

func TestTmcB4RodneyMakesAMutagenOnCombatDamageToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rodney := b27Push(g, me.ID, "Roadkill Rodney", "Artifact Creature — Robot", tmcB4RodneyOracle, "{2}", 2, 1)
	declareAttack(t, g, opp.ID, rodney)
	advanceTo(t, g, game.StepDeclareBlockers)
	finishBlockDeclarations(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if n := tmtMuMutagens(g, me.ID); n != 1 {
		t.Errorf("%d Mutagens after Rodney hit a player, want 1", n)
	}
}

// --- Big Apple, 3 a.m. -------------------------------------------------------

func TestTmcB4BigAppleEntersTappedAndTapsForTheChosenColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b13PlayAs(t, g, 0, "Big Apple, 3 a.m.", "Land", tmcB4BigAppleOracle)
	if !b13Tapped(t, g, id) {
		t.Error("entered untapped")
	}
	prompt := pendingOfKind(g, game.PendingChoiceColor)
	if prompt == nil || !reflect.DeepEqual(prompt.ColorOptions, game.AllColors) {
		t.Fatalf("colour prompt %+v, want every colour", prompt)
	}
	answerColor(t, g, me.ID, "B")
	untap(g, id)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"B"}) {
		t.Errorf("pool = %v, want [B]", got)
	}
}

func tmcB4ActivateBigApple(t *testing.T, g *game.Game, apple uuid.UUID, mana string) error {
	t.Helper()
	me := g.Seats[0]
	floatMana(t, g, me, mana)
	return g.ActivateCatalogAbility(me.ID, apple, 0, game.ActivateAbilityParams{Strict: true})
}

func TestTmcB4BigAppleMakesARatForEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	apple := pushChosenColorPermanent(t, g, me.ID, "Big Apple, 3 a.m.", "Land", tmcB4BigAppleOracle, "B")
	if err := tmcB4ActivateBigApple(t, g, apple, "{C}{C}{C}{C}{C}"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := tmtB1Count(g, me.ID, "Rat"); n != 3 {
		t.Fatalf("%d Rats at a four-player table, want 3", n)
	}
	rat := tmtB1Card(t, g, tmtB1Named(g, "Rat")[0])
	if !rat.IsToken() || rat.Power != 1 || rat.Toughness != 1 || !reflect.DeepEqual(rat.Colors, []string{"B"}) {
		t.Errorf("the Rat is %+v, want a 1/1 black token", rat)
	}
	if !b13Tapped(t, g, apple) {
		t.Error("the ability did not tap the land")
	}

	// An opponent who has left is not one you have.
	untap(g, apple)
	g.WithWriteLock(func() { g.Seats[2].Eliminated = true })
	if err := tmcB4ActivateBigApple(t, g, apple, "{C}{C}{C}{C}{C}"); err != nil {
		t.Fatalf("second activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := tmtB1Count(g, me.ID, "Rat"); n != 5 {
		t.Errorf("%d Rats after the second activation with two opponents, want 5", n)
	}
}

func TestTmcB4BigAppleRefusesWithoutFiveMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	apple := pushChosenColorPermanent(t, g, me.ID, "Big Apple, 3 a.m.", "Land", tmcB4BigAppleOracle, "B")
	if err := tmcB4ActivateBigApple(t, g, apple, "{C}{C}{C}{C}"); err == nil {
		t.Fatal("activated with four mana")
	}
}

// --- Hidden Hideout ----------------------------------------------------------

func TestTmcB4HideoutEntersTappedAndTapsForYourCommandersColors(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	riderGiveCommander(g, me, "{2}{G}")
	id := b13PlayAs(t, g, 0, "Hidden Hideout", "Land", tmcB4HideoutOracle)
	if !b13Tapped(t, g, id) {
		t.Error("entered untapped")
	}
	untap(g, id)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if pick := riderLatestManaPick(g, me.ID); pick == nil || !reflect.DeepEqual(pick.ColorOptions, []string{"G"}) {
		t.Errorf("pick = %+v, want narrowed to [G] by a mono-green commander", pick)
	}
}

func TestTmcB4HideoutAddsNothingWithNoCommander(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Hidden Hideout", "Land", tmcB4HideoutOracle, false)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); len(got) != 0 {
		t.Errorf("pool = %v, want nothing with no commander", got)
	}
}

func TestTmcB4HideoutGivesLifelinkToACreatureWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	hideout := pushCatalogPermanent(g, me.ID, "Hidden Hideout", "Land", tmcB4HideoutOracle, false)
	plain := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	countered := b12Creature(g, me.ID, "Grown Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(countered, game.CounterPlusOne, 1)
		_ = g.AddCounterForEffect(theirs, game.CounterPlusOne, 1)
	})
	for _, bad := range []uuid.UUID{plain, theirs} {
		floatMana(t, g, me, "{C}{C}")
		if err := g.ActivateCatalogAbility(me.ID, hideout, 0, game.ActivateAbilityParams{
			Strict: true, Targets: cardRefs(bad),
		}); err == nil {
			t.Fatalf("targeted %s, which is not a creature you control with a counter", bad)
		}
		me.ManaPool.EmptyPool()
	}
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, hideout, 0, game.ActivateAbilityParams{
		Strict: true, Targets: cardRefs(countered),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	assertKeywords(t, g, countered, "lifelink")
}

// --- Biogenic Ooze -----------------------------------------------------------

func TestTmcB4BiogenicOozeMakesOozesAndGrowsThemAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	ooze := tmtB11Hand(g, me, "Biogenic Ooze", "Creature — Ooze", tmcB4BiogenicOracle, "{3}{G}{G}", 2, 2)
	floatMana(t, g, me, "{C}{C}{C}{G}{G}")
	if err := g.CastSpell(me.ID, ooze, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	tokens := tmtB1Named(g, "Ooze")
	if len(tokens) != 1 {
		t.Fatalf("%d Ooze tokens after it entered, want 1", len(tokens))
	}
	tok := tmtB1Card(t, g, tokens[0])
	if !tok.IsToken() || tok.Power != 2 || tok.Toughness != 2 || !reflect.DeepEqual(tok.Colors, []string{"G"}) || !tok.HasSubtype("Ooze") {
		t.Errorf("the token is %+v, want a 2/2 green Ooze", tok)
	}

	// {1}{G}{G}{G}: another Ooze; three mana is not enough.
	floatMana(t, g, me, "{G}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, ooze, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with three mana")
	}
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, ooze, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	tokens = tmtB1Named(g, "Ooze")
	if len(tokens) != 2 {
		t.Fatalf("%d Ooze tokens after the activation, want 2", len(tokens))
	}

	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	for _, id := range append([]uuid.UUID{ooze}, tokens...) {
		if n := plusOnes(g, id); n != 1 {
			t.Errorf("an Ooze has %d +1/+1 counters after the end step, want 1", n)
		}
	}
	if n := plusOnes(g, bear); n != 0 {
		t.Errorf("the Bear has %d counters, want none (not an Ooze)", n)
	}
}

// --- Steelbane Hydra ---------------------------------------------------------

func TestTmcB4SteelbaneHydraEntersWithXAndSpendsACounterToDestroy(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	hydra := tmtB11Hand(g, me, "Steelbane Hydra", "Creature — Turtle Hydra", tmcB4SteelbaneOracle, "{X}{G}{G}", 0, 0)
	floatMana(t, g, me, "{C}{G}{G}")
	if err := g.CastSpell(me.ID, hydra, game.CastSpellParams{Strict: true, XValue: 1}); err != nil {
		t.Fatalf("cast with X=1: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := plusOnes(g, hydra); n != 1 {
		t.Fatalf("the Hydra has %d +1/+1 counters, want 1", n)
	}
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	floatMana(t, g, me, "{C}{C}{G}")
	if err := g.ActivateCatalogAbility(me.ID, hydra, 0, game.ActivateAbilityParams{
		Strict: true, Targets: cardRefs(bear),
	}); err == nil {
		t.Fatal("targeted a creature that is neither an artifact nor an enchantment")
	}
	if err := g.ActivateCatalogAbility(me.ID, hydra, 0, game.ActivateAbilityParams{
		Strict: true, Targets: cardRefs(rock),
	}); err != nil {
		t.Fatalf("activate at the artifact: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := tmtB1Count(g, opp.ID, "Their Rock"); n != 0 {
		t.Error("the artifact survived")
	}
	if n := tmtB1Count(g, me.ID, "Steelbane Hydra"); n != 0 {
		t.Error("the Hydra paid its last counter and should have died a 0/0")
	}
}

func TestTmcB4SteelbaneHydraCannotActivateWithoutACounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	hydra := b27Push(g, me.ID, "Steelbane Hydra", "Creature — Turtle Hydra", tmcB4SteelbaneOracle, "{X}{G}{G}", 3, 3, "G")
	aura := b12Permanent(g, opp.ID, "Their Shrine", "Enchantment")
	floatMana(t, g, me, "{C}{C}{G}")
	if err := g.ActivateCatalogAbility(me.ID, hydra, 0, game.ActivateAbilityParams{
		Strict: true, Targets: cardRefs(aura),
	}); err == nil {
		t.Fatal("activated with no +1/+1 counter to remove")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(hydra, game.CounterPlusOne, 1) })
	if err := g.ActivateCatalogAbility(me.ID, hydra, 0, game.ActivateAbilityParams{
		Strict: true, Targets: cardRefs(aura),
	}); err != nil {
		t.Fatalf("activate at the enchantment: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := tmtB1Count(g, opp.ID, "Their Shrine"); n != 0 {
		t.Error("the enchantment survived")
	}
	if n := plusOnes(g, hydra); n != 0 {
		t.Errorf("the Hydra kept %d counters, want 0", n)
	}
}
