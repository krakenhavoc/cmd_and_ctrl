package testsl

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lucy_maclean_positively_armed_test.go — #6174: "Whenever a token enters,
// you may have target player other than its controller create a token
// that's a copy of it, then you draw a card if an opponent created a token
// this way. Do this only once each turn."

const oracleLucyMacLean = "1196a4cb-544c-4e7b-91cd-f0820b21a80d"

func lucyTable(t *testing.T) *game.Game {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, g.Seats[0].ID, "Lucy MacLean, Positively Armed", "Legendary Creature — Human Survivor", oracleLucyMacLean, false)
	return g
}

// lucyMakeToken has `owner` create one 2/2 Construct token.
func lucyMakeToken(t *testing.T, g *game.Game, owner uuid.UUID) uuid.UUID {
	t.Helper()
	var made []uuid.UUID
	g.WithWriteLock(func() {
		var err error
		made, err = g.CreateTokensForEffect(owner, game.Card{Name: "Construct", TypeLine: "Token Artifact Creature — Construct",
			Power: 2, Toughness: 2, PrintedPTKnown: true}, 1, game.TokenEntryOptions{})
		if err != nil {
			t.Fatalf("CreateTokensForEffect: %v", err)
		}
	})
	if len(made) != 1 {
		t.Fatalf("made %d tokens", len(made))
	}
	return made[0]
}

func lucyConstructsOf(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsToken() && c.Name == "Construct" && c.Controller == controller {
			n++
		}
	}
	return n
}

// TestLucyMacLeanHasAnotherOpponentCopyTheTokenOnceEachTurn — #6174's
// acceptance: a token entering under player B lets Lucy's controller have
// player C create a copy (C owns and controls it, CR 111.2), Lucy's
// controller draws because C is an opponent, and a second token that turn
// offers nothing.
func TestLucyMacLeanHasAnotherOpponentCopyTheTokenOnceEachTurn(t *testing.T) {
	g := lucyTable(t)
	me, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	hand := me.Hand.Size()

	lucyMakeToken(t, g, b.ID)
	effectstest.PickTriggerTargetPlayer(t, g, me.ID, c.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := lucyConstructsOf(g, c.ID); got != 1 {
		t.Fatalf("player C controls %d copies, want 1", got)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("Lucy's controller drew %d, want 1 (an opponent created the token)", got)
	}

	hand = me.Hand.Size()
	lucyMakeToken(t, g, b.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.OwedMays(g, me.ID) > 0 {
		t.Fatal("a second token this turn offered another copy")
	}
	if got := lucyConstructsOf(g, c.ID); got != 1 {
		t.Errorf("player C controls %d copies after the second token, want 1", got)
	}
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d more this turn", got)
	}
}

// TestLucyMacLeanTargetIsNotTheTokensController — "target player other
// than its controller".
func TestLucyMacLeanTargetIsNotTheTokensController(t *testing.T) {
	g := lucyTable(t)
	me, b := g.Seats[0], g.Seats[1]
	lucyMakeToken(t, g, b.ID)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the token entering asked for no target")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: b.ID}); err == nil {
		t.Fatal("the token's controller was a legal target")
	}
}

// TestLucyMacLeanNoDrawWhenYouCreateTheCopy — you are not an opponent, so
// having yourself create the copy (of an opponent's token) draws nothing.
func TestLucyMacLeanNoDrawWhenYouCreateTheCopy(t *testing.T) {
	g := lucyTable(t)
	me, b := g.Seats[0], g.Seats[1]
	hand := me.Hand.Size()
	lucyMakeToken(t, g, b.ID)
	effectstest.PickTriggerTargetPlayer(t, g, me.ID, me.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := lucyConstructsOf(g, me.ID); got != 1 {
		t.Fatalf("you control %d copies, want 1", got)
	}
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d, want 0", got)
	}
}

// TestLucyMacLeanDecliningKeepsTheTurn — a "no" is not doing it, so the
// next token that turn offers again; your own token entering works too.
func TestLucyMacLeanDecliningKeepsTheTurn(t *testing.T) {
	g := lucyTable(t)
	me, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	lucyMakeToken(t, g, b.ID)
	effectstest.PickTriggerTargetPlayer(t, g, me.ID, c.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := lucyConstructsOf(g, c.ID); got != 0 {
		t.Fatalf("a declined offer made %d copies", got)
	}

	hand := me.Hand.Size()
	lucyMakeToken(t, g, me.ID)
	effectstest.PickTriggerTargetPlayer(t, g, me.ID, c.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := lucyConstructsOf(g, c.ID); got != 1 {
		t.Fatalf("after a declined offer, player C controls %d copies of your token, want 1", got)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}
