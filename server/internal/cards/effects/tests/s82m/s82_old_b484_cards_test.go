package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b484_cards_test.go — Rise of the Eldrazi, batch 484 (S82, #5581).
// Each test plays the real catalog entry through the engine.

const (
	oldB484LustOracle      = "0ba206f3-c3f3-4e53-b072-e5e70feaa00d"
	oldB484GuardOracle     = "3b20e187-c200-40a0-8db2-481f0bf90168"
	oldB484AppetiteOracle  = "52d34007-38c0-4858-8ec8-c622da916277"
	oldB484WakeOracle      = "950e345d-455e-4bd4-be8e-bc3381ac3050"
	oldB484ScuttlerOracle  = "7d1f5f6f-af3d-4b9b-acaf-debf885f660a"
	oldB484BloodlordOracle = "a35e1882-b878-49cf-a06c-be12a06d4977"
	oldB484LizardOracle    = "e241c8c3-359d-4e20-97a6-a873c50f1ef9"
	oldB484SiltOracle      = "f4542dff-1ba1-4b15-beaf-51a7a42782cf"
	oldB484DoppelOracle    = "08528c5e-9a01-42a4-a00b-c1ec5eaae64d"
	oldB484ObserverOracle  = "0b5d49ce-7e02-4de6-8a65-bfc88ac0cf72"
	oldB484GomazoaOracle   = "bc5a7015-c13b-4d92-8afe-50b7956c14dd"
)

func TestOldB484EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB484LustOracle: "Lust for War", oldB484GuardOracle: "Guard Duty",
		oldB484AppetiteOracle: "Demonic Appetite", oldB484WakeOracle: "Luminous Wake",
		oldB484ScuttlerOracle: "Jwari Scuttler", oldB484BloodlordOracle: "Arrogant Bloodlord",
		oldB484LizardOracle: "Lagac Lizard", oldB484SiltOracle: "Nema Siltlurker",
		oldB484DoppelOracle: "Renegade Doppelganger", oldB484ObserverOracle: "Merfolk Observer",
		oldB484GomazoaOracle: "Dormant Gomazoa",
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

func TestOldB484VanillaCreaturesHaveNoAbilities(t *testing.T) {
	for _, tc := range []struct{ oracle, name string }{
		{oldB484ScuttlerOracle, "Jwari Scuttler"}, {oldB484LizardOracle, "Lagac Lizard"}, {oldB484SiltOracle, "Nema Siltlurker"},
	} {
		spec, ok := effects.Lookup(tc.oracle)
		if !ok {
			t.Fatalf("%s is not registered", tc.name)
		}
		if len(spec.Triggered) != 0 || len(spec.Activated) != 0 || len(spec.Static) != 0 || len(spec.PrintedKeywords) != 0 {
			t.Errorf("%s declares abilities, but it is vanilla", tc.name)
		}
	}
}

func TestOldB484GuardDutyGivesDefender(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	host := effectstest.SeedBear(g, me.ID)
	if effectstest.HasEffectiveKeyword(t, g, host, "defender") {
		t.Fatal("the bear has defender before the Aura")
	}
	effectstest.AuraCast(t, g, "Guard Duty", oldB484GuardOracle, host)
	if !effectstest.HasEffectiveKeyword(t, g, host, "defender") {
		t.Error("the enchanted creature has no defender")
	}
}

func TestOldB484LustForWarMustAttackAndHurtsItsControllerWhenTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.SeedBear(g, me.ID)
	effectstest.Unsick(g, mine)
	effectstest.AuraCast(t, g, "Lust for War", oldB484LustOracle, mine)
	life := me.Life
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != mine {
		t.Fatalf("pass with the enchanted creature home = %v, want a refusal naming it", err)
	}
	if err := g.DeclareAttacker(mine, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-3 {
		t.Errorf("life %d, want %d (3 damage to the tapped creature's controller)", me.Life, life-3)
	}
}

func TestOldB484LustForWarDamagesTheCurrentControllerOfTheTappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	theirs := effectstest.SeedBear(g, opp.ID)
	effectstest.AuraCast(t, g, "Lust for War", oldB484LustOracle, theirs)
	life, myLife := opp.Life, me.Life
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventTapCard, CardID: theirs}) })
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 || me.Life != myLife {
		t.Errorf("the creature's controller is at %d (want %d), the Aura's controller at %d (want %d)", opp.Life, life-3, me.Life, myLife)
	}
}

