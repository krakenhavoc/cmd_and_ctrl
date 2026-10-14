package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// emrakul_the_promised_end_test.go — ADR 0167 PR 2: Emrakul, the
// Promised End. The control record, its routing and the turn machinery
// are pinned in game/turn_control_test.go and
// actions/controlled_turn_test.go; these tests pin the card.

const emrakulPromisedEndOracle = "4e7a8817-1a66-45c3-ade9-eac79b40b89f"

// "This spell costs {1} less to cast for each card type among cards in
// your graveyard": an artifact creature is two types, and a second
// creature adds nothing.
func TestEmrakulThePromisedEndCostsLessPerCardTypeInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	price := func() int {
		return selfPricedMV(t, g, me, emrakulPromisedEndOracle, "Legendary Creature — Eldrazi", "{13}")
	}
	if got := price(); got != 13 {
		t.Fatalf("empty graveyard: %d, want 13", got)
	}
	for _, tl := range []string{"Creature — Bear", "Artifact Creature — Golem", "Instant", "Basic Land — Forest", "Creature — Elf"} {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: tl, TypeLine: tl, Owner: me.ID, Controller: me.ID})
	}
	// An opponent's graveyard is not yours.
	opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Sorcery", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	if got := price(); got != 9 {
		t.Errorf("creature, artifact, instant and land in the graveyard: %d, want 9", got)
	}
}

// emrakulCast casts Emrakul from the active seat's hand and targets
// `target` with the cast trigger. It returns Emrakul's instance ID.
func emrakulCast(t *testing.T, g *game.Game, target uuid.UUID) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := putInHand(me, game.Card{Name: "Emrakul, the Promised End", TypeLine: "Legendary Creature — Eldrazi",
		OracleID: emrakulPromisedEndOracle, ManaCost: "{13}", Power: 13, Toughness: 13})
	toMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the cast trigger did not ask for its target opponent")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: target}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	return id
}

// The ruling: "It resolves even if that spell is countered." The cast
// trigger schedules the control above the countered spell.
func TestEmrakulThePromisedEndSchedulesTheControlEvenIfCountered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	id := emrakulCast(t, g, opp.ID)
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(id); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if g.Stack.Contains(id) || g.Battlefield.Contains(id) {
		t.Fatal("Emrakul was not countered")
	}
	var waiting []game.PlayerControl
	g.ReadSnapshot(func() { waiting = g.ScheduledPlayerControlsLocked() })
	if len(waiting) != 1 {
		t.Fatalf("%d controls scheduled, want 1", len(waiting))
	}
	w := waiting[0]
	if w.Controller != me.ID || w.Controlled != opp.ID || w.SourceName != "Emrakul, the Promised End" {
		t.Errorf("scheduled %+v, want %v controlling %v", w, me.ID, opp.ID)
	}
	if got := g.ControllerOfPlayer(opp.ID); got != opp.ID {
		t.Error("the control applies before the opponent's turn")
	}
}

// The whole card: Emrakul resolves, the caster plays the opponent's
// next turn (a land from their hand, through the dispatcher, lands
// under the opponent's control, CR 723.3), and the opponent takes an
// extra turn after it (CR 500.7) that is theirs to play. Then the
// rotation resumes after the opponent.
func TestEmrakulThePromisedEndControlsTheNextTurnThenGivesAnExtraTurn(t *testing.T) {
	g := newCatalogGame(t)
	n := len(g.Seats)
	meSeat := g.Turn.ActiveSeat
	oppSeat := (meSeat + 2) % n
	me, opp := g.Seats[meSeat], g.Seats[oppSeat]
	id := emrakulCast(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Emrakul did not resolve")
	}

	endTurn(t, g)
	assertTurnOf(t, g, (meSeat+1)%n, false, "the next seat's turn")
	if got := g.ControllerOfPlayer(opp.ID); got != opp.ID {
		t.Fatal("the control applies before the opponent's turn")
	}
	endTurn(t, g)
	assertTurnOf(t, g, oppSeat, false, "the controlled turn")
	if got := g.ControllerOfPlayer(opp.ID); got != me.ID {
		t.Fatalf("the opponent's turn is controlled by %v, want the caster", got)
	}

	toMain(t, g)
	land := game.NewCard("Forest", opp.ID)
	land.TypeLine = "Basic Land — Forest"
	g.WithWriteLock(func() { opp.Hand.PushTop(land) })
	raw, err := json.Marshal(map[string]any{"instance_id": land.InstanceID.String()})
	if err != nil {
		t.Fatal(err)
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.TypeCastSpell, Player: opp.ID, Caller: me.ID, Params: raw}); err != nil {
		t.Fatalf("the caster playing the opponent's land: %v", err)
	}
	if c, ok := g.LookupCardForEffect(land.InstanceID); !ok || !g.Battlefield.Contains(land.InstanceID) || c.Controller != opp.ID {
		t.Fatalf("the land is not on the battlefield under the opponent's control: %+v", c)
	}
	// The opponent decides nothing of their own turn (CR 723.5).
	if err := actions.Dispatch(g, actions.Action{Type: actions.TypePassPriority, Player: opp.ID, Caller: opp.ID}); err == nil {
		t.Error("the controlled opponent passed priority for themselves")
	}

	endTurn(t, g)
	assertTurnOf(t, g, oppSeat, true, "the extra turn after the controlled turn")
	if got := g.ControllerOfPlayer(opp.ID); got != opp.ID {
		t.Fatalf("the extra turn is controlled by %v, want the opponent", got)
	}
	endTurn(t, g)
	assertTurnOf(t, g, (oppSeat+1)%n, false, "after the extra turn")
}
