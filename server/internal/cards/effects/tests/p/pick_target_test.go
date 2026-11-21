package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Reclamation Sage -------------------------------------------

func TestReclamationSageYesThenPickDestroysChosen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	sword := effectstest.PushTypedCard(g, opp.ID, "Sword", "Artifact", "{1}")
	anthem := effectstest.PushTypedCard(g, opp.ID, "Anthem", "Enchantment", "{1}{W}{W}")
	myRock := effectstest.PushTypedCard(g, caster.ID, "My Rock", "Artifact", "{1}")
	bear := effectstest.PushTypedCard(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}")

	sageID := effectstest.CastAndResolveCreature(t, g, "Reclamation Sage", "Creature — Elf Shaman", effectstest.ReclamationSageOracle)
	// The target first (CR 603.3d); the "you may" waits for resolution.
	if effectstest.OwedMays(g, caster.ID) != 0 {
		t.Fatalf("the yes/no was asked before the target pick")
	}
	prompt := effectstest.LatestPickTarget(g, caster.ID)
	if prompt == nil {
		t.Fatalf("no pick_target prompt as the trigger went on the stack")
	}
	// Legal set: every artifact / enchantment, own included; not the bear.
	for _, want := range []uuid.UUID{sword, anthem, myRock} {
		if !effectstest.HasID(prompt.PickTargetCards, want) {
			t.Errorf("legal set missing %s", want)
		}
	}
	if effectstest.HasID(prompt.PickTargetCards, bear) {
		t.Errorf("creature offered as a Reclamation Sage target")
	}
	if effectstest.TriggerOnStack(g, sageID) != nil {
		t.Fatalf("trigger reached the stack before the pick")
	}

	effectstest.PickCard(t, g, caster.ID, anthem)
	item := effectstest.TriggerOnStack(g, sageID)
	if item == nil || len(item.Targets) != 1 || item.Targets[0].ID != anthem {
		t.Fatalf("trigger not on the stack with the chosen target: %+v", item)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, caster.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(anthem) || !opp.Graveyard.Contains(anthem) {
		t.Errorf("chosen enchantment not destroyed")
	}
	if !g.Battlefield.Contains(sword) || !g.Battlefield.Contains(myRock) {
		t.Errorf("an un-chosen permanent was destroyed")
	}
}

func TestReclamationSageDeclineNeverPicks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	sword := effectstest.PushTypedCard(g, opp.ID, "Sword", "Artifact", "{1}")
	effectstest.CastAndResolveCreature(t, g, "Reclamation Sage", "Creature — Elf Shaman", effectstest.ReclamationSageOracle)
	effectstest.AnswerLatestTriggerPrompt(t, g, caster.ID, false)
	if effectstest.LatestPickTarget(g, caster.ID) != nil {
		t.Errorf("declined trigger still asked for a target")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(sword) {
		t.Errorf("declined trigger destroyed the artifact")
	}
}

func TestReclamationSageNoLegalTargetSkipsBothPrompts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	effectstest.PushTypedCard(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}") // no artifact / enchantment anywhere
	effectstest.CastAndResolveCreature(t, g, "Reclamation Sage", "Creature — Elf Shaman", effectstest.ReclamationSageOracle)
	if effectstest.OwedMays(g, caster.ID) > 0 || effectstest.LatestPickTarget(g, caster.ID) != nil {
		t.Errorf("trigger with no legal target must be removed without asking (CR 603.3d)")
	}
}

func TestReclamationSageFizzlesWhenChosenTargetLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	sword := effectstest.PushTypedCard(g, opp.ID, "Sword", "Artifact", "{1}")
	other := effectstest.PushTypedCard(g, opp.ID, "Other", "Artifact", "{1}")
	sageID := effectstest.CastAndResolveCreature(t, g, "Reclamation Sage", "Creature — Elf Shaman", effectstest.ReclamationSageOracle)
	effectstest.PickCard(t, g, caster.ID, sword)
	if effectstest.TriggerOnStack(g, sageID) == nil {
		t.Fatalf("trigger not on the stack")
	}
	// In response: the chosen sword is bounced to hand.
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(sword); err != nil {
			t.Fatalf("BounceToHandForEffect: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	mountain := effectstest.PushTypedCard(g, opp.ID, "Mountain", "Basic Land — Mountain", "")
	bear := effectstest.PushTypedCard(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}")
	slimeID := effectstest.CastAndResolveCreature(t, g, "Acidic Slime", "Creature — Ooze", effectstest.AcidicSlimeOracle)
	// Mandatory: straight to the pick, no yes/no.
	if effectstest.OwedMays(g, caster.ID) > 0 {
		t.Errorf("Acidic Slime is mandatory — no yes/no prompt")
	}
	prompt := effectstest.LatestPickTarget(g, caster.ID)
	if prompt == nil {
		t.Fatalf("no pick_target prompt")
	}
	if !effectstest.HasID(prompt.PickTargetCards, mountain) || effectstest.HasID(prompt.PickTargetCards, bear) {
		t.Errorf("legal set = %v (want the land, not the bear)", prompt.PickTargetCards)
	}
	effectstest.PickCard(t, g, caster.ID, mountain)
	if effectstest.TriggerOnStack(g, slimeID) == nil {
		t.Fatalf("trigger not on the stack after the pick")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mountain) || !opp.Graveyard.Contains(mountain) {
		t.Errorf("Acidic Slime did not destroy the chosen land")
	}
}

// --- Eternal Witness --------------------------------------------

func TestEternalWitnessPicksFromOwnGraveyardOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	older := effectstest.PushGraveyardCardForTest(caster, "Older Body")
	newer := effectstest.PushGraveyardCardForTest(caster, "Latest Body")
	theirs := effectstest.PushGraveyardCardForTest(opp, "Their Body")

	witnessID := effectstest.CastAndResolveCreature(t, g, "Eternal Witness", "Creature — Human Shaman", effectstest.EternalWitnessOracle)
	prompt := effectstest.LatestPickTarget(g, caster.ID)
	if prompt == nil {
		t.Fatalf("no pick_target prompt")
	}
	if !effectstest.HasID(prompt.PickTargetCards, older) || !effectstest.HasID(prompt.PickTargetCards, newer) || effectstest.HasID(prompt.PickTargetCards, theirs) {
		t.Errorf("legal set = %v (want both of mine, not theirs)", prompt.PickTargetCards)
	}
	// Choose the OLDER one — the S14 auto-pick could only ever take the top.
	effectstest.PickCard(t, g, caster.ID, older)
	if effectstest.TriggerOnStack(g, witnessID) == nil {
		t.Fatalf("trigger not on the stack")
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, caster.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !caster.Hand.Contains(older) {
		t.Errorf("chosen graveyard card not returned to hand")
	}
	if !caster.Graveyard.Contains(newer) {
		t.Errorf("the un-chosen card left the graveyard")
	}
}

func TestEternalWitnessEmptyGraveyardNoPrompt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[0]
	effectstest.CastAndResolveCreature(t, g, "Eternal Witness", "Creature — Human Shaman", effectstest.EternalWitnessOracle)
	if effectstest.OwedMays(g, caster.ID) > 0 || effectstest.LatestPickTarget(g, caster.ID) != nil {
		t.Errorf("empty graveyard: trigger must be removed without any prompt")
	}
}
