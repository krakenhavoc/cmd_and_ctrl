package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b748_test.go — S82 OLD batch 748 (Stronghold counters, #5627),
// through the real engine.

const (
	oldB748Hydra     = "67a7bfa4-5004-4838-874d-03fa11f051d0"
	oldB748Wall      = "7c852dfe-8238-461e-814e-9667807f2cf5"
	oldB748Pursuit   = "7f5124a4-9ffd-475a-8b77-3ff59a9871a1"
	oldB748Scavs     = "88782913-bfba-4300-b26c-405e6a45537c"
	oldB748Breeder   = "b7b738e1-298d-489d-9939-efe1c983f0ae"
	oldB748Cannibal  = "bf044267-92b9-4983-964f-2bc4ecf76bce"
	oldB748Crovax    = "e1b74ef7-ab6b-4d40-8087-3f3a448b7e0a"
	oldB748Mindwarp  = "efb42ccc-6fed-47f0-84b4-e2a9a42e4dfa"
	oldB748TypeSpike = "Creature — Spike"
)

func TestS82OldB748CardsAreFull(t *testing.T) {
	for _, id := range []string{oldB748Hydra, oldB748Wall, oldB748Pursuit, oldB748Scavs, oldB748Breeder, oldB748Cannibal, oldB748Crovax, oldB748Mindwarp} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Fatalf("%s is not registered", id)
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS82OldB748SpittingHydraEntersWithFourAndSpendsThemOnDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 3)
	hydra := effectstest.B12PlayFromHand(t, g, "Spitting Hydra", "Creature — Hydra", oldB748Hydra, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, hydra); got != 4 {
		t.Fatalf("%d +1/+1 counters on entering, want 4", got)
	}
	effectstest.B06AddMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, hydra, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err == nil {
		t.Fatal("activated for {R} alone")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, hydra, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}, "R")
	if got := effectstest.PlusOnes(g, hydra); got != 3 {
		t.Errorf("%d counters after one activation, want 3", got)
	}
	if got := effectstest.DamageMarkedOn(g, bear); got != 1 {
		t.Errorf("the Bear has %d damage, want 1", got)
	}
}

func TestS82OldB748ShiftingWallEntersWithXCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	wall := effectstest.B12PlayFromHand(t, g, "Shifting Wall", "Artifact Creature — Wall", oldB748Wall, game.CastSpellParams{XValue: 3})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, wall); got != 3 {
		t.Errorf("X = 3 gave %d counters, want 3", got)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, wall, "defender") {
		t.Error("no defender")
	}
	if p, tough := effectstest.PtOf(t, g, wall); p != 3 || tough != 3 {
		t.Errorf("the Wall is %d/%d, want 3/3", p, tough)
	}
}

func TestS82OldB748SpikeBreederMakesSpikeTokensAndMovesCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	breeder := effectstest.B12PlayFromHand(t, g, "Spike Breeder", oldB748TypeSpike, oldB748Breeder, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, breeder); got != 3 {
		t.Fatalf("%d counters on entering, want 3", got)
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.RavB2MustActivate(t, g, me.ID, breeder, 1, game.ActivateAbilityParams{}, "C", "C")
	if got := effectstest.PlusOnes(g, breeder); got != 2 {
		t.Errorf("%d counters after making a token, want 2", got)
	}
	var token uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Spike" && c.Controller == me.ID {
				token = c.InstanceID
			}
		}
	})
	if token == uuid.Nil {
		t.Fatal("no Spike token")
	}
	if c, _ := effectstest.BattlefieldCard(g, token); !c.IsToken() || c.Power != 1 || c.Toughness != 1 || !c.HasColorWord("G") {
		t.Errorf("the token is %+v, want a 1/1 green Spike", c)
	}
	effectstest.RavB2MustActivate(t, g, me.ID, breeder, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(token)}, "C", "C")
	if effectstest.PlusOnes(g, breeder) != 1 || effectstest.PlusOnes(g, token) != 1 {
		t.Errorf("breeder %d, token %d counters, want 1 and 1", effectstest.PlusOnes(g, breeder), effectstest.PlusOnes(g, token))
	}
}

func TestS82OldB748MindwarperMakesAPlayerDiscardAtSorceryTiming(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warper := effectstest.B12PlayFromHand(t, g, "Mindwarper", "Creature — Spirit", oldB748Mindwarp, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, warper); got != 3 {
		t.Fatalf("%d counters on entering, want 3", got)
	}
	effectstest.HandSpell(opp, "Card", "Sorcery", "{9}")
	before := opp.Hand.Size()
	effectstest.RavB2MustActivate(t, g, me.ID, warper, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}, "C", "C", "B")
	if effectstest.DiscardChoiceFor(g, opp.ID) == nil {
		t.Fatal("the target player was not asked to discard")
	}
	effectstest.DiscardFromHand(t, g, opp.ID)
	if opp.Hand.Size() != before-1 {
		t.Errorf("hand %d -> %d, want one discarded", before, opp.Hand.Size())
	}
	if got := effectstest.PlusOnes(g, warper); got != 2 {
		t.Errorf("%d counters left, want 2", got)
	}
	// Not at sorcery timing: the opponent's turn.
	effectstest.AdvanceToMainOf(t, g, 1)
	if err := effectstest.RavB2Activate(t, g, me.ID, warper, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}, "C", "C", "B"); err == nil {
		t.Error("activated on an opponent's turn")
	}
}

