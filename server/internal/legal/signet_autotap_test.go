package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// signet_autotap_test.go is #2455 / #2456's enumerator half, on the
// real catalog cards: the move list asks the planner that now funds a
// Signet's {1} from the player's other sources, so a spell only the
// Signet's colours pay for is offered, and the main phase holding it is
// not dead air to smart autopass (which reads this list). Every cast
// offered is accepted and resolves.

const (
	oracleOrzhovSignet   = "de3dcb5d-775a-479f-99f5-d1883ed9b1b5"
	oracleThrabenInspect = "caa02547-66e3-4e27-a2d3-5e94f3e7a069"
)

func orzhovSignet() game.Card {
	return game.Card{Name: "Orzhov Signet", TypeLine: "Artifact", ManaCost: "{2}", OracleID: oracleOrzhovSignet}
}

func thrabenInspector() game.Card {
	return game.Card{
		Name: "Thraben Inspector", TypeLine: "Creature — Human Soldier", ManaCost: "{W}",
		Power: 1, Toughness: 2, OracleID: oracleThrabenInspect,
	}
}

// castAndResolve dispatches the offered cast on the live game and
// passes priority until the stack (the spell, then anything it
// triggered on entering) is empty.
func castAndResolve(t *testing.T, g *game.Game, seat uuid.UUID, m legal.Move, card uuid.UUID) {
	t.Helper()
	a := actions.Action{Type: actions.Type(m.Type), Player: m.Player, Caller: seat, Params: m.Params}
	if err := actions.Dispatch(g, a); err != nil {
		t.Fatalf("offered cast %q rejected: %v", m.Label, err)
	}
	for i := 0; i < 32 && len(g.Stack.Cards) > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !g.Battlefield.Contains(card) {
		t.Fatal("the cast did not resolve onto the battlefield")
	}
}

// TestSignetAndPlainsCastThrabenInspector: an Orzhov Signet and a
// Plains, Thraben Inspector in hand. The cast is offered and resolves,
// and the Inspector's Clue arrives.
func TestSignetAndPlainsCastThrabenInspector(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	battlefieldCard(g, active, orzhovSignet())
	lands(g, active, "Plains", "Plains", 1)
	inspector := handCard(active, thrabenInspector())
	advanceTo(t, g, game.StepPrecombatMain)

	offered := castMovesFor(legal.EnumerateFor(g, active.ID), inspector)
	if len(offered) == 0 {
		t.Fatal("Thraben Inspector not offered off a Plains and an Orzhov Signet")
	}
	castAndResolve(t, g, active.ID, offered[0], inspector)
}

// TestSignetPaysForItsColoursOffAPlains: a Plains alone pays the
// Inspector, so the case the Signet decides is a {1}{W} — the Plains
// pays the Signet's {1} and the Signet's {W}{B} pays the spell.
func TestSignetPaysForItsColoursOffAPlains(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	battlefieldCard(g, active, orzhovSignet())
	lands(g, active, "Plains", "Plains", 1)
	two := handCard(active, creature("White Two", "{1}{W}", 2, 2))
	advanceTo(t, g, game.StepPrecombatMain)

	offered := castMovesFor(legal.EnumerateFor(g, active.ID), two)
	if len(offered) == 0 {
		t.Fatal("{1}{W} not offered off a Plains and an Orzhov Signet")
	}
	castAndResolve(t, g, active.ID, offered[0], two)
}

// TestTurnThreeSignetMainPhaseIsNotDeadAir is the issue's board: two
// Swamps and an Orzhov Signet, and a white card in hand that only the
// Signet's {W} can pay for. Before #2455 the move list had no cast for
// it, so smart autopass passed the main phase. Now the cast is
// offered, which is what holds the stop, and it is accepted.
func TestTurnThreeSignetMainPhaseIsNotDeadAir(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	lands(g, active, "Swamp", "Swamp", 2)
	battlefieldCard(g, active, orzhovSignet())
	inspector := handCard(active, thrabenInspector())
	twoDrop := handCard(active, creature("White Two", "{1}{W}", 2, 2))
	threeDrop := handCard(active, creature("Orzhov Three", "{1}{W}{B}", 3, 3))
	tooBig := handCard(active, creature("White Four", "{3}{W}", 4, 4))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	for name, id := range map[string]uuid.UUID{"{W}": inspector, "{1}{W}": twoDrop, "{1}{W}{B}": threeDrop} {
		if len(castMovesFor(moves, id)) == 0 {
			t.Errorf("%s not offered off two Swamps and an Orzhov Signet: %v", name, labels(moves))
		}
	}
	// Two Swamps and a Signet are three mana, not four.
	if got := castMovesFor(moves, tooBig); len(got) != 0 {
		t.Errorf("{3}{W} offered off three mana: %v", labels(got))
	}
	dispatchAll(t, g, active.ID, castMovesFor(moves, threeDrop))
	dispatchAll(t, g, active.ID, castMovesFor(moves, inspector))
}

// TestSignetWithNoOtherSourceIsNotOffered: nothing pays the Signet's
// {1}, so the Inspector stays uncastable.
func TestSignetWithNoOtherSourceIsNotOffered(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	battlefieldCard(g, active, orzhovSignet())
	inspector := handCard(active, thrabenInspector())
	advanceTo(t, g, game.StepPrecombatMain)

	if got := castMovesFor(legal.EnumerateFor(g, active.ID), inspector); len(got) != 0 {
		t.Fatalf("Thraben Inspector offered off a lone Signet: %v", labels(got))
	}
}

// TestTwoSignetsAndOneLand: a Swamp pays the first Signet, whose mana
// pays the second, for three mana net. Two Signets alone pay nothing.
func TestTwoSignetsAndOneLand(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	battlefieldCard(g, active, orzhovSignet())
	battlefieldCard(g, active, orzhovSignet())
	inspector := handCard(active, thrabenInspector())
	advanceTo(t, g, game.StepPrecombatMain)
	if got := castMovesFor(legal.EnumerateFor(g, active.ID), inspector); len(got) != 0 {
		t.Fatalf("two Signets with no land were offered as paying for each other: %v", labels(got))
	}

	lands(g, active, "Swamp", "Swamp", 1)
	three := handCard(active, creature("Orzhov Three", "{1}{W}{B}", 3, 3))
	four := handCard(active, creature("Orzhov Four", "{2}{W}{B}", 4, 4))
	moves := legal.EnumerateFor(g, active.ID)
	if got := castMovesFor(moves, four); len(got) != 0 {
		t.Errorf("{2}{W}{B} offered off one land and two Signets: %v", labels(got))
	}
	offered := castMovesFor(moves, three)
	if len(offered) == 0 {
		t.Fatalf("{1}{W}{B} not offered off a Swamp and two Signets: %v", labels(moves))
	}
	castAndResolve(t, g, active.ID, offered[0], three)
}
