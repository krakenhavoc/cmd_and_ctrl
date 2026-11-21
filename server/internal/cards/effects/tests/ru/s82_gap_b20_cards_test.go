package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b20_cards_test.go — Mystery Booster 2 creatures, GAP batch 20 (#5948).

const (
	gapHuntCheetah   = "057e1cfc-d528-49e1-a386-9fb1edc1d135"
	gapStoneDrakeOrc = "7373b50b-a707-40fb-967a-550c31e6ea59"
)

func TestS82GapB20EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapHuntCheetah: "Hunting Cheetah", gapStoneDrakeOrc: "Stone Drake",
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

func TestS82GapB20HuntingCheetahFetchesAForestOnDamageToAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cheetah := effectstest.B12Push(g, me.ID, "Hunting Cheetah", "Creature — Cat", gapHuntCheetah, 2, 3)
	forest := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)

	// Damage to the controller does nothing.
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(cheetah, me.ID, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Size() != 0 || len(g.PendingChoices) != 0 {
		t.Fatal("damage to its own controller triggered the search")
	}

	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(cheetah, opp.ID, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, forest)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(forest) {
		t.Error("the Forest should be in hand")
	}
}

func TestS82GapB20StoneDrakeDistractTapsALandAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.B12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	hand := me.Hand.Size()
	gapStoneDrake(t, g, me, 0)
	effectstest.PickCard(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, land) {
		t.Error("the land should be tapped")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand grew by %d, want 1", got)
	}
}

func TestS82GapB20StoneDrakeEnthrallDiscardsPerSpellCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.D2dSetHand(opp, 3)
	gapStoneDrake(t, g, me, 1)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	// Only the Drake itself was cast this turn: one discard.
	effectstest.DiscardFromHand(t, g, opp.ID)
	if got := opp.Hand.Size(); got != 2 {
		t.Errorf("opponent's hand = %d, want 2 (one spell cast this turn)", got)
	}
}

func gapStoneDrake(t *testing.T, g *game.Game, me *game.Player, mode int) {
	t.Helper()
	effectstest.CastAndResolveCreature(t, g, "Stone Drake", "Creature — Drake", gapStoneDrakeOrc)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatalf("no mode_pick prompt from the ETB: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
}
