package protocol

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reveal_x_view_test.go — #2598. "Reveal X black cards from your hand"
// ships no count (the count is the X the client is about to announce),
// the count_from_x flag, the clause's words, demands_x so the activation
// flow knows an X is owed, and the matching hand cards — to the
// controller alone (#1369).

func seatRevealXSource(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: "Reveal X Source", TypeLine: "Creature — Human Wizard",
		OracleID: "00000000-0000-0000-0000-000000002598",
		Owner:    owner, Controller: owner, Power: 1, Toughness: 1,
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "{1}, Reveal X black cards from your hand: do a thing",
			Cost: game.AbilityCost{
				Mana: "{1}",
				RevealCards: &game.RevealCardsCost{
					RevealCost: game.RevealCost{Color: "B"},
					CountFromX: true,
					Label:      "X black cards",
				},
			},
		}},
	})
	return id
}

func handCardColoured(p *game.Player, name, color string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine = "Sorcery"
	c.Colors = []string{color}
	c.KnownBy = map[uuid.UUID]bool{p.ID: true}
	p.Hand.PushTop(c)
	return c.InstanceID
}

func TestRevealXViewStampsTheFlagAndTheMatchingHandToTheControllerAlone(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var src uuid.UUID
	var black []string
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		src = seatRevealXSource(g, me.ID)
		black = sortedIDs(
			handCardColoured(me, "Black A", "B"),
			handCardColoured(me, "Black B", "B"),
		)
		handCardColoured(me, "White A", "W")
	})
	g.BumpLayerVersionForTest()

	mine := controllerFrameCard(t, g, src).ActivatedAbilities[0]
	if !mine.RevealCostCountFromX || mine.RevealCostN != 0 || mine.RevealCostLabel != "X black cards" {
		t.Errorf("controller row: count_from_x=%v n=%d label=%q, want true / 0 / \"X black cards\"",
			mine.RevealCostCountFromX, mine.RevealCostN, mine.RevealCostLabel)
	}
	if !mine.DemandsX {
		t.Error("demands_x is not set, so the client would never learn an X is owed")
	}
	if got := sortedStrings(mine.RevealCostOptions); !reflect.DeepEqual(got, black) {
		t.Errorf("controller options = %v, want only the black cards %v", got, black)
	}

	theirs := frameCard(t, g, opp.ID.String(), src).ActivatedAbilities[0]
	if !theirs.RevealCostCountFromX || !theirs.DemandsX || theirs.RevealCostLabel != "X black cards" {
		t.Error("the opponent's row lost the printed clause; it names nothing hidden")
	}
	if len(theirs.RevealCostOptions) != 0 {
		t.Errorf("the opponent's row carries the controller's hand: %v", theirs.RevealCostOptions)
	}
}
