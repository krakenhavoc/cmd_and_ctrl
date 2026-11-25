package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b283_cards_test.go — S82 older sets, Shadows over Innistrad
// batch 283 (#5543): delirium cards.

const (
	oldB283Swallower = "6cc99047-91dc-4ef2-99f3-65932b82ca31"
	oldB283Stranger  = "7579e1a2-2747-4e37-9a7a-c13b62190195"
	oldB283Reaper    = "94cce37a-a752-4cde-a5b5-e8a15f8af52c"
	oldB283Tooth     = "b7c165a4-53a2-4117-a452-b9a8fe309c09"
	oldB283Blob      = "bddf0772-503c-42f1-bb86-1dfeeedf4a7e"
	oldB283Brain     = "bef3fca4-7290-416c-a100-1b7abbe7886e"
	oldB283Scribe    = "d2ef475d-6043-4da8-83f0-dffa76052707"
	oldB283Angel     = "f888f3ba-8316-4c10-bf97-873731f7d364"
)

func TestOldB283CardsAreFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB283Swallower: "Soul Swallower", oldB283Stranger: "Kindly Stranger", oldB283Stranger + "#1": "Demon-Possessed Witch",
		oldB283Reaper: "Reaper of Flight Moonsilver", oldB283Tooth: "Tooth Collector", oldB283Blob: "Inexorable Blob",
		oldB283Brain: "Pick the Brain", oldB283Scribe: "Manic Scribe", oldB283Angel: "Angel of Deliverance",
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

func TestOldB283SoulSwallowerGrowsAtYourUpkeepOnlyWithDelirium(t *testing.T) {
	for _, delirium := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		swallower := effectstest.B12Push(g, me.ID, "Soul Swallower", "Creature — Wurm", oldB283Swallower, 3, 3)
		if delirium {
			effectstest.B24SeedTypes(me, 4)
		} else {
			effectstest.B24SeedTypes(me, 3)
		}
		effectstest.AdvanceToUpkeepOf(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if delirium {
			want = 3
		}
		if got := effectstest.CountersOn(g, swallower, game.CounterPlusOne); got != want {
			t.Errorf("delirium=%v: %d counters, want %d", delirium, got, want)
		}
		if !effectstest.HasEffectiveKeyword(t, g, swallower, "trample") {
			t.Error("no trample")
		}
	}
}

func oldB283Stranger2(g *game.Game, owner uuid.UUID) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(), OracleID: oldB283Stranger, Layout: game.LayoutTransform, Owner: owner, Controller: owner,
		Faces: []game.Face{
			{Name: "Kindly Stranger", TypeLine: "Creature — Human", ManaCost: "{2}{B}", Colors: []string{"B"}, Power: 2, Toughness: 3},
			{Name: "Demon-Possessed Witch", TypeLine: "Creature — Human Shaman", Colors: []string{"B"}, Power: 4, Toughness: 3},
		},
	}
	c.SetFace(0)
	return effectstest.PushBattlefieldCardWithTimestamp(g, c)
}

func TestOldB283KindlyStrangerTransformsOnlyWithDeliriumAndMayDestroy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	stranger := oldB283Stranger2(g, me.ID)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.B24SeedTypes(me, 3)
	effectstest.ApaMana(me, "B", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, stranger, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("transformed with three card types in the graveyard")
	}
	effectstest.GyCard(me, "Old Artifact", "Artifact", "", "")
	effectstest.ApaActivate(t, g, me, stranger, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.TransformedPermanent(effectstest.B12Card(t, g, stranger)) {
		t.Fatal("the Stranger did not transform")
	}
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the targeted creature was not destroyed")
	}
	if p, tou := effectstest.EffectivePower(t, g, stranger), effectstest.EffectiveToughness(t, g, stranger); p != 4 || tou != 3 {
		t.Errorf("the Witch is %d/%d, want 4/3", p, tou)
	}
}

func TestOldB283KindlyStrangerMayDeclineToDestroy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	stranger := oldB283Stranger2(g, me.ID)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.B24SeedTypes(me, 4)
	effectstest.ApaMana(me, "B", "C", "C")
	effectstest.ApaActivate(t, g, me, stranger, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Error("the creature was destroyed though the controller declined")
	}
}

