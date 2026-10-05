package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// end_the_turn_cards_test.go — the proof cards for #2165 (CR 724.1, end
// the turn): Sundial of the Infinite, Obeka, Brute Chronologist, Time
// Stop, Glorious End and Ultima. The engine's own tests are
// game/end_turn_test.go.

const (
	sundialOracle     = "79d46e09-2548-440f-9c02-3c3dc39a0cd1"
	obekaOracle       = "21b76a94-d9f3-4c34-ab24-7e89321b04f0"
	timeStopOracle    = "55bc6a7e-2356-48f0-a93a-8bf19044ee4b"
	gloriousEndOracle = "e611a3e0-eb0e-466e-a771-51310eeb34cd"
	ultimaOracle      = "c9a45c28-826e-4e6e-8adc-1368ea1af11a"
	endTurnTestLabel  = "end-the-turn test — exile it at the next end step"
)

// The classic Sundial play: at the beginning of your end step a delayed
// "exile it" trigger goes on the stack, and Sundial ends the turn in
// response. The trigger is exiled with the stack — the creature stays —
// the turn goes straight to cleanup and the next player's turn begins.
func TestSundialOfTheInfiniteEndsTheTurnOverItsEndStepTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sundial := pushCatalogPermanent(g, me.ID, "Sundial of the Infinite", "Artifact", sundialOracle, false)
	unearthed := pushVanillaCreature(g, me.ID, "Unearthed", 3, 3)
	advanceTo(t, g, game.StepPrecombatMain)
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: unearthed})
		if err := (ScheduleDelayedTrigger{
			Label: endTurnTestLabel,
			Cards: []uuid.UUID{unearthed},
			Body:  exileListedCardsBody,
		}).Apply(ctx); err != nil {
			t.Fatalf("schedule: %v", err)
		}
	})
	advanceTo(t, g, game.StepEnd)
	if !stackHasLabel(g, endTurnTestLabel) {
		t.Fatalf("setup: the end-step trigger is not on the stack")
	}
	turn := g.Turn.Seq
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, sundial, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Sundial: %v", err)
	}
	passPriorityAroundTable(t, g)
	discardIfOwed(t, g)

	if g.Turn.Seq != turn+1 || g.Turn.ActiveSeat == me.Seat {
		t.Fatalf("the turn did not end: turn %d seat %d step %s", g.Turn.Seq, g.Turn.ActiveSeat, g.Turn.Step)
	}
	if _, ok := aangCardOnBF(g, unearthed); !ok {
		t.Errorf("the creature was exiled: the trigger should have been exiled from the stack first")
	}
	if c, ok := aangCardOnBF(g, sundial); !ok {
		t.Errorf("Sundial left the battlefield")
	} else if c.Controller != me.ID {
		t.Errorf("Sundial changed hands")
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("%d delayed triggers queued; the one that fired is gone, not re-queued", len(g.DelayedTriggers))
	}
}

// stackHasLabel reports whether a stack item carries the label.
func stackHasLabel(g *game.Game, label string) bool {
	for _, item := range g.StackMeta {
		if item != nil && item.Label == label {
			return true
		}
	}
	return false
}

// "Activate only during your turn."
func TestSundialOfTheInfiniteOnlyDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	sundial := pushCatalogPermanent(g, opp.ID, "Sundial of the Infinite", "Artifact", sundialOracle, false)
	floatMana(t, g, opp, "{C}")
	if err := g.ActivateCatalogAbility(opp.ID, sundial, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("Sundial was activated during an opponent's turn")
	}
	if c, _ := aangCardOnBF(g, sundial); c.Tapped {
		t.Error("the refused activation tapped Sundial")
	}
}

