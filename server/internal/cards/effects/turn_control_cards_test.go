package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// turn_control_cards_test.go — #3857: the cards that say "you control
// target player during that player's next turn" (Mindslaver, Worst
// Fears, Sorin Markov's -7, Construct a Cosmic Cube, Cruel
// Entertainment). The machinery is pinned in game/turn_control_test.go
// and actions/controlled_turn_test.go; these tests pin each card's use
// of it, and that none of them owes the extra turn Emrakul does.

const (
	mindslaverOracle       = "a806f4ee-f48a-46c3-8772-5aaabebe4c7e"
	worstFearsOracle       = "face0a43-6604-4508-b511-81ab0bef7b18"
	cruelEntertainmentOra  = "7747804f-ca49-4a68-b508-b94b27c52b0f"
	constructCosmicCubeOra = "369aa9a6-af8a-432c-adff-ced2bb203dc7"
)

// scheduledControls is the controls still waiting for their turn.
func scheduledControls(g *game.Game) []game.PlayerControl {
	var out []game.PlayerControl
	g.ReadSnapshot(func() { out = g.ScheduledPlayerControlsLocked() })
	return out
}

// playLandAsController has `controller` play a land from `controlled`'s
// hand as them, through the dispatcher, which only a live control
// allows (CR 723.5). It returns the land's ID.
func playLandAsController(t *testing.T, g *game.Game, controller, controlled *game.Player) uuid.UUID {
	t.Helper()
	toMain(t, g)
	land := game.NewCard("Forest", controlled.ID)
	land.TypeLine = "Basic Land — Forest"
	g.WithWriteLock(func() { controlled.Hand.PushTop(land) })
	raw, err := json.Marshal(map[string]any{"instance_id": land.InstanceID.String()})
	if err != nil {
		t.Fatal(err)
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.TypeCastSpell, Player: controlled.ID, Caller: controller.ID, Params: raw}); err != nil {
		t.Fatalf("%s playing %s's land: %v", controller.Name, controlled.Name, err)
	}
	if c, ok := g.LookupCardForEffect(land.InstanceID); !ok || !g.Battlefield.Contains(land.InstanceID) || c.Controller != controlled.ID {
		t.Fatalf("the land is not on the battlefield under %s's control: %+v", controlled.Name, c)
	}
	return land.InstanceID
}

// walkControlledTurnWithNoExtraTurn ends turns until `controlled`'s turn
// begins, checks the controller holds it and can play as them, ends it,
// and checks the control is gone and the table moved on to the seat
// after `controlled` with no extra turn between (the difference from
// Emrakul, the Promised End).
func walkControlledTurnWithNoExtraTurn(t *testing.T, g *game.Game, controller, controlled *game.Player) {
	t.Helper()
	n := len(g.Seats)
	for i := 0; g.Turn.ActiveSeat != controlled.Seat; i++ {
		if i > n {
			t.Fatalf("never reached %s's turn", controlled.Name)
		}
		if got := g.ControllerOfPlayer(controlled.ID); got != controlled.ID {
			t.Fatalf("the control applies before %s's turn: %v", controlled.Name, got)
		}
		endTurn(t, g)
	}
	if g.Turn.Extra {
		t.Fatal("the controlled turn is an extra turn")
	}
	if got := g.ControllerOfPlayer(controlled.ID); got != controller.ID {
		t.Fatalf("%s's turn is controlled by %v, want %s", controlled.Name, got, controller.Name)
	}
	if got := g.ControlUntilOfPlayer(controlled.ID); got != game.ControlUntilTurn {
		t.Errorf("control lasts %q, want a whole turn", got)
	}
	playLandAsController(t, g, controller, controlled)
	// The controlled player decides nothing of their own turn (CR 723.5).
	if err := actions.Dispatch(g, actions.Action{Type: actions.TypePassPriority, Player: controlled.ID, Caller: controlled.ID}); err == nil {
		t.Error("the controlled player passed priority for themselves")
	}

	endTurn(t, g)
	assertTurnOf(t, g, (controlled.Seat+1)%n, false, "after the controlled turn: the rotation resumes, no extra turn")
	if got := g.ControllerOfPlayer(controlled.ID); got != controlled.ID {
		t.Errorf("the control outlived the turn: %v", got)
	}
	if len(g.PlayerControls) != 0 {
		t.Errorf("records left after the controlled turn: %+v", g.PlayerControls)
	}
}

// --- Mindslaver -----------------------------------------------------

