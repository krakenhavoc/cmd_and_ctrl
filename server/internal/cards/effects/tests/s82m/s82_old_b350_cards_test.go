package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b350_cards_test.go — Khans of Tarkir, batch 350 (#5554). Each test
// plays the real catalog entry and asserts what its text says.

const (
	oldB350Krumar      = "00bcfbaf-cd31-4688-973e-4238749051b7"
	oldB350Ramparts    = "0527fc47-2bfc-4d02-90fa-5a1284f89a0a"
	oldB350Leaping     = "1b5c2ba0-3df7-4394-a3f1-8e8330d7ac81"
	oldB350SmokeTeller = "387a5937-cdbd-4e8e-8efb-4d444aa9cbc9"
	oldB350Retribution = "550b8bee-1773-427b-b300-aaceff243e11"
	oldB350Scout       = "08781f2a-0d43-495c-aac8-7ff7a2e5752f"
	oldB350Crane       = "44cf99b0-8a5e-4df9-abc5-552f8b838c56"
	oldB350Secret      = "e753f50d-64ed-4822-a301-57659ee378b6"
	oldB350KinTree     = "fbb74ff8-7381-4de9-bf62-e5d181a187bc"
)

func TestOldB350CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB350Krumar: "Unyielding Krumar", oldB350Ramparts: "Dazzling Ramparts", oldB350Leaping: "Leaping Master",
		oldB350SmokeTeller: "Smoke Teller", oldB350Retribution: "Retribution of the Ancients", oldB350Scout: "Scout the Borders",
		oldB350Crane: "Flying Crane Technique", oldB350Secret: "Rakshasa's Secret", oldB350KinTree: "Kin-Tree Invocation",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB350Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	return g, me, opp
}

// oldB350PassToPrompt passes priority until the spell has resolved and
// asked its question, which stops the table.
func oldB350PassToPrompt(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				return
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
}

func TestOldB350KrumarAndLeapingMasterGainTheirKeywordUntilEndOfTurn(t *testing.T) {
	for _, tc := range []struct{ name, oracle, mana, kw string }{
		{"Unyielding Krumar", oldB350Krumar, "WW", "first strike"},
		{"Leaping Master", oldB350Leaping, "WWW", "flying"},
	} {
		g, me, _ := oldB350Table(t)
		id := effectstest.PoolPush(g, me.ID, tc.name, tc.oracle, 3, 3)
		if effectstest.AvB22Has(t, g, id, tc.kw) {
			t.Fatalf("%s starts with %s", tc.name, tc.kw)
		}
		if err := func() error {
			effectstest.FloatForTest(g, me, tc.mana)
			return g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true})
		}(); err != nil {
			t.Fatalf("%s: activate: %v", tc.name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.AvB22Has(t, g, id, tc.kw) {
			t.Errorf("%s did not gain %s", tc.name, tc.kw)
		}
		effectstest.AdvanceTo(t, g, game.StepUpkeep)
		if effectstest.AvB22Has(t, g, id, tc.kw) {
			t.Errorf("%s still has %s next turn", tc.name, tc.kw)
		}
	}
}

func TestOldB350KrumarRefusesAnUnpayableActivation(t *testing.T) {
	g, me, _ := oldB350Table(t)
	id := effectstest.PoolPush(g, me.ID, "Unyielding Krumar", oldB350Krumar, 3, 3)
	effectstest.FloatForTest(g, me, "BB")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("{1}{W} was paid with black mana")
	}
}

func TestOldB350DazzlingRampartsTapsATargetCreatureAndTapsItself(t *testing.T) {
	g, me, opp := oldB350Table(t)
	wall := effectstest.PoolPush(g, me.ID, "Dazzling Ramparts", oldB350Ramparts, 0, 7)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "WW")
	if err := g.ActivateCatalogAbility(me.ID, wall, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.PermanentTapped(g, victim) {
		t.Error("the target creature is untapped")
	}
	if !effectstest.PermanentTapped(g, wall) {
		t.Error("the Ramparts did not tap for the cost")
	}
	if !effectstest.AvB22Has(t, g, wall, "defender") {
		t.Error("the Ramparts lost defender")
	}
}

