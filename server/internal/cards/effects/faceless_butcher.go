package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Faceless Butcher — Creature — Nightmare Horror {2}{B}{B}, 2/3
// (Torment, #5612):
//
//	"When this creature enters, exile another target creature.
//	 When this creature leaves the battlefield, return the exiled card to
//	 the battlefield under its owner's control."
//
// Fiend Hunter without the "may": the two-trigger form, not Oblivion
// Ring's "until" (CR 610.3). The abilities are linked (CR 607.2a), so
// "the exiled card" is the one the entry trigger exiled, read off the
// event log by that trigger's key (b41ReturnCardsExiledWithToTheBattlefield),
// and only while it is still in exile. A Butcher that comes back is a new
// object with an empty record (CR 400.7). The consequence the Oracle
// rulings state: if the Butcher leaves before its entry trigger resolves,
// the leave trigger finds nothing, and the creature is then exiled for
// good. "Another" keeps the Butcher off its own target list; a token
// exiled this way ceases to exist and does not come back.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5237a105-60c4-4597-b24e-84c12ce7b01f",
		Name:         "Faceless Butcher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters(facelessButcherExileLabel, exileFirstLegalCardTarget),
				Another(TargetCreature("another target creature"))),
			On(game.EventLTB, Self, "Faceless Butcher — return the exiled card to the battlefield under its owner's control",
				b41ReturnCardsExiledWithToTheBattlefield(facelessButcherExileLabel)),
		},
	})
}

// facelessButcherExileLabel is the entry trigger's key, which is what the
// leave trigger reads the linked exile by.
const facelessButcherExileLabel = "Faceless Butcher — exile another target creature"