// The whole card: {4}, {T}, sacrifice; the control waits, covers the
// target's next turn, and the target takes no extra turn after it.
func TestMindslaverControlsTheTargetsNextTurnAndOwesNoExtraTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	pushLands(g, me.ID, 4)
	slaver := pushCatalogPermanent(g, me.ID, "Mindslaver", "Legendary Artifact", mindslaverOracle, false)
	advanceToMain(t, g)

	b16Activate(t, g, me.ID, slaver, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	// The sacrifice is a cost: Mindslaver is gone before the ability resolves.
	if g.Battlefield.Contains(slaver) {
		t.Fatal("Mindslaver was not sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)

	waiting := scheduledControls(g)
	if len(waiting) != 1 {
		t.Fatalf("%d controls scheduled, want 1", len(waiting))
	}
	if w := waiting[0]; w.Controller != me.ID || w.Controlled != opp.ID || w.SourceName != "Mindslaver" || w.ExtraTurn != nil {
		t.Errorf("scheduled %+v, want Mindslaver: %s controls %s with no extra turn", w, me.Name, opp.Name)
	}
	walkControlledTurnWithNoExtraTurn(t, g, me, opp)
}

// "Target player": the controller may point it at themselves, which
// changes nothing.
func TestMindslaverTargetingYourselfControlsNobody(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushLands(g, me.ID, 4)
	slaver := pushCatalogPermanent(g, me.ID, "Mindslaver", "Legendary Artifact", mindslaverOracle, false)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, slaver, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	})
	passPriorityAroundTable(t, g)
	if n := len(scheduledControls(g)); n != 0 {
		t.Errorf("%d controls scheduled by controlling yourself, want 0", n)
	}
}

// --- Worst Fears ----------------------------------------------------

// The spell exiles itself as its last instruction, and the control
// follows the plain shape.
func TestWorstFearsControlsTheTargetsNextTurnAndExilesItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := castCatalogSpell(t, g, "Worst Fears", "Sorcery", worstFearsOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

	if !inExile(g, id) {
		t.Error("Worst Fears did not exile itself")
	}
	for _, c := range me.Graveyard.Cards {
		if c.InstanceID == id {
			t.Error("Worst Fears went to the graveyard instead of exile")
		}
	}
	if w := scheduledControls(g); len(w) != 1 || w[0].Controller != me.ID || w[0].Controlled != opp.ID || w[0].SourceName != "Worst Fears" {
		t.Fatalf("scheduled %+v, want Worst Fears: %s controls %s", w, me.Name, opp.Name)
	}
	walkControlledTurnWithNoExtraTurn(t, g, me, opp)
}

// A countered Worst Fears does nothing and is not exiled.
func TestWorstFearsCounteredControlsNobody(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := castCatalogSpell(t, g, "Worst Fears", "Sorcery", worstFearsOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(id); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if n := len(scheduledControls(g)); n != 0 {
		t.Errorf("%d controls scheduled by a countered spell, want 0", n)
	}
	if inExile(g, id) {
		t.Error("a countered Worst Fears was exiled; it goes to the graveyard")
	}
	found := false
	for _, c := range me.Graveyard.Cards {
		found = found || c.InstanceID == id
	}
	if !found {
		t.Error("a countered Worst Fears is not in the graveyard")
	}
}

// --- Sorin Markov ---------------------------------------------------

// The -7: costs seven loyalty, controls the target's next turn, no
// extra turn.
func TestSorinMarkovMinusSevenControlsTheTargetsNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+3)%len(g.Seats)]
	sorin := pushCatalogPermanent(g, me.ID, "Sorin Markov", "Legendary Planeswalker — Sorin", b39SorinMarkovOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(sorin, game.CounterLoyalty, 7) })
	advanceToMain(t, g)

	loyaltyBefore := loyaltyCount(g, sorin)
	b16Activate(t, g, me.ID, sorin, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	passPriorityAroundTable(t, g)
	if got := loyaltyCount(g, sorin); got != loyaltyBefore-7 {
		t.Errorf("loyalty %d -> %d, want -7", loyaltyBefore, got)
	}
	if w := scheduledControls(g); len(w) != 1 || w[0].Controller != me.ID || w[0].Controlled != opp.ID || w[0].SourceName != "Sorin Markov" {
		t.Fatalf("scheduled %+v, want Sorin Markov: %s controls %s", w, me.Name, opp.Name)
	}
	walkControlledTurnWithNoExtraTurn(t, g, me, opp)
}

func TestSorinMarkovIsFullAndOffersThreeAbilities(t *testing.T) {
	spec, ok := Lookup(b39SorinMarkovOracle)
	if !ok {
		t.Fatal("Sorin Markov is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Sorin Markov is %v with caveats %v, want Full", spec.Completeness, spec.Caveats)
	}
	if len(spec.Activated) != 3 {
		t.Errorf("%d abilities, want +2, -3 and -7", len(spec.Activated))
	}
}

// --- Cruel Entertainment --------------------------------------------

