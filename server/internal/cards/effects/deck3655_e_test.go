package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// deck3655_e_test.go — Sepulchral Primordial and Espers to Magicite
// from the Terra, Herald of Hope deck request (#3655).

const (
	deck3655SepulchralOracle = "480454a0-fb10-432b-a963-8b0c0ffc6680"
	deck3655EspersOracle     = "06f7ec6d-d70e-44d8-bf43-ae667d244ab4"
)

func TestDeck3655SepulchralTakesOneCreatureCardPerOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, oppA, oppB := g.Seats[0], g.Seats[1], g.Seats[2]
	prim := pushCatalogPermanent(g, me.ID, "Sepulchral Primordial", "Creature — Avatar", deck3655SepulchralOracle, false)
	a1 := deck3655GraveyardCreature(oppA, "A One", 1)
	a2 := deck3655GraveyardCreature(oppA, "A Two", 1)
	b1 := deck3655GraveyardCreature(oppB, "B One", 1)
	mine := deck3655GraveyardCreature(me, "Mine", 1)

	spec, _ := Lookup(deck3655SepulchralOracle)
	clause := spec.Triggered[0].TargetsFrom(game.TriggerContext{}, &game.Card{InstanceID: prim, Controller: me.ID}, g)
	if clause == nil {
		t.Fatal("no clause for the entering Primordial")
	}
	var their, own game.Card
	g.ReadSnapshot(func() {
		their, _ = g.LookupCardForEffect(a1)
		own, _ = g.LookupCardForEffect(mine)
	})
	if !clause.CardOK(g, me.ID, their, game.ZoneGraveyard) {
		t.Error("an opponent's creature card must be a legal pick for the first clause")
	}
	if clause.CardOK(g, me.ID, own, game.ZoneGraveyard) {
		t.Error("my own graveyard must not be offered")
	}

	// Resolution takes the still-legal picks under my control.
	item := &game.StackItem{Kind: game.StackItemTriggered, Controller: me.ID, SourceCardID: prim,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a2}, {Kind: game.TargetCard, ID: b1, Slot: 1}}}
	if err := spec.Triggered[0].Effect(g, item); err != nil {
		t.Fatalf("effect: %v", err)
	}
	taken := 0
	for _, c := range g.Battlefield.Cards {
		if (c.Name == "A Two" || c.Name == "B One") && c.Controller == me.ID {
			taken++
		}
	}
	if taken != 2 {
		t.Errorf("took %d creature cards under my control, want 2", taken)
	}
	if !oppA.Graveyard.Contains(a1) {
		t.Error("the unpicked card left its graveyard")
	}
}

func TestDeck3655EspersExilesOpponentGraveyardsAndCopiesAsAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := deck3655GraveyardCreature(opp, "Elder Dragon", 5)
	spell := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Old Spell", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	mine := deck3655GraveyardCreature(me, "My Dead", 1)

	castCatalogSpell(t, g, "Espers to Magicite", "Instant", deck3655EspersOracle, nil)
	passPriorityAroundTable(t, g)
	prompt := chooseCardsChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no choice for the creature card exiled this way")
	}
	if len(prompt.ChooseCards) != 1 || prompt.ChooseCards[0] != dragon {
		t.Errorf("offered %v, want only the opposing creature card", prompt.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, dragon)
	passPriorityAroundTable(t, g)

	if len(opp.Graveyard.Cards) != 0 {
		t.Errorf("the opponent's graveyard still holds %d cards", len(opp.Graveyard.Cards))
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("my own graveyard was exiled")
	}
	var found bool
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elder Dragon" && c.Controller == me.ID && c.IsToken() {
			found = true
			if !c.IsArtifact() || c.IsCreature() {
				t.Errorf("the copy is artifact=%t creature=%t, want an artifact and not a creature", c.IsArtifact(), c.IsCreature())
			}
		}
	}
	if !found {
		t.Error("no token copy of the exiled creature card")
	}
}
