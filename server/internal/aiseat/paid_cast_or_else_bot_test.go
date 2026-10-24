package aiseat_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// paid_cast_or_else_bot_test.go — #3510 through the real card, the real
// enumerator and the real heuristic. Chandra, Torch of Defiance's first +1
// asks a paid_or_else resolution_cast inside its resolution (#4250, CR
// 608.2g); the enumerator offers the paid cast from exile and the
// always-legal decline, the heuristic takes the cast, and the engine
// accepts it.
//
// Not behind AISEAT_GAME_TESTS: one resolution, a few decisions.

const botChandraTorchOracle = "12cc800c-c7be-4811-8207-fd01a99cc892"

func TestHeuristicSeatAnswersChandraTorchAndCastsTheExiledCard(t *testing.T) {
	chandraPaidCast(t, 3510, false)
}

// #3521: a better spell in hand wants the same two mana. The exiled card
// is the one the decline loses, and nothing else can be cast while the
// offer holds the resolution, so it is cast.
func TestHeuristicSeatTakesThePassClosedCastBeforeSpendingTheManaElsewhere(t *testing.T) {
	chandraPaidCast(t, 3521, true)
}

func chandraPaidCast(t *testing.T, seed uint64, rival bool) {
	t.Helper()
	g := newSettledTable(t, seed)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToStep(t, g, game.StepPrecombatMain)
	// Nothing else to spend the mana on: the decision under test is the
	// exiled card against the pass, unless rival asks for a competitor.
	me.Hand.Cards = nil
	me.Command.Cards = nil
	if rival {
		hand := game.Card{InstanceID: uuid.New(), Name: "Runeclaw Bear", TypeLine: "Creature — Bear",
			ManaCost: "{1}{G}", Power: 5, Toughness: 5, Colors: []string{"G"}, Owner: me.ID, Controller: me.ID}
		hand.AddKnowersAll(seatIDs(g))
		me.Hand.PushTop(hand)
	}

	chandra := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: chandra, Name: "Chandra, Torch of Defiance", OracleID: botChandraTorchOracle,
		TypeLine: "Legendary Planeswalker — Chandra", ManaCost: "{2}{R}{R}", Colors: []string{"R"},
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLoyalty: 4},
	})
	for i := 0; i < 2; i++ {
		g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
			Owner: me.ID, Controller: me.ID})
	}
	bears := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: bears, Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Power: 2, Toughness: 2, Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})

	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		if p.ID != me.ID {
			lives[p.ID] = p.Life
		}
	}

	if err := g.ActivateCatalogAbility(me.ID, chandra, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Chandra's +1: %v", err)
	}
	for i := 0; i < 16 && !owesChoice(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !owesChoice(g, me.ID) {
		t.Fatal("Chandra's +1 asked nothing")
	}

	// The prompt is answerable both ways, every move names it, and the
	// decline is the always-legal way out.
	moves := legal.EnumerateFor(g, me.ID)
	var accept, decline bool
	for _, mv := range moves {
		switch {
		case mv.ResolutionCast == "":
			t.Errorf("move %q does not answer the paid_or_else offer", mv.Label)
		case mv.Type == legal.TypeCastSpell && mv.Source == bears:
			accept = true
		case mv.Type == legal.TypeResolveChoice && mv.AlwaysLegal:
			decline = true
		}
	}
	if !accept || !decline {
		t.Fatalf("the paid_or_else prompt offers cast %v and don't %v; moves %v", accept, decline, attackLabels(moves))
	}

	pol := heuristic.New()
	taken := driveChoices(t, g, pol, me.ID, 4)
	if len(taken) != 1 {
		t.Fatalf("answered the prompt with %v", taken)
	}
	if !g.Stack.Contains(bears) {
		t.Fatalf("the bot did not cast the exiled Bears (took %v)", taken)
	}
	if g.ResolutionPaused() {
		t.Error("the +1 is still resolving after the cast")
	}
	for i := 0; i < 16 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !g.Battlefield.Contains(bears) {
		t.Error("the cast Bears did not resolve onto the battlefield")
	}
	for _, p := range g.Seats {
		if want, ok := lives[p.ID]; ok && p.Life != want {
			t.Errorf("the card was cast, so no damage: opponent life %d → %d", want, p.Life)
		}
	}
}
