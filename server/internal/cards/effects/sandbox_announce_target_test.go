package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sandbox_announce_target_test.go — #968 / CR 602.2b from the card
// side. Game.ActivateAbility is the MANUAL announce, the one ADR 0032
// keeps for the ~thousand cards the catalog cannot express: a player
// clicks "activate" on a permanent, types what the ability does and
// picks its targets by hand. It stamped those targets onto the stack
// item and announced nothing, so a card watching for "becomes the
// target of a spell or ability" saw a board where nothing had.
//
// Both cards here are ordinary catalog triggers. Neither knows or
// cares which verb announced the ability — that is the point: one
// fan-out helper, and the manual verb calls it like the other five.

// activatedItemFor returns the stack item ID of the activated ability
// announced from `source`, or uuid.Nil.
func activatedItemFor(g *game.Game, source uuid.UUID) uuid.UUID {
	for id, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemActivated && item.SourceCardID == source {
			return id
		}
	}
	return uuid.Nil
}

// counteredInLog reports whether the log records `stackID` being
// countered.
func counteredInLog(g *game.Game, stackID uuid.UUID) bool {
	for _, ev := range g.Events {
		if ev.Kind == game.EventCounterSpell && ev.Target == stackID {
			return true
		}
	}
	return false
}

// Ward taxes a hand-announced ability exactly as it taxes a cast
// spell (CR 702.21a is "a spell or ability an opponent controls"),
// and a declined payment counters the ability.
func TestWardTriggersOnAHandAnnouncedAbility(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := pushCatalogPermanent(g, me.ID, "Rimeshield Frost Giant",
		"Creature — Giant Warrior", rimeshieldOracle, false)
	machine := pushCreatureToBattlefieldForTest(g, opp.ID, "Uncatalogued Machine")

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.ActivateAbility(opp.ID, machine, game.AbilityParams{
		Label:   "{T}: destroy target creature",
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: giant}},
	}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	ability := activatedItemFor(g, machine)
	if ability == uuid.Nil {
		t.Fatalf("the announced ability is not on the stack")
	}

	// The ward trigger goes on the stack above the ability and
	// resolves first; its resolution asks the ability's CONTROLLER.
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("ward did not trigger off a hand-announced ability")
	}
	if hasPayUnlessFor(g, me.ID) {
		t.Error("the ward permanent's controller must not be asked to pay")
	}

	answerPayUnless(t, g, opp.ID, false)
	if _, ok := g.StackMeta[ability]; ok {
		t.Error("a declined ward must counter the ability — it is still on the stack")
	}
	if !counteredInLog(g, ability) {
		t.Error("no counter event for the warded ability")
	}
	if !g.Battlefield.Contains(giant) {
		t.Error("the warded creature should still be on the battlefield")
	}
}

// Monk Gyatso is the "becomes the target" trigger from the other
// direction: it watches a creature its controller controls, and a
// hand-announced ability pointed at one has to wake it up. (Phantasmal
// Image, the other card of this shape, has no catalog entry yet; this
// is the one the engine can play.)
func TestMonkGyatsoTriggersOnAHandAnnouncedAbility(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gyatso := pushCreatureToBattlefieldForTest(g, me.ID, "Monk Gyatso")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == gyatso {
			g.Battlefield.Cards[i].OracleID = monkGyatsoOracle
		}
	}
	mine := pushCreatureToBattlefieldForTest(g, me.ID, "My Bear")

	aangAdvanceToMain(t, g, 1)
	machine := pushCreatureToBattlefieldForTest(g, opp.ID, "Uncatalogued Machine")
	if err := g.ActivateAbility(opp.ID, machine, game.AbilityParams{
		Label:   "{T}: destroy target creature",
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}},
	}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}

	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	assertAirbent(t, g, mine, me.ID)
}

