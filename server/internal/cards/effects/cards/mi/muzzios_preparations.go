package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Muzzio's Preparations — Conspiracy (Conspiracy, #5557):
//
//	"Hidden agenda (Start the game with this conspiracy face down in the
//	 command zone and secretly choose a card name. You may turn this
//	 conspiracy face up any time and reveal that name.)
//	 Each creature you control with the chosen name enters with an additional
//	 +1/+1 counter on it."
//
// Hidden agenda is the engine's (ADR 0227). Face up, a CR 614.1c replacement
// from the command zone over every creature entering under the conspiracy's
// controller's control with the chosen name (Bioengineered Future's shape):
// the counter is added to the entry as the creature enters, so it goes through
// the counter pipeline, and a creature with another name, or an opponent's with
// the same name, gets none. Face down it has no abilities (CR 315.5b), so a
// creature that entered before the reveal has no counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0e505699-8388-4862-93e0-059cef91cdd4",
		Name:         "Muzzio's Preparations",
		Completeness: effects.CompletenessFull,
		HiddenAgenda: effects.HiddenAgendaOneName,
		Replacements: []game.ReplacementEffect{{
			Watches:   []game.EventKind{game.EventZoneMove},
			AppliesTo: effects.CreatureYouControlWithTheChosenNameEnters,
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.AddCounterAtETB(game.CounterPlusOne, 1)
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Muzzio's Preparations: each creature you control with the chosen name enters with an additional +1/+1 counter on it",
		}},
	})
}