func TestOldB484DemonicAppetitePumpsAndMakesYouSacrificeAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	host := effectstest.SeedBear(g, me.ID)
	other := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.SeedBear(g, opp.ID)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Demonic Appetite", effectstest.AuraTypeLine, oldB484AppetiteOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err == nil {
		t.Fatal("enchanted a creature an opponent controls")
	}
	effectstest.AuraCast(t, g, "Demonic Appetite", oldB484AppetiteOracle, host)
	if p, tough := effectstest.PtOf(t, g, host); p != 5 || tough != 5 {
		t.Errorf("the enchanted creature is %d/%d, want 5/5", p, tough)
	}
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(other) {
		t.Error("the chosen creature was not sacrificed")
	}
	if !g.Battlefield.Contains(host) || !g.Battlefield.Contains(theirs) {
		t.Error("a creature that was not chosen was sacrificed")
	}
}

func TestOldB484LuminousWakeGainsFourOnAttackAndBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.SeedBear(g, opp.ID)
	effectstest.AuraCast(t, g, "Luminous Wake", oldB484WakeOracle, mine)
	effectstest.AuraCast(t, g, "Luminous Wake", oldB484WakeOracle, theirs)
	life := me.Life
	effectstest.DeclareAttack(t, g, opp.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("attacking: life %d, want %d", me.Life, life+4)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(theirs, mine); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+8 {
		t.Errorf("after the enchanted blocker too: life %d, want %d (the Aura's controller gains)", me.Life, life+8)
	}
}

func TestOldB484ArrogantBloodlordDiesAtEndOfCombatOnlyAgainstSmallCreatures(t *testing.T) {
	for _, tc := range []struct {
		name      string
		bloodlord bool // the Bloodlord attacks (and is blocked)
		power     int
		dies      bool
	}{
		{"attacks, blocked by a 1-power creature", true, 1, true},
		{"attacks, blocked by a 2-power creature", true, 2, false},
		{"blocks a 1-power attacker", false, 1, true},
		{"blocks a 2-power attacker", false, 2, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			var lord, other uuid.UUID
			if tc.bloodlord {
				lord = effectstest.B12Push(g, me.ID, "Arrogant Bloodlord", "Creature — Vampire Knight", oldB484BloodlordOracle, 4, 4)
				other = effectstest.B12Push(g, opp.ID, "Small", "Creature — Bear", "", tc.power, 9)
				effectstest.Pr7aBlockedCombat(t, g, opp.ID, lord, other)
			} else {
				other = effectstest.B12Push(g, me.ID, "Small", "Creature — Bear", "", tc.power, 9)
				lord = effectstest.B12Push(g, opp.ID, "Arrogant Bloodlord", "Creature — Vampire Knight", oldB484BloodlordOracle, 4, 4)
				effectstest.Pr7aBlockedCombat(t, g, opp.ID, other, lord)
			}
			effectstest.LockInBlocks(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AdvanceTo(t, g, game.StepEndCombat)
			effectstest.PassPriorityAroundTable(t, g)
			g.RunStateChecksForTest()
			if alive := g.Battlefield.Contains(lord); alive == tc.dies {
				t.Errorf("Bloodlord on the battlefield = %v, want %v", alive, !tc.dies)
			}
		})
	}
}

func TestOldB484RenegadeDoppelgangerCopiesOnceAndMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	doppel := effectstest.B12Push(g, me.ID, "Renegade Doppelganger", "Creature — Shapeshifter", oldB484DoppelOracle, 0, 1)
	// A "no" spends nothing.
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Small", TypeLine: "Creature — Bear", Power: 1, Toughness: 1})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, doppel); p != 0 || tough != 1 {
		t.Fatalf("declined, yet it is %d/%d", p, tough)
	}
	// A "yes" copies the creature that entered.
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Big", TypeLine: "Creature — Ogre", Power: 5, Toughness: 4})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, doppel); p != 5 || tough != 4 {
		t.Fatalf("it is %d/%d after copying the 5/4, want 5/4", p, tough)
	}
	// It lost the ability for the rest of the turn: nothing asks again.
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Huge", TypeLine: "Creature — Ogre", Power: 7, Toughness: 7})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestTriggerPrompt(g, me.ID); c != nil {
		t.Error("asked to copy a second creature in the same turn")
	}
	if p, tough := effectstest.PtOf(t, g, doppel); p != 5 || tough != 4 {
		t.Errorf("it is %d/%d after a second creature entered, want to stay 5/4", p, tough)
	}
}

