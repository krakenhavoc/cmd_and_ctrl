package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// cost_commander_choice_test.go — the enumerator half of #1397.
//
// An activation whose cost moves a commander is now PARKED on a
// CR 903.9 question to the commander's owner before anything is paid
// (ADR 0013 §5af). The enumerator must still offer the activation
// (it is legal; the question is part of making it), and once it is
// parked the owner — who is an OPPONENT when the commander was stolen —
// is offered the yes/no and nobody is offered anything else, because
// the prompt blocks the table.

func TestACostParkedOnAStolenCommanderIsAnsweredByItsOwner(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(me)
	advanceTo(t, g, game.StepPrecombatMain)

	outlet := battlefieldCard(g, me, game.Card{
		Name: "Test Viscera Seer", TypeLine: "Artifact",
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "Sacrifice a creature: nothing",
			Cost: game.AbilityCost{SacrificeOther: &game.TargetSpec{
				Mode: "permanent", Label: "a creature", Zones: []game.ZoneKind{game.ZoneBattlefield},
				CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool { return c.IsCreature() },
				Min:    1, Max: 1,
			}},
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	})
	// The opponent's commander, under my control (a Treachery on it).
	stolen := battlefieldCard(g, opp, game.Card{
		Name: "Their Commander", TypeLine: "Legendary Creature — Angel", Power: 4, Toughness: 4, IsCommander: true,
	})
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == stolen {
			g.Battlefield.Cards[i].Controller = me.ID
		}
	}

	offered := movesFrom(legal.EnumerateFor(g, me.ID), outlet, legal.KindActivate)
	if len(offered) == 0 {
		t.Fatal("the sacrifice outlet was not offered with a creature to pay it")
	}
	m := offered[0]
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(m.Type), Player: me.ID, Caller: me.ID, Params: m.Params}); err != nil {
		t.Fatalf("dispatch the activation: %v", err)
	}
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Chooser != opp.ID {
		t.Fatalf("want one prompt owed by the commander's owner, got %d", len(g.PendingChoices))
	}
	if len(g.StackMeta) != 0 {
		t.Fatal("the activation was paid before the owner answered")
	}

	answers := legal.EnumerateFor(g, opp.ID)
	if len(answers) != 2 {
		t.Fatalf("the owner was offered %d moves, want yes and no: %v", len(answers), labels(answers))
	}
	for _, a := range answers {
		if a.Kind != legal.KindChoice {
			t.Errorf("the owner was offered %q (%s) while the prompt is open", a.Label, a.Type)
		}
	}
	dispatchAll(t, g, opp.ID, answers)
	for _, seat := range g.Seats {
		if seat.ID == opp.ID {
			continue
		}
		for _, other := range legal.EnumerateFor(g, seat.ID) {
			t.Errorf("seat %s offered %q while the owner's prompt blocks the table", seat.Name, other.Label)
		}
	}
}