func TestOldB283ReaperPumpsBySacrificingAnotherCreatureOnlyWithDelirium(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	reaper := effectstest.B12Push(g, me.ID, "Reaper of Flight Moonsilver", "Creature — Angel", oldB283Reaper, 3, 3)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.B24SeedTypes(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, reaper, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err == nil {
		t.Fatal("activated with three card types in the graveyard")
	}
	if !g.Battlefield.Contains(fodder) {
		t.Fatal("the creature was sacrificed for a refused activation")
	}
	effectstest.GyCard(me, "Old Artifact", "Artifact", "", "")
	if err := g.ActivateCatalogAbility(me.ID, reaper, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{reaper}}); err == nil {
		t.Error("the Reaper sacrificed itself (\"another creature\")")
	}
	effectstest.ApaActivate(t, g, me, reaper, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the other creature was not sacrificed")
	}
	if p, tou := effectstest.EffectivePower(t, g, reaper), effectstest.EffectiveToughness(t, g, reaper); p != 5 || tou != 4 {
		t.Errorf("the Reaper is %d/%d, want 5/4", p, tou)
	}
	if !effectstest.HasEffectiveKeyword(t, g, reaper, "flying") {
		t.Error("no flying")
	}
}

func TestOldB283ToothCollectorShrinksOnEntryAndAtOpponentUpkeepsWithDelirium(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 3)
	other := effectstest.B12Creature(g, opp2.ID, "Other", "Creature — Bear", 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastAndResolveCreature(t, g, "Tooth Collector", "Creature — Human Rogue", oldB283Tooth)
	effectstest.AnswerPickTargetIfOpen(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 {
		t.Fatalf("entry: the chosen creature has power %d, want 2", p)
	}
	if p := effectstest.EffectivePower(t, g, other); p != 3 {
		t.Errorf("entry: another creature has power %d, want 3", p)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, theirs); p != 3 {
		t.Errorf("no delirium: power %d at the upkeep, want 3", p)
	}
	effectstest.B24SeedTypes(me, 4)
	effectstest.AdvanceToUpkeepOf(t, g, 2)
	effectstest.AnswerPickTargetIfOpen(t, g, other)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, other); p != 2 {
		t.Errorf("delirium: power %d after the upkeep trigger, want 2", p)
	}
}

func TestOldB283InexorableBlobMakesATappedAttackingOozeOnlyWithDelirium(t *testing.T) {
	for _, delirium := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		blob := effectstest.B12Push(g, me.ID, "Inexorable Blob", "Creature — Ooze", oldB283Blob, 3, 3)
		if delirium {
			effectstest.B24SeedTypes(me, 4)
		}
		effectstest.DeclareAttack(t, g, opp.ID, blob)
		effectstest.PassAndSettleAttackSplit(t, g, me.ID)
		n, tapped := effectstest.CountTokensNamed(g, me.ID, "Ooze")
		if delirium {
			if n != 1 || tapped != 1 {
				t.Fatalf("delirium: %d Ooze tokens (%d tapped), want 1 tapped", n, tapped)
			}
			for _, c := range g.Battlefield.Cards {
				if c.Name == "Ooze" && c.IsToken() {
					if p, tou := effectstest.EffectivePower(t, g, c.InstanceID), effectstest.EffectiveToughness(t, g, c.InstanceID); p != 3 || tou != 3 {
						t.Errorf("the Ooze is %d/%d, want 3/3", p, tou)
					}
					if c.AttackingTarget == uuid.Nil {
						t.Error("the Ooze is not attacking")
					}
				}
			}
		} else if n != 0 {
			t.Errorf("no delirium: %d Ooze tokens", n)
		}
	}
}