// triggeredItemFor returns the stack item ID of the triggered ability
// announced from `source`, or uuid.Nil.
func triggeredItemFor(g *game.Game, source uuid.UUID) uuid.UUID {
	for id, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered && item.SourceCardID == source {
			return id
		}
	}
	return uuid.Nil
}

// #974, the AnnounceTrigger half of the same family. The manual
// TRIGGER announce emitted became-target and then ran no state checks,
// so the ward trigger it harvested sat in PendingTriggers while the
// table passed — and the pass that would have drained it advanced the
// step instead, or dropped the ward under the announced trigger in one
// APNAP batch. Either way the payment prompt arrived after the
// announced trigger had resolved, and declining it countered nothing.
//
// The assertion is the ORDER: ward's prompt is up while the announced
// trigger is still on the stack, and declining the payment counters
// it.
//
// Both rows, because the announced trigger and the ward trigger would
// otherwise share one APNAP batch — and APNAP orders a batch by SEAT.
// The second row is the one that catches it: the ward is on the ACTIVE
// player's permanent, so a single batch would place it first, under
// the non-active announcer's trigger, and the counter would come too
// late. That is why the announce places its own trigger and then
// announces its targets, rather than doing both in one drain.
func TestWardTriggersOnAHandAnnouncedTrigger(t *testing.T) {
	for _, tc := range []struct {
		name                  string
		announcer             int // seat index announcing the trigger
		wardHolder            int // seat index whose permanent has ward
		activeIsTheWardHolder bool
	}{
		{"active player announces at an opponent's ward", 0, 1, false},
		{"an opponent announces at the active player's ward", 1, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			announcer, holder := g.Seats[tc.announcer], g.Seats[tc.wardHolder]
			giant := pushCatalogPermanent(g, holder.ID, "Rimeshield Frost Giant",
				"Creature — Giant Warrior", rimeshieldOracle, false)
			omen := pushCreatureToBattlefieldForTest(g, announcer.ID, "Uncatalogued Omen")

			// Seat 0's precombat main throughout, so the second row
			// really is a non-active announcer.
			aangAdvanceToMain(t, g, 0)
			if (g.Turn.ActiveSeat == tc.wardHolder) != tc.activeIsTheWardHolder {
				t.Fatalf("active seat = %d, which is not the configuration this row is about", g.Turn.ActiveSeat)
			}

			if err := g.AnnounceTrigger(announcer.ID, omen, game.AbilityParams{
				Label:   "when this triggers, destroy target creature",
				Targets: []game.TargetRef{{Kind: game.TargetCard, ID: giant}},
			}); err != nil {
				t.Fatalf("AnnounceTrigger: %v", err)
			}
			announced := triggeredItemFor(g, omen)
			if announced == uuid.Nil {
				t.Fatalf("the announced trigger is not on the stack")
			}
			if len(g.PendingTriggers) != 0 {
				t.Fatalf("%d triggers left queued by the announce", len(g.PendingTriggers))
			}

			// The ward trigger went on the stack ABOVE the announced
			// one, so it resolves first and asks the announcer to pay.
			for i := 0; i < 8 && !hasPayUnlessFor(g, announcer.ID); i++ {
				if err := g.PassPriority(); err != nil {
					t.Fatalf("PassPriority: %v", err)
				}
			}
			if !hasPayUnlessFor(g, announcer.ID) {
				t.Fatal("ward did not trigger off a hand-announced trigger")
			}
			if _, ok := g.StackMeta[announced]; !ok {
				t.Fatal("the announced trigger resolved before the ward's payment was asked for")
			}
			if hasPayUnlessFor(g, holder.ID) {
				t.Error("the ward permanent's controller must not be asked to pay")
			}

			answerPayUnless(t, g, announcer.ID, false)
			if _, ok := g.StackMeta[announced]; ok {
				t.Error("a declined ward must counter the announced trigger — it is still on the stack")
			}
			if !counteredInLog(g, announced) {
				t.Error("no counter event for the warded trigger")
			}
		})
	}
}
