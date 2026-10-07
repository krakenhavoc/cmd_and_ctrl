package protocol

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// discard_mana_value_view_test.go — #2190. "Discard a card with mana
// value X" ships the one-card count, the mana_value_x flag and
// demands_x (so the activation flow knows an X is owed, which the
// picked card then answers), the hand cards that could pay — to the
// controller alone (#1369) — and a spell clause that carries the
// equal-to-X bound with the stack spells' mana values, counting {X} as
// the value chosen for the spell (CR 202.3e).

func seatManaValueDiscardSource(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: "Mana Value Discard Source", TypeLine: "Creature — Eldrazi",
		OracleID: "00000000-0000-0000-0000-000000002190",
		Owner:    owner, Controller: owner, Power: 12, Toughness: 12,
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "Discard a card with mana value X: Counter target spell with mana value X.",
			Cost: game.AbilityCost{
				DiscardCards: &game.DiscardCost{N: 1, Label: "a card with mana value X", ManaValueX: true},
			},
		}},
	})
	return id
}

func TestManaValueDiscardViewStampsTheFlagAndTheHandToTheControllerAlone(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var src uuid.UUID
	var hand []string
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		src = seatManaValueDiscardSource(g, me.ID)
		readable := handCardFor(me, "Hidden Bear", "Creature — Bear")
		readable2 := handCardFor(me, "Hidden Bolt", "Instant")
		hand = sortedIDs(readable, readable2)
		// A card whose cost the engine cannot read cannot pay, so it is
		// not an option.
		split := handCardFor(me, "Fire // Ice", "Instant // Instant")
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == split {
				me.Hand.Cards[i].ManaCost = "{1}{R} // {1}{U}"
			}
		}
	})
	g.BumpLayerVersionForTest()

	mine := controllerFrameCard(t, g, src).ActivatedAbilities[0]
	if !mine.DiscardCostManaValueX || mine.DiscardCostCountFromX || mine.DiscardCostN != 1 || mine.DiscardCostLabel != "a card with mana value X" {
		t.Errorf("controller row: mana_value_x=%v count_from_x=%v n=%d label=%q",
			mine.DiscardCostManaValueX, mine.DiscardCostCountFromX, mine.DiscardCostN, mine.DiscardCostLabel)
	}
	if !mine.DemandsX {
		t.Error("demands_x is not set, so the activation would never announce an X")
	}
	if got := sortedStrings(mine.DiscardCostOptions); !reflect.DeepEqual(got, hand) {
		t.Errorf("controller options = %v, want the readable cards %v", got, hand)
	}

	theirs := frameCard(t, g, opp.ID.String(), src).ActivatedAbilities[0]
	if !theirs.DiscardCostManaValueX || !theirs.DemandsX {
		t.Error("the opponent's row lost the flag; it names nothing hidden")
	}
	if len(theirs.DiscardCostOptions) != 0 {
		t.Errorf("the opponent's row carries the controller's hand: %v", theirs.DiscardCostOptions)
	}
}
