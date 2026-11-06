package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// YouCommittedCrime reads the once-per-announcement event (CR 700.13).
// Target changes and creating copies do not make a new announcement.
func YouCommittedCrime(ev game.Event, source *game.Card) bool {
	return ev.Kind == game.EventBecomesTarget && ev.CrimeCommitted && ev.Actor == source.Controller
}

// WheneverYouCommitACrimeOnceEachTurn is "Whenever you commit a crime,
// <effect>. This ability triggers only once each turn." (Marauding
// Sphinx): the crime announcement (YouCommittedCrime), held to once each
// turn by the per-turn trigger record under `label`, so later crimes that
// turn do not trigger it again.
func WheneverYouCommitACrimeOnceEachTurn(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return YouCommittedCrime(ev, source) && !b11TriggeredThisTurn(g, source.InstanceID, label)
	}, label, effect)
}

// YouCommittedACrimeThisTurn is "if you've committed a crime this turn"
// (Omenport Vigilante, Nimble Brigand): this turn's event log holds a
// crime announcement by `you` (CR 700.13). The crime is committed as the
// spell or ability is put on the stack, so one that was countered or lost
// its target since still counts.
//
// Caller must hold g.mu.
func YouCommittedACrimeThisTurn(g *game.Game, you uuid.UUID) bool {
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind == game.EventBecomesTarget && ev.CrimeCommitted && ev.Actor == you {
			return true
		}
	}
	return false
}
