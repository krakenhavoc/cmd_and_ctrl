package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// paused_exit_tap_cost_test.go — #1427 with real cards. A commander
// that an effect has destroyed stays on the battlefield while its
// owner answers CR 903.9, and no cost may TAP it in that window: one
// card per tap shape — Birds of Paradise ({T} mana), Smuggler's Copter
// (crew), Bennie Bracks (a spell's convoke) and Survivors' Encampment
// ("{T}, Tap an untapped creature you control" on a mana ability).
// Without the gate each of them spent the destroyed commander while
// its owner was still deciding. The engine half is
// game/paused_exit_tap_cost_test.go; the moving costs are #1445's
// paused_exit_cost_test.go.

// toMainPhase walks to the precombat main before anything is
// destroyed: the CR 903.9 prompt blocks AdvanceStep, and each refusal
// here must come from the cost, not from the walk.
func toMainPhase(t *testing.T, g *game.Game) {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

// assertTapRefused checks the refusal paid nothing: the commander is
// untapped and still waiting on its owner, and nothing reached the
// stack or the pool.
func assertTapRefused(t *testing.T, g *game.Game, owner *game.Player, cmdr uuid.UUID, prompt *game.PendingChoice, err error) {
	t.Helper()
	if !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("naming the destroyed commander: err = %v, want ErrChoicePending", err)
	}
	if b20Tapped(t, g, cmdr) {
		t.Error("the refusal tapped the destroyed commander")
	}
	if len(g.StackMeta) != 0 || len(owner.ManaPool) != 0 {
		t.Errorf("stack %d, pool %d — the refusal paid something", len(g.StackMeta), len(owner.ManaPool))
	}
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].ID != prompt.ID {
		t.Fatalf("the refusal disturbed the destroy's prompt (%d pending)", len(g.PendingChoices))
	}
}

func TestBirdsOfParadiseCommanderCannotTapForManaWhileItsOwnerIsAsked(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainPhase(t, g)
	cmdr := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID:  uuid.New(),
		Name:        "Birds of Paradise",
		TypeLine:    "Creature — Bird",
		OracleID:    actBirdsOracle,
		Power:       0,
		Toughness:   1,
		Owner:       me.ID,
		Controller:  me.ID,
		IsCommander: true,
	})
	if err := g.DestroyPermanentForEffect(cmdr); err != nil {
		t.Fatalf("DestroyPermanentForEffect: %v", err)
	}
	if len(g.PendingChoices) != 1 {
		t.Fatalf("want the CR 903.9 prompt, have %d pending choices", len(g.PendingChoices))
	}
	prompt := g.PendingChoices[0]

	assertTapRefused(t, g, me, cmdr, prompt, g.ActivateManaAbility(me.ID, cmdr, 0, game.ManaAbilityParams{}))

	if err := g.ResolveOptionalReplacement(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	if !me.Command.Contains(cmdr) {
		t.Error("the commander did not reach the command zone")
	}
}

func TestSmugglersCopterCannotBeCrewedByADestroyedCommander(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainPhase(t, g)
	copter := pushVehicleForTest(g, me.ID, "Smuggler's Copter", smugglersCopterOracle, 3, 3)
	crewer := pushCrewerForTest(g, me.ID, "Crewer", 1)
	cmdr, prompt := destroyedCommander(t, g, me.ID)

	assertTapRefused(t, g, me, cmdr, prompt,
		g.ActivateCatalogAbility(me.ID, copter, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{cmdr}}))

	// An ordinary creature crews with the prompt still open.
	if err := g.ActivateCatalogAbility(me.ID, copter, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{crewer}}); err != nil {
		t.Fatalf("crewing with an ordinary creature while the prompt is open: %v", err)
	}
	if !b20Tapped(t, g, crewer) {
		t.Error("the ordinary crewer did not tap")
	}
}

func TestBennieBracksCannotConvokeWithADestroyedCommander(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainPhase(t, g)
	soldiers := pushTapCostSoldiers(g, me.ID, 4)
	cmdr, prompt := destroyedCommander(t, g, me.ID)

	bennie, err := castWithTapParams(t, g, "Bennie Bracks, Zoologist", "Legendary Creature — Elf Druid", "{3}{W}",
		b16BennieBracksOracle, game.CastSpellParams{TapIDs: append([]uuid.UUID{cmdr}, soldiers[:3]...)})
	assertTapRefused(t, g, me, cmdr, prompt, err)
	if !me.Hand.Contains(bennie) {
		t.Error("the refused cast left the hand")
	}
	for _, s := range soldiers {
		if tapCostTapped(g, s) {
			t.Error("the refused cast tapped a Soldier")
		}
	}

	// Four ordinary Soldiers convoke it with the prompt still open.
	if _, err := castWithTapParams(t, g, "Bennie Bracks, Zoologist", "Legendary Creature — Elf Druid", "{3}{W}",
		b16BennieBracksOracle, game.CastSpellParams{TapIDs: soldiers}); err != nil {
		t.Fatalf("convoking with ordinary creatures while the prompt is open: %v", err)
	}
}

func TestSurvivorsEncampmentCannotTapADestroyedCommander(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainPhase(t, g)
	land := seedPermanentWithOracle(g, me.ID, "Survivors' Encampment", "Land", b24SurvivorsEncampmentOracle)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	cmdr, prompt := destroyedCommander(t, g, me.ID)

	assertTapRefused(t, g, me, cmdr, prompt,
		g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{TapIDs: []uuid.UUID{cmdr}}))
	if b20Tapped(t, g, land) {
		t.Error("the refusal tapped the land")
	}

	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{TapIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("tapping an ordinary creature while the prompt is open: %v", err)
	}
	if !b20Tapped(t, g, land) || !b20Tapped(t, g, bear) {
		t.Error("the land and the ordinary creature both pay")
	}
}
