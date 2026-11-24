package legal_test

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// commander_color_test.go — CR 903.4b (#5948): the pregame "choose your
// commander's colour" prompt is an ordinary choose_color prompt, so a bot
// seat is offered one move per colour and its answer goes through.
func TestThePregameCommanderColourPromptIsAnswerableByASeat(t *testing.T) {
	g := game.NewGame()
	piper := game.NewCard("The Prismatic Piper", uuid.Nil)
	piper.TypeLine = "Legendary Creature — Shapeshifter"
	piper.ManaCost = "{5}"
	piper.OracleID = "b2aeee90-7b8d-47a6-9167-cc21f7fbfd43"
	piper.IsCommander = true
	deck := []game.Card{piper}
	for range 20 {
		deck = append(deck, game.NewCard("Filler", uuid.Nil))
	}
	if _, err := g.AddPlayer("Bot", deck); err != nil {
		t.Fatal(err)
	}
	other := make([]game.Card, 20)
	for i := range other {
		other[i] = game.NewCard("Filler", uuid.Nil)
	}
	if _, err := g.AddPlayer("Other", other); err != nil {
		t.Fatal(err)
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatal(err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatal(err)
		}
	}
	bot := g.Seats[0].ID
	var colorMoves []legal.Move
	for _, m := range legal.EnumerateFor(g, bot) {
		if m.Kind == legal.KindChoice {
			colorMoves = append(colorMoves, m)
		}
	}
	if len(colorMoves) != 5 {
		t.Fatalf("choice moves = %d (%+v), want one per colour", len(colorMoves), colorMoves)
	}
	for _, m := range legal.EnumerateFor(g, g.Seats[1].ID) {
		if m.Kind == legal.KindChoice {
			t.Fatalf("the other seat was offered %q", m.Label)
		}
	}
	m := colorMoves[0]
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(m.Type), Player: m.Player, Caller: m.Player, Params: m.Params}); err != nil {
		t.Fatalf("dispatch %q: %v", m.Label, err)
	}
	var chosen string
	g.ReadSnapshot(func() { chosen = g.Seats[0].Command.Cards[0].CommanderColor })
	if chosen == "" {
		t.Fatal("the seat's answer did not record a colour")
	}
}
