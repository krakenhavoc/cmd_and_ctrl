package testscr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const doomWhispererOracle = "4a01db2e-cd43-4b1a-a480-169018f82501"

// TestDoomWhispererPaysLifeToSurveilTwo checks the flying/trample
// keywords and the pay-2-life activated ability.
func TestDoomWhispererPaysLifeToSurveilTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Bin Me", "Bin Me Too", "Keep Me")
	life := me.Life

	id := effectstest.PushCatalogPermanent(g, me.ID, "Doom Whisperer", "Creature — Nightmare Demon", doomWhispererOracle, false)

	if !effectstest.EffectiveAbilitiesContain(t, g, id, "flying") || !effectstest.EffectiveAbilitiesContain(t, g, id, "trample") {
		t.Error("Doom Whisperer should have flying and trample")
	}

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if me.Life != life-2 {
		t.Errorf("life after paying for surveil: %d, want %d", me.Life, life-2)
	}
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no surveil-2 prompt: %+v", g.PendingChoices)
	}
	if len(c.ScryCards) != 2 {
		t.Fatalf("surveil 2 should offer two cards: %d", len(c.ScryCards))
	}
}

// TestDoomWhispererCanActivateAgainImmediately — no tap and no
// once-per-turn gate, so a second activation with life to spare is
// legal, as printed.
func TestDoomWhispererCanActivateAgainImmediately(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Doom Whisperer", "Creature — Nightmare Demon", doomWhispererOracle, false)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("first ActivateCatalogAbility: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("second ActivateCatalogAbility should also be legal: %v", err)
	}
}