func TestOldB283PickTheBrainExilesAndSearchesOnlyWithDelirium(t *testing.T) {
	for _, delirium := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		opp := g.Seats[1]
		ids := effectstest.RevealHand(opp,
			game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
			game.Card{Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}"},
			game.Card{Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}"})
		gy := uuid.New()
		opp.Graveyard.PushTop(game.Card{InstanceID: gy, Name: "Divination", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
		lib := uuid.New()
		opp.Library.PushTop(game.Card{InstanceID: lib, Name: "Divination", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
		other := uuid.New()
		opp.Library.PushTop(game.Card{InstanceID: other, Name: "Opt", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
		if delirium {
			effectstest.B24SeedTypes(me, 4)
		}
		effectstest.CastAt(t, g, "Pick the Brain", "Sorcery", oldB283Brain, effectstest.PbPlayer(opp.ID))
		p := effectstest.OpenVariant(t, g)
		for _, id := range p.PickTargetCards {
			if id == ids[0] {
				t.Fatal("a land was offered")
			}
		}
		if err := g.ResolvePendingChoice(p.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
			t.Fatalf("resolve the pick: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Exile.Contains(ids[1]) {
			t.Fatalf("delirium=%v: the chosen card was not exiled", delirium)
		}
		if !delirium {
			if effectstest.SearchChoiceFor(g, me.ID) != nil || g.Exile.Contains(ids[2]) || g.Exile.Contains(gy) {
				t.Error("no delirium: a search happened")
			}
			continue
		}
		if effectstest.SearchChoiceFor(g, me.ID) == nil {
			t.Fatal("delirium: no search prompt")
		}
		effectstest.AnswerSearchByID(t, g, me.ID, ids[2], gy, lib)
		for name, id := range map[string]uuid.UUID{"hand copy": ids[2], "graveyard copy": gy, "library copy": lib} {
			if !g.Exile.Contains(id) {
				t.Errorf("the %s was not exiled", name)
			}
		}
		if g.Exile.Contains(other) || g.Exile.Contains(ids[0]) {
			t.Error("a card with another name was exiled")
		}
	}
}

func TestOldB283ManicScribeMillsOnEntryAndAtOpponentUpkeepsWithDelirium(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	for i := 0; i < 20; i++ {
		for _, p := range []*game.Player{me, opp, opp2} {
			effectstest.SeedLibraryTop(p, "Filler", "Land", "")
		}
	}
	effectstest.AdvanceToMain(t, g)
	g1, g2, mine := opp.Graveyard.Size(), opp2.Graveyard.Size(), me.Graveyard.Size()
	effectstest.CastAndResolveCreature(t, g, "Manic Scribe", "Creature — Human Wizard", oldB283Scribe)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Size() != g1+3 || opp2.Graveyard.Size() != g2+3 || me.Graveyard.Size() != mine {
		t.Errorf("entry: graveyards %d/%d/%d from %d/%d/%d, want each opponent +3", me.Graveyard.Size(), opp.Graveyard.Size(), opp2.Graveyard.Size(), mine, g1, g2)
	}
	g1 = opp.Graveyard.Size()
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Size() != g1 {
		t.Errorf("no delirium: the opponent milled %d at their upkeep", opp.Graveyard.Size()-g1)
	}
	effectstest.B24SeedTypes(me, 4)
	g1, g2 = opp.Graveyard.Size(), opp2.Graveyard.Size()
	effectstest.AdvanceToUpkeepOf(t, g, 2)
	effectstest.PassPriorityAroundTable(t, g)
	// The upkeep player draws in their draw step, not before, so the upkeep mill is the only change.
	if opp2.Graveyard.Size() != g2+3 || opp.Graveyard.Size() != g1 {
		t.Errorf("delirium: graveyards %d/%d from %d/%d, want only the upkeep's player +3", opp.Graveyard.Size(), opp2.Graveyard.Size(), g1, g2)
	}
}

func TestOldB283AngelOfDeliveranceExilesOncePerBatchWithDelirium(t *testing.T) {
	for _, delirium := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		angel := effectstest.B12Push(g, me.ID, "Angel of Deliverance", "Creature — Angel", oldB283Angel, 6, 6)
		reach := func(name string) uuid.UUID {
			return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Spider", Power: 1, Toughness: 3,
				Keywords: []string{"reach"}, Owner: opp.ID, Controller: opp.ID,
			})
		}
		a, b := reach("Blocker A"), reach("Blocker B")
		c := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 1, 1)
		if delirium {
			effectstest.B24SeedTypes(me, 4)
		}
		effectstest.DeclareAttack(t, g, opp.ID, angel)
		// Both blockers block, so the Angel deals damage to two creatures at once.
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		for _, blocker := range []uuid.UUID{a, b} {
			if err := g.DeclareBlocker(blocker, angel); err != nil {
				t.Fatalf("DeclareBlocker: %v", err)
			}
		}
		effectstest.LockInCombatDeclaration(t, g)
		effectstest.AdvanceTo(t, g, game.StepCombatDamage)
		for _, pc := range g.PendingChoices {
			if pc != nil && pc.Kind == game.PendingChoiceDamageAssignment && pc.DamageAssignment != nil {
				var entries []game.DamageAssignmentEntry
				for _, blocker := range pc.DamageAssignment.BlockerIDs {
					entries = append(entries, game.DamageAssignmentEntry{BlockerID: blocker, Amount: pc.DamageAssignment.AttackerPower / len(pc.DamageAssignment.BlockerIDs)})
				}
				if err := g.ResolveDamageAssignment(pc.ID, pc.Chooser, entries, 0); err != nil {
					t.Fatalf("ResolveDamageAssignment: %v", err)
				}
				break
			}
		}
		effectstest.AnswerPickTargetIfOpen(t, g, c)
		effectstest.PassPriorityAroundTable(t, g)
		exiled := 0
		for _, id := range []uuid.UUID{a, b, c} {
			if g.Exile.Contains(id) {
				exiled++
			}
		}
		if delirium && exiled != 1 {
			t.Errorf("delirium: %d creatures exiled by the trigger, want 1", exiled)
		}
		if !delirium && exiled != 0 {
			t.Errorf("no delirium: %d creatures exiled", exiled)
		}
		if delirium && !g.Exile.Contains(c) {
			t.Error("the chosen creature was not the one exiled")
		}
	}
}