// Time Stop on an opponent's turn, in their declare attackers step: the
// spell they cast is exiled with Time Stop, the attack never deals
// damage, and the next player's turn begins.
func TestTimeStopEndsAnOpponentsTurnMidCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := pushVanillaCreature(g, me.ID, "Charger", 5, 5)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	theirs := castInPlace(t, g, me.ID, "Some Instant", "")
	stop := castInPlace(t, g, opp.ID, "Time Stop", timeStopOracle)
	life := opp.Life
	turn := g.Turn.Seq
	passPriorityAroundTable(t, g)
	discardIfOwed(t, g)

	for _, id := range []uuid.UUID{theirs, stop} {
		if !g.Exile.Contains(id) {
			t.Errorf("%s is not in exile (CR 724.1b)", id)
		}
	}
	if opp.Life != life {
		t.Errorf("the defender lost %d life; the turn ended before combat damage", life-opp.Life)
	}
	if c, ok := aangCardOnBF(g, attacker); !ok || c.AttackingTarget != uuid.Nil {
		t.Errorf("the attacker is still in combat")
	}
	if g.Turn.Seq != turn+1 || g.Turn.ActiveSeat != 1 {
		t.Errorf("turn %d seat %d, want the next player's turn", g.Turn.Seq, g.Turn.ActiveSeat)
	}
}

// Glorious End: the turn ends, and "at the beginning of your next end
// step" is the caster's next turn — not this turn's skipped end step,
// and not an opponent's end step in between.
func TestGloriousEndLosesAtYourNextEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	castCatalogSpell(t, g, "Glorious End", "Instant", gloriousEndOracle, nil)
	turn := g.Turn.Seq
	passPriorityAroundTable(t, g)
	discardIfOwed(t, g)
	if g.Turn.Seq != turn+1 {
		t.Fatalf("the turn did not end")
	}
	if g.Seats[me].Eliminated {
		t.Fatal("lost the turn Glorious End ended")
	}
	for seat := 1; seat < len(g.Seats); seat++ {
		advanceToEndStepOf(t, g, (me+seat)%len(g.Seats))
		passPriorityAroundTable(t, g)
		if g.Seats[me].Eliminated {
			t.Fatalf("lost in seat %d's end step; only the caster's own end step counts", (me+seat)%len(g.Seats))
		}
	}
	advanceToEndStepOf(t, g, me)
	passPriorityAroundTable(t, g)
	if !g.Seats[me].Eliminated {
		t.Fatal("did not lose at the beginning of the next end step of the caster's own turn")
	}
}

