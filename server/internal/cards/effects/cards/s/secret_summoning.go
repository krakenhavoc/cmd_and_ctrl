package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Secret Summoning — Conspiracy (Conspiracy, #5557):
//
//	"Hidden agenda (Start the game with this conspiracy face down in the
//	 command zone and secretly choose a card name. You may turn this
//	 conspiracy face up any time and reveal that name.)
//	 Whenever a creature you control with the chosen name enters, you may
//	 search your library for any number of cards with that name, reveal them,
//	 put them into your hand, then shuffle."
//
// Hidden agenda is the engine's (ADR 0227). Face up, an enters trigger
// harvested from the command zone (CR 315.5); the name is the entering
// creature's, carried on the ability, so the search is for that name. The "may"
// is asked as the ability resolves, the search is a real choice from none to
// every match (SearchLibrary.Unbounded), the found cards are revealed, and the
// library is shuffled whether or not anything was found (CR 701.23).
//
// No simplification.
func init() {
	const label = "Secret Summoning — search your library for any number of cards with that name"
	t := effects.Optional(
		effects.On(game.EventETB, effects.S82OldB360ChosenNameCreatureYouControlEnters, label, effects.S82OldB360NamedCardsToHand),
		"Secret Summoning — search your library for any number of cards with that name?")
	t.Build = effects.S82OldB360BuildNamingTheEnteringCreature(label)
	effects.Register(effects.Spec{
		OracleID:     "1342ee23-c3a0-4578-833a-e30a9fc9141c",
		Name:         "Secret Summoning",
		Completeness: effects.CompletenessFull,
		HiddenAgenda: effects.HiddenAgendaOneName,
		Triggered:    []game.TriggeredAbility{t},
	})
}
