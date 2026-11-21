package testsp

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// assertTapRefused checks the refusal paid nothing: the commander is
// untapped and still waiting on its owner, and nothing reached the
// stack or the pool.
func assertTapRefused(t *testing.T, g *game.Game, owner *game.Player, cmdr uuid.UUID, prompt *game.PendingChoice, err error) {
	t.Helper()
	if !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("naming the bounced commander: err = %v, want ErrChoicePending", err)
	}
	if effectstest.B20Tapped(t, g, cmdr) {
		t.Error("the refusal tapped the bounced commander")
	}
	if len(g.StackMeta) != 0 || len(owner.ManaPool) != 0 {
		t.Errorf("stack %d, pool %d — the refusal paid something", len(g.StackMeta), len(owner.ManaPool))
	}
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].ID != prompt.ID {
		t.Fatalf("the refusal disturbed the bounce's prompt (%d pending)", len(g.PendingChoices))
	}
}

func TestBirdsOfParadiseCommanderCannotTapForManaWhileItsOwnerIsAsked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMainPhase(t, g)
	cmdr := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID:  uuid.New(),
		Name:        "Birds of Paradise",
		TypeLine:    "Creature — Bird",
		OracleID:    effectstest.ActBirdsOracle,
		Power:       0,
		Toughness:   1,
		Owner:       me.ID,
		Controller:  me.ID,
		IsCommander: true,
	})
	if err := g.BounceToHandForEffect(cmdr); err != nil {
		t.Fatalf("BounceToHandForEffect: %v", err)
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

func TestSmugglersCopterCannotBeCrewedByABouncedCommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMainPhase(t, g)
	copter := effectstest.PushVehicleForTest(g, me.ID, "Smuggler's Copter", effectstest.SmugglersCopterOracle, 3, 3)
	crewer := effectstest.PushCrewerForTest(g, me.ID, "Crewer", 1)
	cmdr, prompt := bouncedCommander(t, g, me.ID)

	assertTapRefused(t, g, me, cmdr, prompt,
		g.ActivateCatalogAbility(me.ID, copter, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{cmdr}}))

	// An ordinary creature crews with the prompt still open.
	if err := g.ActivateCatalogAbility(me.ID, copter, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{crewer}}); err != nil {
		t.Fatalf("crewing with an ordinary creature while the prompt is open: %v", err)
	}
	if !effectstest.B20Tapped(t, g, crewer) {
		t.Error("the ordinary crewer did not tap")
	}
}

func TestBennieBracksCannotConvokeWithABouncedCommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMainPhase(t, g)
	soldiers := effectstest.PushTapCostSoldiers(g, me.ID, 4)
	cmdr, prompt := bouncedCommander(t, g, me.ID)

	bennie, err := effectstest.CastWithTapParams(t, g, "Bennie Bracks, Zoologist", "Legendary Creature — Elf Druid", "{3}{W}",
		effectstest.B16BennieBracksOracle, game.CastSpellParams{TapIDs: append([]uuid.UUID{cmdr}, soldiers[:3]...)})
	assertTapRefused(t, g, me, cmdr, prompt, err)
	if !me.Hand.Contains(bennie) {
		t.Error("the refused cast left the hand")
	}
	for _, s := range soldiers {
		if effectstest.TapCostTapped(g, s) {
			t.Error("the refused cast tapped a Soldier")
		}
	}

	// Four ordinary Soldiers convoke it with the prompt still open.
	if _, err := effectstest.CastWithTapParams(t, g, "Bennie Bracks, Zoologist", "Legendary Creature — Elf Druid", "{3}{W}",
		effectstest.B16BennieBracksOracle, game.CastSpellParams{TapIDs: soldiers}); err != nil {
		t.Fatalf("convoking with ordinary creatures while the prompt is open: %v", err)
	}
}

func TestSurvivorsEncampmentCannotTapABouncedCommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMainPhase(t, g)
	land := effectstest.SeedPermanentWithOracle(g, me.ID, "Survivors' Encampment", "Land", effectstest.B24SurvivorsEncampmentOracle)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	cmdr, prompt := bouncedCommander(t, g, me.ID)

	assertTapRefused(t, g, me, cmdr, prompt,
		g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{TapIDs: []uuid.UUID{cmdr}}))
	if effectstest.B20Tapped(t, g, land) {
		t.Error("the refusal tapped the land")
	}

	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{TapIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("tapping an ordinary creature while the prompt is open: %v", err)
	}
	if !effectstest.B20Tapped(t, g, land) || !effectstest.B20Tapped(t, g, bear) {
		t.Error("the land and the ordinary creature both pay")
	}
}
