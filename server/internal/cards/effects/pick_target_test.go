package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pick_target_test.go — S20 sub-PR 2: the ETB-destroy / recursion
// cards on real targets. Replaces the S19 auto-picker tests.

const (
	reclamationSageOracle = "032ec6e2-6cc3-4a97-9cc7-3233f5e11904"
	acidicSlimeOracle     = "21f45043-5419-4019-8b6c-e5294bd5f549"
	eternalWitnessOracle  = "30b24e8e-3b0e-4d8e-90f3-f66eb7c1858c"
)

func latestPickTarget(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoicePickTarget && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func pickCard(t *testing.T, g *game.Game, chooser, cardID uuid.UUID) {
	t.Helper()
	p := latestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	if err := g.ResolvePickTarget(p.ID, chooser, game.TargetRef{Kind: game.TargetCard, ID: cardID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

func hasID(ids []uuid.UUID, id uuid.UUID) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

// castAndResolveCreature casts a catalog creature and resolves just
// the spell, leaving whatever it triggered pending.
func castAndResolveCreature(t *testing.T, g *game.Game, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	id := castCatalogSpell(t, g, name, typeLine, oracle, nil)
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	return id
}

// --- Reclamation Sage -------------------------------------------

func TestReclamationSageYesThenPickDestroysChosen(t *testing.T) {
	g := newCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	sword := pushTypedCard(g, opp.ID, "Sword", "Artifact", "{1}")
	anthem := pushTypedCard(g, opp.ID, "Anthem", "Enchantment", "{1}{W}{W}")
	myRock := pushTypedCard(g, caster.ID, "My Rock", "Artifact", "{1}")
	bear := pushTypedCard(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}")

	sageID := castAndResolveCreature(t, g, "Reclamation Sage", "Creature — Elf Shaman", reclamationSageOracle)
	// "You may" first; nothing picked yet.
	if latestPickTarget(g, caster.ID) != nil {
		t.Fatalf("pick prompt must wait for the yes/no")
	}
	answerLatestTriggerPrompt(t, g, caster.ID, true)
	prompt := latestPickTarget(g, caster.ID)
	if prompt == nil {
		t.Fatalf("no pick_target prompt after Yes")
	}
	// Legal set: every artifact / enchantment, own included; not the bear.
	for _, want := range []uuid.UUID{sword, anthem, myRock} {
		if !hasID(prompt.PickTargetCards, want) {
			t.Errorf("legal set missing %s", want)
		}
	}
	if hasID(prompt.PickTargetCards, bear) {
		t.Errorf("creature offered as a Reclamation Sage target")
	}
	if triggerOnStack(g, sageID) != nil {
		t.Fatalf("trigger reached the stack before the pick")
	}

	pickCard(t, g, caster.ID, anthem)
	item := triggerOnStack(g, sageID)
	if item == nil || len(item.Targets) != 1 || item.Targets[0].ID != anthem {
		t.Fatalf("trigger not on the stack with the chosen target: %+v", item)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(anthem) || !opp.Graveyard.Contains(anthem) {
		t.Errorf("chosen enchantment not destroyed")
	}
	if !g.Battlefield.Contains(sword) || !g.Battlefield.Contains(myRock) {
		t.Errorf("an un-chosen permanent was destroyed")
	}
}

func TestReclamationSageDeclineNeverPicks(t *testing.T) {
	g := newCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	sword := pushTypedCard(g, opp.ID, "Sword", "Artifact", "{1}")
	castAndResolveCreature(t, g, "Reclamation Sage", "Creature — Elf Shaman", reclamationSageOracle)
	answerLatestTriggerPrompt(t, g, caster.ID, false)
	if latestPickTarget(g, caster.ID) != nil {
		t.Errorf("declined trigger still asked for a target")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(sword) {
		t.Errorf("declined trigger destroyed the artifact")
	}
}

func TestReclamationSageNoLegalTargetSkipsBothPrompts(t *testing.T) {
	g := newCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	pushTypedCard(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}") // no artifact / enchantment anywhere
	castAndResolveCreature(t, g, "Reclamation Sage", "Creature — Elf Shaman", reclamationSageOracle)
	if latestTriggerPrompt(g, caster.ID) != nil || latestPickTarget(g, caster.ID) != nil {
		t.Errorf("trigger with no legal target must be removed without asking (CR 603.3d)")
	}
}

func TestReclamationSageFizzlesWhenChosenTargetLeaves(t *testing.T) {
	g := newCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	sword := pushTypedCard(g, opp.ID, "Sword", "Artifact", "{1}")
	other := pushTypedCard(g, opp.ID, "Other", "Artifact", "{1}")
	sageID := castAndResolveCreature(t, g, "Reclamation Sage", "Creature — Elf Shaman", reclamationSageOracle)
	answerLatestTriggerPrompt(t, g, caster.ID, true)
	pickCard(t, g, caster.ID, sword)
	if triggerOnStack(g, sageID) == nil {
		t.Fatalf("trigger not on the stack")
	}
	// In response: the chosen sword is bounced to hand.
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(sword); err != nil {
			t.Fatalf("BounceToHandForEffect: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(other) {
		t.Errorf("fizzled trigger must not re-pick another target")
	}
	var fizzled bool
	for _, ev := range g.Events {
		if ev.Kind == game.EventFizzle && ev.Source == sageID {
			fizzled = true
		}
	}
	if !fizzled {
		t.Errorf("no EventFizzle after the chosen target left")
	}
}

// --- Acidic Slime -----------------------------------------------

func TestAcidicSlimeMandatoryPickIncludesLands(t *testing.T) {
	g := newCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	mountain := pushTypedCard(g, opp.ID, "Mountain", "Basic Land — Mountain", "")
	bear := pushTypedCard(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}")
	slimeID := castAndResolveCreature(t, g, "Acidic Slime", "Creature — Ooze", acidicSlimeOracle)
	// Mandatory: straight to the pick, no yes/no.
	if latestTriggerPrompt(g, caster.ID) != nil {
		t.Errorf("Acidic Slime is mandatory — no yes/no prompt")
	}
	prompt := latestPickTarget(g, caster.ID)
	if prompt == nil {
		t.Fatalf("no pick_target prompt")
	}
	if !hasID(prompt.PickTargetCards, mountain) || hasID(prompt.PickTargetCards, bear) {
		t.Errorf("legal set = %v (want the land, not the bear)", prompt.PickTargetCards)
	}
	pickCard(t, g, caster.ID, mountain)
	if triggerOnStack(g, slimeID) == nil {
		t.Fatalf("trigger not on the stack after the pick")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mountain) || !opp.Graveyard.Contains(mountain) {
		t.Errorf("Acidic Slime did not destroy the chosen land")
	}
}

// --- Eternal Witness --------------------------------------------

func TestEternalWitnessPicksFromOwnGraveyardOnly(t *testing.T) {
	g := newCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	older := pushGraveyardCardForTest(caster, "Older Body")
	newer := pushGraveyardCardForTest(caster, "Latest Body")
	theirs := pushGraveyardCardForTest(opp, "Their Body")

	witnessID := castAndResolveCreature(t, g, "Eternal Witness", "Creature — Human Shaman", eternalWitnessOracle)
	answerLatestTriggerPrompt(t, g, caster.ID, true)
	prompt := latestPickTarget(g, caster.ID)
	if prompt == nil {
		t.Fatalf("no pick_target prompt")
	}
	if !hasID(prompt.PickTargetCards, older) || !hasID(prompt.PickTargetCards, newer) || hasID(prompt.PickTargetCards, theirs) {
		t.Errorf("legal set = %v (want both of mine, not theirs)", prompt.PickTargetCards)
	}
	// Choose the OLDER one — the S14 auto-pick could only ever take the top.
	pickCard(t, g, caster.ID, older)
	if triggerOnStack(g, witnessID) == nil {
		t.Fatalf("trigger not on the stack")
	}
	passPriorityAroundTable(t, g)
	if !caster.Hand.Contains(older) {
		t.Errorf("chosen graveyard card not returned to hand")
	}
	if !caster.Graveyard.Contains(newer) {
		t.Errorf("the un-chosen card left the graveyard")
	}
}

func TestEternalWitnessEmptyGraveyardNoPrompt(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	castAndResolveCreature(t, g, "Eternal Witness", "Creature — Human Shaman", eternalWitnessOracle)
	if latestTriggerPrompt(g, caster.ID) != nil || latestPickTarget(g, caster.ID) != nil {
		t.Errorf("empty graveyard: trigger must be removed without any prompt")
	}
}
