package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b468_cards_test.go — Mirrodin Besieged, S82 OLD batch 468 (#5576).
// Every test plays the real catalog entry through the engine.

const (
	oldB468MirranSpy      = "8e5d72a1-6ae0-445d-bac7-62da3bd3a8d6"
	oldB468CorruptedConsc = "c645fab4-51ac-440c-b112-57777c8f4676"
	oldB468TrainingDrone  = "4b40ad5f-b270-44e9-b62f-e94912bc3346"
	oldB468FranticSalvage = "9d2e7099-dcd5-428f-a732-2a82b8d1dd1a"
)

func TestOldB468EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB468MirranSpy:      "Mirran Spy",
		oldB468CorruptedConsc: "Corrupted Conscience",
		oldB468TrainingDrone:  "Training Drone",
		oldB468FranticSalvage: "Frantic Salvage",
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

func TestOldB468MirranSpyUntapsACreatureWhenYouCastAnArtifactSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Mirran Spy", "Creature — Drone", oldB468MirranSpy, 1, 3)
	tapped := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Tapped Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Tapped: true, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.AvB1Hand(me, game.Card{Name: "Mind Stone", TypeLine: "Artifact", ManaCost: "{2}"})
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the artifact: %v", err)
	}
	effectstest.PickCard(t, g, me.ID, tapped)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, tapped) {
		t.Error("the creature is still tapped")
	}
}

func TestOldB468MirranSpyIgnoresNonartifactSpellsAndMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Mirran Spy", "Creature — Drone", oldB468MirranSpy, 1, 3)
	tapped := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Tapped Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Tapped: true, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	sorcery := effectstest.AvB1Hand(me, game.Card{Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}"})
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the sorcery: %v", err)
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a nonartifact spell triggered the Spy")
	}
	effectstest.PassPriorityAroundTable(t, g)
	rock := effectstest.AvB1Hand(me, game.Card{Name: "Mind Stone", TypeLine: "Artifact", ManaCost: "{2}"})
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the artifact: %v", err)
	}
	effectstest.PickCard(t, g, me.ID, tapped)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, tapped) {
		t.Error("declined, yet the creature untapped")
	}
}

func TestOldB468CorruptedConscienceStealsAndGivesInfectUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	aura := effectstest.Enchant(t, g, "Corrupted Conscience", oldB468CorruptedConsc, bear)
	c := effectstest.BattlefieldCardCopy(t, g, bear)
	if c.Controller != me.ID {
		t.Fatalf("the Bear is controlled by %v, want the Aura's controller", c.Controller)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "infect") {
		t.Error("the enchanted creature has no infect")
	}
	effectstest.P1Destroy(t, g, aura)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.BattlefieldCardCopy(t, g, bear); c.Controller != opp.ID {
		t.Error("control did not revert when the Aura left")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "infect") {
		t.Error("infect stayed after the Aura left")
	}
}

func TestOldB468CorruptedConscienceEnchantsOnlyACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	rock := effectstest.B12Permanent(g, g.Seats[1].ID, "Rock", "Artifact")
	if err := effectstest.CastCatalogSpellErr(t, g, "Corrupted Conscience", effectstest.AuraTypeLine, oldB468CorruptedConsc, effectstest.CardRefs(rock)); err == nil {
		t.Fatal("the Aura was cast on a noncreature")
	}
}

func TestOldB468TrainingDroneAttacksAndBlocksOnlyWhileEquipped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drone := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Training Drone", TypeLine: "Artifact Creature — Drone",
		OracleID: oldB468TrainingDrone, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	attacker := effectstest.B12Creature(g, opp.ID, "Their Attacker", "Creature — Bear", 2, 2)
	sword := effectstest.B12Permanent(g, me.ID, "Sword", "Artifact — Equipment")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(drone, opp.ID); err == nil {
		t.Fatal("an unequipped Training Drone attacked")
	}
	oldB465Block(t, g, attacker, drone, false)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(sword, game.TargetRef{Kind: game.TargetCard, ID: drone}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
	oldB465Block(t, g, attacker, drone, true)
	if err := g.DeclareAttacker(drone, opp.ID); err != nil {
		t.Fatalf("the equipped Drone could not attack: %v", err)
	}
	g.WithWriteLock(func() { _ = g.UnattachForEffect(sword) })
	oldB465Block(t, g, attacker, drone, false)
}

func oldB465Block(t *testing.T, g *game.Game, attacker, blocker uuid.UUID, want bool) {
	t.Helper()
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		a, _ := g.LookupCardForEffect(attacker)
		b, _ := g.LookupCardForEffect(blocker)
		if r := g.BlockPairRefusalLocked(&a, &b); r.Legal() != want {
			t.Errorf("%s may block = %v, want %v (%q)", b.Name, r.Legal(), want, r.Reason)
		}
	})
}

func TestOldB468FranticSalvagePutsArtifactsOnTopThenDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	a := uuid.New()
	b := uuid.New()
	creature := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: a, Name: "Rock A", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: b, Name: "Rock B", TypeLine: "Artifact Creature — Golem", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: creature, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	if err := effectstest.CastCatalogSpellErr(t, g, "Frantic Salvage", "Instant", oldB468FranticSalvage, effectstest.CardRefs(creature)); err == nil {
		t.Fatal("a nonartifact card was targeted")
	}
	hand, lib := me.Hand.Size(), me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Frantic Salvage", "Instant", oldB468FranticSalvage, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.PutInLibraryPrompt(g)
	if prompt == nil {
		t.Fatal("no ordering prompt for the two artifact cards")
	}
	if err := g.ResolvePutInLibrary(prompt.ID, me.ID, nil, []uuid.UUID{b, a}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// b was put first on top, then a: the draw takes b, the first of the pile.
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
	if me.Library.Size() != lib+1 {
		t.Errorf("library %d → %d, want two put back and one drawn", lib, me.Library.Size())
	}
	if me.Graveyard.Contains(a) && me.Graveyard.Contains(b) {
		t.Error("the artifact cards stayed in the graveyard")
	}
	if !me.Graveyard.Contains(creature) {
		t.Error("the untargeted creature card left the graveyard")
	}
}

func TestOldB468FranticSalvageWithNoTargetsJustDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Frantic Salvage", "Instant", oldB468FranticSalvage, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
}
