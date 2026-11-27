package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b688_test.go — Mercadian Masques creatures, batch 688 (#5619). Every card is the
// real catalog entry played through the engine.

const (
	d688Jhovall   = "8db78c62-e0b9-4ed7-b29f-238f23f79bc4"
	d688ChoManno  = "91af5e35-b3b8-43ce-b1ea-997ed74e4ad2"
	d688Lithophg  = "b4eb3d7e-a234-4ed5-8611-fcb818686fcf"
	d688Volunteer = "bd0031c3-ff88-4bc5-82a0-2187c60a7d33"
	d688Raider    = "ca011fb1-fdc6-4eec-9344-c247339292ab"
	d688Outrigger = "cd57638e-8444-4522-8638-4514dec8b3ed"
	d688Fountain  = "e6527ba3-3293-4a29-a2f9-7a81f8f27b7c"
	d688Satyr     = "fba1329d-d8bf-4c5e-9cf1-5410a42f732e"
	d688Wall      = "Creature — Wall"
)

func TestS82OldB688EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		d688Jhovall: "Wild Jhovall", d688ChoManno: "Cho-Manno, Revolutionary", d688Lithophg: "Lithophage",
		d688Volunteer: "Fresh Volunteers", d688Raider: "Saprazzan Raider", d688Outrigger: "Saprazzan Outrigger",
		d688Fountain: "Fountain Watch", d688Satyr: "Lumbering Satyr",
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

func TestS82OldB688VanillaCreaturesHaveNoAbilities(t *testing.T) {
	for _, id := range []string{d688Jhovall, d688Volunteer} {
		spec, _ := effects.Lookup(id)
		if len(spec.Triggered) != 0 || len(spec.Static) != 0 || len(spec.Activated) != 0 || len(spec.Replacements) != 0 {
			t.Errorf("%s has abilities", spec.Name)
		}
	}
}

func TestS82OldB688ChoMannoPreventsAllDamageToItselfOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cho := effectstest.B12Push(g, me.ID, "Cho-Manno, Revolutionary", "Legendary Creature — Human Rebel", d688ChoManno, 2, 2)
	other := effectstest.SeedBear(g, me.ID)
	effectstest.DealDamageTo(t, g, uuid.New(), cho, 9)
	effectstest.DealDamageTo(t, g, uuid.New(), other, 2)
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(cho) {
		t.Error("Cho-Manno died to damage")
	}
	if g.Battlefield.Contains(other) {
		t.Error("another creature was protected")
	}
}

func TestS82OldB688LithophageTaxesItsControllersUpkeep(t *testing.T) {
	t.Run("sacrifices a Mountain", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		litho := effectstest.B12Push(g, me.ID, "Lithophage", "Creature — Insect", d688Lithophg, 7, 7)
		mountain := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
		forest := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		effectstest.AdvanceToUpkeepOf(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.Sweep05OptionKey(t, g, me.ID, "option-pick/sacrifice-this-unless-you-sacrifice-a-mountain")
		if len(c.PickOptions) != 2 {
			t.Fatalf("%d options, want 2", len(c.PickOptions))
		}
		if err := g.ResolveOptionPick(c.ID, me.ID, 0); err != nil {
			t.Fatalf("pick the Mountain: %v", err)
		}
		pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
		if pick == nil || !effectstest.HasID(pick.ChooseCards, mountain) || effectstest.HasID(pick.ChooseCards, forest) {
			t.Fatalf("the sacrifice must offer the Mountain only: %+v", pick)
		}
		effectstest.AnswerChooseCards(t, g, me.ID, mountain)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(mountain) || !g.Battlefield.Contains(litho) {
			t.Error("the Mountain should be gone and the Lithophage kept")
		}
	})
	t.Run("no Mountain sacrifices itself", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		litho := effectstest.B12Push(g, me.ID, "Lithophage", "Creature — Insect", d688Lithophg, 7, 7)
		effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		effectstest.AdvanceToUpkeepOf(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(litho) {
			t.Error("the Lithophage stayed with no Mountain to sacrifice")
		}
	})
	t.Run("declines and sacrifices itself", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		litho := effectstest.B12Push(g, me.ID, "Lithophage", "Creature — Insect", d688Lithophg, 7, 7)
		mountain := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		effectstest.AdvanceToUpkeepOf(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.Sweep05OptionKey(t, g, me.ID, "option-pick/sacrifice-this-unless-you-sacrifice-a-mountain")
		if err := g.ResolveOptionPick(c.ID, me.ID, 1); err != nil {
			t.Fatalf("pick the Lithophage: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(litho) || !g.Battlefield.Contains(mountain) {
			t.Error("the Lithophage should be gone and the Mountain kept")
		}
	})
}

func TestS82OldB688SaprazzanRaiderReturnsToHandWhenBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raider := effectstest.B12Push(g, me.ID, "Saprazzan Raider", "Creature — Merfolk", d688Raider, 1, 2)
	wall := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Wall", TypeLine: d688Wall, Power: 0, Toughness: 4})
	effectstest.DeclareAttack(t, g, opp.ID, raider)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, raider); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(raider) || !me.Hand.Contains(raider) {
		t.Error("the Raider did not return to its owner's hand when blocked")
	}
}