func TestOldB350SmokeTellerLooksAtAFaceDownCreatureOnlyForItsController(t *testing.T) {
	g, me, opp := oldB350Table(t)
	other := g.Seats[2]
	teller := effectstest.PoolPush(g, me.ID, "Smoke Teller", oldB350SmokeTeller, 2, 2)
	hidden := effectstest.PoolPush(g, opp.ID, "Sheoldred, the Apocalypse", "", 4, 5)
	var turned []uuid.UUID
	g.WithWriteLock(func() { turned = g.TurnFaceDownForEffect(uuid.Nil, hidden) })
	if len(turned) != 1 {
		t.Fatal("the creature was not turned face down")
	}
	faceUp := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	// A face-up creature is not a legal target.
	effectstest.FloatForTest(g, me, "UU")
	if err := g.ActivateCatalogAbility(me.ID, teller, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: faceUp}}}); err == nil {
		t.Error("a face-up creature was targeted")
	}
	g.WithWriteLock(func() { me.ManaPool = nil })

	c := effectstest.CardByID(g, hidden)
	if c.IsKnownTo(me.ID) {
		t.Fatal("the activator already knew the face-down creature")
	}
	effectstest.FloatForTest(g, me, "UU")
	if err := g.ActivateCatalogAbility(me.ID, teller, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: hidden}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c = effectstest.CardByID(g, hidden)
	if !c.IsKnownTo(me.ID) || !c.IsKnownTo(opp.ID) || c.IsKnownTo(other.ID) {
		t.Errorf("knowers: me %v controller %v bystander %v, want me and the controller only",
			c.IsKnownTo(me.ID), c.IsKnownTo(opp.ID), c.IsKnownTo(other.ID))
	}
}

