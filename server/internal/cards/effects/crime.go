package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// YouCommittedCrime reads the once-per-announcement event (CR 700.13).
// Target changes and creating copies do not make a new announcement.
func YouCommittedCrime(ev game.Event, source *game.Card) bool {
	return ev.Kind == game.EventBecomesTarget && ev.CrimeCommitted && ev.Actor == source.Controller
}
