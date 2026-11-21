package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestDeck3655FalconSacrificesToReturnAnyCreatureCard(t *testing.T) {
	for _, sacrifice := range []bool{true, false} {
		name := "decline"
		if sacrifice {
			name = "sacrifice"
		}
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			falcon := effectstest.PushCatalogPermanent(g, me.ID, "The Falcon, Airship Restored", "Legendary Creature — Vehicle", effectstest.Deck3655FalconOracle, false)
			big := effectstest.Deck3655GraveyardCreature(me, "Big Dragon", 9)

			effectstest.DealCombatDamageToPlayer(g, falcon, opp.ID, 4)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerMayChoice(t, g, me.ID, sacrifice)
			if sacrifice {
				effectstest.AnswerPickTarget(t, g, big)
			}
			effectstest.PassPriorityAroundTable(t, g)

			if !sacrifice {
				if !g.Battlefield.Contains(falcon) || g.Battlefield.Contains(big) {
					t.Error("declining changed the board")
				}
				return
			}
			if g.Battlefield.Contains(falcon) {
				t.Error("The Falcon was not sacrificed")
			}
			if !g.Battlefield.Contains(big) {
				t.Error("the creature card did not return")
			}
		})
	}
}

func TestDeck3655FalconShape(t *testing.T) {
	spec, _ := effects.Lookup(effectstest.Deck3655FalconOracle)
	var graveyard, crew bool
	for _, a := range spec.Activated {
		if len(a.Zones) == 1 && a.Zones[0] == game.ZoneGraveyard {
			graveyard = true
		}
		if a.Cost.Crew == 2 {
			crew = true
		}
	}
	if !graveyard || !crew {
		t.Errorf("graveyard ability %t, crew 2 %t, want both", graveyard, crew)
	}
}

func TestDeck3655RejoinTheFightLetsEachOpponentChooseOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.Deck3655GraveyardCreature(me, "Alpha", 1)
	b := effectstest.Deck3655GraveyardCreature(me, "Beta", 1)
	c := effectstest.Deck3655GraveyardCreature(me, "Gamma", 1)
	d := effectstest.Deck3655GraveyardCreature(me, "Delta", 1)

	effectstest.CastCatalogSpell(t, g, "Rejoin the Fight", "Sorcery", effectstest.Deck3655RejoinOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	picks := []uuid.UUID{a, b, c}
	for i, id := range picks {
		chooser := g.Seats[i+1]
		prompt := effectstest.ChooseCardsChoiceFor(g, chooser.ID)
		if prompt == nil {
			t.Fatalf("opponent %d was not asked", i+1)
		}
		for _, offered := range prompt.ChooseCards {
			for _, taken := range picks[:i] {
				if offered == taken {
					t.Errorf("opponent %d was offered a card already chosen", i+1)
				}
			}
		}
		effectstest.AnswerChooseCards(t, g, chooser.ID, id)
	}
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range picks {
		if !g.Battlefield.Contains(id) {
			t.Errorf("chosen card %s did not return", id)
		}
	}
	if g.Battlefield.Contains(d) {
		t.Error("an unchosen creature card returned")
	}
	for _, card := range g.Battlefield.Cards {
		if (card.InstanceID == a || card.InstanceID == b || card.InstanceID == c) && card.Controller != me.ID {
			t.Errorf("%s returned under %s, not the caster", card.Name, card.Controller)
		}
	}
}

// "An opponent chooses": the controller names WHICH opponent, so the
// answer is not the next seat in turn order (#3846).
func TestDeck3655CoinOfFateControllerNamesTheChoosingOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	libBefore := len(g.Seats[0].Library.Cards)
	me, one, _ := effectstest.CoinOfFateActivation(t, g)
	pick := effectstest.PendingOfKind(g, game.PendingChoiceOptionPick)
	if pick == nil {
		t.Fatal("the controller was not asked which opponent chooses")
	}
	if pick.Chooser != me.ID {
		t.Errorf("the prompt is for %v, want the controller", pick.Chooser)
	}
	if len(pick.PickOptions) != 3 {
		t.Errorf("offered %d seats, want the three opponents", len(pick.PickOptions))
	}
	for _, opt := range pick.PickOptions {
		if opt.Player == me.ID {
			t.Error("the controller offered themself as the opponent")
		}
	}
	// Name the LAST seat: neither seat 1 (next in turn order) nor anyone
	// else may be asked to choose.
	last := g.Seats[3]
	effectstest.AnswerChosenPlayer(t, g, me.ID, last.ID)
	if effectstest.ChooseCardsChoiceFor(g, g.Seats[1].ID) != nil {
		t.Error("the next opponent in turn order was asked, not the one named")
	}
	if effectstest.ChooseCardsChoiceFor(g, last.ID) == nil {
		t.Fatal("the named opponent was not asked to choose")
	}
	effectstest.AnswerChooseCards(t, g, last.ID, one)
	effectstest.FinishCoinOfFate(t, g, me, libBefore)
}

// With a single opponent there is nobody to name: no prompt, and that
// opponent chooses.
func TestDeck3655CoinOfFateSkipsTheNamingPromptWithOneOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	libBefore := len(g.Seats[0].Library.Cards)
	g.WithWriteLock(func() {
		g.Seats[1].Eliminated = true
		g.Seats[2].Eliminated = true
	})
	last := g.Seats[3]
	me, one, _ := effectstest.CoinOfFateActivation(t, g)
	if effectstest.PendingOfKind(g, game.PendingChoiceOptionPick) != nil {
		t.Fatal("the controller was asked to name an opponent with only one left")
	}
	if effectstest.ChooseCardsChoiceFor(g, last.ID) == nil {
		t.Fatal("the only opponent was not asked to choose")
	}
	effectstest.AnswerChooseCards(t, g, last.ID, one)
	effectstest.FinishCoinOfFate(t, g, me, libBefore)
}