// Ultima: every artifact and creature is destroyed, lands survive, and
// the dies triggers the wipe causes cease to exist with the turn
// (CR 724.1a) — Zulaport Cutthroat drains nobody.
func TestUltimaWipesAndItsDiesTriggersCeaseToExist(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	zulaport := pushCatalogPermanent(g, me.ID, "Zulaport Cutthroat", "Creature — Human Rogue", zulaportOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	rock := pushPermanentForTest(g, opp.ID, "Mind Stone", "", "Artifact")
	land := pushLand(g, opp.ID, "Island")
	lifeMe, lifeOpp := me.Life, opp.Life
	ultima := castCatalogSpell(t, g, "Ultima", "Sorcery", ultimaOracle, nil)
	turn := g.Turn.Seq
	passPriorityAroundTable(t, g)
	discardIfOwed(t, g)

	for _, id := range []uuid.UUID{zulaport, bear, rock} {
		if _, ok := aangCardOnBF(g, id); ok {
			t.Errorf("%s survived the wipe", id)
		}
	}
	if _, ok := aangCardOnBF(g, land); !ok {
		t.Errorf("the land was destroyed")
	}
	if me.Life != lifeMe || opp.Life != lifeOpp {
		t.Errorf("life moved (%d→%d, %d→%d): the wipe's dies triggers ceased to exist with the turn (CR 724.1a)",
			lifeMe, me.Life, lifeOpp, opp.Life)
	}
	if !g.Exile.Contains(ultima) {
		t.Errorf("Ultima was not exiled with the stack")
	}
	if len(g.PendingChoices) != 0 || len(g.PendingTriggers) != 0 {
		t.Errorf("%d prompts and %d triggers survived the turn", len(g.PendingChoices), len(g.PendingTriggers))
	}
	if g.Turn.Seq != turn+1 {
		t.Errorf("the turn did not end")
	}
}

// Obeka asks the player whose turn it is, not its controller: activated
// on an opponent's turn, the opponent decides — "no" leaves the turn
// alone — and on its controller's own turn, "yes" ends it.
func TestObekaAsksThePlayerWhoseTurnItIs(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	mine := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	obeka := pushCatalogPermanent(g, mine.ID, "Obeka, Brute Chronologist", "Legendary Creature — Ogre Wizard", obekaOracle, false)
	advanceTo(t, g, game.StepPrecombatMain)
	turn := g.Turn.Seq
	if err := g.ActivateCatalogAbility(mine.ID, obeka, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Obeka: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, mine.ID) != nil {
		t.Fatal("Obeka's controller was asked; the player whose turn it is decides")
	}
	c := latestConfirmFor(g, active.ID)
	if c == nil {
		t.Fatal("the active player was not asked")
	}
	if err := g.ResolveConfirm(c.ID, active.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if g.Turn.Seq != turn || g.Turn.Step != game.StepPrecombatMain {
		t.Fatalf("declining ended the turn: turn %d %s", g.Turn.Seq, g.Turn.Step)
	}

	// Obeka's controller's own turn: they ask themselves, and yes ends it.
	advanceToStepOf(t, g, mine.Seat, game.StepPrecombatMain)
	turn = g.Turn.Seq
	if err := g.ActivateCatalogAbility(mine.ID, obeka, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Obeka on its controller's turn: %v", err)
	}
	passPriorityAroundTable(t, g)
	c = latestConfirmFor(g, mine.ID)
	if c == nil {
		t.Fatal("Obeka's controller, now the active player, was not asked")
	}
	if err := g.ResolveConfirm(c.ID, mine.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	discardIfOwed(t, g)
	if g.Turn.Seq != turn+1 || g.Turn.ActiveSeat == mine.Seat {
		t.Errorf("yes did not end the turn: turn %d seat %d step %s", g.Turn.Seq, g.Turn.ActiveSeat, g.Turn.Step)
	}
	if g.TurnEndPending {
		t.Errorf("the answer's boundary left TurnEndPending set")
	}
}

// Final Fortune's loss is bound to the extra turn's end step. Ending
// that turn with Sundial skips the end step, so the trigger never
// fires and is swept as the turn ends: the player does not lose
// (the Final Fortune ruling — skipping the end step skips the loss).
func TestSundialSkipsFinalFortunesEndStepLoss(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	sundial := pushCatalogPermanent(g, g.Seats[me].ID, "Sundial of the Infinite", "Artifact", sundialOracle, false)
	castCatalogSpell(t, g, "Final Fortune", "Instant", finalFortuneOracle, nil)
	passPriorityAroundTable(t, g)
	endTurn(t, g)
	assertTurnOf(t, g, me, true, "the extra turn")
	advanceTo(t, g, game.StepPrecombatMain)
	floatMana(t, g, g.Seats[me], "{C}")
	if err := g.ActivateCatalogAbility(g.Seats[me].ID, sundial, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Sundial: %v", err)
	}
	passPriorityAroundTable(t, g)
	discardIfOwed(t, g)
	if g.Seats[me].Eliminated {
		t.Fatal("lost although the extra turn's end step never began")
	}
	assertTurnOf(t, g, (me+1)%len(g.Seats), false, "after the ended extra turn")
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("the bound loss trigger was not swept: %d queued", len(g.DelayedTriggers))
	}
}

// discardIfOwed answers the cleanup step's hand-size discard (CR 514.1)
// in the turn that ended, with the first cards in hand. newCatalogGame's
// active seat holds eight cards after its draw, so the ended turn's
// cleanup step asks for one before the next turn can begin.
func discardIfOwed(t *testing.T, g *game.Game) {
	t.Helper()
	if g.Turn.Step != game.StepCleanup {
		return
	}
	for _, p := range g.Seats {
		n := g.DiscardPending[p.ID]
		if n <= 0 {
			continue
		}
		ids := make([]uuid.UUID, 0, n)
		for i := 0; i < n && i < len(p.Hand.Cards); i++ {
			ids = append(ids, p.Hand.Cards[i].InstanceID)
		}
		if err := g.DiscardSelection(p.ID, ids); err != nil {
			t.Fatalf("DiscardSelection: %v", err)
		}
	}
}
