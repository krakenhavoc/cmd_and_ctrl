package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// paused_exit_cost_test.go — #1445 with real cards. A commander that an
// effect has destroyed stays on the battlefield while its owner answers
// CR 903.9, and no sacrifice cost may eat it in that window: one card
// per announcement site — Viscera Seer (an activated ability), Ashnod's
// Altar (a mana ability), Village Rites (a spell's additional cost).
// Without the gate each of them sacrificed the commander for value and
// the destroy's prompt was withdrawn as stale. The engine half is
// game/paused_exit_cost_test.go.

// seedCommanderCreature puts a commander creature onto the battlefield
// under `controller`.
func seedCommanderCreature(g *game.Game, controller uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID:  uuid.New(),
		Name:        "Some Commander",
		TypeLine:    "Legendary Creature — Human",
		Power:       3,
		Toughness:   3,
		Owner:       controller,
		Controller:  controller,
		IsCommander: true,
	})
}

// destroyedCommander destroys a fresh commander by effect and returns
// it with its owner's open CR 903.9 prompt.
func destroyedCommander(t *testing.T, g *game.Game, owner uuid.UUID) (uuid.UUID, *game.PendingChoice) {
	t.Helper()
	cmdr := seedCommanderCreature(g, owner)
	if err := g.DestroyPermanentForEffect(cmdr); err != nil {
		t.Fatalf("DestroyPermanentForEffect: %v", err)
	}
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceOptionalReplacement {
		t.Fatalf("want exactly one CR 903.9 prompt, have %d pending choices", len(g.PendingChoices))
	}
	if _, ok := battlefieldCard(g, cmdr); !ok {
		t.Fatal("the destroyed commander left before its owner answered — the premise is gone")
	}
	return cmdr, g.PendingChoices[0]
}

func TestVisceraSeerCannotEatADestroyedCommanderWhileItsOwnerIsAsked(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seer := pushCatalogPermanent(g, me.ID, "Viscera Seer", "Creature — Vampire Wizard", visceraSeerOracle, false)
	fodder := seedCreature(g, "Doomed Traveler", me.ID)
	cmdr, prompt := destroyedCommander(t, g, me.ID)

	if err := g.ActivateCatalogAbility(me.ID, seer, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{cmdr}}); !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("Viscera Seer naming the destroyed commander: err = %v, want ErrChoicePending", err)
	}
	if len(g.StackMeta) != 0 || len(g.PendingChoices) != 1 {
		t.Fatalf("stack %d, pending %d — the refusal paid or asked something", len(g.StackMeta), len(g.PendingChoices))
	}
	if err := g.ResolveOptionalReplacement(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	if !me.Command.Contains(cmdr) {
		t.Error("the commander did not reach the command zone")
	}
	if err := g.ActivateCatalogAbility(me.ID, seer, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("the Seer after the answer: %v", err)
	}
}

func TestAshnodsAltarCannotEatADestroyedCommanderWhileItsOwnerIsAsked(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	altar := seedPermanentWithOracle(g, me.ID, "Ashnod's Altar", "Artifact", ashnodsAltarOracle)
	fodder := seedCreature(g, "Doomed Traveler", me.ID)
	cmdr, prompt := destroyedCommander(t, g, me.ID)

	if err := g.ActivateManaAbility(me.ID, altar, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{cmdr}}); !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("Ashnod's Altar naming the destroyed commander: err = %v, want ErrChoicePending", err)
	}
	if len(me.ManaPool) != 0 {
		t.Fatalf("pool = %d mana — a destroyed commander paid for mana", len(me.ManaPool))
	}
	if err := g.ResolveOptionalReplacement(prompt.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	if !me.Graveyard.Contains(cmdr) {
		t.Error("the declined commander is not in the graveyard")
	}
	if err := g.ActivateManaAbility(me.ID, altar, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("the Altar after the answer: %v", err)
	}
	if len(me.ManaPool) != 2 {
		t.Errorf("pool = %d mana, want 2", len(me.ManaPool))
	}
}

func TestVillageRitesCannotSacrificeADestroyedCommanderWhileItsOwnerIsAsked(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fodder := seedCreature(g, "Doomed Traveler", me.ID)
	// To the main phase first: the prompt below blocks AdvanceStep, and
	// the refusal this test is about must come from the cost, not from
	// the walk to a step where an instant can be cast.
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	cmdr, prompt := destroyedCommander(t, g, me.ID)

	if _, err := tryCastWithSacrifice(g, "Village Rites", "Instant", villageRitesOracle, []uuid.UUID{cmdr}); !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("Village Rites naming the destroyed commander: err = %v, want ErrChoicePending", err)
	}
	if len(g.StackMeta) != 0 || len(g.PendingChoices) != 1 {
		t.Fatalf("stack %d, pending %d — the refusal cast or asked something", len(g.StackMeta), len(g.PendingChoices))
	}
	if err := g.ResolveOptionalReplacement(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	if !me.Command.Contains(cmdr) {
		t.Error("the commander did not reach the command zone")
	}
	castWithSacrifice(t, g, "Village Rites", "Instant", villageRitesOracle, fodder)
}