func TestOldB350RetributionRemovesASplitOfCountersAndShrinksTheTarget(t *testing.T) {
	g, me, opp := oldB350Table(t)
	a := effectstest.PoolPush(g, me.ID, "Counter Bear A", "", 2, 2)
	b := effectstest.PoolPush(g, me.ID, "Counter Bear B", "", 2, 2)
	effectstest.PfAddCounters(g, a, "+1/+1", 2)
	effectstest.PfAddCounters(g, b, "+1/+1", 1)
	source := effectstest.PushCatalogPermanent(g, me.ID, "Retribution of the Ancients", "Enchantment", oldB350Retribution, false)
	victim := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	effectstest.FloatForTest(g, me, "B")
	if err := g.ActivateCatalogAbility(me.ID, source, 0, game.ActivateAbilityParams{Strict: true,
		CounterSourceIDs: []uuid.UUID{a, b}, CounterCounts: []int{2, 1},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindBattlefieldCardForTest(g, a).Counters["+1/+1"]; got != 0 {
		t.Errorf("bear A has %d counters, want 0", got)
	}
	if got := effectstest.FindBattlefieldCardForTest(g, b).Counters["+1/+1"]; got != 0 {
		t.Errorf("bear B has %d counters, want 0", got)
	}
	if p, tough := effectstest.EffectivePower(t, g, victim), effectstest.EffectiveToughness(t, g, victim); p != 2 || tough != 2 {
		t.Errorf("target is %d/%d, want 2/2 (-3/-3)", p, tough)
	}
	effectstest.AdvanceTo(t, g, game.StepUpkeep)
	if p := effectstest.EffectivePower(t, g, victim); p != 5 {
		t.Errorf("target power %d after the turn, want 5", p)
	}
}

func TestOldB350RetributionCannotRemoveMoreCountersThanYouHave(t *testing.T) {
	g, me, opp := oldB350Table(t)
	a := effectstest.PoolPush(g, me.ID, "Counter Bear A", "", 2, 2)
	effectstest.PfAddCounters(g, a, "+1/+1", 1)
	source := effectstest.PushCatalogPermanent(g, me.ID, "Retribution of the Ancients", "Enchantment", oldB350Retribution, false)
	victim := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	effectstest.FloatForTest(g, me, "B")
	if err := g.ActivateCatalogAbility(me.ID, source, 0, game.ActivateAbilityParams{Strict: true,
		CounterSourceIDs: []uuid.UUID{a}, CounterCounts: []int{2},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err == nil {
		t.Error("two counters were removed from a creature with one")
	}
}

func TestOldB350ScoutTheBordersTakesACreatureOrLandAndMillsTheRest(t *testing.T) {
	g, me, _ := oldB350Table(t)
	g.WithWriteLock(func() { me.Library.Cards = nil })
	var deep []uuid.UUID
	for i := 0; i < 3; i++ {
		deep = append(deep, effectstest.SeedLibraryTop(me, "Deep Card", "Sorcery", ""))
	}
	land := effectstest.SeedLibraryTop(me, "Top Forest", "Basic Land — Forest", "")
	spell := effectstest.SeedLibraryTop(me, "Top Spell", "Sorcery", "{1}")
	creature := effectstest.SeedLibraryTop(me, "Top Bear", "Creature — Bear", "{1}{G}")
	// The top five are the creature, the spell, the land and two of the
	// deep cards; one deep card stays in the library.
	handBefore := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Scout the Borders", "Sorcery", oldB350Scout, nil)
	oldB350PassToPrompt(t, g)
	prompt := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no choose-cards prompt for the revealed cards")
	}
	for _, c := range []uuid.UUID{creature, land} {
		if !effectstest.HasID(prompt.ChooseCards, c) {
			t.Errorf("a creature or land card %s was not offered", c)
		}
	}
	if effectstest.HasID(prompt.ChooseCards, spell) {
		t.Error("a sorcery card was offered")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	inHand := func(id uuid.UUID) bool {
		for _, c := range me.Hand.Cards {
			if c.InstanceID == id {
				return true
			}
		}
		return false
	}
	inGrave := func(id uuid.UUID) bool {
		for _, c := range me.Graveyard.Cards {
			if c.InstanceID == id {
				return true
			}
		}
		return false
	}
	if !inHand(land) || me.Hand.Size() != handBefore+1 {
		t.Errorf("the land is in hand = %v, hand grew by %d, want 1", inHand(land), me.Hand.Size()-handBefore)
	}
	for _, id := range []uuid.UUID{creature, spell, deep[2], deep[1]} {
		if !inGrave(id) {
			t.Errorf("card %s from the revealed five is not in the graveyard", id)
		}
	}
	if inGrave(deep[0]) || me.Library.Size() != 1 {
		t.Errorf("the sixth card left the library: in graveyard %v, library %d", inGrave(deep[0]), me.Library.Size())
	}
}

func TestOldB350FlyingCraneTechniqueUntapsAndGrantsOnlyTheCreaturesYouHaveNow(t *testing.T) {
	g, me, _ := oldB350Table(t)
	tapped := effectstest.B12Creature(g, me.ID, "Tapped Bear", "Creature — Bear", 2, 2)
	effectstest.ArnTap(t, g, tapped)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Flying Crane Technique", "Instant", oldB350Crane, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PermanentTapped(g, tapped) {
		t.Error("the creature is still tapped")
	}
	for _, kw := range []string{"flying", "double strike"} {
		if !effectstest.AvB22Has(t, g, tapped, kw) {
			t.Errorf("your creature did not gain %s", kw)
		}
		if effectstest.AvB22Has(t, g, theirs, kw) {
			t.Errorf("an opponent's creature gained %s", kw)
		}
	}
	later := effectstest.B12Creature(g, me.ID, "Later Bear", "Creature — Bear", 2, 2)
	if effectstest.AvB22Has(t, g, later, "flying") {
		t.Error("a creature that arrived after resolution gained flying")
	}
}

func TestOldB350RakshasasSecretDiscardsTwoAndMillsTwo(t *testing.T) {
	g, me, opp := oldB350Table(t)
	var cards []uuid.UUID
	for i := 0; i < 3; i++ {
		id := uuid.New()
		cards = append(cards, id)
		opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Their Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	}
	oppHand := opp.Hand.Size()
	library, grave := me.Library.Size(), me.Graveyard.Size()
	effectstest.CastCatalogSpell(t, g, "Rakshasa's Secret", "Sorcery", oldB350Secret,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	oldB350PassToPrompt(t, g)
	effectstest.AnswerDiscard(t, g, opp.ID, cards[0], cards[1])
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != oppHand-2 {
		t.Errorf("opponent hand %d, want %d", opp.Hand.Size(), oppHand-2)
	}
	if me.Library.Size() != library-2 || me.Graveyard.Size() < grave+2 {
		t.Errorf("library %d (was %d), graveyard %d (was %d): want two milled", me.Library.Size(), library, me.Graveyard.Size(), grave)
	}
}

func TestOldB350KinTreeInvocationMakesATokenAsBigAsYourGreatestToughness(t *testing.T) {
	g, me, _ := oldB350Table(t)
	effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 3, 3)
	effectstest.B12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 6)
	effectstest.B12Creature(g, g.Seats[1].ID, "Their Colossus", "Creature — Giant", 9, 9)
	effectstest.CastCatalogSpell(t, g, "Kin-Tree Invocation", "Sorcery", oldB350KinTree, nil)
	effectstest.PassPriorityAroundTable(t, g)
	tokens := effectstest.AvB34Tokens(g, me.ID, "Spirit Warrior")
	if len(tokens) != 1 {
		t.Fatalf("%d Spirit Warrior tokens, want 1", len(tokens))
	}
	if p, tough := effectstest.EffectivePower(t, g, tokens[0].InstanceID), effectstest.EffectiveToughness(t, g, tokens[0].InstanceID); p != 6 || tough != 6 {
		t.Errorf("token is %d/%d, want 6/6", p, tough)
	}
	if got := tokens[0].Colors; len(got) != 2 {
		t.Errorf("token colours %v, want black and green", got)
	}
}
