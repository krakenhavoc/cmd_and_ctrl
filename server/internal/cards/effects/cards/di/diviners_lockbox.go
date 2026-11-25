package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Diviner's Lockbox — Artifact {4} (Core Set 2020, #5518):
//
//	"{1}, {T}: Choose a card name, then reveal the top card of your library. If
//	 that card has the chosen name, sacrifice this artifact and draw three
//	 cards. Activate only as a sorcery."
//
// The name is chosen as the ability resolves (it is not a target and can't be
// responded to), then the top card is revealed and compared. A Lockbox that
// has left the battlefield is not sacrificed, but the three cards are still
// drawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e00810cd-8863-433b-aa7a-683b8f308a9b",
		Name:         "Diviner's Lockbox",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}, {T}: Choose a card name, then reveal the top card of your library. If that card has the chosen name, sacrifice this artifact and draw three cards. Activate only as a sorcery.",
			Purpose:      game.Purpose{Draws: 3},
			Cost:         effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			SorcerySpeed: true,
			Effect:       effects.OldB154LockboxChooseName,
		}},
	})
}
