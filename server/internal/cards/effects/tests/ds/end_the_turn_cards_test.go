package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The classic Sundial play: at the beginning of your end step a delayed
// "exile it" trigger goes on the stack, and Sundial ends the turn in
// response. The trigger is exiled with the stack — the creature stays —
// the turn goes straight to cleanup and the next player's turn begins.
func TestSundialOfTheInfiniteEndsTheTurnOverItsEndStepTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sundial := effectstest.PushCatalogPermanent(g, me.ID, "Sundial of the Infinite", "Artifact", effectstest.SundialOracle, false)
	unearthed := effectstest.PushVanillaCreature(g, me.ID, "Unearthed", 3, 3)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: unearthed})
		if err := (effects.ScheduleDelayedTrigger{
			Label: effectstest.EndTurnTestLabel,
			Cards: []uuid.UUID{unearthed},
			Body:  effects.ExileListedCardsBody,
		}).Apply(ctx); err != nil {
			t.Fatalf("schedule: %v", err)
		}
	})
	effectstest.AdvanceTo(t, g, game.StepEnd)
	if !stackHasLabel(g, effectstest.EndTurnTestLabel) {
		t.Fatalf("setup: the end-step trigger is not on the stack")
	}
	turn := g.Turn.Seq
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, sundial, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Sundial: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardIfOwed(t, g)

	if g.Turn.Seq != turn+1 || g.Turn.ActiveSeat == me.Seat {
		t.Fatalf("the turn did not end: turn %d seat %d step %s", g.Turn.Seq, g.Turn.ActiveSeat, g.Turn.Step)
	}
	if _, ok := effectstest.AangCardOnBF(g, unearthed); !ok {
		t.Errorf("the creature was exiled: the trigger should have been exiled from the stack first")
	}
	if c, ok := effectstest.AangCardOnBF(g, sundial); !ok {
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
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	sundial := effectstest.PushCatalogPermanent(g, opp.ID, "Sundial of the Infinite", "Artifact", effectstest.SundialOracle, false)
	effectstest.FloatMana(t, g, opp, "{C}")
	if err := g.ActivateCatalogAbility(opp.ID, sundial, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("Sundial was activated during an opponent's turn")
	}
	if c, _ := effectstest.AangCardOnBF(g, sundial); c.Tapped {
		t.Error("the refused activation tapped Sundial")
	}
}

// Time Stop on an opponent's turn, in their declare attackers step: the
// spell they cast is exiled with Time Stop, the attack never deals
// damage, and the next player's turn begins.
func TestTimeStopEndsAnOpponentsTurnMidCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Charger", 5, 5)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	theirs := effectstest.CastInPlace(t, g, me.ID, "Some Instant", "")
	stop := effectstest.CastInPlace(t, g, opp.ID, "Time Stop", effectstest.TimeStopOracle)
	life := opp.Life
	turn := g.Turn.Seq
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardIfOwed(t, g)

	for _, id := range []uuid.UUID{theirs, stop} {
		if !g.Exile.Contains(id) {
			t.Errorf("%s is not in exile (CR 724.1b)", id)
		}
	}
	if opp.Life != life {
		t.Errorf("the defender lost %d life; the turn ended before combat damage", life-opp.Life)
	}
	if c, ok := effectstest.AangCardOnBF(g, attacker); !ok || c.AttackingTarget != uuid.Nil {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	effectstest.CastCatalogSpell(t, g, "Glorious End", "Instant", effectstest.GloriousEndOracle, nil)
	turn := g.Turn.Seq
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardIfOwed(t, g)
	if g.Turn.Seq != turn+1 {
		t.Fatalf("the turn did not end")
	}
	if g.Seats[me].Eliminated {
		t.Fatal("lost the turn Glorious End ended")
	}
	for seat := 1; seat < len(g.Seats); seat++ {
		effectstest.AdvanceToEndStepOf(t, g, (me+seat)%len(g.Seats))
		effectstest.PassPriorityAroundTable(t, g)
		if g.Seats[me].Eliminated {
			t.Fatalf("lost in seat %d's end step; only the caster's own end step counts", (me+seat)%len(g.Seats))
		}
	}
	effectstest.AdvanceToEndStepOf(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Seats[me].Eliminated {
		t.Fatal("did not lose at the beginning of the next end step of the caster's own turn")
	}
}

// Ultima: every artifact and creature is destroyed, lands survive, and
// the dies triggers the wipe causes cease to exist with the turn
// (CR 724.1a) — Zulaport Cutthroat drains nobody.
func TestUltimaWipesAndItsDiesTriggersCeaseToExist(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	zulaport := effectstest.PushCatalogPermanent(g, me.ID, "Zulaport Cutthroat", "Creature — Human Rogue", effectstest.ZulaportOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	rock := effectstest.PushPermanentForTest(g, opp.ID, "Mind Stone", "", "Artifact")
	land := effectstest.PushLand(g, opp.ID, "Island")
	lifeMe, lifeOpp := me.Life, opp.Life
	ultima := effectstest.CastCatalogSpell(t, g, "Ultima", "Sorcery", effectstest.UltimaOracle, nil)
	turn := g.Turn.Seq
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardIfOwed(t, g)

	for _, id := range []uuid.UUID{zulaport, bear, rock} {
		if _, ok := effectstest.AangCardOnBF(g, id); ok {
			t.Errorf("%s survived the wipe", id)
		}
	}
	if _, ok := effectstest.AangCardOnBF(g, land); !ok {
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
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	mine := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	obeka := effectstest.PushCatalogPermanent(g, mine.ID, "Obeka, Brute Chronologist", "Legendary Creature — Ogre Wizard", effectstest.ObekaOracle, false)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	turn := g.Turn.Seq
	if err := g.ActivateCatalogAbility(mine.ID, obeka, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Obeka: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, mine.ID) != nil {
		t.Fatal("Obeka's controller was asked; the player whose turn it is decides")
	}
	c := effectstest.LatestConfirmFor(g, active.ID)
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
	effectstest.AdvanceToStepOf(t, g, mine.Seat, game.StepPrecombatMain)
	turn = g.Turn.Seq
	if err := g.ActivateCatalogAbility(mine.ID, obeka, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Obeka on its controller's turn: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c = effectstest.LatestConfirmFor(g, mine.ID)
	if c == nil {
		t.Fatal("Obeka's controller, now the active player, was not asked")
	}
	if err := g.ResolveConfirm(c.ID, mine.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	effectstest.DiscardIfOwed(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	sundial := effectstest.PushCatalogPermanent(g, g.Seats[me].ID, "Sundial of the Infinite", "Artifact", effectstest.SundialOracle, false)
	effectstest.CastCatalogSpell(t, g, "Final Fortune", "Instant", effectstest.FinalFortuneOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, me, true, "the extra turn")
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.FloatMana(t, g, g.Seats[me], "{C}")
	if err := g.ActivateCatalogAbility(g.Seats[me].ID, sundial, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Sundial: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardIfOwed(t, g)
	if g.Seats[me].Eliminated {
		t.Fatal("lost although the extra turn's end step never began")
	}
	effectstest.AssertTurnOf(t, g, (me+1)%len(g.Seats), false, "after the ended extra turn")
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("the bound loss trigger was not swept: %d queued", len(g.DelayedTriggers))
	}
}
