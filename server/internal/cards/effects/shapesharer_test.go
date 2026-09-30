package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// shapesharer_test.go — #1723's first seam at the card level:
// BecomeCopy's new CopyUntilYourNextTurn duration (CR 611.2b). The
// invariant that matters is the one "until end of turn" and "until
// your next turn" are easy to confuse on: the copy must survive the
// END of the turn it was made on (which only ends an UntilEndOfTurn
// effect) and revert only once the controller's OWN next turn begins,
// after every other seat has had a turn.
const shapesharerOracle = "400c9fd8-c307-4b27-af6a-73e100717881"

func TestShapesharerBecomesACopyUntilYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	startSeat := g.Turn.ActiveSeat

	shifter := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Shapesharer",
		TypeLine:   "Creature — Shapeshifter",
		OracleID:   shapesharerOracle,
		Power:      1, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	model := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Grizzly Bears",
		TypeLine:   "Creature — Bear",
		Power:      2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})

	if err := g.ActivateCatalogAbility(me.ID, shifter, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: shifter},
			{Kind: game.TargetCard, ID: model},
		},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)

	card := battlefieldCardFor(g, shifter)
	if card == nil {
		t.Fatal("Shapesharer left the battlefield")
	}
	if card.Name != "Grizzly Bears" || card.Power != 2 || card.Toughness != 2 {
		t.Fatalf("post-activation = %q %d/%d, want Grizzly Bears 2/2", card.Name, card.Power, card.Toughness)
	}

	// The end of the current turn's cleanup — an UntilEndOfTurn copy
	// would revert here. This one must not.
	advanceToNextSeatsTurn(t, g)
	card = battlefieldCardFor(g, shifter)
	if card == nil {
		t.Fatal("Shapesharer left the battlefield")
	}
	if card.Name != "Grizzly Bears" {
		t.Errorf("copy reverted at the end of the turn it was made on (name = %q); "+
			"\"until your next turn\" (CR 611.2b) must survive it", card.Name)
	}

	// Cycle through every other seat's turn until it is the
	// controller's turn again.
	for g.Turn.ActiveSeat != startSeat {
		advanceToNextSeatsTurn(t, g)
	}
	card = battlefieldCardFor(g, shifter)
	if card == nil {
		t.Fatal("Shapesharer left the battlefield")
	}
	if card.Name != "Shapesharer" || card.Power != 1 || card.Toughness != 1 {
		t.Errorf("post-revert = %q %d/%d, want Shapesharer 1/1 once the controller's next turn began",
			card.Name, card.Power, card.Toughness)
	}
}