func TestS82OldB748PursuitOfKnowledgeStudiesInsteadOfDrawingThenDrawsSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	pok := effectstest.BfCardWithOracle(g, me, "Pursuit of Knowledge", "Enchantment", oldB748Pursuit)
	hand, lib := me.Hand.Size(), me.Library.Size()
	effectstest.DrawN(g, me, 2)
	effectstest.AnswerOptional(t, g, me, true) // card one: a study counter
	effectstest.AnswerOptional(t, g, me, false)
	if got := effectstest.CounterOn(g, pok, "study"); got != 1 {
		t.Errorf("%d study counters, want 1", got)
	}
	if me.Hand.Size() != hand+1 || me.Library.Size() != lib-1 {
		t.Errorf("hand %d library %d, want one card drawn of two", me.Hand.Size(), me.Library.Size())
	}
	// Not enough counters yet.
	if err := g.ActivateCatalogAbility(me.ID, pok, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with one study counter")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(pok, "study", 2) })
	hand = me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, pok, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate with three counters: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(pok) {
		t.Error("the enchantment was not sacrificed")
	}
	if me.Hand.Size() != hand+7 {
		t.Errorf("hand %d -> %d, want seven drawn", hand, me.Hand.Size())
	}
}

func TestS82OldB748SkeletonScavengersPayPerCounterAndGrowWhenTheShieldIsUsed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	skel := effectstest.B12PlayFromHand(t, g, "Skeleton Scavengers", "Creature — Skeleton", oldB748Scavs, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, skel); got != 1 {
		t.Fatalf("%d counters on entering, want 1", got)
	}
	// One counter: {1}.
	effectstest.RavB2MustActivate(t, g, me.ID, skel, 0, game.ActivateAbilityParams{}, "C")
	if n := g.RegenerationShieldsOn(skel); n != 1 {
		t.Fatalf("%d regeneration shields, want 1", n)
	}
	effectstest.P1Destroy(t, g, skel)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(skel) {
		t.Fatal("the shield did not save it")
	}
	if got := effectstest.PlusOnes(g, skel); got != 2 {
		t.Errorf("%d counters after regenerating, want 2", got)
	}
	// Two counters: {2}; one mana is not enough.
	if err := effectstest.RavB2Activate(t, g, me.ID, skel, 0, game.ActivateAbilityParams{}, "C"); err == nil {
		t.Fatal("activated for {1} with two counters")
	}
}

func TestS82OldB748SkeletonScavengersGetNoCounterFromAnUnusedShield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	skel := effectstest.B12PlayFromHand(t, g, "Skeleton Scavengers", "Creature — Skeleton", oldB748Scavs, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.RavB2MustActivate(t, g, me.ID, skel, 0, game.ActivateAbilityParams{}, "C")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, skel); got != 1 {
		t.Errorf("%d counters with the shield unused, want 1", got)
	}
}

func TestS82OldB748CannibalizeLetsTheCasterPickWhichIsExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "Theirs A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Theirs B", "Creature — Bear", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	if err := effectstest.CastCatalogSpellErr(t, g, "Cannibalize", "Sorcery", oldB748Cannibal, effectstest.CardRefs(a, mine)); err == nil {
		t.Fatal("two creatures with different controllers were accepted")
	}
	effectstest.CastCatalogSpell(t, g, "Cannibalize", "Sorcery", oldB748Cannibal, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if p == nil {
		t.Fatal("the caster was not asked which to exile")
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{a}); err != nil {
		t.Fatalf("choosing: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) {
		t.Error("the chosen creature was not exiled")
	}
	if !g.Battlefield.Contains(b) || effectstest.PlusOnes(g, b) != 2 {
		t.Errorf("the other creature: on battlefield %v, %d counters, want 2", g.Battlefield.Contains(b), effectstest.PlusOnes(g, b))
	}
}

// oldB748Sacrifice answers the open sacrifice-a-creature offer with `card`.
func oldB748Sacrifice(t *testing.T, g *game.Game, chooser, card uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == chooser {
			if err := g.ResolvePayUnlessWithCards(c.ID, chooser, true, []uuid.UUID{card}); err != nil {
				t.Fatalf("ResolvePayUnlessWithCards: %v", err)
			}
			return
		}
	}
	t.Fatal("no sacrifice offer was open")
}

func TestS82OldB748CrovaxGrowsWhenYouSacrificeAndShrinksWhenYouDont(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	crovax := effectstest.B12PlayFromHand(t, g, "Crovax the Cursed", "Legendary Creature — Vampire Noble", oldB748Crovax, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, crovax); got != 4 {
		t.Fatalf("%d counters on entering, want 4", got)
	}
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	// Turn 2 for me: sacrifice the Fodder.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	oldB748Sacrifice(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Fatal("the creature was not sacrificed")
	}
	if got := effectstest.PlusOnes(g, crovax); got != 5 {
		t.Errorf("%d counters after sacrificing, want 5", got)
	}
	// The next upkeep: decline.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, crovax); got != 4 {
		t.Errorf("%d counters after declining, want 4", got)
	}
}

func TestS82OldB748CrovaxWithNoOtherCreatureCanSacrificeItselfOrShrinks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	crovax := effectstest.B12PlayFromHand(t, g, "Crovax the Cursed", "Legendary Creature — Vampire Noble", oldB748Crovax, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("Crovax is a creature and may be sacrificed; the offer was not made")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, crovax); got != 3 {
		t.Errorf("%d counters after declining, want 3", got)
	}
}

func TestS82OldB748CrovaxGainsFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	crovax := effectstest.B12PlayFromHand(t, g, "Crovax the Cursed", "Legendary Creature — Vampire Noble", oldB748Crovax, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasKeywordOnBattlefield(t, g, crovax, "flying") {
		t.Fatal("flying before the ability")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, crovax, 0, game.ActivateAbilityParams{}, "B")
	if !effectstest.HasKeywordOnBattlefield(t, g, crovax, "flying") {
		t.Error("no flying after {B}")
	}
}
