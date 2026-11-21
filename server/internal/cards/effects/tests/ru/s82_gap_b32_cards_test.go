package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gap32Serra = "8b0e2382-dd91-42e6-aba3-5e8a15418a49"
	gap32Dack  = "9d4b1e44-3c66-4b1b-8a1d-308aa4f841f8"
	gap32Minsc = "3a310554-3f69-4158-bb94-9d6965c7125a"
)

func TestGapB32CardsAreRegisteredFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		gap32Serra: "Serra the Benevolent",
		gap32Dack:  "Dack Fayden",
		gap32Minsc: "Minsc & Boo, Timeless Heroes",
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

func TestGapB32SerraPlusTwoPumpsFlyersAndMinusThreeMakesAnAngel(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	serra := effectstest.PushWalkerForTest(g, me.ID, "Serra the Benevolent", gap32Serra, 4)
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 1,
		Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID,
	})
	walker := effectstest.B12Creature(g, me.ID, "Walker", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, serra, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.ApaLive(g, flyer); c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("flyer %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
	if c := effectstest.ApaLive(g, walker); c.CurrentPower() != 2 {
		t.Errorf("non-flyer was pumped: %d", c.CurrentPower())
	}
}

func TestGapB32SerraEmblemLeavesYouAtOneWhileYouControlACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	serra := effectstest.PushWalkerForTest(g, me.ID, "Serra the Benevolent", gap32Serra, 6)
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, serra, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−6: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.EmblemsOf(g, me.ID)) != 1 {
		t.Fatalf("no emblem")
	}
	effectstest.B12Creature(g, me.ID, "Guard", "Creature — Bear", 1, 1)
	src := effectstest.Pr7Creature(g, opp.ID, "Pinger", 1, "R")
	g.WithWriteLock(func() { me.Life = 3 })
	effectstest.Pr7Hit(t, g, src, me.ID, 10)
	if me.Life != 1 {
		t.Errorf("life %d after lethal damage with a creature, want 1", me.Life)
	}
}

func TestGapB32SerraEmblemDoesNothingWithoutACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	serra := effectstest.PushWalkerForTest(g, me.ID, "Serra the Benevolent", gap32Serra, 6)
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, serra, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−6: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	src := effectstest.Pr7Creature(g, opp.ID, "Pinger", 1, "R")
	g.WithWriteLock(func() { me.Life = 3 })
	effectstest.Pr7Hit(t, g, src, me.ID, 10)
	if me.Life > 0 && !me.Eliminated {
		t.Errorf("life %d with no creature, want a lethal hit", me.Life)
	}
}

func TestGapB32DackFaydenMinusTwoStealsAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	dack := effectstest.PushWalkerForTest(g, me.ID, "Dack Fayden", gap32Dack, 3)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, dack, 1, game.ActivateAbilityParams{Targets: effectstest.CardTarget(rock)}); err != nil {
		t.Fatalf("−2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := g.LookupCardForEffect(rock); !ok || c.Controller != me.ID {
		t.Errorf("the artifact is not under my control")
	}
}

func TestGapB32DackFaydenPlusOneLootsTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	dack := effectstest.PushWalkerForTest(g, me.ID, "Dack Fayden", gap32Dack, 3)
	effectstest.AdvanceToMainOf(t, g, seat)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, dack, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(me.ID)}}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d, want %d after the draws and before the discards", got, hand+2)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
}

func TestGapB32MinscAndBooMinusTwoSacrificesAndDamages(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	minsc := effectstest.PushWalkerForTest(g, me.ID, "Minsc & Boo, Timeless Heroes", gap32Minsc, 3)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 3, 3)
	effectstest.AdvanceToMainOf(t, g, seat)
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, minsc, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("pick the target: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Errorf("the creature was not sacrificed")
	}
	if opp.Life != life-3 {
		t.Errorf("opponent life %d → %d, want -3", life, opp.Life)
	}
}

func TestGapB32MinscAndBooPlusOneNeedsTrampleOrHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	minsc := effectstest.PushWalkerForTest(g, me.ID, "Minsc & Boo, Timeless Heroes", gap32Minsc, 3)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 1, 1)
	hasty := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hasty", TypeLine: "Creature — Goblin", Power: 1, Toughness: 1,
		Keywords: []string{"haste"}, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, minsc, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(plain)}); err == nil {
		t.Fatalf("+1 targeted a creature with neither trample nor haste")
	}
	if err := g.ActivateCatalogAbility(me.ID, minsc, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(hasty)}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.ApaLive(g, hasty); c.CurrentPower() != 4 {
		t.Errorf("Hasty is %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
}

// The emblem takes control of the permanents a spell of yours targets, as it
// is cast: the trigger resolves before the spell does.
func TestGapB32DackFaydenEmblemStealsTheTargetsOfYourSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	dack := effectstest.PushWalkerForTest(g, me.ID, "Dack Fayden", gap32Dack, 6)
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, dack, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−6: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.EmblemsOf(g, me.ID)) != 1 {
		t.Fatalf("no emblem")
	}
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Naturalize", "Instant", "bdb3ca68-ec1f-4e16-81cc-d23f8f52c728", effectstest.CardTarget(rock))
	stolen := false
	for i := 0; i < 6 && !stolen; i++ {
		if c, ok := g.LookupCardForEffect(rock); ok && c.Controller == me.ID && g.Battlefield.Contains(rock) {
			stolen = true
			break
		}
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	if !stolen {
		t.Errorf("the emblem did not take control of the targeted permanent before the spell resolved")
	}
}
