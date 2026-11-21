package testscr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// deck3655_a_test.go — the first slice of the Terra, Herald of Hope
// deck request (#3655): Terra herself.

const deck3655TerraOracle = "25d40d28-1b46-4055-959a-45719040ed30"

func TestDeck3655TerraMillsTwoAndFliesAtYourCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	terra := effectstest.PushCatalogPermanent(g, me.ID, "Terra, Herald of Hope", "Legendary Creature — Human Wizard Warrior", deck3655TerraOracle, false)
	lib, gy := len(me.Library.Cards), len(me.Graveyard.Cards)

	effectstest.AdvanceToStepOf(t, g, 0, game.StepBeginCombat)
	if effectstest.TriggerOnStack(g, terra) == nil {
		t.Fatal("no Trance trigger at the beginning of combat")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := lib - len(me.Library.Cards); got != 2 {
		t.Errorf("milled %d cards, want 2", got)
	}
	if got := len(me.Graveyard.Cards) - gy; got != 2 {
		t.Errorf("graveyard grew by %d, want 2", got)
	}
	var flies bool
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(terra)
		flies = game.HasKeyword(&c, "flying")
	})
	if !flies {
		t.Error("Terra did not gain flying")
	}
}

func TestDeck3655TerraPaysTwoToReturnASmallCreatureTapped(t *testing.T) {
	for _, pay := range []bool{true, false} {
		name := "decline"
		if pay {
			name = "pay"
		}
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			terra := effectstest.PushCatalogPermanent(g, me.ID, "Terra, Herald of Hope", "Legendary Creature — Human Wizard Warrior", deck3655TerraOracle, false)
			small := effectstest.Deck3655GraveyardCreature(me, "Small Bear", 3)
			big := effectstest.Deck3655GraveyardCreature(me, "Big Bear", 4)

			effectstest.DealCombatDamageToPlayer(g, terra, opp.ID, 3)
			effectstest.PassPriorityAroundTable(t, g)
			if pay {
				effectstest.B06AddMana(me, "R", "R")
			}
			effectstest.AnswerPayUnless(t, g, me.ID, pay)
			if pay {
				// Only the power-3 card is offered; the power-4 one is not.
				pick := effectstest.PickTargetPrompt(t, g)
				if len(pick.PickTargetCards) != 1 || pick.PickTargetCards[0] != small {
					t.Fatalf("offered targets %v, want only the power-3 creature", pick.PickTargetCards)
				}
				effectstest.AnswerPickTarget(t, g, small)
			}
			effectstest.PassPriorityAroundTable(t, g)

			if !pay {
				if g.Battlefield.Contains(small) || !me.Graveyard.Contains(small) {
					t.Fatal("declining the payment still returned a creature")
				}
				return
			}
			if !g.Battlefield.Contains(small) {
				t.Fatal("the power-3 creature did not return")
			}
			if !me.Graveyard.Contains(big) || g.Battlefield.Contains(big) {
				t.Error("the power-4 creature must stay in the graveyard")
			}
			var tapped bool
			g.ReadSnapshot(func() {
				c, _ := g.LookupCardForEffect(small)
				tapped = c.Tapped
			})
			if !tapped {
				t.Error("the creature returned untapped")
			}
		})
	}
}
