package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// mkm_disguise_x_cards_test.go — #5249: disguise {X} (CR 107.3d,
// CR 702.168e) and Aurelia's Vindicator on it. The X is chosen as the
// turn-face-up special action is taken, paid as generic mana, and read
// by the turned-face-up trigger as its target count. Each test plays the
// real catalog entry.

const aureliasVindicatorOracle = "4f452ce7-1166-4810-afed-ea5e156a6f52"

func TestAureliasVindicatorIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(aureliasVindicatorOracle)
	if !ok {
		t.Fatal("Aurelia's Vindicator is not registered")
	}
	if spec.Name != "Aurelia's Vindicator" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
	alt := game.FaceDownCastFor(aureliasVindicatorOracle)
	if alt == nil || alt.FaceDown == nil || alt.FaceDown.Kind != game.FaceDownDisguised || alt.FaceDown.FaceUpCost != "{X}{3}{W}" {
		t.Errorf("disguise declaration = %+v, want disguise {X}{3}{W}", alt)
	}
}

// --- fixtures ------------------------------------------------------------------

// avFaceDown casts Aurelia's Vindicator face down for the active player,
// resolves it, and empties that player's pool.
func avFaceDown(t *testing.T, g *game.Game) (*game.Player, uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := mkmMPCastFaceDown(t, g, "Aurelia's Vindicator", "Creature — Angel", aureliasVindicatorOracle, 4, 2)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	return me, id
}

// avTurnUp gives the player `mana` and turns the Vindicator face up for X.
func avTurnUp(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, x int, mana ...string) error {
	t.Helper()
	g.WithWriteLock(func() { me.ManaPool.AddMana(manaTokens(mana...)...) })
	return g.PerformSpecialAction(me.ID, id, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true, XValue: x})
}

// avGraveCreature puts a creature card into owner's graveyard.
func avGraveCreature(g *game.Game, owner *game.Player, name string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		owner.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: owner.ID, Controller: owner.ID})
	})
	return id
}

// --- the trigger ---------------------------------------------------------------

// The headline: X = 2 paid, so the turned-face-up trigger exiles up to two
// other creatures and/or creature cards in graveyards, and when the
// Vindicator leaves they return to their owners' hands.
func TestAureliasVindicatorExilesUpToTheChosenXAndReturnsThemToHand(t *testing.T) {
	g := newCatalogGame(t)
	me, id := avFaceDown(t, g)
	opp := mkmMPOpponents(g)[0]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, opp.ID, "Their Other Bear", "Creature — Bear", 2, 2)
	dead := avGraveCreature(g, opp, "Dead Bear")

	if err := avTurnUp(t, g, me, id, 2, "W", "C", "C", "C", "C", "C"); err != nil {
		t.Fatalf("turn face up for X=2: %v", err)
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("%d mana left, want 0: X=2 costs {2}{3}{W}", got)
	}
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if p.PickTargetMin != 0 || p.PickTargetMax != 2 {
		t.Errorf("\"up to X\" with X=2: min %d max %d, want 0 and 2", p.PickTargetMin, p.PickTargetMax)
	}
	if hasID(p.PickTargetCards, id) {
		t.Error("the Vindicator itself is offered: the clause says other")
	}
	if !hasID(p.PickTargetCards, bear) || !hasID(p.PickTargetCards, dead) {
		t.Error("a creature on the battlefield and a creature card in a graveyard are both offered")
	}
	b17PickCards(t, g, me.ID, bear, dead)
	passPriorityAroundTable(t, g)
	if !inExile(g, bear) || !inExile(g, dead) {
		t.Fatal("the two chosen cards are exiled")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("the creature not chosen stays")
	}
	c := findCardAnywhere(t, g, id)
	if !hasEffectiveKeyword(t, g, id, "flying") || !hasEffectiveKeyword(t, g, id, "lifelink") || c.CurrentPower() != 4 {
		t.Error("face up it is a 4/2 with flying and lifelink")
	}

	b27Kill(g, id)
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(bear) || !opp.Hand.Contains(dead) {
		t.Error("the exiled cards return to their owner's hand when the Vindicator leaves")
	}
}

// X = 0 targets nothing: the Vindicator comes up for {3}{W} and exiles
// nothing.
func TestAureliasVindicatorAtXZeroExilesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, id := avFaceDown(t, g)
	opp := mkmMPOpponents(g)[0]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	if err := avTurnUp(t, g, me, id, 0, "W", "C", "C", "C"); err != nil {
		t.Fatalf("turn face up for X=0: %v", err)
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("X=0 asks for targets")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) || findCardAnywhere(t, g, id).FaceDown {
		t.Error("X=0 turns the Vindicator up and exiles nothing")
	}
}

