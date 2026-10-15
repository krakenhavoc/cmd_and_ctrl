package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// deck3655_a_test.go — the first slice of the Terra, Herald of Hope
// deck request (#3655): Terra herself.

const deck3655TerraOracle = "25d40d28-1b46-4055-959a-45719040ed30"

func deck3655GraveyardCreature(p *game.Player, name string, power int) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Bear",
		Power: power, Toughness: 2, Owner: p.ID, Controller: p.ID,
	})
	return id
}

func TestDeck3655TerraMillsTwoAndFliesAtYourCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	terra := pushCatalogPermanent(g, me.ID, "Terra, Herald of Hope", "Legendary Creature — Human Wizard Warrior", deck3655TerraOracle, false)
	lib, gy := len(me.Library.Cards), len(me.Graveyard.Cards)

	advanceToStepOf(t, g, 0, game.StepBeginCombat)
	if triggerOnStack(g, terra) == nil {
		t.Fatal("no Trance trigger at the beginning of combat")
	}
	passPriorityAroundTable(t, g)
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
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			terra := pushCatalogPermanent(g, me.ID, "Terra, Herald of Hope", "Legendary Creature — Human Wizard Warrior", deck3655TerraOracle, false)
			small := deck3655GraveyardCreature(me, "Small Bear", 3)
			big := deck3655GraveyardCreature(me, "Big Bear", 4)

			dealCombatDamageToPlayer(g, terra, opp.ID, 3)
			passPriorityAroundTable(t, g)
			if pay {
				b06AddMana(me, "R", "R")
			}
			answerPayUnless(t, g, me.ID, pay)
			if pay {
				// Only the power-3 card is offered; the power-4 one is not.
				pick := pickTargetPrompt(t, g)
				if len(pick.PickTargetCards) != 1 || pick.PickTargetCards[0] != small {
					t.Fatalf("offered targets %v, want only the power-3 creature", pick.PickTargetCards)
				}
				answerPickTarget(t, g, small)
			}
			passPriorityAroundTable(t, g)

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
