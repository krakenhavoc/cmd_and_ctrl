package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// deck3655_b_test.go — the second slice of the Terra, Herald of Hope
// deck request (#3655).

const (
	deck3655CyanOracle   = "5a2fa420-c27f-4c14-b285-e0bb2828e053"
	deck3655GauOracle    = "9e96d12b-8381-4706-8089-3521aa2b5422"
	deck3655ShadowOracle = "13ca49a2-8a78-4be4-976a-b684086864bb"
)

func TestDeck3655CyanCostsLessPerCreatureCardInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	q := game.CostQuery{Game: g, Controller: me.ID, Card: game.Card{InstanceID: uuid.New()}}
	if got := creatureCardsInYourGraveyard(q); got != 0 {
		t.Fatalf("empty graveyard reduces by %d", got)
	}
	deck3655GraveyardCreature(me, "One", 1)
	deck3655GraveyardCreature(me, "Two", 1)
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	if got := creatureCardsInYourGraveyard(q); got != 2 {
		t.Errorf("reduction = %d, want 2 (the sorcery does not count)", got)
	}
	spec, _ := Lookup(deck3655CyanOracle)
	if len(spec.SelfCostModifiers) != 1 {
		t.Errorf("self cost modifiers = %d, want 1", len(spec.SelfCostModifiers))
	}
}

func TestDeck3655CyanGrowsWhenCreatureCardsLeaveYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cyan := pushCatalogPermanent(g, me.ID, "Cyan, Vengeful Samurai", "Legendary Creature — Human Samurai", deck3655CyanOracle, false)
	one := deck3655GraveyardCreature(me, "One", 1)
	two := deck3655GraveyardCreature(me, "Two", 1)

	for _, id := range []uuid.UUID{one, two} {
		if err := g.MoveCardByID(
			game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
			game.ZoneRef{Kind: game.ZoneBattlefield}, id,
		); err != nil {
			t.Fatalf("move: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if got := battlefieldCardCopy(t, g, cyan).Counters[game.CounterPlusOne]; got < 1 {
		t.Errorf("Cyan has %d counters after creature cards left the graveyard, want at least 1", got)
	} else {
		t.Logf("two cards moved one after the other made %d counter(s)", got)
	}
}

func TestDeck3655GauAttackCounterAndEndStepDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gau := pushCatalogPermanent(g, me.ID, "Gau, Feral Youth", "Legendary Creature — Human Berserker", deck3655GauOracle, false)
	g.WithWriteLock(func() {
		g.Battlefield.Cards[len(g.Battlefield.Cards)-1].Power = 2
	})

	// No card left the graveyard: nothing at the end step.
	before := opp.Life
	advanceToStepOf(t, g, 0, game.StepEnd)
	passPriorityAroundTable(t, g)
	if opp.Life != before {
		t.Fatalf("opponent lost %d with no card leaving the graveyard", before-opp.Life)
	}

	// Next turn: a card leaves the graveyard, then the end step damages.
	advanceToStepOf(t, g, 1, game.StepPrecombatMain)
	advanceToStepOf(t, g, 0, game.StepPrecombatMain)
	back := deck3655GraveyardCreature(me, "Dead", 1)
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, back,
	); err != nil {
		t.Fatalf("move: %v", err)
	}
	before = opp.Life
	advanceToStepOf(t, g, 0, game.StepEnd)
	if triggerOnStack(g, gau) == nil {
		t.Fatal("no end-step trigger after a card left the graveyard")
	}
	passPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 2 {
		t.Errorf("opponent lost %d, want Gau's power 2", got)
	}
}

func TestDeck3655ShadowSacrificesForCardsAndDrain(t *testing.T) {
	for _, accept := range []bool{true, false} {
		name := "decline"
		if accept {
			name = "sacrifice"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			shadow := pushCatalogPermanent(g, me.ID, "Shadow, Mysterious Assassin", "Legendary Creature — Human Assassin", deck3655ShadowOracle, false)
			fodder := pushCostedPermanentForTest(g, me.ID, "Fodder", "Artifact", "{3}")
			land := pushCostedPermanentForTest(g, me.ID, "Plains", "Basic Land — Plains", "")
			hand, life := me.Hand.Size(), opp.Life

			dealCombatDamageToPlayer(g, shadow, opp.ID, 3)
			passPriorityAroundTable(t, g)
			pick := chooseCardsChoiceFor(g, me.ID)
			if pick == nil {
				t.Fatal("no sacrifice prompt")
			}
			for _, id := range pick.ChooseCards {
				if id == land || id == shadow {
					t.Errorf("offered %s, which is a land or Shadow itself", id)
				}
			}
			if accept {
				answerChooseCards(t, g, me.ID, fodder)
			} else {
				answerChooseCards(t, g, me.ID)
			}
			passPriorityAroundTable(t, g)

			if !accept {
				if !g.Battlefield.Contains(fodder) || me.Hand.Size() != hand || opp.Life != life {
					t.Error("declining changed the board, the hand or the life")
				}
				return
			}
			if g.Battlefield.Contains(fodder) {
				t.Error("the artifact was not sacrificed")
			}
			if got := me.Hand.Size() - hand; got != 2 {
				t.Errorf("drew %d, want 2", got)
			}
			if got := life - opp.Life; got != 3 {
				t.Errorf("opponent lost %d, want the artifact's mana value 3", got)
			}
		})
	}
}
