package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// graveyard_return_together_test.go — #1867. A card that returns
// several cards from a graveyard to the battlefield puts them there in
// ONE event, so each newcomer sees the others enter (CR 603.6a: "Each
// time an event puts one or more permanents onto the battlefield, all
// permanents on the battlefield (including the newcomers) are checked
// for any enters-the-battlefield triggers that match the event").
//
// The observer is Soul Warden ("Whenever another creature enters, you
// gain 1 life"). Two Wardens returned together trigger each other: two
// triggers, 2 life. Returned one at a time, only the first sees the
// second: one trigger, 1 life.

// gtWarden puts a Soul Warden card into p's graveyard.
func GtWarden(p *game.Player) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: "Soul Warden", OracleID: B04SoulWardenOracle,
		TypeLine: "Creature — Human Cleric", Power: 1, Toughness: 1,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

// gtSettle resolves the stack, answering the CR 603.3b order prompt
// two simultaneous triggers of one controller ask for.
func GtSettle(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 8; i++ {
		PassPriorityAroundTable(t, g)
		if TriggerOrderPrompt(g) == nil {
			return
		}
		AnswerTriggerOrderInOfferedOrder(t, g)
	}
	t.Fatal("the stack did not settle")
}