func TestS82OldB688SaprazzanOutriggerGoesOnTopOfTheLibraryAtEndOfCombat(t *testing.T) {
	for _, blocks := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		var out uuid.UUID
		var owner *game.Player
		if !blocks {
			out = effectstest.B12Push(g, me.ID, "Saprazzan Outrigger", "Creature — Merfolk", d688Outrigger, 5, 5)
			owner = me
			effectstest.DeclareAttack(t, g, opp.ID, out)
		} else {
			attacker := effectstest.SeedBear(g, me.ID)
			out = effectstest.B12Push(g, opp.ID, "Saprazzan Outrigger", "Creature — Merfolk", d688Outrigger, 5, 5)
			owner = opp
			effectstest.DeclareAttack(t, g, opp.ID, attacker)
			effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
			if err := g.DeclareBlocker(out, attacker); err != nil {
				t.Fatalf("DeclareBlocker: %v", err)
			}
			effectstest.LockInBlocks(t, g)
		}
		if !g.Battlefield.Contains(out) {
			t.Fatal("the Outrigger left before end of combat")
		}
		effectstest.AdvanceTo(t, g, game.StepEndCombat)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(out) {
			t.Errorf("blocks=%v: the Outrigger stayed on the battlefield", blocks)
		}
		top, err := owner.Library.Top()
		if err != nil || top.InstanceID != out {
			t.Errorf("blocks=%v: the Outrigger is not on top of its owner's library", blocks)
		}
	}
}

func TestS82OldB688FountainWatchGivesShroudToYourArtifactsAndEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Fountain Watch", "Creature — Human Cleric", d688Fountain, 2, 4)
	rock := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	ench := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Ward", TypeLine: "Enchantment"})
	bear := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Their Rock", TypeLine: "Artifact"})
	for id, want := range map[uuid.UUID]bool{rock: true, ench: true, bear: false, theirs: false} {
		if got := effectstest.HasEffectiveKeyword(t, g, id, "shroud"); got != want {
			t.Errorf("shroud on %s = %v, want %v", id, got, want)
		}
	}
}

func TestS82OldB688LumberingSatyrGivesEveryCreatureForestwalk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	satyr := effectstest.B12Push(g, me.ID, "Lumbering Satyr", "Creature — Satyr Beast", d688Satyr, 5, 4)
	mine := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.SeedBear(g, opp.ID)
	rock := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	for id, want := range map[uuid.UUID]bool{satyr: true, mine: true, theirs: true, rock: false} {
		if got := effectstest.HasEffectiveKeyword(t, g, id, "forestwalk"); got != want {
			t.Errorf("forestwalk on %s = %v, want %v", id, got, want)
		}
	}
}
