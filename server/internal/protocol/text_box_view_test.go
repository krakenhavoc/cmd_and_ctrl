package protocol

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestCardViewOfAnExchangedTextBox pins ADR 0160 on the wire: each
// permanent names the object whose text box it has, its keyword chips
// are that text's, and its name and P/T are its own.
func TestCardViewOfAnExchangedTextBox(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	flier := game.Card{
		InstanceID: uuid.New(), Name: "Fixture Flier", TypeLine: "Creature — Bird",
		Keywords: []string{"flying"}, Power: 1, Toughness: 1, Owner: owner, Controller: owner,
	}
	bear := game.Card{
		InstanceID: uuid.New(), Name: "Fixture Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	}
	g.Battlefield.PushTop(flier)
	g.Battlefield.PushTop(bear)
	ok := false
	g.WithWriteLock(func() {
		ok = g.ExchangeTextBoxesForEffect(uuid.Nil, flier.InstanceID, bear.InstanceID, game.IndefiniteDuration(), "test")
	})
	if !ok {
		t.Fatal("the exchange did not happen")
	}
	f, b := cardViewByID(t, g, flier.InstanceID), cardViewByID(t, g, bear.InstanceID)
	if f.TextBoxFrom != "Fixture Bear" || b.TextBoxFrom != "Fixture Flier" {
		t.Errorf("text_box_from = %q / %q", f.TextBoxFrom, b.TextBoxFrom)
	}
	if slices.Contains(f.Abilities, "flying") || !slices.Contains(b.Abilities, "flying") {
		t.Errorf("abilities: flier %v, bear %v; flying belongs to the bear now", f.Abilities, b.Abilities)
	}
	if f.Name != "Fixture Flier" || b.Power != 2 {
		t.Errorf("name and P/T stay: %q, %d", f.Name, b.Power)
	}
}

// TestCopyTargetPromptSaysWhenItExchangesText: the Deadpool prompt is a
// copy_target with exchanges_text set.
func TestCopyTargetPromptSaysWhenItExchangesText(t *testing.T) {
	g := buildActiveGame(t)
	chooser := g.Seats[0].ID
	g.WithWriteLock(func() {
		g.QueueChoiceForEffect(game.PendingChoice{
			Kind: game.PendingChoiceCopyTarget, Chooser: chooser, Count: 1, ExchangesText: true,
		})
	})
	v := ViewOfGame(g)
	if len(v.PendingChoices) != 1 || !v.PendingChoices[0].ExchangesText {
		t.Fatalf("pending choices %+v, want one with exchanges_text", v.PendingChoices)
	}
}
