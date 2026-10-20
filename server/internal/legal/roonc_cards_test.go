package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// roonc_cards_test.go — #4128 slice c, the enumerator's half: the moves a
// bot is offered for Lonis, Cryptozoologist (a variable-sacrifice
// activation with a player target, then a put out of an OPPONENT'S
// library) and Surprise Deployment (a spell that may only be cast in
// combat). Every enumerated move must be accepted by the dispatcher.

const (
	rooncLonisOracle  = "0ee06ed0-717a-4bfa-b634-d00e013c1f16"
	rooncDeployOracle = "61a1748c-e772-4338-9edd-720cd5d2fc12"
)

func rooncMovesFor(moves []legal.Move, typ string, source uuid.UUID) []legal.Move {
	var out []legal.Move
	for _, m := range moves {
		if m.Type == typ && m.Source == source {
			out = append(out, m)
		}
	}
	return out
}

// Lonis, with Clues on the board, is offered {T}, Sacrifice X Clues for a
// positive X aimed at an opponent, and the answer to its put prompt (the
// nonland card of mana value X or less, out of the opponent's library) is
// an enumerated move that the engine accepts.
func TestRooncLonisEnumeratesTheActivationAndThePutAnswer(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(me)
	lonis := battlefieldCard(g, me, game.Card{Name: "Lonis, Cryptozoologist",
		TypeLine: "Legendary Creature — Snake Elf Scout", OracleID: rooncLonisOracle, Power: 1, Toughness: 2})
	for i := 0; i < 2; i++ {
		battlefieldCard(g, me, game.Card{Name: "Clue", TypeLine: "Token Artifact — Clue"})
	}
	// The opponent's top two cards: a two-drop creature over a land.
	var bear uuid.UUID
	g.WithWriteLock(func() {
		land := game.Card{InstanceID: uuid.New(), Name: "Mountain", TypeLine: "Basic Land — Mountain", Owner: opp.ID, Controller: opp.ID}
		b := game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
			Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID}
		bear = b.InstanceID
		opp.Library.PushTop(land)
		opp.Library.PushTop(b)
	})
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, me.ID)
	acts := rooncMovesFor(moves, legal.TypeActivateAbility, lonis)
	if len(acts) == 0 {
		t.Fatalf("Lonis's activation is not offered with two Clues out: %v", labels(moves))
	}
	dispatchAll(t, g, me.ID, acts)

	// Activate for X=2 through the engine, resolve it, and read the
	// put prompt's answers off the enumerator.
	clues := []uuid.UUID{}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Clue" {
			clues = append(clues, c.InstanceID)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, lonis, 0, game.ActivateAbilityParams{
		XValue:       2,
		SacrificeIDs: clues,
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for i := 0; i < 16 && len(g.PendingChoices) == 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	moves = legal.EnumerateFor(g, me.ID)
	var answers, picksBear int
	for _, m := range moves {
		if m.Type != legal.TypeResolveChoice {
			continue
		}
		answers++
		if strings.Contains(string(m.Params), bear.String()) {
			picksBear++
		}
	}
	if answers == 0 {
		t.Fatalf("the put prompt has no enumerated answers: %v", labels(moves))
	}
	if picksBear == 0 {
		t.Errorf("taking the Bear out of the opponent's library is not an offered answer: %v", labels(moves))
	}
	dispatchAll(t, g, me.ID, moves)
}

// "Cast this spell only during combat": not offered in a main phase,
// offered at the beginning of combat, and the offer is accepted.
func TestRooncSurpriseDeploymentIsOfferedOnlyDuringCombat(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	clearHand(me)
	spell := handCard(me, game.Card{Name: "Surprise Deployment", TypeLine: "Instant",
		ManaCost: "{3}{W}", OracleID: rooncDeployOracle})
	for i := 0; i < 4; i++ {
		battlefieldCard(g, me, basic("Plains", "Plains"))
	}
	advanceTo(t, g, game.StepPrecombatMain)
	if got := castMovesFor(legal.EnumerateFor(g, me.ID), spell); len(got) != 0 {
		t.Fatalf("offered in a main phase: %v", labels(got))
	}
	advanceTo(t, g, game.StepBeginCombat)
	moves := legal.EnumerateFor(g, me.ID)
	got := castMovesFor(moves, spell)
	if len(got) == 0 {
		t.Fatalf("not offered at the beginning of combat: %v", labels(moves))
	}
	dispatchAll(t, g, me.ID, got)
}