// Two controls, one each way, and neither the caster's: seat 1 controls
// seat 2's turn and seat 2 controls seat 1's, with no extra turns.
func TestCruelEntertainmentSwapsControlOfTheTwoPlayers(t *testing.T) {
	g := newCatalogGame(t)
	n := len(g.Seats)
	if n < 4 {
		t.Fatalf("setup: %d seats", n)
	}
	caster := g.Seats[g.Turn.ActiveSeat]
	first, second := g.Seats[(g.Turn.ActiveSeat+1)%n], g.Seats[(g.Turn.ActiveSeat+2)%n]
	castCatalogSpell(t, g, "Cruel Entertainment", "Sorcery", cruelEntertainmentOra, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: first.ID},
		{Kind: game.TargetPlayer, ID: second.ID},
	})
	passPriorityAroundTable(t, g)

	waiting := scheduledControls(g)
	if len(waiting) != 2 {
		t.Fatalf("%d controls scheduled, want 2: %+v", len(waiting), waiting)
	}
	seen := map[[2]uuid.UUID]bool{}
	for _, w := range waiting {
		seen[[2]uuid.UUID{w.Controller, w.Controlled}] = true
		if w.ExtraTurn != nil || w.SourceName != "Cruel Entertainment" {
			t.Errorf("record %+v: want Cruel Entertainment with no extra turn", w)
		}
	}
	if !seen[[2]uuid.UUID{first.ID, second.ID}] || !seen[[2]uuid.UUID{second.ID, first.ID}] {
		t.Errorf("records %+v: want each of the two controlling the other", waiting)
	}
	if got := g.ControllerOfPlayer(caster.ID); got != caster.ID {
		t.Error("the caster is controlled by their own spell")
	}

	// First's turn is the second player's to play...
	endTurn(t, g)
	assertTurnOf(t, g, first.Seat, false, "the first player's turn")
	if got := g.ControllerOfPlayer(first.ID); got != second.ID {
		t.Fatalf("the first player's turn is controlled by %v, want the second player", got)
	}
	endTurn(t, g)
	// ...and second's turn is the first player's, with no extra turn between.
	assertTurnOf(t, g, second.Seat, false, "the second player's turn")
	if got := g.ControllerOfPlayer(second.ID); got != first.ID {
		t.Fatalf("the second player's turn is controlled by %v, want the first player", got)
	}
	endTurn(t, g)
	assertTurnOf(t, g, (second.Seat+1)%n, false, "after both controlled turns")
	if len(g.PlayerControls) != 0 {
		t.Errorf("records left: %+v", g.PlayerControls)
	}
}

// "Another target player": the same player twice is not a legal cast.
func TestCruelEntertainmentNeedsTwoDifferentPlayers(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	err := castCatalogSpellErr(t, g, "Cruel Entertainment", "Sorcery", cruelEntertainmentOra, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID},
		{Kind: game.TargetPlayer, ID: opp.ID},
	})
	if err == nil {
		t.Error("Cruel Entertainment was cast at one player twice")
	}
}

// --- Construct a Cosmic Cube ----------------------------------------

// The second draw of a turn makes a Villain and the seventh plan
// counter sacrifices the Cube; its reflexive trigger asks for an
// opponent and controls their next turn.
func TestConstructACosmicCubeSeventhCounterControlsAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	advanceToMain(t, g)
	cube := pushCatalogPermanent(g, me.ID, "Construct a Cosmic Cube", "Enchantment — Plan", constructCosmicCubeOra, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(cube, "plan", 6) })
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cube) {
		t.Fatal("the Cube was sacrificed at six counters")
	}

	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(me.ID, 2); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Villain"); n != 1 {
		t.Fatalf("%d Villain tokens, want 1", n)
	}
	if g.Battlefield.Contains(cube) {
		t.Fatal("the Cube was not sacrificed at the seventh plan counter")
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the reflexive trigger did not ask for a target opponent")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if w := scheduledControls(g); len(w) != 1 || w[0].Controller != me.ID || w[0].Controlled != opp.ID || w[0].ExtraTurn != nil {
		t.Fatalf("scheduled %+v, want %s controlling %s with no extra turn", w, me.Name, opp.Name)
	}
	walkControlledTurnWithNoExtraTurn(t, g, me, opp)
}

// Only the crossing placement pays off: six counters and a draw make a
// token and a counter, and the Cube stays.
func TestConstructACosmicCubeBelowSevenJustMakesVillains(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	cube := pushCatalogPermanent(g, me.ID, "Construct a Cosmic Cube", "Enchantment — Plan", constructCosmicCubeOra, false)
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(me.ID, 2); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Villain"); n != 1 {
		t.Errorf("%d Villain tokens, want 1", n)
	}
	if c, ok := battlefieldCard(g, cube); !ok || c.Counters["plan"] != 1 {
		t.Errorf("Cube = %+v, want it on the battlefield with one plan counter", c)
	}
	if n := len(scheduledControls(g)); n != 0 {
		t.Errorf("%d controls scheduled before the seventh counter", n)
	}
	if v := TokenCard("2/1 black Villain with menace"); v.Power != 2 || v.Toughness != 1 || len(v.Keywords) != 1 || v.Keywords[0] != "menace" {
		t.Errorf("Villain template = %+v", v)
	}
}
