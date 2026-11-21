package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// surveil_test.go — CR 701.25 surveil, and the surveil-land cycle
// that finally does what its type line says.
//
// Surveil is scry with one word changed, so the tests that matter are
// the ones that would still pass if it had been built as scry:
//
//   - the cards that leave the top land in the GRAVEYARD, not under
//     the library. That is the entire keyword. A reanimator deck
//     surveils to fill a graveyard, and bottoming instead is the
//     difference between a resource and a card that is gone.
//   - they emit an EventZoneMove on the way, library to graveyard, and
//     not the EventMill a mill emits: surveil is not a mill (CR
//     701.25a, #2809). A "whenever a card is put into your graveyard
//     from your library" payoff watches both kinds and still sees it;
//     a "whenever one or more cards are milled" payoff does not.
//   - surveil is LOOK AT, not reveal — same privacy rule as scry.
//   - the scry resolver must refuse a surveil prompt and vice versa,
//     because both answers carry a top_order and only the other key
//     tells them apart.

const UndercitySewersOracle = "08d80efc-9542-4ba2-824c-c8615d8d07f2"

func SurveilChoiceFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceSurveil && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// graveyardNames reads the graveyard bottom-to-top.
func GraveyardNames(p *game.Player) []string {
	out := make([]string, 0, p.Graveyard.Size())
	for _, c := range p.Graveyard.Cards {
		out = append(out, c.Name)
	}
	return out
}
