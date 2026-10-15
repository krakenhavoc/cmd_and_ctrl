package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// deck3655_c_test.go — The Falcon, Rejoin the Fight and Coin of Fate
// from the Terra, Herald of Hope deck request (#3655).

const (
	deck3655FalconOracle = "9f4093a6-83f1-4c1d-8b03-8fe15647fb11"
	deck3655RejoinOracle = "d22ea32b-2128-4dcc-8b1c-c4c52f0b2f45"
	deck3655CoinOracle   = "e43b478b-4cb8-42b1-a5b0-0b1421b84852"
)

func TestDeck3655FalconSacrificesToReturnAnyCreatureCard(t *testing.T) {
	for _, sacrifice := range []bool{true, false} {
		name := "decline"
		if sacrifice {
			name = "sacrifice"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			falcon := pushCatalogPermanent(g, me.ID, "The Falcon, Airship Restored", "Legendary Creature — Vehicle", deck3655FalconOracle, false)
			big := deck3655GraveyardCreature(me, "Big Dragon", 9)

			dealCombatDamageToPlayer(g, falcon, opp.ID, 4)
			passPriorityAroundTable(t, g)
			answerMayChoice(t, g, me.ID, sacrifice)
			if sacrifice {
				answerPickTarget(t, g, big)
			}
			passPriorityAroundTable(t, g)

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
	spec, _ := Lookup(deck3655FalconOracle)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := deck3655GraveyardCreature(me, "Alpha", 1)
	b := deck3655GraveyardCreature(me, "Beta", 1)
	c := deck3655GraveyardCreature(me, "Gamma", 1)
	d := deck3655GraveyardCreature(me, "Delta", 1)

	castCatalogSpell(t, g, "Rejoin the Fight", "Sorcery", deck3655RejoinOracle, nil)
	passPriorityAroundTable(t, g)

	picks := []uuid.UUID{a, b, c}
	for i, id := range picks {
		chooser := g.Seats[i+1]
		prompt := chooseCardsChoiceFor(g, chooser.ID)
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
		answerChooseCards(t, g, chooser.ID, id)
	}
	passPriorityAroundTable(t, g)

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

// coinOfFateActivation exiles two creature cards from `me`'s graveyard and
// resolves Coin of Fate's ability, leaving whatever prompt it queues open.
func coinOfFateActivation(t *testing.T, g *game.Game) (me *game.Player, one, two uuid.UUID) {
	t.Helper()
	me = g.Seats[0]
	coin := pushCatalogPermanent(g, me.ID, "Coin of Fate", "Artifact", deck3655CoinOracle, false)
	one = deck3655GraveyardCreature(me, "One", 1)
	two = deck3655GraveyardCreature(me, "Two", 1)
	for _, id := range []uuid.UUID{one, two} {
		if err := g.MoveCardByID(
			game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
			game.ZoneRef{Kind: game.ZoneExile}, id,
		); err != nil {
			t.Fatalf("exile: %v", err)
		}
	}
	spec, _ := Lookup(deck3655CoinOracle)
	item := &game.StackItem{Kind: game.StackItemActivated, Controller: me.ID, SourceCardID: coin,
		Paid: game.PaidCost{Exiled: []uuid.UUID{one, two}}}
	if err := spec.Activated[0].Effect(g, item); err != nil {
		t.Fatalf("effect: %v", err)
	}
	return me, one, two
}

// finishCoinOfFate checks the cards were split as `one` chosen: bottom of
// the library, `two` back tapped, and the crown taken.
func finishCoinOfFate(t *testing.T, g *game.Game, me *game.Player, libBefore int) {
	t.Helper()
	if len(me.Library.Cards) != libBefore+1 {
		t.Errorf("library grew by %d, want 1", len(me.Library.Cards)-libBefore)
	}
	var returnedTapped bool
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Two" {
			returnedTapped = c.Tapped
		}
	}
	if !returnedTapped {
		t.Error("the other exiled card did not return tapped")
	}
	if g.Monarch != me.ID {
		t.Error("the controller did not become the monarch")
	}
}

// "An opponent chooses": the controller names WHICH opponent, so the
// answer is not the next seat in turn order (#3846).
func TestDeck3655CoinOfFateControllerNamesTheChoosingOpponent(t *testing.T) {
	g := newCatalogGame(t)
	libBefore := len(g.Seats[0].Library.Cards)
	me, one, _ := coinOfFateActivation(t, g)
	pick := pendingOfKind(g, game.PendingChoiceOptionPick)
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
	answerChosenPlayer(t, g, me.ID, last.ID)
	if chooseCardsChoiceFor(g, g.Seats[1].ID) != nil {
		t.Error("the next opponent in turn order was asked, not the one named")
	}
	if chooseCardsChoiceFor(g, last.ID) == nil {
		t.Fatal("the named opponent was not asked to choose")
	}
	answerChooseCards(t, g, last.ID, one)
	finishCoinOfFate(t, g, me, libBefore)
}

// With a single opponent there is nobody to name: no prompt, and that
// opponent chooses.
func TestDeck3655CoinOfFateSkipsTheNamingPromptWithOneOpponent(t *testing.T) {
	g := newCatalogGame(t)
	libBefore := len(g.Seats[0].Library.Cards)
	g.WithWriteLock(func() {
		g.Seats[1].Eliminated = true
		g.Seats[2].Eliminated = true
	})
	last := g.Seats[3]
	me, one, _ := coinOfFateActivation(t, g)
	if pendingOfKind(g, game.PendingChoiceOptionPick) != nil {
		t.Fatal("the controller was asked to name an opponent with only one left")
	}
	if chooseCardsChoiceFor(g, last.ID) == nil {
		t.Fatal("the only opponent was not asked to choose")
	}
	answerChooseCards(t, g, last.ID, one)
	finishCoinOfFate(t, g, me, libBefore)
}