// An X the player cannot pay is refused before anything is spent.
func TestAureliasVindicatorRefusesAnUnaffordableX(t *testing.T) {
	g := newCatalogGame(t)
	me, id := avFaceDown(t, g)
	if err := avTurnUp(t, g, me, id, 3, "W", "C", "C", "C", "C", "C"); err == nil {
		t.Fatal("X=3 with six mana was taken")
	}
	if !findCardAnywhere(t, g, id).FaceDownIsPermanent() || me.ManaPool.Total() != 6 {
		t.Errorf("the refused action changed the board: face down %v, %d mana left",
			findCardAnywhere(t, g, id).FaceDownIsPermanent(), me.ManaPool.Total())
	}
}

// Turned up by an effect, nothing chose an X: the trigger exiles nothing.
func TestAureliasVindicatorTurnedUpByAnEffectExilesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, id := avFaceDown(t, g)
	opp := mkmMPOpponents(g)[0]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	var turned bool
	g.WithWriteLock(func() { turned = g.TurnFaceUpForEffect(uuid.Nil, me.ID, id) })
	if !turned {
		t.Fatal("the effect did not turn it up")
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("an effect's turn face up asks for targets")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("an effect's turn face up exiled a creature")
	}
}

// A restore taken after it turns face up, with the trigger on the stack
// and its two targets chosen, keeps the X the player chose: the restored
// trigger re-derives its clause from the event's X and exiles both.
func TestAureliasVindicatorTriggerRestoresWithTheChosenX(t *testing.T) {
	g := newCatalogGame(t)
	me, id := avFaceDown(t, g)
	opp := mkmMPOpponents(g)[0]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	dead := avGraveCreature(g, opp, "Dead Bear")
	if err := avTurnUp(t, g, me, id, 2, "W", "C", "C", "C", "C", "C"); err != nil {
		t.Fatalf("turn face up for X=2: %v", err)
	}
	b04WaitForPick(t, g, me.ID)
	b17PickCards(t, g, me.ID, bear, dead)
	if len(g.StackMeta) == 0 {
		t.Fatal("the trigger is not on the stack")
	}

	restored := restoreRoundTrip(t, g, true)
	if c := findCardAnywhere(t, restored, id); c.FaceDown {
		t.Fatal("the restored Vindicator is face down")
	}
	passPriorityAroundTable(t, restored)
	if !inExile(restored, bear) || !inExile(restored, dead) {
		t.Error("the restored trigger exiles both targets the X=2 clause allowed")
	}
	rOpp := restored.PlayerByIDForEffect(opp.ID)
	if rOpp == nil {
		t.Fatal("the opponent is not on the restored table")
	}
	b27Kill(restored, id)
	passPriorityAroundTable(t, restored)
	if !rOpp.Hand.Contains(bear) || !rOpp.Hand.Contains(dead) {
		t.Error("on the restored table the leave trigger still returns them to hand")
	}
}

// --- the legal move and the view -------------------------------------------------

// The enumerator offers the largest X the seat can pay, with the X open
// from 0 to it (ADR 0122 §6.2); the view stamps the same ceiling on the
// row and marks it as demanding an X.
func TestAureliasVindicatorTurnUpOffersTheMaxX(t *testing.T) {
	g := newCatalogGame(t)
	me, id := avFaceDown(t, g)
	g.WithWriteLock(func() { me.ManaPool.AddMana(manaTokens("W", "C", "C", "C", "C", "C", "C")...) })

	var move *legal.Move
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type == legal.TypeSpecialAction && m.Source == id {
			move = &m
			break
		}
	}
	if move == nil {
		t.Fatal("no turn_face_up move for the Vindicator")
	}
	var params struct {
		Kind   string `json:"kind"`
		XValue int    `json:"x_value"`
	}
	if err := json.Unmarshal(move.Params, &params); err != nil {
		t.Fatal(err)
	}
	if params.Kind != "turn_face_up" || params.XValue != 3 {
		t.Errorf("move params %s, want turn_face_up at x_value 3 (seven mana pay {3}{3}{W})", move.Params)
	}
	if v := move.Value; v == nil || v.Kind != legal.ValueX || v.Min == nil || *v.Min != 0 || v.Max == nil || *v.Max != 3 {
		t.Errorf("move value %+v, want X open from 0 to 3", move.Value)
	}

	row := fdTurnUpRow(t, g, me.ID, id)
	if !row.DemandsX || row.XMax == nil || *row.XMax != 3 {
		t.Errorf("view row %+v, want demands_x with x_max 3", row)
	}

	// The move is one the engine accepts.
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true, AutoTap: true, XValue: params.XValue}); err != nil {
		t.Fatalf("the offered move was refused: %v", err)
	}
}
