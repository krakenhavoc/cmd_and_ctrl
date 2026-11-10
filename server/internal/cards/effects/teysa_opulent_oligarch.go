package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Teysa, Opulent Oligarch — Legendary Creature — Human Advisor {1}{W}{B}
// (Murders at Karlov Manor, #4469):
//
//	"Deathtouch
//	 At the beginning of your end step, investigate for each opponent who
//	 lost life this turn.
//	 Whenever a Clue you control is put into a graveyard from the
//	 battlefield, create a 1/1 white and black Spirit creature token with
//	 flying. This ability triggers only once each turn."
//
// The count is each opponent whose turn tally shows life lost, one that has
// since lost the game included (the ruling), read as the ability resolves.
// The second ability sees any Clue artifact, a token or not (a Clue is an
// artifact type, CR 111.10f), as it last existed: a Clue that was yours when
// it went to the graveyard, whatever it is now. It triggers once each turn
// (TriggersOnlyOnceEachTurn): a second Clue the same turn does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d4f5748f-bb0e-4c48-8802-45a7ad5a80a4",
		Name:            "Teysa, Opulent Oligarch",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Teysa — investigate for each opponent who lost life this turn",
				func(g *game.Game, item *game.StackItem) error {
					n := opponentsWhoLostLifeThisTurn(g, item.Controller)
					if n == 0 {
						return nil
					}
					return CreateToken{Template: ClueToken(), N: n}.Apply(NewContext(g, item))
				}),
			TriggersOnlyOnceEachTurn(On(game.EventLTB, aClueYouControlWasPutIntoAGraveyard,
				"Teysa — create a 1/1 white and black Spirit creature token with flying",
				Do(CreateToken{Template: TokenCard("1/1 white and black Spirit with flying"), N: 1}))),
		},
	})
}

// opponentsWhoLostLifeThisTurn is the number of opponents of `player` whose
// turn tally shows any life lost, a player who has since left the game
// counted (CR 800.4, the Teysa ruling).
func opponentsWhoLostLifeThisTurn(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, p := range g.Seats {
		if p != nil && p.ID != player && g.TurnTallyFor(p.ID).LifeLost > 0 {
			n++
		}
	}
	return n
}

// aClueYouControlWasPutIntoAGraveyard is "whenever a Clue you control is put
// into a graveyard from the battlefield": a leaves-the-battlefield event to a
// graveyard, for an object that was a Clue under the source's controller as
// it last existed (CR 603.10a).
func aClueYouControlWasPutIntoAGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
		return false
	}
	if was, known := ev.WasSubtype("Clue"); !known || !was {
		return false
	}
	controller, known := ev.LeftUnderControlOf()
	return known && controller == source.Controller
}