func TestOldB484RenegadeDoppelgangerIgnoresCreaturesOtherThanYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	doppel := effectstest.B12Push(g, me.ID, "Renegade Doppelganger", "Creature — Shapeshifter", oldB484DoppelOracle, 0, 1)
	effectstest.EnterCard(t, g, opp.ID, game.Card{Name: "Theirs", TypeLine: "Creature — Ogre", Power: 5, Toughness: 4})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestTriggerPrompt(g, me.ID); c != nil {
		t.Error("asked about an opponent's creature")
	}
	if p, tough := effectstest.PtOf(t, g, doppel); p != 0 || tough != 1 {
		t.Errorf("it is %d/%d, want 0/1", p, tough)
	}
}

func TestOldB484MerfolkObserverLooksWithoutMovingAnything(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	top := uuid.New()
	opp.Library.PushTop(game.Card{InstanceID: top, Name: "Top", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	size := opp.Library.Size()
	effectstest.PlsB1Cast(t, g, "Merfolk Observer", "Creature — Merfolk Rogue", oldB484ObserverOracle, "", 2, 1)
	pick := effectstest.PassUntilPickTarget(t, g, me.ID)
	if pick == nil {
		t.Fatal("no target was asked for")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, effectstest.PlayerTarget(opp.ID)); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != size {
		t.Errorf("the library has %d cards, want %d", opp.Library.Size(), size)
	}
	if c, err := opp.Library.Top(); err != nil || c.InstanceID != top {
		t.Error("the top card is not the one that was there")
	}
}

func TestOldB484DormantGomazoaStaysTappedUntilYouAreTargetedBySpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gomazoa := effectstest.EnterFromHand(t, g, me.ID, "Dormant Gomazoa", "Creature — Jellyfish", oldB484GomazoaOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, gomazoa, "flying") {
		t.Error("no flying")
	}
	if !effectstest.TappedForTest(t, g, gomazoa) {
		t.Fatal("it did not enter tapped")
	}
	// It does not untap in its controller's untap step.
	effectstest.B12ToMyNextUpkeep(t, g)
	if !effectstest.TappedForTest(t, g, gomazoa) {
		t.Fatal("it untapped during its controller's untap step")
	}
	target := func(who uuid.UUID, kind game.StackItemKind) {
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventBecomesTarget, Target: who, Actor: opp.ID, TargeterKind: kind, StackItemID: uuid.New()})
		})
		effectstest.PassPriorityAroundTable(t, g)
	}
	// An ability, or another player, does not do it.
	target(me.ID, game.StackItemActivated)
	target(opp.ID, game.StackItemSpell)
	if c := effectstest.LatestTriggerPrompt(g, me.ID); c != nil || !effectstest.TappedForTest(t, g, gomazoa) {
		t.Fatal("untapped by an ability or by another player becoming a target")
	}
	// A spell targeting you does, if you say yes.
	target(me.ID, game.StackItemSpell)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedForTest(t, g, gomazoa) {
		t.Fatal("untapped although the player said no")
	}
	target(me.ID, game.StackItemSpell)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TappedForTest(t, g, gomazoa) {
		t.Error("still tapped after being targeted by a spell and saying yes")
	}
}

// A Doppelganger that became a copy of another Doppelganger has the ability
// again through the copy, but "it loses this ability for the rest of the
// turn": only the other one still asks.
func TestOldB484RenegadeDoppelgangerLosesTheAbilityEvenAsACopyOfADoppelganger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	first := effectstest.B12Push(g, me.ID, "Renegade Doppelganger", "Creature — Shapeshifter", oldB484DoppelOracle, 0, 1)
	second := effectstest.EnterFromHand(t, g, me.ID, "Renegade Doppelganger", "Creature — Shapeshifter", oldB484DoppelOracle)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(first) || !g.Battlefield.Contains(second) {
		t.Fatal("a Doppelganger left the battlefield")
	}
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Small", TypeLine: "Creature — Bear", Power: 1, Toughness: 1})
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestTriggerPrompt(g, me.ID)
	for prompt != nil {
		if prompt.Source == first {
			t.Fatal("the Doppelganger that already became a copy this turn asked again")
		}
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
		effectstest.PassPriorityAroundTable(t, g)
		prompt = effectstest.LatestTriggerPrompt(g, me.ID)
	}
}
